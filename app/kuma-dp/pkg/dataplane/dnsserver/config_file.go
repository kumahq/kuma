package dnsserver

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	kuma_dp "github.com/kumahq/kuma/pkg/config/app/kuma-dp"
)

func GenerateConfigFile(cfg kuma_dp.DNS, config []byte) (string, error) {
	configFile := filepath.Join(cfg.ConfigDir, "Corefile")
	if err := writeFile(configFile, config, 0600); err != nil {
		return "", errors.Wrap(err, "failed to persist Envoy bootstrap config on disk")
	}
	return configFile, nil
}

func writeFile(filename string, data []byte, perm os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return err
	}
	return os.WriteFile(filename, data, perm)
}
