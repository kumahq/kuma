package dp_server

import (
	"github.com/kumahq/kuma/pkg/core/runtime"
)

func SetupServer(rt runtime.Runtime) error {
	if err := rt.Add(rt.DpServer()); err != nil {
		return err
	}
	return nil
}
