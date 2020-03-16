package config

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	config_proto "github.com/Kong/kuma/pkg/config/app/kumactl/v1alpha1"
	util_files "github.com/Kong/kuma/pkg/util/files"
	util_proto "github.com/Kong/kuma/pkg/util/proto"
)

var DefaultConfigFile = filepath.Join(os.Getenv("HOME"), ".kumactl", "config")

func Load(file string, cfg *config_proto.Configuration) error {
	configFile := DefaultConfigFile
	if file != "" {
		if util_files.FileExists(file) {
			configFile = file
		} else {
			return errors.Errorf("Failed to access configuration file %q", file)
		}
	}
	if util_files.FileExists(configFile) {
		if contents, err := ioutil.ReadFile(configFile); err != nil {
			return errors.Wrapf(err, "Failed to read configuration from file %q", configFile)
		} else if err := util_proto.FromYAML(contents, cfg); err != nil {
			return errors.Wrapf(err, "Failed to parse configuration from file %q", configFile)
		}
	}
	if err := cfg.Validate(); err != nil {
		return errors.Wrapf(err, "Failed to load invalid configuration from file %q", configFile)
	}
	return nil
}

func Save(file string, cfg *config_proto.Configuration) error {
	if err := cfg.Validate(); err != nil {
		return errors.Wrapf(err, "Failed to save invalid configuration: %s", cfg)
	}
	contents, err := util_proto.ToYAML(cfg)
	if err != nil {
		return errors.Wrapf(err, "Failed to format configuration: %#v", cfg)
	}
	configFile := DefaultConfigFile
	if file != "" {
		configFile = file
	}
	dir := filepath.Dir(configFile)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, os.ModeDir|0755); err != nil {
			return errors.Wrapf(err, "Failed to create a directory %q", dir)
		}
	}
	if err := ioutil.WriteFile(configFile, contents, 0600); err != nil {
		return errors.Wrapf(err, "Failed to write configuration into file %q", configFile)
	}
	return nil
}
