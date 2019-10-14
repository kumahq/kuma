package token_server

import (
	"github.com/Kong/kuma/pkg/config"
	"github.com/pkg/errors"
)

func DefaultDataplaneTokenServerConfig() *DataplaneTokenServerConfig {
	return &DataplaneTokenServerConfig{
		Port: 5679,
	}
}

// Dataplane Token Server configuration
type DataplaneTokenServerConfig struct {
	// Port of the server
	Port uint32 `yaml:"port" envconfig:"kuma_dataplane_token_server_port"`
}

var _ config.Config = &DataplaneTokenServerConfig{}

func (i *DataplaneTokenServerConfig) Validate() error {
	if i.Port > 65535 {
		return errors.New("Port must be in the range [0, 65535]")
	}
	return nil
}
