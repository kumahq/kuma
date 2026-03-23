package config

const SanitizedValue = "*****"

type Config interface {
	Sanitize()
	Validate() error
	PostProcess() error
}

var _ Config = BaseConfig{}

type BaseConfig struct{}

func (BaseConfig) Sanitize()          {}
func (BaseConfig) PostProcess() error { return nil }
func (BaseConfig) Validate() error    { return nil }
