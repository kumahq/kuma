package config

import (
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

func Load(file string) (*Config, error) {
	cfg := DefaultConfig()
	if file == "" {
		core.Log.Info("Skipping reading config from file")
	} else if err := loadFromFile(file, &cfg); err != nil {
		return nil, err
	}

	if err := loadFromEnv(&cfg); err != nil {
		return nil, err
	}

	if err := cfg.Validate(); err != nil {
		return nil, errors.Wrapf(err, "Invalid configuration")
	}
	return &cfg, nil
}

func loadFromFile(file string, cfg *Config) error {
	if !fileExists(file) {
		return errors.Errorf("Failed to access configuration file %q", file)
	}
	if contents, err := ioutil.ReadFile(file); err != nil {
		return errors.Wrapf(err, "Failed to read configuration from file %q", file)
	} else if err := yaml.Unmarshal(contents, &cfg); err != nil {
		return errors.Wrapf(err, "Failed to parse configuration from file %q", file)
	}
	return nil
}

func loadFromEnv(config *Config) error {
	return envconfig.Process("", config)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
