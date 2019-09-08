package universal

import (
	"time"

	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config"
)

var _ config.Config = &UniversalDiscoveryConfig{}

type UniversalDiscoveryConfig struct {
	// Interval for which the underlying resource store will be checked for changes
	PollingInterval time.Duration `yaml:"pollingInterval" envconfig:"kuma_discovery_universal_polling_interval"`
}

func (u UniversalDiscoveryConfig) Validate() error {
	return nil
}

func DefaultUniversalDiscoveryConfig() *UniversalDiscoveryConfig {
	return &UniversalDiscoveryConfig{
		PollingInterval: time.Second,
	}
}
