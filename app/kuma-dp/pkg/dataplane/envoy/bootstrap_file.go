package envoy

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"

	kuma_dp "github.com/Kong/kuma/pkg/config/app/kuma-dp"
	util_proto "github.com/Kong/kuma/pkg/util/proto"
)

func GenerateBootstrapFile(cfg kuma_dp.DataplaneRuntime, config proto.Message) (string, error) {
	switch v := config.(type) {
	case (interface{ Validate() error }):
		if err := v.Validate(); err != nil {
			return "", errors.Wrap(err, "Envoy bootstrap config is not valid")
		}
	}
	data, err := util_proto.ToYAML(config)
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal Envoy config")
	}
	configFile := filepath.Join(cfg.ConfigDir, "bootstrap.yaml")
	if err := writeFile(configFile, data, 0600); err != nil {
		return "", errors.Wrap(err, "failed to persist Envoy bootstrap config on disk")
	}
	return configFile, nil
}

func writeFile(filename string, data []byte, perm os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return err
	}
	return ioutil.WriteFile(filename, data, perm)
}
