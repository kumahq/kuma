package gui_server

import (
	"errors"

	"github.com/Kong/kuma/pkg/config"
)

// Web GUI Server configuration
type GuiServerConfig struct {
	// Port on which the server is exposed
	Port uint32 `yaml:"port" envconfig:"kuma_gui_server_port"`
	// URL of the Api Server that requests with /api prefix will be redirected to. By default autoconfigured to http://locahost:port_of_api_server
	ApiServerUrl string `yaml:"apiServerUrl" envconfig:"kuma_gui_server_api_server_url"`
	// Config of the GUI itself
	GuiConfig *GuiConfig `yaml:"-"` // DEPRECATED, will be removed when GUI is switched to use /api URL
}

func (g *GuiServerConfig) Sanitize() {
}

func (g *GuiServerConfig) Validate() error {
	if g.Port > 65535 {
		return errors.New("Port must be in the range [0, 65535]")
	}
	return nil
}

var _ config.Config = &GuiServerConfig{}

func DefaultGuiServerConfig() *GuiServerConfig {
	return &GuiServerConfig{
		Port:      5683,
		GuiConfig: &GuiConfig{},
	}
}

// Not yet exposed via YAML and env vars on purpose. All of those are autoconfigured
type GuiConfig struct {
	ApiUrl      string
	Environment string
}
