package gc

import (
	"time"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/core/runtime"
)

func Setup(rt runtime.Runtime) error {
	if err := setupCollector(rt); err != nil {
		return err
	}
	if err := setupFinalizer(rt); err != nil {
		return err
	}
	return nil
}

func setupCollector(rt runtime.Runtime) error {
	// Dataplane GC is run only on Universal because on Kubernetes Dataplanes are bounded by ownership to Pods.
	// Therefore on K8S offline dataplanes are cleaned up quickly enough to not run this.
	if rt.Config().Environment != config_core.UniversalEnvironment {
		return nil
	}
	return rt.Add(
		NewCollector(rt.ResourceManager(), 1*time.Minute, rt.Config().Runtime.Universal.DataplaneCleanupAge),
	)
}

func setupFinalizer(rt runtime.Runtime) error {
	return rt.Add(
		NewFinalizer(rt.ResourceManager(), 1*time.Minute),
	)
}
