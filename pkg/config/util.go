package config

import (
	"gopkg.in/yaml.v2"
)

func FromYAML(content []byte, cfg Config) error {
	return yaml.Unmarshal(content, cfg)
}

func ToYAML(cfg Config) ([]byte, error) {
	return yaml.Marshal(cfg)
}
