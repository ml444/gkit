package main

import (
	"reflect"
	"testing"

	"github.com/ml444/gkit/cmd/protoc-gen-go-http/pluck"
)

func Test_pluckFields(t *testing.T) {
	type args struct {
		v interface{}
	}
	ct := "application/json"
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "Test pluckFields empty",
			args: args{
				v: &pluck.RequestHeaders{},
			},
			want: map[string]string{},
		},
		{
			name: "Test pluckFields",
			args: args{
				v: &pluck.RequestHeaders{ContentType: &ct},
			},
			want: map[string]string{"Content-Type": "application/json"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pluckFields(tt.args.v); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("pluckFields() = %v, want %v", got, tt.want)
			}
		})
	}
}
