package report

import (
	"os"
	"path/filepath"
	"testing"
)

// TestAddFileToReportEntryOutsideGinkgo runs outside a Ginkgo spec (no
// RunSpecs), the same path the gateway API conformance suite hits on failure.
// ginkgo.AddReportEntry used to panic there; the file should now be persisted
// directly to BaseDir instead.
func TestAddFileToReportEntryOutsideGinkgo(t *testing.T) {
	dir := t.TempDir()
	baseDir := filepath.Join(dir, "results")
	oldBase := BaseDir
	BaseDir = baseDir
	t.Cleanup(func() { BaseDir = oldBase })

	// Would panic before the fix.
	AddFileToReportEntry("kuma-1/debug info.txt", "boom")

	// name is sanitized to a flat filename via files.ToValidUnixFilename.
	got, err := os.ReadFile(filepath.Join(baseDir, "kuma-1-debug_info.txt"))
	if err != nil {
		t.Fatalf("expected report file written to BaseDir: %v", err)
	}
	if string(got) != "boom" {
		t.Fatalf("unexpected report content: got %q, want %q", got, "boom")
	}
}
