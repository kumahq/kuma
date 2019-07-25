package config

type Config interface {
	Validate() error
}
