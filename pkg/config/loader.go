package config

import (
	"fmt"
	"os"

	"github.com/kelseyhightower/envconfig"
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
		return fmt.Errorf("Invalid configuration: %w", err)
	}
	return nil
}

func loadFromFile(file string, cfg Config) error {
	if !fileExists(file) {
		return fmt.Errorf("Failed to access configuration file %q", file)
	}
	if contents, err := os.ReadFile(file); err != nil {
		return fmt.Errorf("Failed to read configuration from file %q: %w", file, err)
	} else if err := yaml.Unmarshal(contents, cfg); err != nil {
		return fmt.Errorf("Failed to parse configuration from file %q: %w", file, err)
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
