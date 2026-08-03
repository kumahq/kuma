package config

const SanitizedValue = "*****"

type Config interface {
	Sanitize()
	Validate() error
	PostProcess() error
}

var _ Config = BaseConfig{}

type BaseConfig struct{}

func (c BaseConfig) Sanitize()          {}
func (c BaseConfig) PostProcess() error { return nil }
func (c BaseConfig) Validate() error    { return nil }
