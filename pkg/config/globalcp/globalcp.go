package globalcp

import "github.com/Kong/kuma/pkg/config"

var _ config.Config = &GlobalCPConfig{}

// Global CP configuration
type GlobalCPConfig struct {
	// Registered Loca CPs for this Global CP
	LocalCPs map[string]string `yaml:"localCPs,omitempty"`
}

func (a *GlobalCPConfig) Sanitize() {
}

func (a *GlobalCPConfig) Validate() error {
	return nil
}

func DefaultGlobalCPConfig() *GlobalCPConfig {
	return &GlobalCPConfig{
		LocalCPs: map[string]string{},
	}
}
