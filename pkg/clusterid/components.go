package clusterid

import (
	"github.com/pkg/errors"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
)

func Setup(rt core_runtime.Runtime) error {
	creator := &clusterIDCreator{configManager: rt.ConfigManager()}
	reader := &clusterIDReader{rt: rt}
	switch rt.Config().Mode {
	case config_core.Standalone, config_core.Global:
		if err := rt.Add(creator, reader); err != nil {
			return err
		}
		return nil
	case config_core.Zone:
		if err := rt.Add(reader); err != nil {
			return err
		}
		return nil
	default:
		return errors.Errorf("unknown mode of the CP %s", rt.Config().Mode)
	}
}
