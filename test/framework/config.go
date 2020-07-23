package framework

import (
	"gopkg.in/yaml.v2"

	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	"github.com/kumahq/kuma/pkg/config/mode"
)

func addGlobal(rawYAML, lbAdress string) (string, error) {
	cfg := kuma_cp.Config{}
	err := yaml.Unmarshal([]byte(rawYAML), &cfg)
	if err != nil {
		return "", err
	}

	if cfg.Mode == nil {
		cfg.Mode = mode.DefaultModeConfig()
		cfg.Mode.Mode = mode.Global
	}

	if cfg.Mode.Global == nil {
		cfg.Mode.Global = mode.DefaultGlobalConfig()
	}

	if lbAdress != "" {
		cfg.Mode.Global.LBAddress = lbAdress
	}

	yamlBytes, err := yaml.Marshal(&cfg)
	if err != nil {
		return "", err
	}

	return string(yamlBytes), nil
}
