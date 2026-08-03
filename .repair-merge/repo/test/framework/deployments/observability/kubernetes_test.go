package observability

import (
	"strings"
	"testing"
)

func TestRenderJaegerManifestSubstitutesNamespace(t *testing.T) {
	manifest, err := renderJaegerManifest("observability-test")
	if err != nil {
		t.Fatalf("render Jaeger manifest: %v", err)
	}

	if !strings.Contains(manifest, "namespace: observability-test") {
		t.Fatalf("expected rendered manifest to contain namespace, got:\n%s", manifest)
	}
	if !strings.Contains(manifest, "kind: Namespace") || !strings.Contains(manifest, "name: observability-test") {
		t.Fatalf("expected rendered manifest to create namespace, got:\n%s", manifest)
	}
	if strings.Contains(manifest, "{{ .Namespace }}") {
		t.Fatalf("expected rendered manifest to substitute namespace template, got:\n%s", manifest)
	}
}
