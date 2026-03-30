package main

import (
	"strings"
	"testing"

	"github.com/go-sum/forge/internal/health"
)

func TestRenderHealthHumanQuietModeShowsOnlyFailures(t *testing.T) {
	report := health.Report{
		Checks: []health.Result{
			{Name: "config_load", Status: health.StatusPass, Message: "loaded"},
			{Name: "database_connectivity", Status: health.StatusFail, Message: "dial tcp: connection refused"},
			{Name: "required_relations", Status: health.StatusSkip, Message: "skipped because database connectivity failed"},
		},
	}

	got := renderHealthHuman(report, false)

	if strings.Contains(got, "config_load") {
		t.Fatalf("quiet output should omit passing checks: %q", got)
	}
	if strings.Contains(got, "required_relations") {
		t.Fatalf("quiet output should omit skipped checks: %q", got)
	}
	if !strings.Contains(got, "database_connectivity: dial tcp: connection refused") {
		t.Fatalf("quiet output missing failure: %q", got)
	}
}

func TestRenderHealthHumanVerboseModeShowsEveryCheck(t *testing.T) {
	report := health.Report{
		Checks: []health.Result{
			{Name: "config_load", Status: health.StatusPass, Message: "loaded"},
			{Name: "database_connectivity", Status: health.StatusFail, Message: "dial tcp: connection refused"},
			{Name: "required_relations", Status: health.StatusSkip, Message: "skipped because database connectivity failed"},
		},
	}

	got := renderHealthHuman(report, true)

	for _, want := range []string{
		"PASS config_load: loaded",
		"FAIL database_connectivity: dial tcp: connection refused",
		"SKIP required_relations: skipped because database connectivity failed",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("verbose output missing %q in %q", want, got)
		}
	}
}
