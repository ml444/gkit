package ratelimit

import (
	"reflect"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/middleware"
)

func TestNewRateLimiter(t *testing.T) {
	type args struct {
		period uint32
		limit  uint32
	}
	tests := []struct {
		name string
		args args
		want *RateLimiter
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewRateLimiter(tt.args.period, tt.args.limit); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewRateLimiter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRateLimiter_Allow(t *testing.T) {
	type fields struct {
		period uint32
		limit  uint32
		count  uint32
		offset uint32
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "100/s-true",
			fields: fields{
				period: uint32(time.Duration(time.Second).Seconds()),
				limit:  10,
				count:  9,
				offset: uint32(time.Now().Unix()),
			},
			want: true,
		},
		{
			name: "100/m-true",
			fields: fields{
				period: uint32(time.Duration(time.Second * 60).Seconds()),
				limit:  100,
				count:  0,
				offset: uint32(time.Now().Unix()),
			},
			want: true,
		},
		{
			name: "100/m-false",
			fields: fields{
				period: uint32(time.Duration(time.Second * 60).Seconds()),
				limit:  100,
				count:  100,
				offset: uint32(time.Now().Unix()),
			},
			want: false,
		},
		{
			name: "100/h-ok",
			fields: fields{
				period: uint32(time.Duration(time.Hour * 1).Seconds()),
				limit:  100,
				count:  99,
				offset: uint32(time.Now().Unix()),
			},
			want: true,
		},
		{
			name: "100/h-false",
			fields: fields{
				period: uint32(time.Duration(time.Hour * 1).Seconds()),
				limit:  100,
				count:  100,
				offset: uint32(time.Now().Unix()),
			},
			want: false,
		},
		{
			name: "100/m-delay",
			fields: fields{
				period: uint32(time.Duration(time.Minute * 1).Seconds()),
				limit:  100,
				count:  100,
				offset: uint32(time.Now().Unix() - 60),
			},
			want: false,
		},
		{
			name: "100/m-delay-ok",
			fields: fields{
				period: uint32(time.Duration(time.Minute * 1).Seconds()),
				limit:  100,
				count:  100,
				offset: uint32(time.Now().Unix() - 61),
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RateLimiter{
				period: tt.fields.period,
				limit:  tt.fields.limit,
				count:  tt.fields.count,
				offset: tt.fields.offset,
			}
			if got := r.Allow(); got != tt.want {
				t.Errorf("RateLimiter.Allow() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRateLimitSettings_WalkAllow(t *testing.T) {
	type fields struct {
		EnabledRouteMap map[string][]*RateLimiter
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
			s := &RateLimitSettings{
				specificMethodMap: tt.fields.EnabledRouteMap,
			}
			if got := s.WalkAllow(tt.args.key); got != tt.want {
				t.Errorf("RateLimitSettings.WalkAllow() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServer(t *testing.T) {
	type args struct {
		routes []*RateLimit
	}
	tests := []struct {
		name string
		args args
		want middleware.Middleware
	}{
		{
			name: "1",
			args: args{
				routes: []*RateLimit{{
					All:        true,
					Regexp:     false,
					FullMethod: "",
					Cycles:     []*Cycle{{Period: 10, Limit: 2}},
				}},
			},
			want: func(middleware.Handler) middleware.Handler {
				return nil
			},
		},
		{
			name: "regexp",
			args: args{
				routes: []*RateLimit{{
					All:        false,
					Regexp:     true,
					FullMethod: "/abc/\\w+",
					Cycles:     []*Cycle{{Period: 10, Limit: 2}},
				}},
			},
			want: func(middleware.Handler) middleware.Handler {
				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Server(tt.args.routes...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Server() = %v, want %v", got, tt.want)
			}
		})
	}
}
