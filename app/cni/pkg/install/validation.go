package install

import (
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/util/files"
)

func lookForValidConfig(files []string, checkerFn func(string) error) (string, bool) {
	for _, file := range files {
		err := checkerFn(file)
		if err != nil {
			log.Info("error occurred testing config file", "file", file)
		} else {
			return file, true
		}
	}
	return "", false
}

func isValidConfFile(file string) error {
	parsed, err := parseFileToHashMap(file)
	if err != nil {
		return errors.Wrap(err, "could not unmarshal conf file")
	}

	configType, ok := parsed["type"]
	if ok {
		log.V(1).Info("config valid", "file", file, "type", configType)
		return nil
	}
	return errors.Errorf(`config file %v not valid - does not contain "type" field`, file)
}

func isValidConflistFile(file string) error {
	parsed, err := parseFileToHashMap(file)
	if err != nil {
		return errors.Wrap(err, "could not unmarshal conflist file")
	}

	configName, hasName := parsed["name"]
	plugins, hasPlugins := parsed["plugins"]
	if hasName && hasPlugins {
		log.V(1).Info("config valid", "file", file, "name", configName, "plugins", plugins)
		return nil
	}

	return errors.Errorf(`config file %v not valid - does not contain "name" and "plugin" fields`, file)
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
		err := isValidConflistFile(cniConfPath)
		if err != nil {
			return errors.Wrap(err, "chained plugin requires a valid conflist file")
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
		err := isValidConfFile(cniConfPath)
		if err != nil {
			return errors.Wrap(err, "standalone plugin requires a valid conf file")
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
