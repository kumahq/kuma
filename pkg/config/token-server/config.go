package token_server

import (
	"errors"
	"github.com/Kong/kuma/pkg/config"
)

func DefaultDataplaneTokenServerConfig() *DataplaneTokenServerConfig {
	return &DataplaneTokenServerConfig{
		Port: 5679,
	}
}

// Dataplane Token Server configuration
type DataplaneTokenServerConfig struct {
	// Port of the server
	Port int `yaml:"port" envconfig:"kuma_dataplane_token_server_port"`
}

var _ config.Config = &DataplaneTokenServerConfig{}

func (i *DataplaneTokenServerConfig) Validate() error {
	if i.Port < 0 {
		return errors.New("Port cannot be negative")
	}
	return nil
}

