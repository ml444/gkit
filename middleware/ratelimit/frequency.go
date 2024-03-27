package ratelimit

/*
The frequency limit has three matching methods (all, regular, and exact) to match
the full-method value of the API.
`All` counts the total number of hits for all APIs.
There are no independent statistical restrictions on specific matching APIs.
`Exact` and `Regular` use specific APIs as statistical units.
If `exact` and `regular` hit the same interface, the exact limit takes precedence.

频率限制有三种匹配方式（全部、正则、精确）来匹配API的full-method的值
`全部`针对命中的所有API的总数统计。不分别对匹配的具体API做独立的统计限制。
`精确`和`正则`是以具体的接口为统计单元。
如果`精确`和`正则`命中同一个接口，以精确的限制为主。
*/

import (
	"context"
	"regexp"
	"sort"
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
	// Unit is milliseconds
	Period time.Duration
	Limit  uint64
}
type LimitCfg struct {
	// MatchKindTerm: 	Exactly matches the value of the field `Paths`.
	// MatchKindRegexp: The field `Pattern` is used as a regular expression.
	// MatchKindAll: 	All methods match this limit. The field `Paths` can't be set.
	Kind    MatchKind
	Pattern string

	// example: /helloworld.Greeter/SayHello
	// example: /v1/sayhello/{name}
	Paths []string
	// The same matching rule can set multiple frequency
	// example: [{Period: 60, Limit: 10}, {Period: 3600, Limit: 100}]
	Freqs []*Frequency
}

type periodLimit struct {
	period    int64
	limit     uint64
	timestamp *atomic.Int64
	count     *atomic.Uint64
}

func newPeriodLimit(period int64, limit uint64) *periodLimit {
	tsNow := &atomic.Int64{}
	tsNow.Add(time.Now().UnixMilli())
	return &periodLimit{
		period:    period,
		limit:     limit,
		timestamp: tsNow,
		count:     &atomic.Uint64{},
	}
}

func (p *periodLimit) reset(now int64) {
	p.count.Store(1)
	p.timestamp.Store(now)
}

func (p *periodLimit) over(now int64) bool {
	count := p.count.Add(1)
	if p.timestamp.Load()+p.period > now {
		return count > p.limit
	}
	p.reset(now)
	return false
}

type freqLimiter struct {
	// total  *atomic.Uint64
	limits []*periodLimit
}

func newFreqLimiter() *freqLimiter {
	return &freqLimiter{}
}

func (rl *freqLimiter) sortLimits() {
	sort.SliceStable(rl.limits, func(i, j int) bool {
		return rl.limits[i].period < rl.limits[j].period
	})
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
	allMethod         *freqLimiter
	reMap             map[*regexp.Regexp]*freqLimiter
	specificMethodMap map[string]*freqLimiter
	checkedMethod     map[string]bool
}

func (s *limitSet) WalkAllow(key string) bool {
	if s.allMethod != nil {
		if !s.allMethod.Allow() {
			return false
		}
	}
	limiter, ok := s.specificMethodMap[key]
	if ok {
		return limiter.Allow()
	}
	if len(s.reMap) <= 0 || s.checkedMethod[key] {
		return true
	}
	s.checkedMethod[key] = true
	for re, reLimiter := range s.reMap {
		if !re.MatchString(key) {
			continue
		}
		limiter, ok = s.specificMethodMap[key]
		if !ok {
			limiter = newFreqLimiter()
			s.specificMethodMap[key] = limiter
		}
		for _, limit := range reLimiter.limits {
			limiter.limits = append(
				limiter.limits,
				newPeriodLimit(limit.period, limit.limit),
			)
		}
	}
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
		// Sort by period from small to large
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

func FrequencyLimit(rls ...*LimitCfg) middleware.Middleware {
	lSet := newLimitSet(rls...)
	return func(handler middleware.ServiceHandler) middleware.ServiceHandler {
		return func(ctx context.Context, req interface{}) (rsp interface{}, err error) {
			tr, ok := transport.FromContext(ctx)
			if ok && !lSet.WalkAllow(tr.GetOperation()) {
				return nil, ErrLimitExceed
			}
			return handler(ctx, req)
		}
	}
}
