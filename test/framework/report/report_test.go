package report

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
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

// TestDumpReportKeepsOtherSuites covers the case of a job running several suites
// against the same BaseDir: the suites that run after the failing one used to
// move its whole directory away, so the failure was uploaded with no debug data.
func TestDumpReportKeepsOtherSuites(t *testing.T) {
	baseDir := filepath.Join(t.TempDir(), "results")
	oldBase, oldDumpOnSuccess := BaseDir, DumpOnSuccess
	BaseDir, DumpOnSuccess = baseDir, false
	t.Cleanup(func() { BaseDir, DumpOnSuccess = oldBase, oldDumpOnSuccess })

	failed := func(suite string) ginkgo.Report {
		return ginkgo.Report{
			SuiteDescription: suite,
			SpecReports: []types.SpecReport{{
				LeafNodeType:               types.NodeTypeIt,
				LeafNodeText:               "spec",
				State:                      types.SpecStateFailed,
				CapturedGinkgoWriterOutput: suite + " output",
			}},
		}
	}

	DumpReport(failed("E2E Helm Suite"))
	// A later suite in the same job must not touch the helm suite's dump.
	DumpReport(failed("E2E CNI Suite"))

	for _, suite := range []string{"E2E_Helm_Suite", "E2E_CNI_Suite"} {
		got, err := os.ReadFile(filepath.Join(baseDir, suite, "spec", "combined.log"))
		if err != nil {
			t.Fatalf("expected report for suite %s: %v", suite, err)
		}
		if len(got) == 0 {
			t.Fatalf("expected non-empty report for suite %s", suite)
		}
	}
}
