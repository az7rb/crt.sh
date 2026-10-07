package main

import "testing"

func TestVersionLess(t *testing.T) {
	tests := []struct {
		a, b string
		want bool
	}{
		{"3.0.3", "3.0.4", true},
		{"3.0.4", "3.0.3", false},
		{"3.0.3", "3.0.3", false},
		{"3.0.9", "3.0.10", true},
		{"3.0.10", "3.0.9", false},
		{"2.9.9", "3.0.0", true},
		{"3.1", "3.1.1", true},
		{"v3.0.3", "3.0.4", true},
		{"3.0.3", "v3.0.4", true},
		{"v3.0.4", "3.0.3", false},
		{" 3.0.3\n", "3.0.3", false},
		{"3.0.3", " v3.0.4 ", true},
		{"3.0.4-rc1", "3.0.3", false},
		{"3.0.3", "3.0.4-rc1", true},
	}
	for _, tt := range tests {
		if got := versionLess(tt.a, tt.b); got != tt.want {
			t.Errorf("versionLess(%q, %q) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
}
