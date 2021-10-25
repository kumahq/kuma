package rbac

import (
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/config"
)

const StaticType = "static"

func DefaultRBACConfig() RBACConfig {
	return RBACConfig{
		Type: StaticType,
		Static: RBACStaticConfig{
			AdminResources: AdminResourcesRBACStaticConfig{
				Users:  []string{"admin"},
				Groups: []string{"admin"},
			},
			GenerateDPToken: GenerateDPTokenRBACStaticConfig{
				Users:  []string{"admin"},
				Groups: []string{"admin"},
			},
			GenerateUserToken: GenerateUserTokenRBACStaticConfig{
				Users:  []string{"admin"},
				Groups: []string{"admin"},
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
	// AdminResources defines an access to admin resources (Secret/GlobalSecret)
	AdminResources AdminResourcesRBACStaticConfig `yaml:"adminResources"`
	// GenerateDPToken defines an access to generating dataplane token
	GenerateDPToken GenerateDPTokenRBACStaticConfig `yaml:"generateDpToken"`
	// GenerateDPToken defines an access to generating user token
	GenerateUserToken GenerateUserTokenRBACStaticConfig `yaml:"generateUserToken"`
}

type AdminResourcesRBACStaticConfig struct {
	// List of users that are allowed to access admin resources
	Users []string `yaml:"users" envconfig:"KUMA_RBAC_STATIC_ADMIN_RESOURCES_USERS"`
	// List of groups that are allowed to access admin resources
	Groups []string `yaml:"groups" envconfig:"KUMA_RBAC_STATIC_ADMIN_RESOURCES_GROUPS"`
}

type GenerateDPTokenRBACStaticConfig struct {
	// List of users that are allowed to generate dataplane token
	Users []string `yaml:"users" envconfig:"KUMA_RBAC_STATIC_GENERATE_DP_TOKEN_USERS"`
	// List of groups that are allowed to generate dataplane token
	Groups []string `yaml:"groups" envconfig:"KUMA_RBAC_STATIC_GENERATE_DP_TOKEN_GROUPS"`
}

type GenerateUserTokenRBACStaticConfig struct {
	// List of users that are allowed to generate user token
	Users []string `yaml:"users" envconfig:"KUMA_RBAC_STATIC_GENERATE_USER_TOKEN_USERS"`
	// List of groups that are allowed to generate user token
	Groups []string `yaml:"groups" envconfig:"KUMA_RBAC_STATIC_GENERATE_USER_TOKEN_GROUPS"`
}
