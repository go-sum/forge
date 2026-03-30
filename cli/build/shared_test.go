package main

import "testing"

func TestResolveVersion(t *testing.T) {
	tests := []struct {
		name           string
		pkgName        string
		defaultVersion string
		envKey         string
		envValue       string
		want           string
	}{
		{
			name:           "returns default when env var is unset",
			pkgName:        "htmx",
			defaultVersion: "1.9.10",
			want:           "1.9.10",
		},
		{
			name:           "env var overrides default",
			pkgName:        "htmx",
			defaultVersion: "1.9.10",
			envKey:         "HTMX_VERSION",
			envValue:       "2.0.0",
			want:           "2.0.0",
		},
		{
			name:           "env var with whitespace is trimmed",
			pkgName:        "alpine",
			defaultVersion: "3.0.0",
			envKey:         "ALPINE_VERSION",
			envValue:       "  3.1.0  ",
			want:           "3.1.0",
		},
		{
			name:           "empty env var falls back to default",
			pkgName:        "htmx",
			defaultVersion: "1.9.10",
			envKey:         "HTMX_VERSION",
			envValue:       "",
			want:           "1.9.10",
		},
		{
			name:           "whitespace-only env var falls back to default",
			pkgName:        "htmx",
			defaultVersion: "1.9.10",
			envKey:         "HTMX_VERSION",
			envValue:       "   ",
			want:           "1.9.10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envKey != "" {
				t.Setenv(tt.envKey, tt.envValue)
			}
			got := resolveVersion(tt.pkgName, tt.defaultVersion)
			if got != tt.want {
				t.Fatalf("resolveVersion(%q, %q) = %q, want %q", tt.pkgName, tt.defaultVersion, got, tt.want)
			}
		})
	}
}
