package clusters

import (
	"github.com/kumahq/kuma/pkg/core/runtime"
)

func SetupServer(rt runtime.Runtime) error {
	return rt.Add(rt.Clusters())
}
