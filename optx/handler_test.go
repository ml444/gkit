package optx

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"testing"
)

func Test_noneHandler_Apply(t *testing.T) {
	type fields struct {
		fn func() error
	}
	type args struct {
		in0 interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "ok",
			fields: fields{fn: func() error {
				fmt.Println("===> ok <====")
				return nil
			}},
			args:    args{in0: ""},
			wantErr: false,
		},
		{
			name: "err",
			fields: fields{fn: func() error {
				return errors.New("test error")
			}},
			args:    args{in0: ""},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewNone(tt.fields.fn)
			if err := h.Apply(tt.args.in0); (err != nil) != tt.wantErr {
				t.Errorf("noneHandler.Apply() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_boolHandler_Apply(t *testing.T) {
	type fields struct {
		fn func(bool) error
	}
	type args struct {
		v interface{}
	}
	boolFunc := func(val bool) error {
		return nil
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "ok",
			fields:  fields{fn: boolFunc},
			args:    args{v: true},
			wantErr: false,
		},
		{
			name:    "ok-string",
			fields:  fields{fn: boolFunc},
			args:    args{v: "true"},
			wantErr: false,
		},
		{
			name:    "err-string",
			fields:  fields{fn: boolFunc},
			args:    args{v: "foo"},
			wantErr: true,
		},
		{
			name:    "err-int",
			fields:  fields{fn: boolFunc},
			args:    args{v: 123},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewBool(tt.fields.fn)
			if err := h.Apply(tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("boolHandler.Apply() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_stringHandler_Apply(t *testing.T) {
	type fields struct {
		fn         func(string) error
		ignoreZero bool
	}
	type args struct {
		v interface{}
	}
	s := "testString"
	fn := func(val string) error {
		if val == "" {
			return errors.New("test error")
		}
		if val == s {
			return errors.New("valueError")
		}
		return nil
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{name: "ok", fields: fields{fn: fn}, args: args{v: s}, wantErr: false},
		{name: "ok-ptr_string", fields: fields{fn: fn}, args: args{v: &s}, wantErr: false}, // pointer
		{name: "ok-ignoreZero", fields: fields{fn: fn, ignoreZero: true}, args: args{v: ""}, wantErr: false},
		{name: "err-empty-string", fields: fields{fn: fn}, args: args{v: ""}, wantErr: true},
		{name: "err-int", fields: fields{fn: fn}, args: args{v: 123}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewString(tt.fields.fn, tt.fields.ignoreZero)
			if err := h.Apply(tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("stringHandler.Apply() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_stringListHandler_Apply(t *testing.T) {
	type fields struct {
		fn         func([]string) error
		ignoreZero bool
	}
	type args struct {
		v interface{}
	}
	fn := func(val []string) error {
		if len(val) == 0 {
			return errors.New("test error")
		}
		return nil
	}
	var ss []string
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{name: "ok", fields: fields{fn: fn, ignoreZero: false}, args: args{v: []string{"foo", "bar"}}, wantErr: false},
		{name: "ok-empty-string", fields: fields{fn: fn, ignoreZero: true}, args: args{v: ""}, wantErr: false},
		{name: "ok-empty-[]string", fields: fields{fn: fn, ignoreZero: true}, args: args{v: []string{}}, wantErr: false},
		{name: "ok-new([]string)", fields: fields{fn: fn, ignoreZero: true}, args: args{v: new([]string)}, wantErr: false},
		{name: "ok-nil([]string)", fields: fields{fn: fn, ignoreZero: true}, args: args{v: ss}, wantErr: false},
		{name: "err-int", fields: fields{fn: fn, ignoreZero: false}, args: args{v: 123}, wantErr: true},
		{name: "err-[]int32", fields: fields{fn: fn, ignoreZero: false}, args: args{v: []int32{1, 2, 3}}, wantErr: true},
		{name: "err-empty-string", fields: fields{fn: fn, ignoreZero: false}, args: args{v: ""}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &stringListHandler{
				fn:         tt.fields.fn,
				ignoreZero: tt.fields.ignoreZero,
			}
			if err := h.Apply(tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("stringListHandler.Apply() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_int32Handler_Apply(t *testing.T) {
	type fields struct {
		fn         func(int32) error
		ignoreZero bool
	}
	type args struct {
		v interface{}
	}
	var v int32 = 123
	fn := func(val int32) error {
		if val == 0 {
			return errors.New("test error")
		}
		if val != v {
			return errors.New("valueError")
		}
		return nil
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{name: "ok", fields: fields{fn, false}, args: args{v: v}, wantErr: false},
		{name: "ok-ptr", fields: fields{fn, false}, args: args{v: &v}, wantErr: false},
		{name: "ok-string", fields: fields{fn, false}, args: args{v: "123"}, wantErr: false},
		{name: "ok-int32-zero", fields: fields{fn, true}, args: args{v: int32(0)}, wantErr: false},
		{name: "ok-string-empty", fields: fields{fn, true}, args: args{v: ""}, wantErr: false},
		{name: "err-int32-zero", fields: fields{fn, false}, args: args{v: int32(0)}, wantErr: true},
		{name: "err-string-empty", fields: fields{fn, false}, args: args{v: ""}, wantErr: true},
		{name: "err-outOfRange", fields: fields{fn, false}, args: args{v: fmt.Sprintf("%d", math.MaxInt64)}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewInt32(tt.fields.fn, tt.fields.ignoreZero)
			if err := h.Apply(tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("int32Handler.Apply() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_int32ListHandler_Apply(t *testing.T) {
	type fields struct {
		fn         func([]int32) error
		ignoreZero bool
	}
	type args struct {
		v interface{}
	}
	v := []int32{1, 2, 3}
	fn := func(val []int32) error {
		if len(val) == 0 {
			return errors.New("test error")
		}
		if !reflect.DeepEqual(val, v) {
			return errors.New("value error")
		}
		return nil
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{name: "ok", fields: fields{fn, false}, args: args{v: v}, wantErr: false},
		{name: "ok-ptr", fields: fields{fn, false}, args: args{v: &v}, wantErr: false},
		{name: "ok-string", fields: fields{fn, false}, args: args{v: "1,2,3"}, wantErr: false},
		{name: "ok-ignore-len0", fields: fields{fn, true}, args: args{v: []int32{}}, wantErr: false},
		{name: "ok-ignore-ptr-nil", fields: fields{fn, true}, args: args{v: new([]int32)}, wantErr: false},
		{name: "ok-ignore-empty-string", fields: fields{fn, true}, args: args{v: ""}, wantErr: false},
		{name: "err-len0", fields: fields{fn, false}, args: args{v: []int32{}}, wantErr: true},
		{name: "err-ptr-nil", fields: fields{fn, false}, args: args{v: new([]int32)}, wantErr: true},
		{name: "err-empty-string", fields: fields{fn, false}, args: args{v: ""}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewInt32List(tt.fields.fn, tt.fields.ignoreZero)
			if err := h.Apply(tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("int32ListHandler.Apply() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_int32RangeHandler_Apply(t *testing.T) {
	type fields struct {
		fn         func(begin, end int32) error
		ignoreZero bool
	}
	type args struct {
		v interface{}
	}
	v := [2]int32{1, 99}
	fn := func(begin, end int32) error {
		if begin == 0 && end == 0 {
			return errors.New("test error")
		}
		if !reflect.DeepEqual([2]int32{begin, end}, v) {
			t.Log("==> value: ", begin, end)
			return errors.New("value error")
		}
		return nil
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{name: "ok", fields: fields{fn, false}, args: args{v: v}, wantErr: false},
		{name: "ok-ptr", fields: fields{fn, false}, args: args{v: &v}, wantErr: false},
		{name: "ok-string", fields: fields{fn, false}, args: args{v: "1,99"}, wantErr: false},
		{name: "ok-ignore-len0", fields: fields{fn, true}, args: args{v: []int32{}}, wantErr: false},
		{name: "ok-ignore-ptr-nil", fields: fields{fn, true}, args: args{v: new([]int32)}, wantErr: false},
		{name: "ok-ignore-empty-string", fields: fields{fn, true}, args: args{v: ""}, wantErr: false},
		{name: "err-len0", fields: fields{fn, false}, args: args{v: []int32{}}, wantErr: true},
		{name: "err-ptr-nil", fields: fields{fn, false}, args: args{v: new([]int32)}, wantErr: true},
		{name: "err-empty-string", fields: fields{fn, false}, args: args{v: ""}, wantErr: true},
		{name: "err-outOfRange-string", fields: fields{fn, false}, args: args{v: fmt.Sprintf("%d,%d", math.MinInt64, math.MaxInt64)}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &int32RangeHandler{
				fn:         tt.fields.fn,
				ignoreZero: tt.fields.ignoreZero,
			}
			if err := h.Apply(tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("int32RangeHandler.Apply() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_int64Handler_Apply(t *testing.T) {
	type fields struct {
		fn         func(int64) error
		ignoreZero bool
	}
	type args struct {
		v interface{}
	}
	var v int64 = 123
	fn := func(val int64) error {
		if val == 0 {
			return errors.New("test error")
		}
		if val != v {
			return errors.New("valueError")
		}
		return nil
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{name: "ok", fields: fields{fn, false}, args: args{v: v}, wantErr: false},
		{name: "ok-ptr", fields: fields{fn, false}, args: args{v: &v}, wantErr: false},
		{name: "ok-string", fields: fields{fn, false}, args: args{v: "123"}, wantErr: false},
		{name: "ok-int32-zero", fields: fields{fn, true}, args: args{v: int64(0)}, wantErr: false},
		{name: "ok-string-empty", fields: fields{fn, true}, args: args{v: ""}, wantErr: false},
		{name: "err-int32-zero", fields: fields{fn, false}, args: args{v: int64(0)}, wantErr: true},
		{name: "err-string-empty", fields: fields{fn, false}, args: args{v: ""}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &int64Handler{
				fn:         tt.fields.fn,
				ignoreZero: tt.fields.ignoreZero,
			}
			if err := h.Apply(tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("int64Handler.Apply() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_int64ListHandler_Apply(t *testing.T) {
	type fields struct {
		fn         func([]int64) error
		ignoreZero bool
	}
	type args struct {
		v interface{}
	}
	v := []int64{1, 2, 3}
	fn := func(val []int64) error {
		if len(val) == 0 {
			return errors.New("test error")
		}
		if !reflect.DeepEqual(val, v) {
			return errors.New("value error")
		}
		return nil
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{name: "ok", fields: fields{fn, false}, args: args{v: v}, wantErr: false},
		{name: "ok-ptr", fields: fields{fn, false}, args: args{v: &v}, wantErr: false},
		{name: "ok-string", fields: fields{fn, false}, args: args{v: "1,2,3"}, wantErr: false},
		{name: "ok-ignore-len0", fields: fields{fn, true}, args: args{v: []int64{}}, wantErr: false},
		{name: "ok-ignore-ptr-nil", fields: fields{fn, true}, args: args{v: new([]int64)}, wantErr: false},
		{name: "ok-ignore-empty-string", fields: fields{fn, true}, args: args{v: ""}, wantErr: false},
		{name: "err-len0", fields: fields{fn, false}, args: args{v: []int64{}}, wantErr: true},
		{name: "err-ptr-nil", fields: fields{fn, false}, args: args{v: new([]int64)}, wantErr: true},
		{name: "err-empty-string", fields: fields{fn, false}, args: args{v: ""}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &int64ListHandler{
				fn:         tt.fields.fn,
				ignoreZero: tt.fields.ignoreZero,
			}
			if err := h.Apply(tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("int64ListHandler.Apply() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_int64RangeHandler_Apply(t *testing.T) {
	type fields struct {
		fn         func(begin, end int64) error
		ignoreZero bool
	}
	type args struct {
		v interface{}
	}
	v := [2]int64{1, 99}
	fn := func(begin, end int64) error {
		if begin == 0 && end == 0 {
			return errors.New("test error")
		}
		if !reflect.DeepEqual([2]int64{begin, end}, v) {
			return errors.New("value error")
		}
		return nil
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{name: "ok", fields: fields{fn, false}, args: args{v: v}, wantErr: false},
		{name: "ok-ptr", fields: fields{fn, false}, args: args{v: &v}, wantErr: false},
		{name: "ok-string", fields: fields{fn, false}, args: args{v: "1,99"}, wantErr: false},
		{name: "ok-ignore-len0", fields: fields{fn, true}, args: args{v: []string{}}, wantErr: false},
		{name: "ok-ignore-ptr-nil", fields: fields{fn, true}, args: args{v: new([]string)}, wantErr: false},
		{name: "ok-ignore-empty-string", fields: fields{fn, true}, args: args{v: ""}, wantErr: false},
		{name: "err-len0", fields: fields{fn, false}, args: args{v: []string{}}, wantErr: true},
		{name: "err-ptr-nil", fields: fields{fn, false}, args: args{v: new([]string)}, wantErr: true},
		{name: "err-empty-string", fields: fields{fn, false}, args: args{v: ""}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &int64RangeHandler{
				fn:         tt.fields.fn,
				ignoreZero: tt.fields.ignoreZero,
			}
			if err := h.Apply(tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("int64RangeHandler.Apply() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_uint32Handler_Apply(t *testing.T) {
	type fields struct {
		fn         func(uint32) error
		ignoreZero bool
	}
	type args struct {
		v interface{}
	}
	var v uint32 = 123
	fn := func(val uint32) error {
		if val == 0 {
			return errors.New("test error")
		}
		if val != v {
			return errors.New("valueError")
		}
		return nil
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{name: "ok", fields: fields{fn, false}, args: args{v: v}, wantErr: false},
		{name: "ok-ptr", fields: fields{fn, false}, args: args{v: &v}, wantErr: false},
		{name: "ok-string", fields: fields{fn, false}, args: args{v: "123"}, wantErr: false},
		{name: "ok-int32-zero", fields: fields{fn, true}, args: args{v: uint32(0)}, wantErr: false},
		{name: "ok-string-empty", fields: fields{fn, true}, args: args{v: ""}, wantErr: false},
		{name: "err-int32-zero", fields: fields{fn, false}, args: args{v: uint32(0)}, wantErr: true},
		{name: "err-string-empty", fields: fields{fn, false}, args: args{v: ""}, wantErr: true},
		{name: "err-outOfRange", fields: fields{fn, false}, args: args{v: fmt.Sprintf("%d", math.MaxInt64)}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewUint32(tt.fields.fn, tt.fields.ignoreZero)
			if err := h.Apply(tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("handlerUint32.Apply() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_uint32ListHandler_Apply(t *testing.T) {
	type fields struct {
		fn         func([]uint32) error
		ignoreZero bool
	}
	type args struct {
		v interface{}
	}
	v := []uint32{1, 2, 3}
	fn := func(val []uint32) error {
		if len(val) == 0 {
			return errors.New("test error")
		}
		if !reflect.DeepEqual(val, v) {
			return errors.New("value error")
		}
		return nil
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{name: "ok", fields: fields{fn, false}, args: args{v: v}, wantErr: false},
		{name: "ok-ptr", fields: fields{fn, false}, args: args{v: &v}, wantErr: false},
		{name: "ok-string", fields: fields{fn, false}, args: args{v: "1,2,3"}, wantErr: false},
		{name: "ok-ignore-len0", fields: fields{fn, true}, args: args{v: []uint32{}}, wantErr: false},
		{name: "ok-ignore-ptr-nil", fields: fields{fn, true}, args: args{v: new([]uint32)}, wantErr: false},
		{name: "ok-ignore-empty-string", fields: fields{fn, true}, args: args{v: ""}, wantErr: false},
		{name: "err-len0", fields: fields{fn, false}, args: args{v: []uint32{}}, wantErr: true},
		{name: "err-ptr-nil", fields: fields{fn, false}, args: args{v: new([]uint32)}, wantErr: true},
		{name: "err-empty-string", fields: fields{fn, false}, args: args{v: ""}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewUint32List(tt.fields.fn, tt.fields.ignoreZero)
			if err := h.Apply(tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("uint32ListHandler.Apply() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_uint32RangeHandler_Apply(t *testing.T) {
	type fields struct {
		fn         func(begin, end uint32) error
		ignoreZero bool
	}
	type args struct {
		v interface{}
	}
	v := [2]uint32{1, 99}
	fn := func(begin, end uint32) error {
		if begin == 0 && end == 0 {
			return errors.New("test error")
		}
		if !reflect.DeepEqual([2]uint32{begin, end}, v) {
			t.Log("==> value: ", begin, end)
			return errors.New("value error")
		}
		return nil
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{name: "ok", fields: fields{fn, false}, args: args{v: v}, wantErr: false},
		{name: "ok-ptr", fields: fields{fn, false}, args: args{v: &v}, wantErr: false},
		{name: "ok-string", fields: fields{fn, false}, args: args{v: "1,99"}, wantErr: false},
		{name: "ok-ignore-len0", fields: fields{fn, true}, args: args{v: []uint32{}}, wantErr: false},
		{name: "ok-ignore-ptr-nil", fields: fields{fn, true}, args: args{v: new([]uint32)}, wantErr: false},
		{name: "ok-ignore-empty-string", fields: fields{fn, true}, args: args{v: ""}, wantErr: false},
		{name: "err-len0", fields: fields{fn, false}, args: args{v: []uint32{}}, wantErr: true},
		{name: "err-ptr-nil", fields: fields{fn, false}, args: args{v: new([]uint32)}, wantErr: true},
		{name: "err-empty-string", fields: fields{fn, false}, args: args{v: ""}, wantErr: true},
		{name: "err-outOfRange-string-max", fields: fields{fn, false}, args: args{v: fmt.Sprintf("0,%d", math.MaxInt64)}, wantErr: true},
		{name: "err-outOfRange-string-min", fields: fields{fn, false}, args: args{v: fmt.Sprintf("%d,0", math.MinInt64)}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &uint32RangeHandler{
				fn:         tt.fields.fn,
				ignoreZero: tt.fields.ignoreZero,
			}
			if err := h.Apply(tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("uint32RangeHandler.Apply() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_uint64Handler_Apply(t *testing.T) {
	type fields struct {
		fn         func(uint64) error
		ignoreZero bool
	}
	type args struct {
		v interface{}
	}
	v := uint64(1)
	fn := func(val uint64) error {
		if val == 0 {
			return errors.New("test error")
		}
		if !reflect.DeepEqual(val, v) {
			return errors.New("value error")
		}
		return nil
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{name: "ok", fields: fields{fn, false}, args: args{v: v}, wantErr: false},
		{name: "ok-ptr", fields: fields{fn, false}, args: args{v: &v}, wantErr: false},
		{name: "ok-string", fields: fields{fn, false}, args: args{v: "1"}, wantErr: false},
		{name: "ok-ignore-len0", fields: fields{fn, true}, args: args{v: uint64(0)}, wantErr: false},
		{name: "ok-ignore-ptr-nil", fields: fields{fn, true}, args: args{v: new(uint64)}, wantErr: false},
		{name: "ok-ignore-empty-string", fields: fields{fn, true}, args: args{v: ""}, wantErr: false},
		{name: "err-len0", fields: fields{fn, false}, args: args{v: uint64(0)}, wantErr: true},
		{name: "err-ptr-nil", fields: fields{fn, false}, args: args{v: new(uint64)}, wantErr: true},
		{name: "err-empty-string", fields: fields{fn, false}, args: args{v: ""}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &uint64Handler{
				fn:         tt.fields.fn,
				ignoreZero: tt.fields.ignoreZero,
			}
			if err := h.Apply(tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("uint64Handler.Apply() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_uint64ListHandler_Apply(t *testing.T) {
	type fields struct {
		fn         func([]uint64) error
		ignoreZero bool
	}
	type args struct {
		v interface{}
	}
	v := []uint64{1, 2, 3}
	fn := func(val []uint64) error {
		if len(val) == 0 {
			return errors.New("test error")
		}
		if !reflect.DeepEqual(val, v) {
			return errors.New("value error")
		}
		return nil
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{name: "ok", fields: fields{fn, false}, args: args{v: v}, wantErr: false},
		{name: "ok-ptr", fields: fields{fn, false}, args: args{v: &v}, wantErr: false},
		{name: "ok-string", fields: fields{fn, false}, args: args{v: "1,2,3"}, wantErr: false},
		{name: "ok-ignore-len0", fields: fields{fn, true}, args: args{v: []uint64{}}, wantErr: false},
		{name: "ok-ignore-ptr-nil", fields: fields{fn, true}, args: args{v: new([]uint64)}, wantErr: false},
		{name: "ok-ignore-empty-string", fields: fields{fn, true}, args: args{v: ""}, wantErr: false},
		{name: "err-len0", fields: fields{fn, false}, args: args{v: []uint64{}}, wantErr: true},
		{name: "err-ptr-nil", fields: fields{fn, false}, args: args{v: new([]uint64)}, wantErr: true},
		{name: "err-empty-string", fields: fields{fn, false}, args: args{v: ""}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &uint64ListHandler{
				fn:         tt.fields.fn,
				ignoreZero: tt.fields.ignoreZero,
			}
			if err := h.Apply(tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("uint64ListHandler.Apply() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_uint64RangeHandler_Apply(t *testing.T) {
	type fields struct {
		fn         func(begin, end uint64) error
		ignoreZero bool
	}
	type args struct {
		v interface{}
	}
	v := [2]uint64{1, 99}
	fn := func(begin, end uint64) error {
		if begin == 0 && end == 0 {
			return errors.New("test error")
		}
		if !reflect.DeepEqual([2]uint64{begin, end}, v) {
			t.Log("==> value: ", begin, end)
			return errors.New("value error")
		}
		return nil
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{name: "ok", fields: fields{fn, false}, args: args{v: v}, wantErr: false},
		{name: "ok-ptr", fields: fields{fn, false}, args: args{v: &v}, wantErr: false},
		{name: "ok-string", fields: fields{fn, false}, args: args{v: "1,99"}, wantErr: false},
		{name: "ok-ignore-len0", fields: fields{fn, true}, args: args{v: []uint64{}}, wantErr: false},
		{name: "ok-ignore-ptr-nil", fields: fields{fn, true}, args: args{v: new([]uint64)}, wantErr: false},
		{name: "ok-ignore-empty-string", fields: fields{fn, true}, args: args{v: ""}, wantErr: false},
		{name: "err-len0", fields: fields{fn, false}, args: args{v: []uint64{}}, wantErr: true},
		{name: "err-ptr-nil", fields: fields{fn, false}, args: args{v: new([]uint64)}, wantErr: true},
		{name: "err-empty-string", fields: fields{fn, false}, args: args{v: ""}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewUint64Range(tt.fields.fn, tt.fields.ignoreZero)
			if err := h.Apply(tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("uint64RangeHandler.Apply() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
