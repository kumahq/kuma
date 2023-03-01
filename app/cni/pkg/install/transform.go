package install

import (
	"encoding/json"

	"github.com/pkg/errors"
)

func transformJsonConfig(kumaCniConfig string, hostCniConfig []byte) ([]byte, error) {
	parsed, err := parseBytesToHashMap(hostCniConfig)
	if err != nil {
		return nil, err
	}

	kumaCniConfigParsed, err := parseBytesToHashMap([]byte(kumaCniConfig))
	if err != nil {
		return nil, err
	}

	_, hasType := parsed["type"]
	if hasType {
		newConfig := map[string]interface{}{
			"name":       "k8s-pod-network",
			"cniVersion": "0.3.0",
			"plugins":    []interface{}{parsed, kumaCniConfigParsed},
		}
		return json.MarshalIndent(newConfig, "", "  ")
	} else {
		pluginsArray, err := getPluginsArray(parsed)
		if err != nil {
			return nil, err
		}

		pluginsArray, err = removeKumaCniConfig(pluginsArray)
		if err != nil {
			return nil, err
		}

		pluginsArray = append(pluginsArray, kumaCniConfigParsed)

		parsed["plugins"] = pluginsArray
	}

	marshaled, err := json.MarshalIndent(parsed, "", "  ")
	if err != nil {
		return nil, err
	}
	return marshaled, nil
}

func getPluginsArray(parsed map[string]interface{}) ([]interface{}, error) {
	plugins, ok := parsed["plugins"]
	if !ok {
		return nil, errors.New("config does not have 'plugins' field")
	}

	pluginsArray, ok := plugins.([]interface{})
	if !ok {
		return nil, errors.New("config's 'plugins' field is not an array")
	}

	return pluginsArray, nil
}

func removeKumaCniConfig(pluginsArray []interface{}) ([]interface{}, error) {
	kumaCniConfigIndex, err := findKumaCniConfigIndex(pluginsArray)
	if err != nil {
		return nil, err
	}
	if kumaCniConfigIndex >= 0 {
		pluginsArray = append(pluginsArray[:kumaCniConfigIndex], pluginsArray[kumaCniConfigIndex+1:]...)
	}
	return pluginsArray, nil
}

func findKumaCniConfigIndex(pluginsArray []interface{}) (int, error) {
	kumaCniConfigIndex := -1
	for i, p := range pluginsArray {
		plugin, ok := p.(map[string]interface{})
		if !ok {
			return -1, errors.New("plugin is not an object")
		}

		pluginType, ok := plugin["type"]
		if !ok {
			continue
		}

		if pluginType == "kuma-cni" {
			kumaCniConfigIndex = i
		}
	}

	return kumaCniConfigIndex, nil
}

func revertConfigContents(configBytes []byte) ([]byte, error) {
	parsed, err := parseBytesToHashMap(configBytes)
	if err != nil {
		return nil, err
	}
	pluginsArray, err := getPluginsArray(parsed)
	if err != nil {
		return nil, err
	}
	pluginsArray, err = removeKumaCniConfig(pluginsArray)
	if err != nil {
		return nil, err
	}

	parsed["plugins"] = pluginsArray

	marshaled, err := json.MarshalIndent(parsed, "", "  ")
	if err != nil {
		return nil, err
	}
	return marshaled, nil
}
