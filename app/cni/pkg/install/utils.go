package install

import (
	"encoding/json"
	"io/ioutil"

	"github.com/go-logr/logr"
)

func parseFileToHashMap(file string) (map[string]interface{}, error) {
	contents, err := ioutil.ReadFile(file)
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

func SetLogLevel(logger *logr.Logger, level string) {
	switch level {
	case "off":
		*logger = logr.Discard()
	case "debug":
		*logger = logger.V(1)
	default:
		*logger = logger.V(0)
	}
}
