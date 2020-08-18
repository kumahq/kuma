package gc

import (
	"time"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/runtime"
)

var (
	gcLog = core.Log.WithName("garbage-collector")
)

func Setup(rt runtime.Runtime) error {
	if rt.Config().Environment != config_core.UniversalEnvironment {
		return nil
	}
	return rt.Add(NewCollector(rt.ResourceManager(), 2*time.Second, rt.Config().Runtime.Universal.DataplaneCleanupAge))
}
