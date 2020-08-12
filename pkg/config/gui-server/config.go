package gui_server

import (
	"github.com/kumahq/kuma/pkg/config"
)

// Web GUI Server configuration
type GuiServerConfig struct {
	// URL of the Api Server that requests with /api prefix will be redirected to. By default autoconfigured to http://locahost:port_of_api_server
	ApiServerUrl string `yaml:"apiServerUrl" envconfig:"kuma_gui_server_api_server_url"`
}

func (g *GuiServerConfig) Sanitize() {
}

func (g *GuiServerConfig) Validate() error {
	return nil
}

var _ config.Config = &GuiServerConfig{}

func DefaultGuiServerConfig() *GuiServerConfig {
	return &GuiServerConfig{}
}

// Not yet exposed via YAML and env vars on purpose. All of those are autoconfigured
type GuiConfig struct {
	ApiUrl      string
	Environment string
}
