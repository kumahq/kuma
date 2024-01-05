package policies

import "github.com/kumahq/kuma/pkg/config"

type PoliciesConfig struct {
	config.BaseConfig

	PluginPoliciesEnabled []string `json:"-" envconfig:"KUMA_PLUGIN_POLICIES_ENABLED" default:""`
}

