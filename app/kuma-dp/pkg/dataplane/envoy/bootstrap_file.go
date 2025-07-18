package envoy

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	kuma_dp "github.com/kumahq/kuma/pkg/config/app/kuma-dp"
)

func GenerateBootstrapFile(cfg kuma_dp.DataplaneRuntime, config []byte) (string, error) {
	configFile := filepath.Join(cfg.WorkDir, "bootstrap.yaml")
	if err := writeFile(configFile, config, 0o600); err != nil {
		return "", errors.Wrap(err, "failed to persist Envoy bootstrap config on disk")
	}
	return configFile, nil
}

func writeFile(filename string, data []byte, perm os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
		return err
	}
	return os.WriteFile(filename, data, perm)
}
