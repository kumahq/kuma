package globalcp

import "github.com/Kong/kuma/pkg/config"

var _ config.Config = &GlobalCPConfig{}

// Global CP configuration
type GlobalCPConfig struct {
	// Registered Local CP name to sync URL for this Global CP
	// The sync URL is used in the synchronisation process between Global and Local CPs
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
