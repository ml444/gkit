package httpx

import (
	"net/http"
	"reflect"
	"testing"
)

func Test_getAcceptLanguage(t *testing.T) {
	type args struct {
		r *http.Request
	}
	tests := []struct {
		name      string
		args      args
		wantLangs []string
	}{
		{
			name: "test-ok",
			args: args{r: &http.Request{Header: map[string][]string{
				"Accept-Language": {
					"zh",
					"en-US",
					"rs;q=0.5",
					" rs;q=0.5",
					"es-DC ;q=0.5",
				},
			}}},
			wantLangs: []string{"zh", "en-US", "rs", "rs", "es-DC"},
		},
		{
			name: "test-*",
			args: args{r: &http.Request{Header: map[string][]string{
				"Accept-Language": {
					";q=0.5",
					"*;q=0.5",
					" *;q=0.5",
				},
			}}},
			wantLangs: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotLangs := getAcceptLanguage(tt.args.r); !reflect.DeepEqual(gotLangs, tt.wantLangs) {
				t.Errorf("getAcceptLanguage() = %v, want %v", gotLangs, tt.wantLangs)
			}
		})
	}
}
