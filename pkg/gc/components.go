package gc

import (
	"time"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"

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
	switch rt.Config().Environment {
	// Dataplane GC is run only on Universal because on Kubernetes Dataplanes are bounded by ownership to Pods.
	// Therefore on K8S offline dataplanes are cleaned up quickly enough to not run this.
	case config_core.UniversalEnvironment:
		return rt.Add(
			NewCollector(rt.ResourceManager(), 1*time.Minute, rt.Config().Runtime.Universal.DataplaneCleanupAge),
		)
	default:
		return nil
	}
}

func setupFinalizer(rt runtime.Runtime) error {
	switch rt.Config().Mode {
	case config_core.Standalone:
		return rt.Add(
			NewSubscriptionFinalizer(
				rt.ResourceManager(),
				func() *time.Ticker { return time.NewTicker(rt.Config().Metrics.Dataplane.IdleTimeout / 2) },
				mesh.DataplaneInsightType,
			),
		)
	case config_core.Zone:
		return rt.Add(
			NewSubscriptionFinalizer(
				rt.ResourceManager(),
				func() *time.Ticker { return time.NewTicker(rt.Config().Metrics.Dataplane.IdleTimeout / 2) },
				mesh.DataplaneInsightType, mesh.ZoneIngressInsightType),
		)
	case config_core.Global:
		return rt.Add(
			NewSubscriptionFinalizer(
				rt.ResourceManager(),
				func() *time.Ticker { return time.NewTicker(rt.Config().Metrics.Zone.IdleTimeout / 2) },
				system.ZoneInsightType),
		)
	default:
		return errors.Errorf("unknown Kuma CP mode %v", rt.Config().Mode)
	}
}
