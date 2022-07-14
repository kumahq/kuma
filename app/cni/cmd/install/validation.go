package main

import (
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/util/files"
)

func isValidConfFile(file string) bool {
	parsed, err := parseFileToHashMap(file)
	if err != nil {
		log.Error(err, "could not unmarshal config file")
		return false
	}

	configType, ok := parsed["type"]
	if ok {
		log.V(1).Info("config valid", "file", file, "type", configType)
		return true
	}
	log.V(1).Info("config not valid", "file", file)
	return false
}

func isValidConflistFile(file string) bool {
	parsed, err := parseFileToHashMap(file)
	if err != nil {
		log.Error(err, "could not unmarshal config file")
		return false
	}

	configName, hasName := parsed["name"]
	plugins, hasPlugins := parsed["plugins"]
	if hasName && hasPlugins {
		log.V(1).Info("config valid", "file", file, "name", configName, "plugins", plugins)
		return true
	}

	return false
}

func checkInstall(cniConfPath string, isPluginChained bool) error {
	if !files.FileExists(cniConfPath) {
		return errors.New("cni config file does not exist")
	}

	parsed, err := parseFileToHashMap(cniConfPath)
	if err != nil {
		return err
	}

	if isPluginChained {
		if !isValidConflistFile(cniConfPath) {
			return errors.New("chained plugin requires a valid conflist file")
		}
		plugins, err := getPluginsArray(parsed)
		if err != nil {
			return err
		}
		index, err := findKumaCniConfigIndex(plugins)
		if err != nil {
			return err
		}
		if index >= 0 {
			return nil
		} else {
			return errors.New("chained plugin config file does not contain kuma-cni plugin")
		}
	} else {
		if !isValidConfFile(cniConfPath) {
			return errors.New("chained plugin requires a valid conflist file")
		}
		pluginType, ok := parsed["type"]
		if !ok {
			return errors.New("cni config was modified and does not have a type")
		}
		if pluginType == "kuma-cni" {
			return nil
		} else {
			return errors.New("config file does not contain kuma-cni configuration")
		}
	}
}
