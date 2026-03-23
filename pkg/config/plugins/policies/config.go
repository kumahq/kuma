package policies

type Config struct {
	Enabled []string `json:"pluginPoliciesEnabled" envconfig:"KUMA_PLUGIN_POLICIES_ENABLED" default:""`
}

func (*Config) PostProcess() error {
	return nil
}

func (*Config) Sanitize() {
}

func (*Config) Validate() error {
	return nil
}

func Default() *Config {
	return &Config{
		Enabled: DefaultEnabled,
	}
}
