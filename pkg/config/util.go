package config

import (
	ghodss_yaml "github.com/ghodss/yaml"
	"gopkg.in/yaml.v2"
)

func FromYAML(content []byte, cfg Config) error {
	return yaml.Unmarshal(content, cfg)
}

func ToYAML(cfg Config) ([]byte, error) {
	return yaml.Marshal(cfg)
}

func ToJson(cfg Config) ([]byte, error) {
	yamlBytes, err := ToYAML(cfg)
	if err != nil {
		return nil, err
	}
	// there is no easy way to convert yaml to json using gopkg.in/yaml.v2
	return ghodss_yaml.YAMLToJSON(yamlBytes)
}
