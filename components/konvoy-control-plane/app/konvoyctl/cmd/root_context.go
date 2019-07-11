package cmd

import (
	"github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/pkg/config"
	config_proto "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config/app/konvoyctl/v1alpha1"
)

func (rc *rootContext) LoadConfig() error {
	return config.Load(rc.args.configFile, &rc.runtime.config)
}

func (rc *rootContext) SaveConfig() error {
	return config.Save(rc.args.configFile, &rc.runtime.config)
}

func (rc *rootContext) Config() *config_proto.Configuration {
	return &rc.runtime.config
}
