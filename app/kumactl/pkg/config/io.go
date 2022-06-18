package config

import (
	"fmt"
	"os"
	"path/filepath"

	config_proto "github.com/kumahq/kuma/pkg/config/app/kumactl/v1alpha1"
	util_files "github.com/kumahq/kuma/pkg/util/files"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var DefaultConfigFile = filepath.Join(os.Getenv("HOME"), ".kumactl", "config")

func Load(file string, cfg *config_proto.Configuration) error {
	configFile := DefaultConfigFile
	if file != "" {
		if util_files.FileExists(file) {
			configFile = file
		} else {
			return fmt.Errorf("Failed to access configuration file %q", file)
		}
	}
	if util_files.FileExists(configFile) {
		if contents, err := os.ReadFile(configFile); err != nil {
			return fmt.Errorf("Failed to read configuration from file %q: %w", configFile, err)
		} else if err := util_proto.FromYAML(contents, cfg); err != nil {
			return fmt.Errorf("Failed to parse configuration from file %q: %w", configFile, err)
		}
	}
	return nil
}

func Save(file string, cfg *config_proto.Configuration) error {
	contents, err := util_proto.ToYAML(cfg)
	if err != nil {
		return fmt.Errorf("Failed to format configuration: %#v: %w", cfg, err)
	}
	configFile := DefaultConfigFile
	if file != "" {
		configFile = file
	}
	dir := filepath.Dir(configFile)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, os.ModeDir|0o755); err != nil {
			return fmt.Errorf("Failed to create a directory %q: %w", dir, err)
		}
	}
	if err := os.WriteFile(configFile, contents, 0o600); err != nil {
		return fmt.Errorf("Failed to write configuration into file %q: %w", configFile, err)
	}
	return nil
}
