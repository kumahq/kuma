package universal_logs

import (
	"os"
	"path"
	"sync"
	"time"

	"github.com/kennygrant/sanitize"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
	k8s_strings "k8s.io/utils/strings"
)

var (
	timePrefix = time.Now().Local().Format("060102_150405")
	pathSet    = map[string]struct{}{}
	mutex      sync.RWMutex
)

// CreateLogsPath constructs a path for logs based on the Ginkgo node hierarchy.
// If the CreateLogsPath function is called before the suite starts (like SyncBeforeSuite),
// then the result is "/{basePath}/{timePrefix}".
// If the function is called inside the suite, then the result is
// a concatenation of the basePath, timePrefix, and the Ginkgo node hierarchy.
//
// For example, if the spec has the following hierarchy:
//
//	Describe("spec 1", func() {
//	    Context("ctx 1", func() {
//	        BeforeAll(func() {
//	            lp := CreateLogsPath("/tmp") // lp == "/tmp/060102_150405/spec-1/ctx-1"
//	        })
//	        Context("ctx 2", func() {
//	            It("it 1", func() {
//	                lp := CreateLogsPath("/tmp") // lp == "/tmp/060102_150405/spec-1/ctx-1/ctx-2/it-1"
//	            })
//	        })
//	    })
//	})
//
// Additionally the function adds logs path to the Ginkgo report. Cleanup function is using this information
// to remove logs for successfully passed tests.
func CreateLogsPath(basePath string) string {
	sr := ginkgo.CurrentSpecReport()

	if len(sr.SpecEvents) == 0 {
		p := withTimePrefix(basePath)
		addPath(p)
		return p
	}

	lastEvent := sr.SpecEvents[len(sr.SpecEvents)-1]

	var logsPath []string

	switch lastEvent.NodeType {
	case types.NodeTypeBeforeAll:
		// BeforeAll is a special case from the SpecReport perspective it's similar to BeforeEach
		// but executed only once before the first It(). That's why we want to take into account
		// only nodes that surround the BeforeAll node.
		idx, ok := find(sr.ContainerHierarchyTexts, lastEvent.Message)
		if !ok {
			panic("ContainerHierarchyTexts doesn't contain BeforeAll node")
		}
		logsPath = append(sr.ContainerHierarchyTexts[:idx], lastEvent.Message)
	default:
		logsPath = append(sr.ContainerHierarchyTexts, sr.LeafNodeText)
	}

	sanitizedPath := []string{}
	for _, p := range logsPath {
		sanitizedPath = append(sanitizedPath, k8s_strings.ShortenString(sanitize.Name(p), 243))
	}

	// add only a root level ginkgo.Describe() directory for a cleanup,
	// i.e "/tmp/060102_150405/mesh-traffic-permissions"
	addPath(path.Join(append([]string{basePath, timePrefix}, sanitizedPath[:1]...)...))

	return path.Join(append([]string{basePath, timePrefix}, sanitizedPath...)...)
}

func CleanupIfSuccess(basePath string, report ginkgo.Report) {
	suiteFailed := false

	for _, sr := range report.SpecReports {
		if sr.Failed() {
			suiteFailed = true
		}

		if !sr.Failed() && len(sr.ContainerHierarchyTexts) != 0 {
			for _, re := range sr.ReportEntries {
				_ = os.RemoveAll(re.Name)
			}
		}
	}

	if !suiteFailed {
		_ = os.RemoveAll(withTimePrefix(basePath))
	}
}

func find(slice []string, s string) (int, bool) {
	for idx, ss := range slice {
		if s == ss {
			return idx, true
		}
	}
	return 0, false
}

func addPath(path string) {
	mutex.Lock()
	if _, ok := pathSet[path]; !ok {
		pathSet[path] = struct{}{}
		ginkgo.AddReportEntry(path)
	}
	mutex.Unlock()
}

func withTimePrefix(basePath string) string {
	return path.Join(basePath, timePrefix)
}
