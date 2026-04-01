package main

import (
	"errors"
	"os/exec"
	"sort"
	"testing"
)

func TestServicesToStop(t *testing.T) {
	tests := []struct {
		name     string
		baseline map[string]struct{}
		current  map[string]struct{}
		want     []string
	}{
		{
			name:     "new services detected",
			baseline: map[string]struct{}{"a": {}, "b": {}},
			current:  map[string]struct{}{"a": {}, "b": {}, "c": {}},
			want:     []string{"c"},
		},
		{
			name:     "no new services",
			baseline: map[string]struct{}{"a": {}, "b": {}},
			current:  map[string]struct{}{"a": {}, "b": {}},
			want:     nil,
		},
		{
			name:     "empty baseline stops all current",
			baseline: map[string]struct{}{},
			current:  map[string]struct{}{"a": {}, "b": {}},
			want:     []string{"a", "b"},
		},
		{
			name:     "empty current stops nothing",
			baseline: map[string]struct{}{"a": {}},
			current:  map[string]struct{}{},
			want:     nil,
		},
		{
			name:     "both empty",
			baseline: map[string]struct{}{},
			current:  map[string]struct{}{},
			want:     nil,
		},
		{
			name:     "baseline service removed externally",
			baseline: map[string]struct{}{"a": {}, "b": {}},
			current:  map[string]struct{}{"a": {}, "c": {}},
			want:     []string{"c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := servicesToStop(tt.baseline, tt.current)
			sort.Strings(got)
			sort.Strings(tt.want)

			if len(got) == 0 && len(tt.want) == 0 {
				return
			}
			if len(got) != len(tt.want) {
				t.Fatalf("servicesToStop() = %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("servicesToStop() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestExitCodeFromError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{
			name: "nil returns 0",
			err:  nil,
			want: 0,
		},
		{
			name: "generic error returns 1",
			err:  errors.New("something went wrong"),
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := exitCodeFromError(tt.err)
			if got != tt.want {
				t.Fatalf("exitCodeFromError() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestExitCodeFromError_ViaExec(t *testing.T) {
	// Use a real process to get a genuine ExitError with exit code 2.
	cmd := exec.Command("sh", "-c", "exit 2")
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected error from exit 2 command")
	}
	got := exitCodeFromError(err)
	if got != 2 {
		t.Fatalf("exitCodeFromError() = %d, want 2", got)
	}
}

func TestMapKeys(t *testing.T) {
	m := map[string]struct{}{"x": {}, "y": {}, "z": {}}
	keys := mapKeys(m)
	sort.Strings(keys)
	want := []string{"x", "y", "z"}
	if len(keys) != len(want) {
		t.Fatalf("mapKeys() = %v, want %v", keys, want)
	}
	for i := range keys {
		if keys[i] != want[i] {
			t.Fatalf("mapKeys() = %v, want %v", keys, want)
		}
	}
}
