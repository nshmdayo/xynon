package proxy

import "testing"

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"/api/v1/users/123", "/api/v1/users/{id}"},
		{"/api/v1/users/123/profile", "/api/v1/users/{id}/profile"},
		{"/api/v1/items/abc-def", "/api/v1/items/abc-def"},
		{"/api/v1/items/550e8400-e29b-41d4-a716-446655440000", "/api/v1/items/{id}"},
		{"/", "/"},
		{"/123", "/{id}"},
	}

	for _, tt := range tests {
		got := normalizePath(tt.input)
		if got != tt.want {
			t.Errorf("normalizePath(%q) = %q; want %q", tt.input, got, tt.want)
		}
	}
}
