package access

import (
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/config"
)

const StaticType = "static"

func DefaultAccessConfig() AccessConfig {
	return AccessConfig{
		Type: StaticType,
		Static: StaticAccessConfig{
			AdminResources: AdminResourcesStaticAccessConfig{
				Users:  []string{"mesh-system:admin"},
				Groups: []string{"mesh-system:admin"},
			},
			GenerateDPToken: GenerateDPTokenStaticAccessConfig{
				Users:  []string{"mesh-system:admin"},
				Groups: []string{"mesh-system:admin"},
			},
			GenerateUserToken: GenerateUserTokenStaticAccessConfig{
				Users:  []string{"mesh-system:admin"},
				Groups: []string{"mesh-system:admin"},
			},
		},
	}
}

// AccessConfig defines a configuration for acccess control
type AccessConfig struct {
	// Type of the access strategy (available values: "static")
	Type string `yaml:"type" envconfig:"KUMA_ACCESS_TYPE"`
	// Configuration of static access strategy
	Static StaticAccessConfig `yaml:"static"`
}

func (r AccessConfig) Sanitize() {
}

func (r AccessConfig) Validate() error {
	if r.Type == "" {
		return errors.New("Type has to be defined")
	}
	return nil
}

var _ config.Config = &AccessConfig{}

// StaticAccessConfig a static access strategy configuration
type StaticAccessConfig struct {
	// AdminResources defines an access to admin resources (Secret/GlobalSecret)
	AdminResources AdminResourcesStaticAccessConfig `yaml:"adminResources"`
	// GenerateDPToken defines an access to generating dataplane token
	GenerateDPToken GenerateDPTokenStaticAccessConfig `yaml:"generateDpToken"`
	// GenerateDPToken defines an access to generating user token
	GenerateUserToken GenerateUserTokenStaticAccessConfig `yaml:"generateUserToken"`
}

type AdminResourcesStaticAccessConfig struct {
	// List of users that are allowed to access admin resources
	Users []string `yaml:"users" envconfig:"KUMA_ACCESS_STATIC_ADMIN_RESOURCES_USERS"`
	// List of groups that are allowed to access admin resources
	Groups []string `yaml:"groups" envconfig:"KUMA_ACCESS_STATIC_ADMIN_RESOURCES_GROUPS"`
}

type GenerateDPTokenStaticAccessConfig struct {
	// List of users that are allowed to generate dataplane token
	Users []string `yaml:"users" envconfig:"KUMA_ACCESS_STATIC_GENERATE_DP_TOKEN_USERS"`
	// List of groups that are allowed to generate dataplane token
	Groups []string `yaml:"groups" envconfig:"KUMA_ACCESS_STATIC_GENERATE_DP_TOKEN_GROUPS"`
}

type GenerateUserTokenStaticAccessConfig struct {
	// List of users that are allowed to generate user token
	Users []string `yaml:"users" envconfig:"KUMA_ACCESS_STATIC_GENERATE_USER_TOKEN_USERS"`
	// List of groups that are allowed to generate user token
	Groups []string `yaml:"groups" envconfig:"KUMA_ACCESS_STATIC_GENERATE_USER_TOKEN_GROUPS"`
}
