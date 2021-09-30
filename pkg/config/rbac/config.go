package rbac

import (
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/config"
)

func DefaultRBACConfig() RBACConfig {
	return RBACConfig{
		Type: "static",
		Static: RBACStaticConfig{
			AdminUsers: []string{
				"admin",
			},
			AdminGroups: []string{
				"admin",
			},
		},
	}
}

// RBACConfig defines a configuration for RBAC
type RBACConfig struct {
	// Type of RBAC strategy (available values: "static")
	Type string `yaml:"type" envconfig:"KUMA_RBAC_TYPE"`
	// Configuration of static RBAC strategy
	Static RBACStaticConfig `yaml:"static"`
}

func (r RBACConfig) Sanitize() {
}

func (r RBACConfig) Validate() error {
	if r.Type == "" {
		return errors.New("Type has to be defined")
	}
	return nil
}

var _ config.Config = &RBACConfig{}

// RBACStaticConfig a static RBAC strategy configuration
type RBACStaticConfig struct {
	// List of users that will be assigned an admin role
	AdminUsers []string `yaml:"adminUsers" envconfig:"KUMA_RBAC_STATIC_ADMIN_USERS"`
	// List of groups that will be assigned an admin role
	AdminGroups []string `yaml:"adminGroups" envconfig:"KUMA_RBAC_STATIC_ADMIN_GROUPS"`
}
