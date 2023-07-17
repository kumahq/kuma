package install

import (
	"encoding/json"
	"os"

	"github.com/go-logr/logr"

	"github.com/kumahq/kuma/pkg/core"
	kuma_log "github.com/kumahq/kuma/pkg/log"
)

func parseFileToHashMap(file string) (map[string]interface{}, error) {
	contents, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	return parseBytesToHashMap(contents)
}

func parseBytesToHashMap(bytes []byte) (map[string]interface{}, error) {
	var parsed map[string]interface{}
	err := json.Unmarshal(bytes, &parsed)
	if err != nil {
		return nil, err
	}
	return parsed, nil
}

func CreateNewLogger(name string, logLevel kuma_log.LogLevel, logFormat kuma_log.LogFormat) logr.Logger {
	// kubelet expects a specific JSON on stdout, so we're using stderr in CNI
	return core.NewLoggerTo(os.Stderr, logLevel, logFormat).WithName(name)
}

func SetLogUtils(logger *logr.Logger, level string, name string, format string) error {
	logFormat, errFormat := kuma_log.ParseLogFormat(format)
	logLevel, errLevel := kuma_log.ParseLogLevel(level)
	if errFormat != nil {
		return errFormat
	}
	
	if errLevel != nil {
		return errLevel
	}

	*logger = CreateNewLogger(name, logLevel, logFormat)
	return nil
}
