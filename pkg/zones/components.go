package zones

import (
	"github.com/kumahq/kuma/pkg/core/runtime"
)

func SetupServer(rt runtime.Runtime) error {
	return rt.Add(rt.Zones())
}
