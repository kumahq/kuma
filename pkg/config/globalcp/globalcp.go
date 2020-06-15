package globalcp

import (
	"net/url"

	"github.com/pkg/errors"

	"github.com/Kong/kuma/pkg/config"
)

var _ config.Config = &GlobalCPConfig{}

// Global CP configuration
type GlobalCPConfig struct {
	// Registered Local CP name to sync URL for this Global CP
	// The sync URL is used in the synchronisation process between Global and Local CPs
	LocalCPs map[string]string `yaml:"localCPs,omitempty"`
}

func (g *GlobalCPConfig) Sanitize() {
}

func (g *GlobalCPConfig) Validate() error {
	for name, localcpurl := range g.LocalCPs {
		_, err := url.ParseRequestURI(localcpurl)
		if err != nil {
			return errors.Wrapf(err, "Local CP %s has invalid url %s", name, localcpurl)
		}
	}
	return nil
}

func DefaultGlobalCPConfig() *GlobalCPConfig {
	return &GlobalCPConfig{
		LocalCPs: map[string]string{},
	}
}
