package ratelimit

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ml444/gkit/transport"
	"github.com/ml444/gkit/transport/httpx"
)

func TestNewRateLimiter(t *testing.T) {
	type args struct {
		period int64
		limit  uint64
	}
	tests := []struct {
		name string
		args args
		want *periodLimit
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newPeriodLimit(tt.args.period, tt.args.limit); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewRateLimiter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func getInt64(v int64) *atomic.Int64 {
	i64V := &atomic.Int64{}
	i64V.Store(v)
	return i64V
}

func getUint64(v uint64) *atomic.Uint64 {
	u64V := &atomic.Uint64{}
	u64V.Store(v)
	return u64V
}

func TestRateLimiter_over(t *testing.T) {
	type fields struct {
		period    int64
		limit     uint64
		count     *atomic.Uint64
		timestamp *atomic.Int64
	}

	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "10/s-ok",
			fields: fields{
				period:    int64(time.Duration(time.Second).Milliseconds()),
				limit:     10,
				count:     getUint64(0),
				timestamp: getInt64(time.Now().UnixMilli()),
			},
			want: false,
		},
		{
			name: "100/m-ok",
			fields: fields{
				period:    int64(time.Duration(time.Second * 60).Milliseconds()),
				limit:     100,
				count:     &atomic.Uint64{},
				timestamp: getInt64(time.Now().UnixMilli()),
			},
			want: false,
		},
		{
			name: "100/m-over",
			fields: fields{
				period:    int64(time.Duration(time.Minute * 1).Milliseconds()),
				limit:     100,
				count:     getUint64(100),
				timestamp: getInt64(time.Now().Add(-time.Second * 59).UnixMilli()),
			},
			want: true,
		},
		{
			name: "100/h-ok",
			fields: fields{
				period:    int64(time.Duration(time.Hour * 1).Milliseconds()),
				limit:     100,
				count:     &atomic.Uint64{},
				timestamp: getInt64(time.Now().UnixMilli()),
			},
			want: false,
		},
		{
			name: "100/h-over",
			fields: fields{
				period:    int64(time.Duration(time.Hour * 1).Milliseconds()),
				limit:     100,
				count:     getUint64(100),
				timestamp: getInt64(time.Now().UnixMilli()),
			},
			want: true,
		},
		{
			name: "100/m-delay-ok",
			fields: fields{
				period:    int64(time.Duration(time.Minute * 1).Milliseconds()),
				limit:     100,
				count:     getUint64(100),
				timestamp: getInt64(time.Now().Add(-time.Second * 61).UnixMilli()),
			},
			want: false,
		},
		{
			name: "100/m-delay-over",
			fields: fields{
				period:    int64(time.Duration(time.Minute * 1).Milliseconds()),
				limit:     100,
				count:     getUint64(100),
				timestamp: getInt64(time.Now().UnixMilli() - 59000),
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &periodLimit{
				period:    tt.fields.period,
				limit:     tt.fields.limit,
				count:     tt.fields.count,
				timestamp: tt.fields.timestamp,
			}
			now := time.Now().UnixMilli()
			if got := r.over(now); got != tt.want {
				t.Logf("period: %d, ts: %v, now: %v, count: %v \n", r.period, r.timestamp.Load(), now, r.count.Load())
				t.Errorf("%s periodLimit.over() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func Test_limitSettings_WalkAllow(t *testing.T) {
	type fields struct {
		allMethod         *freqLimiter
		reMap             map[*regexp.Regexp]*freqLimiter
		specificMethodMap map[string]*freqLimiter
		checkedMethod     map[string]bool
	}
	type args struct {
		key string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &limitSet{
				allMethod:         tt.fields.allMethod,
				reMap:             tt.fields.reMap,
				specificMethodMap: tt.fields.specificMethodMap,
				checkedMethod:     tt.fields.checkedMethod,
			}
			if got := s.WalkAllow(tt.args.key); got != tt.want {
				t.Errorf("RateLimitSettings.WalkAllow() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_newLimitSet(t *testing.T) {
	type args struct {
		rls []*LimitCfg
	}
	tests := []struct {
		name string
		args args
		want *limitSet
	}{
		{
			name: "all",
			args: args{
				rls: []*LimitCfg{{
					Kind:  MatchKindAll,
					Freqs: []*Frequency{{Period: time.Millisecond * 10, Limit: 2}},
				}},
			},
			want: &limitSet{
				specificMethodMap: map[string]*freqLimiter{},
				allMethod: &freqLimiter{
					limits: []*periodLimit{{
						period:    10,
						limit:     2,
						count:     getUint64(0),
						timestamp: getInt64(time.Now().UnixMilli()),
					}},
				},
			},
		},
		{
			name: "regexp",
			args: args{
				rls: []*LimitCfg{{
					Kind:    MatchKindRegular,
					Pattern: "/abc/\\w+",
					Freqs:   []*Frequency{{Period: time.Millisecond * 10, Limit: 2}},
				}},
			},
			want: &limitSet{
				specificMethodMap: map[string]*freqLimiter{},
				reMap: map[*regexp.Regexp]*freqLimiter{
					regexp.MustCompile(`/abc/\w+`): {
						// re: regexp.MustCompile(`/abc/\w+`),
						limits: []*periodLimit{{
							period:    10,
							limit:     2,
							count:     getUint64(0),
							timestamp: getInt64(time.Now().UnixMilli()),
						}},
					},
				},
				checkedMethod: map[string]bool{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newLimitSet(tt.args.rls...)
			if tt.name == "regexp" {
				for re, limiter := range got.reMap {
					t.Logf("%+v, %+v \n", re, limiter.limits[0].count.Load())
				}
				for re, limiter := range tt.want.reMap {
					t.Logf("%+v, %+v \n", re, limiter.limits[0].count.Load())
				}
			} else {
				if !reflect.DeepEqual(got, tt.want) {
					var g, w interface{}
					g = got
					w = tt.want

					t.Errorf("newLimitSet() = %+v, want %+v", g, w)
				}
			}
		})
	}
}

func TestServer(t *testing.T) {
	rls := []*LimitCfg{
		{
			Kind:  MatchKindAll,
			Freqs: []*Frequency{{Period: time.Millisecond, Limit: 1}},
		},
		{
			Kind:    MatchKindRegular,
			Pattern: `/user.abc/\w+/\d+`,
			Freqs: []*Frequency{
				{time.Millisecond * 30, 2},
				{time.Millisecond * 100, 4},
				{time.Millisecond * 50, 3},
			},
		},
		{
			Kind:  MatchKindExact,
			Paths: []string{"/user.abc/foo/123"},
			Freqs: []*Frequency{
				{time.Millisecond * 20, 10},
				{5 * time.Millisecond, 4},
				{10 * time.Millisecond, 6},
			},
		},
	}
	h := func(ctx context.Context, req interface{}) (rsp interface{}, err error) {
		// t.Log("===> doing something...", req)
		return nil, nil
	}
	mw := FrequencyLimit(rls...)
	run := func(ctx context.Context, allow, n int) {
		fmt.Println("==> running:", allow, n)
		allowCount := 0
		limitCount := 0
		limit := n - allow
		for i := 0; i < n; i++ {
			_, err := mw(h)(ctx, fmt.Sprintf("allow[%d], limit[%d]", allow, limit))
			if err != nil {
				if !errors.Is(err, ErrLimitExceed) {
					fmt.Println(err)
					return
				} else {
					limitCount++
				}
			} else {
				allowCount++
			}
			time.Sleep(time.Microsecond * 100)
		}
		if allowCount != allow || limitCount != limit {
			fmt.Printf(" allow: %d != %d, limit: %d != %d \n", allowCount, allow, limitCount, limit)
		} else {
			fmt.Printf("pass!!! allow: %d, limit: %d \n", allow, limit)
		}
	}

	ctx1 := transport.ToContext(context.Background(), &httpx.Transport{
		BaseTransport: transport.BaseTransport{
			Endpoint:  "",
			Operation: "/user.abc/efg/123",
		},
	})
	go run(ctx1, 8, 1960)

	ctx2 := transport.ToContext(context.Background(), &httpx.Transport{
		BaseTransport: transport.BaseTransport{
			Endpoint:  "",
			Operation: "/user.abc/foo/123",
		},
	})
	run(ctx2, 260, 5000)
}
