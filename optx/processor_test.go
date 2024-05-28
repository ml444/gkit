package optx

import (
	"fmt"
	"math"
	"reflect"
	"testing"

	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
)

func TestProcessor_SetHandler(t *testing.T) {
	type fields struct {
		keyMap     map[int32]string
		handlerMap map[string]Handler
	}
	type args struct {
		key interface{}
		h   Handler
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Processor{
				keyMap:     tt.fields.keyMap,
				handlerMap: tt.fields.handlerMap,
			}
			p.SetHandler(tt.args.key, tt.args.h)
		})
	}
}

func TestProcessor_GetHandler(t *testing.T) {
	type fields struct {
		keyMap     map[int32]string
		handlerMap map[string]Handler
	}
	type args struct {
		key interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		wantH  Handler
		wantOk bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Processor{
				keyMap:     tt.fields.keyMap,
				handlerMap: tt.fields.handlerMap,
			}
			gotH, gotOk := p.GetHandler(tt.args.key)
			if !reflect.DeepEqual(gotH, tt.wantH) {
				t.Errorf("Processor.GetHandler() gotH = %v, want %v", gotH, tt.wantH)
			}
			if gotOk != tt.wantOk {
				t.Errorf("Processor.GetHandler() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func TestProcessor_AddHandle(t *testing.T) {
	type fields struct {
		keyMap     map[int32]string
		handlerMap map[string]Handler
	}
	type args struct {
		key interface{}
		cb  func(val interface{}) error
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Processor
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Processor{
				keyMap:     tt.fields.keyMap,
				handlerMap: tt.fields.handlerMap,
			}
			if got := p.AddHandle(tt.args.key, tt.args.cb); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Processor.AddHandle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProcessor_AddNone(t *testing.T) {
	type fields struct {
		keyMap     map[int32]string
		handlerMap map[string]Handler
	}
	type args struct {
		key interface{}
		cb  func() error
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Processor
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Processor{
				keyMap:     tt.fields.keyMap,
				handlerMap: tt.fields.handlerMap,
			}
			if got := p.AddNone(tt.args.key, tt.args.cb); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Processor.AddNone() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProcessor_AddBool(t *testing.T) {
	type fields struct {
		keyMap     map[int32]string
		handlerMap map[string]Handler
	}
	type args struct {
		key interface{}
		cb  func(val bool) error
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Processor
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Processor{
				keyMap:     tt.fields.keyMap,
				handlerMap: tt.fields.handlerMap,
			}
			if got := p.AddBool(tt.args.key, tt.args.cb); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Processor.AddBool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProcessor_AddString(t *testing.T) {
	type fields struct {
		keyMap     map[int32]string
		handlerMap map[string]Handler
	}
	type args struct {
		key interface{}
		cb  func(val string) error
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Processor
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Processor{
				keyMap:     tt.fields.keyMap,
				handlerMap: tt.fields.handlerMap,
			}
			if got := p.AddString(tt.args.key, tt.args.cb); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Processor.AddString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProcessor_AddStringIgnoreZero(t *testing.T) {
	type fields struct {
		keyMap     map[int32]string
		handlerMap map[string]Handler
	}
	type args struct {
		key interface{}
		cb  func(val string) error
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Processor
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Processor{
				keyMap:     tt.fields.keyMap,
				handlerMap: tt.fields.handlerMap,
			}
			if got := p.AddStringIgnoreZero(tt.args.key, tt.args.cb); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Processor.AddStringIgnoreZero() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProcessor_AddStringList(t *testing.T) {
	type fields struct {
		keyMap     map[int32]string
		handlerMap map[string]Handler
	}
	type args struct {
		key interface{}
		cb  func(val []string) error
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Processor
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Processor{
				keyMap:     tt.fields.keyMap,
				handlerMap: tt.fields.handlerMap,
			}
			if got := p.AddStringList(tt.args.key, tt.args.cb); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Processor.AddStringList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProcessor_AddStringListIgnoreZero(t *testing.T) {
	type fields struct {
		keyMap     map[int32]string
		handlerMap map[string]Handler
	}
	type args struct {
		key interface{}
		cb  func(val []string) error
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Processor
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Processor{
				keyMap:     tt.fields.keyMap,
				handlerMap: tt.fields.handlerMap,
			}
			if got := p.AddStringListIgnoreZero(tt.args.key, tt.args.cb); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Processor.AddStringListIgnoreZero() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProcessor_AddInt32(t *testing.T) {
	type fields struct {
		keyMap     map[int32]string
		handlerMap map[string]Handler
	}
	type args struct {
		key interface{}
		cb  func(val int32) error
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Processor
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Processor{
				keyMap:     tt.fields.keyMap,
				handlerMap: tt.fields.handlerMap,
			}
			if got := p.AddInt32(tt.args.key, tt.args.cb); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Processor.AddInt32() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProcessor_AddInt32IgnoreZero(t *testing.T) {
	type fields struct {
		keyMap     map[int32]string
		handlerMap map[string]Handler
	}
	type args struct {
		key interface{}
		cb  func(val int32) error
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Processor
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Processor{
				keyMap:     tt.fields.keyMap,
				handlerMap: tt.fields.handlerMap,
			}
			if got := p.AddInt32IgnoreZero(tt.args.key, tt.args.cb); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Processor.AddInt32IgnoreZero() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProcessor_AddInt32List(t *testing.T) {
	type fields struct {
		keyMap     map[int32]string
		handlerMap map[string]Handler
	}
	type args struct {
		key interface{}
		cb  func(valList []int32) error
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Processor
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Processor{
				keyMap:     tt.fields.keyMap,
				handlerMap: tt.fields.handlerMap,
			}
			if got := p.AddInt32List(tt.args.key, tt.args.cb); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Processor.AddInt32List() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProcessor_AddInt32ListIgnoreZero(t *testing.T) {
	type fields struct {
		keyMap     map[int32]string
		handlerMap map[string]Handler
	}
	type args struct {
		key interface{}
		cb  func(valList []int32) error
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Processor
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Processor{
				keyMap:     tt.fields.keyMap,
				handlerMap: tt.fields.handlerMap,
			}
			if got := p.AddInt32ListIgnoreZero(tt.args.key, tt.args.cb); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Processor.AddInt32ListIgnoreZero() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProcessor_AddInt32Range(t *testing.T) {
	type fields struct {
		keyMap     map[int32]string
		handlerMap map[string]Handler
	}
	type args struct {
		key interface{}
		cb  func(begin, end int32) error
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Processor
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Processor{
				keyMap:     tt.fields.keyMap,
				handlerMap: tt.fields.handlerMap,
			}
			if got := p.AddInt32Range(tt.args.key, tt.args.cb); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Processor.AddInt32Range() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProcessor_AddInt32RangeIgnoreZero(t *testing.T) {
	type fields struct {
		keyMap     map[int32]string
		handlerMap map[string]Handler
	}
	type args struct {
		key interface{}
		cb  func(begin, end int32) error
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Processor
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Processor{
				keyMap:     tt.fields.keyMap,
				handlerMap: tt.fields.handlerMap,
			}
			if got := p.AddInt32RangeIgnoreZero(tt.args.key, tt.args.cb); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Processor.AddInt32RangeIgnoreZero() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProcessor_AddUint32(t *testing.T) {
	type fields struct {
		keyMap     map[int32]string
		handlerMap map[string]Handler
	}
	type args struct {
		key interface{}
		cb  func(val uint32) error
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Processor
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Processor{
				keyMap:     tt.fields.keyMap,
				handlerMap: tt.fields.handlerMap,
			}
			if got := p.AddUint32(tt.args.key, tt.args.cb); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Processor.AddUint32() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProcessor_AddUint32IgnoreZero(t *testing.T) {
	type fields struct {
		keyMap     map[int32]string
		handlerMap map[string]Handler
	}
	type args struct {
		key interface{}
		cb  func(val uint32) error
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Processor
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Processor{
				keyMap:     tt.fields.keyMap,
				handlerMap: tt.fields.handlerMap,
			}
			if got := p.AddUint32IgnoreZero(tt.args.key, tt.args.cb); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Processor.AddUint32IgnoreZero() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProcessor_AddUint32List(t *testing.T) {
	type fields struct {
		keyMap     map[int32]string
		handlerMap map[string]Handler
	}
	type args struct {
		key interface{}
		cb  func(valList []uint32) error
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Processor
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Processor{
				keyMap:     tt.fields.keyMap,
				handlerMap: tt.fields.handlerMap,
			}
			if got := p.AddUint32List(tt.args.key, tt.args.cb); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Processor.AddUint32List() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProcessor_AddUint32ListIgnoreZero(t *testing.T) {
	type fields struct {
		keyMap     map[int32]string
		handlerMap map[string]Handler
	}
	type args struct {
		key interface{}
		cb  func(valList []uint32) error
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Processor
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Processor{
				keyMap:     tt.fields.keyMap,
				handlerMap: tt.fields.handlerMap,
			}
			if got := p.AddUint32ListIgnoreZero(tt.args.key, tt.args.cb); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Processor.AddUint32ListIgnoreZero() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProcessor_AddUint32Range(t *testing.T) {
	type fields struct {
		keyMap     map[int32]string
		handlerMap map[string]Handler
	}
	type args struct {
		key interface{}
		cb  func(begin, end uint32) error
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Processor
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Processor{
				keyMap:     tt.fields.keyMap,
				handlerMap: tt.fields.handlerMap,
			}
			if got := p.AddUint32Range(tt.args.key, tt.args.cb); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Processor.AddUint32Range() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProcessor_AddUint32RangeIgnoreZero(t *testing.T) {
	type fields struct {
		keyMap     map[int32]string
		handlerMap map[string]Handler
	}
	type args struct {
		key interface{}
		cb  func(begin, end uint32) error
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Processor
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Processor{
				keyMap:     tt.fields.keyMap,
				handlerMap: tt.fields.handlerMap,
			}
			if got := p.AddUint32RangeIgnoreZero(tt.args.key, tt.args.cb); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Processor.AddUint32RangeIgnoreZero() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProcessor_AddInt64(t *testing.T) {
	type fields struct {
		keyMap     map[int32]string
		handlerMap map[string]Handler
	}
	type args struct {
		key interface{}
		cb  func(val int64) error
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Processor
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Processor{
				keyMap:     tt.fields.keyMap,
				handlerMap: tt.fields.handlerMap,
			}
			if got := p.AddInt64(tt.args.key, tt.args.cb); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Processor.AddInt64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProcessor_AddInt64IgnoreZero(t *testing.T) {
	type fields struct {
		keyMap     map[int32]string
		handlerMap map[string]Handler
	}
	type args struct {
		key interface{}
		cb  func(val int64) error
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Processor
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Processor{
				keyMap:     tt.fields.keyMap,
				handlerMap: tt.fields.handlerMap,
			}
			if got := p.AddInt64IgnoreZero(tt.args.key, tt.args.cb); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Processor.AddInt64IgnoreZero() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProcessor_AddInt64List(t *testing.T) {
	type fields struct {
		keyMap     map[int32]string
		handlerMap map[string]Handler
	}
	type args struct {
		key interface{}
		cb  func(valList []int64) error
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Processor
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Processor{
				keyMap:     tt.fields.keyMap,
				handlerMap: tt.fields.handlerMap,
			}
			if got := p.AddInt64List(tt.args.key, tt.args.cb); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Processor.AddInt64List() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProcessor_AddInt64ListIgnoreZero(t *testing.T) {
	type fields struct {
		keyMap     map[int32]string
		handlerMap map[string]Handler
	}
	type args struct {
		key interface{}
		cb  func(valList []int64) error
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Processor
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Processor{
				keyMap:     tt.fields.keyMap,
				handlerMap: tt.fields.handlerMap,
			}
			if got := p.AddInt64ListIgnoreZero(tt.args.key, tt.args.cb); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Processor.AddInt64ListIgnoreZero() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProcessor_AddInt64Range(t *testing.T) {
	type fields struct {
		keyMap     map[int32]string
		handlerMap map[string]Handler
	}
	type args struct {
		key interface{}
		cb  func(begin, end int64) error
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Processor
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Processor{
				keyMap:     tt.fields.keyMap,
				handlerMap: tt.fields.handlerMap,
			}
			if got := p.AddInt64Range(tt.args.key, tt.args.cb); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Processor.AddInt64Range() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProcessor_AddInt64RangeIgnoreZero(t *testing.T) {
	type fields struct {
		keyMap     map[int32]string
		handlerMap map[string]Handler
	}
	type args struct {
		key interface{}
		cb  func(begin, end int64) error
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Processor
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Processor{
				keyMap:     tt.fields.keyMap,
				handlerMap: tt.fields.handlerMap,
			}
			if got := p.AddInt64RangeIgnoreZero(tt.args.key, tt.args.cb); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Processor.AddInt64RangeIgnoreZero() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProcessor_AddUint64(t *testing.T) {
	type fields struct {
		keyMap     map[int32]string
		handlerMap map[string]Handler
	}
	type args struct {
		key interface{}
		cb  func(val uint64) error
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Processor
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Processor{
				keyMap:     tt.fields.keyMap,
				handlerMap: tt.fields.handlerMap,
			}
			if got := p.AddUint64(tt.args.key, tt.args.cb); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Processor.AddUint64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProcessor_AddUint64IgnoreZero(t *testing.T) {
	type fields struct {
		keyMap     map[int32]string
		handlerMap map[string]Handler
	}
	type args struct {
		key interface{}
		cb  func(val uint64) error
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Processor
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Processor{
				keyMap:     tt.fields.keyMap,
				handlerMap: tt.fields.handlerMap,
			}
			if got := p.AddUint64IgnoreZero(tt.args.key, tt.args.cb); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Processor.AddUint64IgnoreZero() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProcessor_AddUint64List(t *testing.T) {
	type fields struct {
		keyMap     map[int32]string
		handlerMap map[string]Handler
	}
	type args struct {
		key interface{}
		cb  func(valList []uint64) error
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Processor
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Processor{
				keyMap:     tt.fields.keyMap,
				handlerMap: tt.fields.handlerMap,
			}
			if got := p.AddUint64List(tt.args.key, tt.args.cb); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Processor.AddUint64List() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProcessor_AddUint64ListIgnoreZero(t *testing.T) {
	type fields struct {
		keyMap     map[int32]string
		handlerMap map[string]Handler
	}
	type args struct {
		key interface{}
		cb  func(valList []uint64) error
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Processor
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Processor{
				keyMap:     tt.fields.keyMap,
				handlerMap: tt.fields.handlerMap,
			}
			if got := p.AddUint64ListIgnoreZero(tt.args.key, tt.args.cb); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Processor.AddUint64ListIgnoreZero() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProcessor_AddUint64Range(t *testing.T) {
	type fields struct {
		keyMap     map[int32]string
		handlerMap map[string]Handler
	}
	type args struct {
		key interface{}
		cb  func(begin, end uint64) error
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Processor
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Processor{
				keyMap:     tt.fields.keyMap,
				handlerMap: tt.fields.handlerMap,
			}
			if got := p.AddUint64Range(tt.args.key, tt.args.cb); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Processor.AddUint64Range() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProcessor_AddUint64RangeIgnoreZero(t *testing.T) {
	type fields struct {
		keyMap     map[int32]string
		handlerMap map[string]Handler
	}
	type args struct {
		key interface{}
		cb  func(begin, end uint64) error
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Processor
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Processor{
				keyMap:     tt.fields.keyMap,
				handlerMap: tt.fields.handlerMap,
			}
			if got := p.AddUint64RangeIgnoreZero(tt.args.key, tt.args.cb); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Processor.AddUint64RangeIgnoreZero() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_toInt32(t *testing.T) {
	type args struct {
		key interface{}
	}
	tests := []struct {
		name string
		args args
		want int32
	}{
		{name: "int32", args: args{key: int32(123)}, want: 123},
		{name: "int64", args: args{key: int64(123)}, want: 123},
		{name: "uint32", args: args{key: uint32(123)}, want: 123},
		{name: "uint64", args: args{key: uint64(123)}, want: 123},
		{name: "string", args: args{key: "123"}, want: 123},
		{name: "bool", args: args{key: true}, want: 1},
		{name: "bool", args: args{key: false}, want: 0},
		{name: "float32", args: args{key: float32(123.123)}, want: 123},
		{name: "float64", args: args{key: float64(123.123)}, want: 123},
		{name: "nil", args: args{key: nil}, want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toInt32(tt.args.key); got != tt.want {
				t.Errorf("toInt32() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_toStr(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "int32", args: args{v: int32(123)}, want: "123"},
		{name: "int64", args: args{v: int64(123)}, want: "123"},
		{name: "uint32", args: args{v: uint32(123)}, want: "123"},
		{name: "uint64", args: args{v: uint64(123)}, want: "123"},
		{name: "string", args: args{v: "123"}, want: "123"},
		{name: "bool", args: args{v: true}, want: "true"},
		{name: "float32", args: args{v: float32(123.123)}, want: "123.123"},
		{name: "float64", args: args{v: float64(123.123)}, want: "123.123"},
		{name: "nil", args: args{v: nil}, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toStr(tt.args.v); got != tt.want {
				t.Errorf("toStr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Benchmark_toInt32(b *testing.B) {
	for i := 0; i < b.N; i++ {
		toInt32(int32RangeKey)
	}
}

func Benchmark_enumToInt32(b *testing.B) {
	for i := 0; i < b.N; i++ {
		enumToInt32(int32RangeKey)
	}
}

const (
	noneKey optionKey = iota
	boolKey
	int32Key
	int32ListKey
	int32RangeKey
	int64Key
	int64ListKey
	int64RangeKey
	uint32Key
	uint32ListKey
	uint32RangeKey
	uint64Key
	uint64ListKey
	uint64RangeKey
)

var _map = map[optionKey]string{
	noneKey:        "NoneV",
	boolKey:        "BoolV",
	int32Key:       "Int32V",
	int32ListKey:   "Int32List",
	int32RangeKey:  "Int32Range",
	int64Key:       "Int64V",
	int64ListKey:   "Int64List",
	int64RangeKey:  "Int64Range",
	uint32Key:      "Uint32V",
	uint32ListKey:  "Uint32List",
	uint32RangeKey: "Uint32Range",
	uint64Key:      "Uint64V",
	uint64ListKey:  "Uint64List",
	uint64RangeKey: "Uint64Range",
}

type optionKey int32

func (x optionKey) Enum() *optionKey {
	p := new(optionKey)
	*p = x
	return p
}

func (x optionKey) String() string {
	return _map[x]
}

func (x optionKey) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

type testData struct {
	NoneV       *string
	BoolV       *bool
	Int32V      int32
	Int32List   []int32
	Int32Range  [2]int32
	Int64V      int64
	Int64List   []int64
	Int64Range  [2]int64
	Uint32V     uint32
	Uint32List  []uint32
	Uint32Range [2]uint32
	Uint64V     uint64
	Uint64List  []uint64
	Uint64Range [2]uint64
}

func getDefaultTestData() *testData {
	none := "None"
	boolV := true
	return &testData{
		NoneV:       &none,
		BoolV:       &boolV,
		Int32V:      math.MaxInt32,
		Int32List:   []int32{math.MinInt32, 0, math.MaxInt32},
		Int32Range:  [2]int32{math.MinInt32, math.MaxInt32},
		Int64V:      math.MaxInt64,
		Int64List:   []int64{math.MinInt64, 0, math.MaxInt64},
		Int64Range:  [2]int64{math.MinInt64, math.MaxInt64},
		Uint32V:     math.MaxUint32,
		Uint32List:  []uint32{0, 99, math.MaxUint32},
		Uint32Range: [2]uint32{0, math.MaxUint32},
		Uint64V:     math.MaxUint64,
		Uint64List:  []uint64{0, 456, math.MaxUint64},
		Uint64Range: [2]uint64{0, math.MaxUint64},
	}
}

func getTestOptions() *Options {
	opts := NewOptions(
		noneKey, "None",
		boolKey, true,
		int32Key, math.MaxInt32,
		int32ListKey, []int32{math.MinInt32, 0, math.MaxInt32},
		int32RangeKey, [2]int32{math.MinInt32, math.MaxInt32},
		int64Key, fmt.Sprintf("%d", math.MaxInt64),
		int64ListKey, fmt.Sprintf("%d , %d,%d", math.MinInt64, 0, math.MaxInt64), // check space
		int64RangeKey, fmt.Sprintf("%d,%d,%d", math.MinInt64, math.MaxInt64, 0), // test mutli value
	).
		AddOpt(uint32Key, math.MaxUint32).
		AddOpt(uint32ListKey, []uint32{0, 99, math.MaxUint32}).
		AddOpt(uint32RangeKey, [2]uint32{0, math.MaxUint32}).
		AddOpt(uint64Key, fmt.Sprintf("%d", uint64(math.MaxUint64))).
		AddOpt(uint64ListKey, fmt.Sprintf("%d , %d,%d", 0, 456, uint64(math.MaxUint64))).
		AddOpt(uint64RangeKey, fmt.Sprintf("%d,%d,%d", 0, uint64(math.MaxUint64), 456))

	return opts
}

func getProcessor(data *testData) *Processor {
	p := NewProcessor().
		AddNone(noneKey, func() error {
			v := "None"
			data.NoneV = &v
			return nil
		}).
		AddBool(boolKey, func(val bool) error {
			data.BoolV = &val
			return nil
		}).
		AddInt32(int32Key, func(val int32) error {
			data.Int32V = val
			return nil
		}).
		AddInt32List(int32ListKey, func(val []int32) error {
			data.Int32List = val
			return nil
		}).
		AddInt32Range(int32RangeKey, func(begin, end int32) error {
			data.Int32Range = [2]int32{begin, end}
			return nil
		}).
		AddInt64(int64Key, func(val int64) error {
			data.Int64V = val
			return nil
		}).
		AddInt64List(int64ListKey, func(val []int64) error {
			data.Int64List = val
			return nil
		}).
		AddInt64Range(int64RangeKey, func(begin, end int64) error {
			data.Int64Range = [2]int64{begin, end}
			return nil
		}).
		AddUint32(uint32Key, func(val uint32) error {
			data.Uint32V = val
			return nil
		}).
		AddUint32List(uint32ListKey, func(val []uint32) error {
			data.Uint32List = val
			return nil
		}).
		AddUint32Range(uint32RangeKey, func(begin, end uint32) error {
			data.Uint32Range = [2]uint32{begin, end}
			return nil
		}).
		AddUint64(uint64Key, func(val uint64) error {
			data.Uint64V = val
			return nil
		}).
		AddUint64List(uint64ListKey, func(val []uint64) error {
			data.Uint64List = val
			return nil
		}).
		AddUint64Range(uint64RangeKey, func(begin, end uint64) error {
			data.Uint64Range = [2]uint64{begin, end}
			return nil
		})

	return p
}

func TestEnumProcessor(t *testing.T) {
	data := &testData{}
	p := getProcessor(data)
	err := p.ProcessOptions(getTestOptions())
	if err != nil {
		t.Error(err.Error())
	}
	defaultData := getDefaultTestData()
	if !reflect.DeepEqual(data, defaultData) {
		t.Log(data)
		t.Log(defaultData)
		t.Errorf("result is error")
	}
}

func TestStructProcessor(t *testing.T) {
	data := &testData{}
	defaultData := getDefaultTestData()
	p := getProcessor(data)
	p.SkipZero()
	err := p.ProcessStruct(defaultData)
	if err != nil {
		t.Error(err.Error())
	}
	if !reflect.DeepEqual(data, defaultData) {
		t.Log(data)
		t.Log(defaultData)
		t.Errorf("result is error")
	}
}
