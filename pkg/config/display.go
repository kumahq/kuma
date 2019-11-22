package config

import (
	"encoding/json"
	ghodss_yaml "github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"reflect"
)

func ConfigForDisplay(cfg Config) (Config, error) {
	// copy config so we don't override values, because nested structs in config are pointers
	newCfg, err := copyConfig(cfg)
	if err != nil {
		return nil, err
	}
	newCfg.Sanitize()
	return newCfg, nil
}

func ConfigForDisplayYAML(cfg Config) ([]byte, error) {
	cfgForDisplay, err := ConfigForDisplay(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "could not prepare config for display")
	}
	return yaml.Marshal(&cfgForDisplay)
}

func ConfigForDisplayJSON(cfg Config) ([]byte, error) {
	yamlBytes, err := ConfigForDisplayYAML(cfg)
	if err != nil {
		return nil, err
	}
	// there is no easy way to convert yaml to json using gopkg.in/yaml.v2
	return ghodss_yaml.YAMLToJSON(yamlBytes)
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
