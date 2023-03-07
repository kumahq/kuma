package universal_logs

import (
	"path"
	"sync"
	"time"

	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/kennygrant/sanitize"
	"k8s.io/utils/strings"
)

var (
	timePrefix = time.Now().Local().Format("060102_150405")
	paths      = map[string]string{}
	mutex      sync.RWMutex
)

func GenAndSavePath(logsPath, specName string) {
	mutex.Lock()
	defer mutex.Unlock()

	// Let's be sure we won't exceed max filename length
	fileName := strings.ShortenString(sanitize.Name(specName), 243)

	paths[specName] = path.Join(
		logsPath,
		timePrefix,
		fileName+"-"+random.UniqueId(),
	)
}

func GetPath(logsPath, specName string) string {
	mutex.RLock()
	defer mutex.RUnlock()

	id, ok := paths[specName]
	if !ok {
		return path.Join(logsPath, timePrefix)
	}

	return id
}
