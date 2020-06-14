package globalcp

import "github.com/Kong/kuma/pkg/core/runtime"

func SetupServer(rt runtime.Runtime) error {
	globalCPPoller, err := NewGlobalCPPoller(rt.Config().GlobalCP.LocalCPs)
	if err != nil {
		return err
	}

	return rt.Add(globalCPPoller)
}
