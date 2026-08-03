package config

import (
	"sigs.k8s.io/yaml"
)

func FromYAML(content []byte, cfg Config) error {
	return yaml.Unmarshal(content, cfg)
}

func ToYAML(cfg Config) ([]byte, error) {
	return yaml.Marshal(cfg)
}

// ToJson converts through YAML, because we only have `yaml` tags on Config.
// This JSON cannot be parsed by json.Unmarshal because durations are marshaled by yaml to pretty form like "1s".
// To change it to simple json.Marshal we need to add `json` tag everywhere.
func ToJson(cfg Config) ([]byte, error) {
	yamlBytes, err := ToYAML(cfg)
	if err != nil {
		return nil, err
	}
	// there is no easy way to convert yaml to json using gopkg.in/yaml.v2
	return yaml.YAMLToJSON(yamlBytes)
}
