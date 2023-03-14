package universal_logs

import (
	"path"
	"sync"
	"time"

	"github.com/kennygrant/sanitize"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
	"k8s.io/utils/strings"
)

var (
	timePrefix = time.Now().Local().Format("060102_150405")
	pathSet    = map[string]struct{}{}
	mutex      sync.RWMutex
)

func UniversalLogPath(basePath string) string {
	return path.Join(basePath, timePrefix)
}

func LogsPath(basePath string) string {
	sr := ginkgo.CurrentSpecReport()

	if len(sr.SpecEvents) == 0 {
		p := UniversalLogPath(basePath)
		addPath(p)
		return p
	}

	lastEvent := sr.SpecEvents[len(sr.SpecEvents)-1]

	var logsPath []string

	switch lastEvent.NodeType {
	case types.NodeTypeBeforeAll:
		idx, ok := find(sr.ContainerHierarchyTexts, lastEvent.Message)
		if !ok {
			panic("")
		}
		logsPath = append(sr.ContainerHierarchyTexts[:idx], lastEvent.Message)
	default:
		logsPath = append(sr.ContainerHierarchyTexts, sr.LeafNodeText)
	}

	sanitizedPath := []string{}
	for _, p := range logsPath {
		sanitizedPath = append(sanitizedPath, strings.ShortenString(sanitize.Name(p), 243))
	}

	p := path.Join(append([]string{basePath, timePrefix}, sanitizedPath...)...)
	addPath(p)
	return p
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
