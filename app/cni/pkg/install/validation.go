package install

import (
	"github.com/pkg/errors"
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
