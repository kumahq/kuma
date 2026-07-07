package framework

import (
	"reflect"
	"testing"
)

func TestFilteredReproEnvRedactsUnknownAllowedKeys(t *testing.T) {
	got := filteredReproEnvFrom([]string{
		"KUMA_K8S_TYPE=kind",
		"KUMA_DB_URL=postgres://user:pass@db",
		"GITHUB_SHA=abc123",
		"GITHUB_TOKEN=secret",
		"IGNORED_KEY=ignored",
	})

	want := map[string]string{
		"GITHUB_SHA":    "abc123",
		"GITHUB_TOKEN":  redactedReproValue,
		"KUMA_DB_URL":   redactedReproValue,
		"KUMA_K8S_TYPE": "kind",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected filtered repro env %v, got %v", want, got)
	}
}

func TestFilteredReproEnvFromHandlesExplicitAndPrefixedKeys(t *testing.T) {
	got := filteredReproEnvFrom([]string{
		"ARCH=arm64",
		"KUBECONFIG=/tmp/kubeconfig",
		"GITHUB_WORKFLOW=ci",
		"KUMA_LICENSE=license",
		"NOT_INCLUDED=value",
	})

	want := map[string]string{
		"ARCH":            "arm64",
		"GITHUB_WORKFLOW": "ci",
		"KUBECONFIG":      "/tmp/kubeconfig",
		"KUMA_LICENSE":    redactedReproValue,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected filtered repro env %v, got %v", want, got)
	}
}
