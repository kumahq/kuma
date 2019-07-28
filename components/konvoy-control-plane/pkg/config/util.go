package config

import (
	"gopkg.in/yaml.v2"
)

func ToYAML(cfg Config) ([]byte, error) {
	return yaml.Marshal(cfg)
}
