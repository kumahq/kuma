package report

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"

	"github.com/kumahq/kuma/pkg/util/files"
)

var (
	BaseDir       = "results"
	DumpOnSuccess = false
)

func stagingDir() string {
	return path.Join(BaseDir, "..", "kuma-test-staging")
}

// AddFileToReportEntry adds a file to the report. The file will be copied to the report directory.
// It's an alternative to ginkgo.AddReportEntry so that not all logs are kept in memory.
func AddFileToReportEntry(name string, content interface{}) {
	base := stagingDir()
	if err := os.MkdirAll(base, 0o755); err != nil {
		logf("[WARNING]: Error creating staging directory %s: %v", base, err)
	}
	tmp, err := os.CreateTemp(base, "report-*")
	if err != nil {
		logf("[WARNING]: could not create temporary report %v", err)
		return
	}
	defer tmp.Close()

	switch c := content.(type) {
	case string:
		_, err = tmp.WriteString(c)
	case []byte:
		_, err = tmp.Write(c)
	default:
		_, err = tmp.WriteString(fmt.Sprintf("%v", c))
	}
	if err != nil {
		logf("[WARNING]: could not write to temporary report %v", err)
		return
	}
	ginkgo.AddReportEntry(name, tmp.Name(), ginkgo.ReportEntryVisibilityNever)
}

// DumpReport dumps the report to the disk.
func DumpReport(report ginkgo.Report) {
	ginkgo.GinkgoHelper()
	basePath := BaseDir
	if files.FileExists(basePath) {
		tmpDir := path.Join(os.TempDir(), fmt.Sprintf("kuma-%04d", ginkgo.GinkgoRandomSeed()))
		logf("Report already exists in %q, moving to tmpDir: %q", basePath, tmpDir)
		if err := os.Rename(BaseDir, tmpDir); err != nil {
			logf("[WARNING]: failed to move %q to %q deleting it! %v", basePath, tmpDir, err)
			if err := os.RemoveAll(basePath); err != nil {
				logf("[WARNING]: failed to remove %q %v", basePath, err)
			}
		}
	}
	logf("saving report to %q DumpOnSuccess: %v", basePath, DumpOnSuccess)
	writeEntry := func(path string, data string) {
		err := os.MkdirAll(filepath.Dir(path), 0o755)
		if err != nil {
			logf("[WARNING]: failed to create directory %q: %v", path, err)
			return
		}
		// If the value is a file that actually exists let's simply move it in
		if files.FileExists(data) {
			err = os.Rename(data, path)
		} else {
			err = os.WriteFile(path, []byte(data), 0o600)
		}
		if err != nil {
			logf("[WARNING]: failed to write file %q: %v", path, err)
		}
	}
	for _, entry := range report.SpecReports {
		if entry.State == types.SpecStatePending || entry.State == types.SpecStateSkipped {
			continue
		}
		if DumpOnSuccess || entry.Failed() {
			entryPath := path.Join(basePath, files.ToValidUnixFilename(entry.FullText()))
			writeEntry(path.Join(entryPath, "combined.log"), entry.CombinedOutput())
			f := &strings.Builder{}
			_, _ = fmt.Fprintf(f, "Entry[%s]: %s\n", entry.LeafNodeType, entry.FullText())
			_, _ = fmt.Fprintf(f, "State: %s\n", entry.State)
			_, _ = fmt.Fprintf(f, "Duration: %s\n", entry.RunTime)
			_, _ = fmt.Fprintf(f, "Start: %s End: %s\n", entry.StartTime, entry.EndTime)
			_, _ = fmt.Fprintf(f, "SpecEvents:\n")
			for _, e := range entry.SpecEvents {
				_, _ = fmt.Fprintf(f, "%s\n", e.GomegaString())
			}
			writeEntry(path.Join(entryPath, "report.txt"), f.String())

			for _, e := range entry.ReportEntries {
				writeEntry(path.Join(entryPath, e.Name), e.StringRepresentation())
			}
		}
	}
	if err := os.RemoveAll(stagingDir()); err != nil {
		logf("[WARNING]: failed to remove staging directory %q: %v", stagingDir(), err)
	}
	logf("saved report to %q", basePath)
}

func logf(c string, args ...interface{}) {
	ginkgo.GinkgoWriter.Printf(c, args...)
	ginkgo.GinkgoWriter.Println("")
}
