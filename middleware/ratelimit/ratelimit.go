package ratelimit

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
)

var ErrLimitExceed = errors.New(429, "RATELIMIT", "service unavailable due to rate limit exceeded")

type Cycle struct {
	// Unit is seconds
	Period uint32
	Limit  uint32
}
type RateLimit struct {
	// If this value is true, all methods match this rate limit.
	// And when true, there is no need to set the FullMethod value
	All bool
	// If the value is true, a regular expression is used to match,
	// and the value of the expression is the value of FullMethod.
	Regexp bool `json:"regexp"`
	// Set rate limit based on specified fullname method
	// example: /helloworld.Greeter/SayHello
	FullMethod string   `json:"method"`
	Cycles     []*Cycle `json:"cycles"`
}

type RateLimiter struct {
	period uint32
	limit  uint32
	count  uint32
	offset uint32
}

func NewRateLimiter(period, limit uint32) *RateLimiter {
	return &RateLimiter{
		period: period,
		limit:  limit,
		count:  0,
		offset: uint32(time.Now().Unix()),
	}
}

func (r *RateLimiter) Allow() bool {
	now := uint32(time.Now().Unix())
	if r.offset+r.period < now {
		r.offset = now
		r.count = 1
		return true
	}
	r.count++
	return r.count <= r.limit
}

type RateLimitSettings struct {
	allMethod         []*RateLimiter
	specificMethodMap map[string][]*RateLimiter
	reMap             map[*regexp.Regexp][]*RateLimiter
	checkedMethod     map[string]bool
}

func (s *RateLimitSettings) WalkAllow(key string) bool {
	if len(s.allMethod) > 0 {
		for _, r := range s.allMethod {
			if !r.Allow() {
				return false
			}
		}
	}
	limiterList, ok := s.specificMethodMap[key]
	if !ok {
		if len(s.reMap) > 0 && !s.checkedMethod[key] {
			s.checkedMethod[key] = true
			for re, limiters := range s.reMap {
				if re.MatchString(key) {
					limiterList = limiters
					s.specificMethodMap[key] = limiters
				}
			}
		} else {
			return true
		}
	}
	for _, r := range limiterList {
		if !r.Allow() {
			return false
		}
	}
	return true
}

func Server(routes ...*RateLimit) middleware.Middleware {
	rlSt := &RateLimitSettings{
		specificMethodMap: map[string][]*RateLimiter{},
	}
	for _, ratelimitCfg := range routes {
		if ratelimitCfg.All {
			for _, cycle := range ratelimitCfg.Cycles {
				if cycle.Period == 0 || cycle.Limit == 0 {
					continue
				}
				rlSt.allMethod = append(rlSt.allMethod,
					NewRateLimiter(cycle.Period, cycle.Limit))
			}
			continue
		}
		var limiter []*RateLimiter
		for _, cycle := range ratelimitCfg.Cycles {
			if cycle.Period == 0 || cycle.Limit == 0 {
				continue
			}
			limiter = append(limiter, NewRateLimiter(cycle.Period, cycle.Limit))
		}
		if ratelimitCfg.Regexp {
			if rlSt.reMap == nil {
				rlSt.reMap = make(map[*regexp.Regexp][]*RateLimiter)
				rlSt.checkedMethod = map[string]bool{}
			}
			rlSt.reMap[regexp.MustCompile(ratelimitCfg.FullMethod)] = limiter
			fmt.Printf("%v", rlSt.reMap)
		} else {
			rlSt.specificMethodMap[ratelimitCfg.FullMethod] = limiter
		}
	}

	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			tr, ok := transport.FromServerContext(ctx)
			if ok && !rlSt.WalkAllow(tr.Operation()) {
				return nil, ErrLimitExceed
			}
			return handler(ctx, req)
		}
	}
}
