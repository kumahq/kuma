package config

const SanitizedValue = "*****"

type Config interface {
	Sanitize()
	Validate() error
}
