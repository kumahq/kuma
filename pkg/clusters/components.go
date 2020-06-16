package clusters

import (
	"github.com/Kong/kuma/pkg/core/runtime"
)

func SetupServer(rt runtime.Runtime) error {
	return rt.Add(rt.Clusters())
}
