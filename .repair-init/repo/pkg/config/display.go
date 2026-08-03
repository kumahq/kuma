package config

import (
	"encoding/json"
	"os"
	"reflect"

	"sigs.k8s.io/yaml"
)

func ConfigForDisplay(cfg Config) (string, error) {
	// copy config so we don't override values, because nested structs in config are pointers
	newCfg, err := copyConfig(cfg)
	if err != nil {
		return "", err
	}
	newCfg.Sanitize()
	b, err := json.Marshal(newCfg)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func copyConfig(cfg Config) (Config, error) {
	cfgBytes, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}

	newCfg := reflect.New(reflect.TypeOf(cfg).Elem()).Interface().(Config)
	if err := json.Unmarshal(cfgBytes, newCfg); err != nil {
		return nil, err
	}
	return newCfg, nil
}

func DumpToFile(filename string, cfg Config) error {
	if err := cfg.Validate(); err != nil {
		return err
	}

	b, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(filename, b, 0o600)
}
