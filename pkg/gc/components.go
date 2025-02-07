package gc

import (
	"time"

	"github.com/pkg/errors"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/runtime"
)

func Setup(rt runtime.Runtime) error {
	if err := setupDpCollector(rt); err != nil {
		return err
	}
	if err := setupZoneResourceCollector(rt); err != nil {
		return err
	}
	if err := setupFinalizer(rt); err != nil {
		return err
	}
	return nil
}

func setupDpCollector(rt runtime.Runtime) error {
	if rt.Config().Environment != config_core.UniversalEnvironment || rt.Config().Mode == config_core.Global {
		// Dataplane GC is run only on Universal because on Kubernetes Dataplanes are bounded by ownership to Pods.
		// Therefore, on K8S offline dataplanes are cleaned up quickly enough to not run this.
		return nil
	}
	collector, err := NewCollector(
		rt.ResourceManager(),
		func() *time.Ticker { return time.NewTicker(1 * time.Minute) },
		rt.Config().Runtime.Universal.DataplaneCleanupAge.Duration,
		rt.Metrics(),
		"dp",
		[]InsightToResource{
			{
				Insight:  mesh.DataplaneInsightType,
				Resource: mesh.DataplaneType,
			},
		},
	)
	if err != nil {
		return err
	}
	return rt.Add(collector)
}

func setupZoneResourceCollector(rt runtime.Runtime) error {
	if rt.Config().Environment != config_core.UniversalEnvironment || rt.Config().Mode == config_core.Global {
		// ZoneIngress/ZoneEgress GC is run only on Universal because on Kubernetes ZoneIngress/ZoneEgress are bounded by ownership to Pods.
		// Therefore, on K8S offline dataplanes are cleaned up quickly enough to not run this.
		return nil
	}
	collector, err := NewCollector(
		rt.ResourceManager(),
		func() *time.Ticker { return time.NewTicker(1 * time.Minute) },
		rt.Config().Runtime.Universal.ZoneResourceCleanupAge.Duration,
		rt.Metrics(),
		"zone",
		[]InsightToResource{
			{
				Insight:  mesh.ZoneEgressInsightType,
				Resource: mesh.ZoneEgressType,
			},
			{
				Insight:  mesh.ZoneIngressInsightType,
				Resource: mesh.ZoneIngressType,
			},
		},
	)
	if err != nil {
		return err
	}
	return rt.Add(collector)
}

func setupFinalizer(rt runtime.Runtime) error {
	var newTicker func() *time.Ticker
	var resourceTypes []model.ResourceType

	switch rt.Config().Mode {
	case config_core.Zone:
		newTicker = func() *time.Ticker {
			return time.NewTicker(rt.Config().Metrics.Dataplane.IdleTimeout.Duration)
		}
		resourceTypes = []model.ResourceType{
			mesh.DataplaneInsightType,
			mesh.ZoneIngressInsightType,
			mesh.ZoneEgressInsightType,
		}
	case config_core.Global:
		newTicker = func() *time.Ticker {
			return time.NewTicker(rt.Config().Metrics.Zone.IdleTimeout.Duration)
		}
		resourceTypes = []model.ResourceType{
			system.ZoneInsightType,
		}
	default:
		return errors.Errorf("unknown Kuma CP mode %s", rt.Config().Mode)
	}

	finalizer, err := NewSubscriptionFinalizer(rt.ResourceManager(), rt.Tenants(), newTicker, rt.Metrics(), rt.Extensions(), rt.Config().Store.Upsert, resourceTypes...)
	if err != nil {
		return err
	}
	return rt.Add(finalizer)
}
