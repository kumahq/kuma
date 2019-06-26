package config

import "github.com/kelseyhightower/envconfig"

const envPrefix = "konvoy"

func Load(spec interface{}) error {
	return envconfig.Process(envPrefix, spec)
}
