package policies

import (
	"github.com/kumahq/kuma/pkg/config"
)

var _ config.Config = &PoliciesConfig{}

// Pluggable policies configuration
type PoliciesConfig struct {
	// Pluggable policies and their order. Order of policies
	// is important and consider setting those policies
	// that create resources should be at the beginning so
	// the policies applying configuration will edit all the
	// resources.
	Enabled []string `json:"enabled" envconfig:"kuma_policies_enabled"`
}

func DefaultPoliciesConfig() PoliciesConfig {
	return PoliciesConfig{
		Enabled: []string{
			"meshaccesslog",
			"meshtrace",
			"meshratelimit",
			"meshtimeout",
			"meshtrafficpermission",
		},
	}
}

func (c *PoliciesConfig) Sanitize() {
}

func (c *PoliciesConfig) Validate() error {
	return nil
}
