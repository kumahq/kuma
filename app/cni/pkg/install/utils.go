package install

import (
	"encoding/json"
	"os"
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
