package xds

import (
	"github.com/pkg/errors"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/xds/bootstrap"
	"github.com/kumahq/kuma/pkg/xds/server"
)

func Setup(rt core_runtime.Runtime) error {
	if rt.Config().Mode == config_core.Global {
		return nil
	}
	if err := server.RegisterXDS(rt); err != nil {
		return errors.Wrap(err, "could not register XDS")
	}
	if err := bootstrap.RegisterBootstrap(rt); err != nil {
		return errors.Wrap(err, "could not register Bootstrap")
	}
	return nil
}
