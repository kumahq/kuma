package universal_logs

import (
	"os"
	"path"
	"time"

	"github.com/kennygrant/sanitize"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
	k8s_strings "k8s.io/utils/strings"

	"github.com/kumahq/kuma/pkg/core"
)

var timePrefix = time.Now().Local().Format("060102_150405")

func CurrentLogsPath(basePath string) string {
	if lp := GetLogsPath(ginkgo.CurrentSpecReport(), basePath); lp.Spec != "" {
		return lp.Spec
	} else if lp.Describe != "" {
		return lp.Describe
	} else {
		return lp.Root
	}
}

type LogsPaths struct {
	// Root contains path of the root of logging directory i.e. /tmp/060102_150405/
	Root string
	// Describe contains path of the 'Describe' logging directory i.e. /tmp/060102_150405/meshtrafficpermission
	Describe string
	// Spec contains path of the spec like 'BeforeAll', 'It' etc., it can be quite long and complicated
	// i.e //tmp/060102_150405/meshtrafficpermission/http/with-mtls/should-work/. You shouldn't rely on 'spec' when performing cleanup.
	Spec string
}

// GetLogsPath returns a struct with 3 paths – rootPath, describePath and specPath.
// All e2e tests in Kuma have the following structure:
//
//	<env>_suite:
//		Describe("<feature1>", ...)
//			It("should do smth1...")
//			It("should do smth2...")
//		Describe("<feature2>", ...)
//
// Some components (like kuma-cp) are shared across env, other apps (like test-server) could be deployed per Describe or per It.
// We'd like to reflect hierarchical structure when collecting logs from running containers, that's why for writing logs we're using specPath.
func GetLogsPath(sr ginkgo.SpecReport, basePath string) LogsPaths {
	result := LogsPaths{
		Root: withTimePrefix(basePath),
	}

	if len(sr.SpecEvents) == 0 || len(sr.ContainerHierarchyTexts) == 0 {
		return result
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
	result.Describe = path.Join(append([]string{result.Root}, sanitizedPath[:1]...)...)
	result.Spec = path.Join(append([]string{result.Root}, sanitizedPath...)...)

	return result
}

// CleanupIfSuccess removes logs for successfully passed Describe specs by using describePath.
// If all tests passed then we'll remove rootPath dir.
func CleanupIfSuccess(basePath string, report ginkgo.Report) {
	specFailedByLogsPath := map[string]bool{}

	for _, sr := range report.SpecReports {
		lp := GetLogsPath(sr, basePath)
		if lp.Describe == "" {
			continue
		}
		specFailedByLogsPath[lp.Describe] = specFailedByLogsPath[lp.Describe] || sr.Failed()
	}

	suiteFailed := false
	for logsPath, failed := range specFailedByLogsPath {
		if failed {
			suiteFailed = true
		} else {
			core.Log.Info("cleanup after 'Describe'", "describePath", logsPath)
			_ = os.RemoveAll(logsPath)
		}
	}

	if !suiteFailed {
		core.Log.Info("suite didn't fail, so cleanup everything")
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

func withTimePrefix(basePath string) string {
	return path.Join(basePath, timePrefix)
}
