package framework

import (
	"net/url"

	yaml "gopkg.in/yaml.v2"

	kuma_cp "github.com/Kong/kuma/pkg/config/app/kuma-cp"
	"github.com/Kong/kuma/pkg/config/mode"
)

func addGlobal(rawYAML, lbAdress, rawURL string) (resultYaml string, err error) {
	cfg := kuma_cp.Config{}
	err = yaml.Unmarshal([]byte(rawYAML), &cfg)
	if err != nil {
		return
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

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return
	}

	cfg.Mode.Global.Zones = append(cfg.Mode.Global.Zones, &mode.ZoneConfig{
		Remote:  mode.EndpointConfig{Address: rawURL},
		Ingress: mode.EndpointConfig{Address: parsed.Host},
	})

	yamlBytes, err := yaml.Marshal(&cfg)
	if err != nil {
		return
	}

	resultYaml = string(yamlBytes)
	return
}
