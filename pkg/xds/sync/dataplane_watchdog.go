package sync

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/xds/cache/mesh"
)

type DataplaneWatchdogDependencies struct {
	dataplaneProxyBuilder *DataplaneProxyBuilder
	dataplaneReconciler   SnapshotReconciler
	ingressProxyBuilder   *IngressProxyBuilder
	ingressReconciler     SnapshotReconciler
	xdsContextBuilder     *xdsContextBuilder
	meshCache             *mesh.Cache
	metadataTracker       DataplaneMetadataTracker
}

type DataplaneWatchdog struct {
	DataplaneWatchdogDependencies
	key core_model.ResourceKey
	log logr.Logger

	// state of watchdog
	lastHash         string // last Mesh hash that was used to **successfully** generate Reconcile Envoy config
	dpType           mesh_proto.ProxyType
	proxyTypeSettled bool
}

func NewDataplaneWatchdog(deps DataplaneWatchdogDependencies, dpKey core_model.ResourceKey) *DataplaneWatchdog {
	return &DataplaneWatchdog{
		DataplaneWatchdogDependencies: deps,
		key:                           dpKey,
		log:                           core.Log.WithValues("key", dpKey),
		proxyTypeSettled:              false,
	}
}

func (d *DataplaneWatchdog) Sync() error {
	ctx := context.Background()
	metadata := d.metadataTracker.Metadata(d.key)
	if metadata == nil {
		return errors.New("metadata cannot be nil")
	}

	if d.dpType == "" {
		d.dpType = metadata.GetProxyType()
	}
	// backwards compatibility
	if d.dpType == mesh_proto.DataplaneProxyType && !d.proxyTypeSettled {
		dataplane := core_mesh.NewDataplaneResource()
		if err := d.dataplaneProxyBuilder.CachingResManager.Get(ctx, dataplane, store.GetBy(d.key)); err != nil {
			return err
		}
		if dataplane.Spec.IsIngress() {
			d.dpType = mesh_proto.IngressProxyType
		}
		d.proxyTypeSettled = true
	}
	switch d.dpType {
	case mesh_proto.DataplaneProxyType:
		return d.syncDataplane()
	case mesh_proto.IngressProxyType:
		return d.syncIngress()
	default:
		// It might be a case that dp type is not yet inferred because there is no Dataplane definition yet.
		return nil
	}
}

func (d *DataplaneWatchdog) Cleanup() error {
	proxyID := core_xds.FromResourceKey(d.key)
	switch d.dpType {
	case mesh_proto.DataplaneProxyType:
		return d.dataplaneReconciler.Clear(&proxyID)
	case mesh_proto.IngressProxyType:
		return d.ingressReconciler.Clear(&proxyID)
	default:
		return nil
	}
}

// syncDataplane syncs state of the Dataplane.
// It uses Mesh Hash to decide if we need to regenerate configuration or not.
func (d *DataplaneWatchdog) syncDataplane() error {
	snapshotHash, err := d.meshCache.GetHash(context.Background(), d.key.Mesh)
	if err != nil {
		return err
	}
	if d.lastHash != "" && snapshotHash == d.lastHash {
		// Kuma policies (including Dataplanes and Mesh) has not change therefore there is no need to regenerate configuration.
		return nil
	}
	d.log.V(1).Info("snapshot hash updated, reconcile", "prev", d.lastHash, "current", snapshotHash)

	envoyCtx, err := d.xdsContextBuilder.buildMeshedContext(d.key, d.lastHash)
	if err != nil {
		return err
	}
	proxy, err := d.dataplaneProxyBuilder.Build(d.key, envoyCtx)
	if err != nil {
		return err
	}
	if err := d.dataplaneReconciler.Reconcile(*envoyCtx, proxy); err != nil {
		return err
	}
	d.lastHash = snapshotHash
	return nil
}

// syncIngress synces state of Ingress Dataplane. Notice that it does not use Mesh Hash yet because Ingress supports many Meshes.
func (d *DataplaneWatchdog) syncIngress() error {
	envoyCtx, err := d.xdsContextBuilder.buildContext(d.key)
	if err != nil {
		return err
	}
	proxy, err := d.ingressProxyBuilder.build(d.key)
	if err != nil {
		return err
	}
	return d.ingressReconciler.Reconcile(*envoyCtx, proxy)
}
