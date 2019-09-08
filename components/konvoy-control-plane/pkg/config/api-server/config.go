package api_server

import (
	"errors"

	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config"
)

var _ config.Config = &ApiServerConfig{}

// API Server configuration
type ApiServerConfig struct {
	// Port of the API Server
	Port int `yaml:"port" envconfig:"kuma_api_server_port"`
	// If true, then API Server will operate in read only mode (serving GET requests)
	ReadOnly bool `yaml:"readOnly" envconfig:"kuma_api_server_read_only"`
	// Path on which the Open API docs will be exposed
	ApiDocsPath string `yaml:"apiDocsPath" envconfig:"kuma_api_server_api_docs_path"`
}

func (a *ApiServerConfig) Validate() error {
	if a.Port < 0 {
		return errors.New("Port cannot be negative")
	}
	if len(a.ApiDocsPath) < 1 {
		return errors.New("ApiDocsPath should not be empty")
	}
	return nil
}

func DefaultApiServerConfig() *ApiServerConfig {
	return &ApiServerConfig{
		Port:        5681,
		ReadOnly:    false,
		ApiDocsPath: "/apidocs.json",
	}
}
