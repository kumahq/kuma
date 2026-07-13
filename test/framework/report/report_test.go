package report

import (
	"os"
	"path/filepath"
	"testing"
)

// TestAddFileToReportEntryOutsideGinkgo reproduces the crash that took down the
// Gateway API conformance suite: it runs as a plain `testing` test and, on
// failure, t.Cleanup -> DebugKube -> DumpState calls AddFileToReportEntry. There
// is no running Ginkgo spec, so ginkgo.AddReportEntry used to panic
// ("...not during the Run phase..."), masking the real failure and losing the
// debug artifacts. The file should now be persisted directly to BaseDir instead.
//
// This test itself runs outside a Ginkgo spec (no RunSpecs), so it exercises the
// exact code path that panicked.
func TestAddFileToReportEntryOutsideGinkgo(t *testing.T) {
	dir := t.TempDir()
	oldBase := BaseDir
	BaseDir = dir
	t.Cleanup(func() { BaseDir = oldBase })

	// Would panic before the fix.
	AddFileToReportEntry("kuma-1/debug info.txt", "boom")

	// name is sanitized to a flat filename via files.ToValidUnixFilename.
	got, err := os.ReadFile(filepath.Join(dir, "kuma-1-debug_info.txt"))
	if err != nil {
		t.Fatalf("expected report file written to BaseDir: %v", err)
	}
	if string(got) != "boom" {
		t.Fatalf("unexpected report content: got %q, want %q", got, "boom")
	}
}
