package gui_server

import (
	"errors"
	"github.com/Kong/kuma/pkg/config"
	"os"
	"path/filepath"
)

// Web GUI Server configuration
type GuiServerConfig struct {
	// Port on which the server is exposed
	Port uint32 `yaml:"port" envconfig:"kuma_gui_server_port"`
	// Directory from which GUI is served. All files in this directory are exposed.
	Directory string `yaml:"directory" envconfig:"kuma_gui_server_directory"`
}

func (g *GuiServerConfig) Validate() error {
	if g.Port > 65535 {
		return errors.New("Port must be in the range [0, 65535]")
	}
	return nil
}

var _ config.Config = &GuiServerConfig{}

func DefaultGuiServerConfig() *GuiServerConfig {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)
	return &GuiServerConfig{
		Port:      5683,
		Directory: filepath.Join(exPath, "gui"),
	}
}
