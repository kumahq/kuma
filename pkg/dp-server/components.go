package dp_server

import (
	config_core "github.com/kumahq/kuma/v2/pkg/config/core"
	"github.com/kumahq/kuma/v2/pkg/core/runtime"
)

func SetupServer(rt runtime.Runtime) error {
	if rt.Config().Mode == config_core.Global {
		return nil
	}
	if err := rt.Add(rt.DpServer()); err != nil {
		return err
	}
	return nil
}
