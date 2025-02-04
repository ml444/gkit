package dbx

import (
	"testing"
)

func Test_isNonEmptySlice(t *testing.T) {
	type args struct {
		slice interface{}
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"1", args{slice: []string{"a", "b"}}, true},
		{"2", args{slice: []uint64{}}, false},
		{"3", args{slice: nil}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isNonEmptySlice(tt.args.slice); got != tt.want {
				t.Errorf("isNonEmptySlice() = %v, want %v", got, tt.want)
			}
		})
	}
}
func Benchmark_isNonEmptySlice(b *testing.B) {
	for i := 0; i < b.N; i++ {
		isNonEmptySlice([]string{"a", "b"})
	}
}
