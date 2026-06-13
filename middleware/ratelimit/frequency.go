package ratelimit

/*
The frequency limit has three matching methods (all, regular, and exact) to match
the full-method value of the API.
*/

import (
	"context"
	"regexp"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ml444/gkit/errorx"
	"github.com/ml444/gkit/middleware"
	"github.com/ml444/gkit/transport"
)

var ErrLimitExceed = errorx.CreateError(429, 40029, "RATELIMIT: service unavailable due to frequency limit exceeded")

type MatchKind uint8

const (
	MatchKindExact   MatchKind = 0
	MatchKindRegular MatchKind = 1
	MatchKindAll     MatchKind = 3
)

type Frequency struct {
	Period time.Duration
	Limit  uint64
}

type LimitCfg struct {
	Kind    MatchKind
	Pattern string
	Paths   []string
	Freqs   []*Frequency
}

type periodLimit struct {
	period    int64
	limit     uint64
	timestamp *atomic.Int64
	count     *atomic.Uint64
}

func newPeriodLimit(period int64, limit uint64) *periodLimit {
	tsNow := &atomic.Int64{}
	tsNow.Store(time.Now().UnixMilli())
	return &periodLimit{
		period:    period,
		limit:     limit,
		timestamp: tsNow,
		count:     &atomic.Uint64{},
	}
}

func (p *periodLimit) over(now int64) bool {
	ts := p.timestamp.Load()
	if ts+p.period <= now {
		// Window expired: let exactly one goroutine claim the reset via CAS,
		// so concurrent callers at the boundary do not all reset the window.
		if p.timestamp.CompareAndSwap(ts, now) {
			p.count.Store(1)
			return false
		}
	}
	count := p.count.Add(1)
	return count > p.limit
}

type freqLimiter struct {
	limits []*periodLimit
}

func newFreqLimiter() *freqLimiter {
	return &freqLimiter{}
}

func (rl *freqLimiter) Allow() bool {
	now := time.Now().UnixMilli()
	for _, p := range rl.limits {
		if p.over(now) {
			return false
		}
	}
	return true
}

type limitSet struct {
	mu                sync.RWMutex
	allMethod         *freqLimiter
	reMap             map[*regexp.Regexp]*freqLimiter
	specificMethodMap map[string]*freqLimiter
	checkedMethod     map[string]bool
}

// resolveRegexLocked lazily materializes a per-key limiter from regex configs.
// It must mutate the maps, so it takes the write lock and double-checks to stay
// correct under concurrency.
func (s *limitSet) resolveRegexLocked(key string) *freqLimiter {
	s.mu.Lock()
	defer s.mu.Unlock()
	if l, ok := s.specificMethodMap[key]; ok {
		return l
	}
	if s.checkedMethod[key] {
		return nil
	}
	s.checkedMethod[key] = true
	var limiter *freqLimiter
	for re, reLimiter := range s.reMap {
		if !re.MatchString(key) {
			continue
		}
		if limiter == nil {
			limiter = newFreqLimiter()
		}
		for _, limit := range reLimiter.limits {
			limiter.limits = append(limiter.limits, newPeriodLimit(limit.period, limit.limit))
		}
	}
	if limiter != nil {
		s.specificMethodMap[key] = limiter
	}
	return limiter
}

func (s *limitSet) WalkAllow(key string) bool {
	// allMethod is set only at construction; its Allow() is itself atomic-safe.
	if s.allMethod != nil && !s.allMethod.Allow() {
		return false
	}
	// Read phase: look up an already-resolved limiter under the read lock only.
	s.mu.RLock()
	limiter, ok := s.specificMethodMap[key]
	checked := s.checkedMethod[key]
	hasRe := len(s.reMap) > 0
	s.mu.RUnlock()
	if ok {
		return limiter.Allow()
	}
	if !hasRe || checked {
		return true
	}
	// Write phase: populate from regex matches under the write lock.
	limiter = s.resolveRegexLocked(key)
	if limiter == nil {
		return true
	}
	return limiter.Allow()
}

func newLimitSet(cfgs ...*LimitCfg) *limitSet {
	rlSet := &limitSet{
		specificMethodMap: map[string]*freqLimiter{},
	}
	for _, limitCfg := range cfgs {
		if len(limitCfg.Freqs) == 0 {
			continue
		}
		var limits []*periodLimit
		for _, cycle := range limitCfg.Freqs {
			if cycle.Period == 0 || cycle.Limit == 0 {
				continue
			}
			limits = append(limits, newPeriodLimit(cycle.Period.Milliseconds(), cycle.Limit))
		}
		sort.SliceStable(limits, func(i, j int) bool {
			return limits[i].period < limits[j].period
		})
		switch limitCfg.Kind {
		case MatchKindAll:
			allLimiter := rlSet.allMethod
			if allLimiter == nil {
				allLimiter = newFreqLimiter()
				rlSet.allMethod = allLimiter
			}
			allLimiter.limits = append(allLimiter.limits, limits...)
		case MatchKindRegular:
			if rlSet.checkedMethod == nil {
				rlSet.checkedMethod = map[string]bool{}
			}
			if rlSet.reMap == nil {
				rlSet.reMap = make(map[*regexp.Regexp]*freqLimiter)
			}
			limiter := newFreqLimiter()
			limiter.limits = append(limiter.limits, limits...)
			rlSet.reMap[regexp.MustCompile(limitCfg.Pattern)] = limiter
		default:
			for _, path := range limitCfg.Paths {
				if path == "" {
					continue
				}
				limiter, ok := rlSet.specificMethodMap[path]
				if !ok {
					limiter = newFreqLimiter()
					rlSet.specificMethodMap[path] = limiter
				}
				limiter.limits = append(limiter.limits, limits...)
			}
		}
	}
	return rlSet
}

// FrequencyLimit returns middleware that limits request frequency by transport path.
func FrequencyLimit(rls ...*LimitCfg) middleware.Middleware {
	return FrequencyLimitWithOptions(rls)
}

// FrequencyLimitWithOptions is like FrequencyLimit with extra configuration.
func FrequencyLimitWithOptions(rls []*LimitCfg, opts ...Option) middleware.Middleware {
	o := applyOptions(opts)
	// If a distributed/custom Store is configured via WithStore, use the
	// store-backed limiter so the option is actually honored.
	if o.Store != nil {
		return FrequencyLimitWithStore(o.Store, rls, opts...)
	}
	lSet := newLimitSet(rls...)
	return func(handler middleware.ServiceHandler) middleware.ServiceHandler {
		return func(ctx context.Context, req interface{}) (rsp interface{}, err error) {
			tr, ok := transport.FromContext(ctx)
			if !ok {
				if o.FailOpen {
					return handler(ctx, req)
				}
				return nil, ErrLimitExceed
			}
			if !lSet.WalkAllow(tr.Path()) {
				return nil, ErrLimitExceed
			}
			return handler(ctx, req)
		}
	}
}

// FrequencyLimitWithStore uses Store for distributed limiting while preserving LimitCfg matching.
func FrequencyLimitWithStore(store Store, rls []*LimitCfg, opts ...Option) middleware.Middleware {
	o := applyOptions(opts)
	if store == nil {
		store = NewMemoryStore()
	}
	o.Store = store
	return func(handler middleware.ServiceHandler) middleware.ServiceHandler {
		return func(ctx context.Context, req interface{}) (rsp interface{}, err error) {
			tr, ok := transport.FromContext(ctx)
			if !ok {
				if o.FailOpen {
					return handler(ctx, req)
				}
				return nil, ErrLimitExceed
			}
			key := tr.Path()
			for _, cfg := range rls {
				if !matchCfg(cfg, key) {
					continue
				}
				for _, f := range cfg.Freqs {
					storeKey := RateLimitKey(o.ServiceName, key, f.Period)
					allowed, allowErr := o.Store.Allow(ctx, storeKey, f.Period, f.Limit)
					if allowErr != nil {
						if o.FailOpen {
							continue
						}
						return nil, allowErr
					}
					if !allowed {
						return nil, ErrLimitExceed
					}
				}
			}
			return handler(ctx, req)
		}
	}
}

func matchCfg(cfg *LimitCfg, key string) bool {
	switch cfg.Kind {
	case MatchKindAll:
		return true
	case MatchKindRegular:
		if cfg.Pattern == "" {
			return false
		}
		re, err := regexp.Compile(cfg.Pattern)
		return err == nil && re.MatchString(key)
	default:
		for _, p := range cfg.Paths {
			if p == key {
				return true
			}
		}
		return false
	}
}
