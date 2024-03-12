package policies

type Config struct {
	Enabled []string `json:"pluginPoliciesEnabled" envconfig:"KUMA_PLUGIN_POLICIES_ENABLED" default:""`
}

func (c *Config) PostProcess() error {
	return nil
}

func (c *Config) Sanitize() {
}

func (c *Config) Validate() error {
	return nil
}
