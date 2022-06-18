package envoy

import (
	"fmt"
	"os"
	"path/filepath"

	kuma_dp "github.com/kumahq/kuma/pkg/config/app/kuma-dp"
)

func GenerateBootstrapFile(cfg kuma_dp.DataplaneRuntime, config []byte) (string, error) {
	configFile := filepath.Join(cfg.ConfigDir, "bootstrap.yaml")
	if err := writeFile(configFile, config, 0o600); err != nil {
		return "", fmt.Errorf("failed to persist Envoy bootstrap config on disk: %w", err)
	}
	return configFile, nil
}

func writeFile(filename string, data []byte, perm os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
		return err
	}
	return os.WriteFile(filename, data, perm)
}
