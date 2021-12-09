package config

import (
	"os"

	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"

	"github.com/kumahq/kuma/pkg/core"
)

func Load(file string, cfg Config) error {
	if file == "" {
		core.Log.Info("Skipping reading config from file")
	} else if err := loadFromFile(file, cfg); err != nil {
		return err
	}

	if err := loadFromEnv(cfg); err != nil {
		return err
	}

	if err := cfg.Validate(); err != nil {
		return errors.Wrapf(err, "Invalid configuration")
	}
	return nil
}

func loadFromFile(file string, cfg Config) error {
	if !fileExists(file) {
		return errors.Errorf("Failed to access configuration file %q", file)
	}
	if contents, err := os.ReadFile(file); err != nil {
		return errors.Wrapf(err, "Failed to read configuration from file %q", file)
	} else if err := yaml.Unmarshal(contents, cfg); err != nil {
		return errors.Wrapf(err, "Failed to parse configuration from file %q", file)
	}
	return nil
}

func loadFromEnv(config Config) error {
	return envconfig.Process("", config)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
