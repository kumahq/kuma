package install

import (
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/util/files"
)

func lookForValidConfig(files []string, checkerFn func(string) error) (string, bool) {
	for _, file := range files {
		if err := checkerFn(file); err != nil {
			log.Info("error occurred testing config file", "file", file)
			continue
		}

		return file, true
	}

	return "", false
}

func isValidConfFile(file string) error {
	parsed, err := parseFileToHashMap(file)
	if err != nil {
		return errors.Wrap(err, "failed to parse configuration file")
	}

	if configType, ok := parsed["type"]; ok {
		log.V(1).Info("configuration validated", "file", file, "type", configType)
		return nil
	}

	return errors.Errorf(`configuration file "%s" missing "type" field`, file)
}

func isValidConflistFile(file string) error {
	parsed, err := parseFileToHashMap(file)
	if err != nil {
		return errors.Wrap(err, "failed to parse conflist file")
	}

	var missingFields []string

	configName, ok := parsed["name"]
	if !ok {
		missingFields = append(missingFields, "name")
	}

	plugins, ok := parsed["plugins"]
	if !ok {
		missingFields = append(missingFields, "plugins")
	}

	if len(missingFields) > 0 {
		return errors.Errorf("conflist file %s missing required fields: %+v", file, missingFields)
	}

	log.V(1).Info("conflist validated", "file", file, "name", configName, "plugins", plugins)

	return nil
}

func checkInstall(cniConfPath string, isPluginChained bool) error {
	if !files.FileExists(cniConfPath) {
		return errors.Errorf("cni config file does not exist at the specified path: %s", cniConfPath)
	}

	parsed, err := parseFileToHashMap(cniConfPath)
	if err != nil {
		return errors.Wrap(err, "failed to parse cni config file")
	}

	if isPluginChained {
		if err := isValidConflistFile(cniConfPath); err != nil {
			return errors.Wrap(err, "chained plugin requires a valid conflist file format")
		}

		plugins, err := getPluginsArray(parsed)
		if err != nil {
			return errors.Wrap(err, "failed to retrieve plugins array from cni config")
		}

		if index, err := findKumaCniConfigIndex(plugins); err != nil {
			return errors.Wrap(err, "failed to find kuma-cni plugin in chained config file")
		} else if index < 0 {
			return errors.New("kuma-cni plugin is missing in the chained config file")
		}

		return nil
	}

	if err := isValidConfFile(cniConfPath); err != nil {
		return errors.Wrap(err, "standalone plugin requires a valid conf file format")
	}

	if pluginType, ok := parsed["type"]; !ok {
		return errors.New("cni config is missing the required 'type' field")
	} else if pluginType != "kuma-cni" {
		return errors.New("cni config 'type' field is not set to 'kuma-cni'")
	}

	return nil
}
