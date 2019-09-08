package config

import "github.com/kelseyhightower/envconfig"

const envPrefix = "kuma"

func Load(spec interface{}) error {
	return envconfig.Process(envPrefix, spec)
}
