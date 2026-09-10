package v1alpha1

// Config is an entity that represents dynamic configuration that is stored in
// underlying storage. For now it's used only for internal mechanisms.
type Config struct {
	// Configuration that is stored, for example as JSON.
	Config string `json:"config,omitempty"`
}

func (c *Config) GetConfig() string {
	if c == nil {
		return ""
	}
	return c.Config
}
