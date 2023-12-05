package clusterid

import (
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
)

func Setup(rt core_runtime.Runtime) error {
	creator := &clusterIDCreator{configManager: rt.ConfigManager()}
	reader := &clusterIDReader{rt: rt}
	if err := rt.Add(reader); err != nil {
		return err
	}
	if !rt.Config().IsFederatedZoneCP() {
		if err := rt.Add(creator); err != nil {
			return err
		}
	}
	return nil
}
