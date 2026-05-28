package httpx

import "testing"

func TestJoinPath(t *testing.T) {
	tests := []struct {
		prefix, path, want string
	}{
		{"", "/users", "/users"},
		{"/api/v1", "/users", "/api/v1/users"},
		{"/api/v1/", "/users", "/api/v1/users"},
		{"/api/v1", "users", "/api/v1/users"},
		{"/api/v1", "", "/api/v1"},
	}
	for _, tt := range tests {
		if got := JoinPath(tt.prefix, tt.path); got != tt.want {
			t.Errorf("JoinPath(%q, %q) = %q, want %q", tt.prefix, tt.path, got, tt.want)
		}
	}
}
