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
	// Dataplane GC is run only on Universal because on Kubernetes Dataplanes are bounded by ownership to Pods.
	// Therefore on K8S offline dataplanes are cleaned up quickly enough to not run this.
	if rt.Config().Environment != config_core.UniversalEnvironment {
		return nil
	}
	return rt.Add(NewCollector(rt.ResourceManager(), 1*time.Minute, rt.Config().Runtime.Universal.DataplaneCleanupAge))
}
