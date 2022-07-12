package main

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
		plugins, ok := parsed["plugins"]
		if !ok {
			return nil, errors.New("config does not have 'plugins' field")
		}

		pluginsArray, ok := plugins.([]interface{})
		if !ok {
			return nil, errors.New("config's 'plugins' field is not an array")
		}

		kumaCniConfigIndex := -1
		for i, p := range pluginsArray {
			plugin, ok := p.(map[string]interface{})
			if !ok {
				return nil, errors.New("plugin is not an object")
			}

			pluginType, ok := plugin["type"]
			if !ok {
				continue
			}

			if pluginType == "kuma-cni" {
				kumaCniConfigIndex = i
			}
		}

		if kumaCniConfigIndex > 0 {
			pluginsArray = append(pluginsArray[:kumaCniConfigIndex], pluginsArray[kumaCniConfigIndex+1:]...)
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
