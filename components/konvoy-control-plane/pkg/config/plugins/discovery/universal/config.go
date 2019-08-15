package universal

import (
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config"
	"time"
)

var _ config.Config = &UniversalDiscoveryConfig{}

type UniversalDiscoveryConfig struct {
	// Interval for which the underlying resource store will be checked for changes
	PollingInterval time.Duration `yaml:"pollingInterval" envconfig:"konvoy_discovery_universal_polling_interval"`
}

func (u UniversalDiscoveryConfig) Validate() error {
	return nil
}

func DefaultUniversalDiscoveryConfig() *UniversalDiscoveryConfig {
	return &UniversalDiscoveryConfig{
		PollingInterval: time.Second,
	}
}
