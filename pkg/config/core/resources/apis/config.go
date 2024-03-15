package apis

type Config struct {
	Enabled []string `json:"enabled" envconfig:"KUMA_CORE_RESOURCES_ENABLED" default:""`
}

func (c *Config) PostProcess() error {
	return nil
}

func (c *Config) Sanitize() {
}

func (c *Config) Validate() error {
	return nil
}
