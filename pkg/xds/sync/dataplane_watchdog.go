package sync

import (
	"context"

	"github.com/go-logr/logr"

	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/store"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/xds/cache/mesh"
)

type DataplaneWatchdogDependencies struct {
	resManager            manager.ResourceManager
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
	key      core_model.ResourceKey
	streamId int64
	log      logr.Logger

	// state of watchdog
	lastHash         string // last Mesh hash that was used to **successfully** generate Reconcile Envoy config
	dpType           mesh_proto.ProxyType
	proxyTypeSettled bool
}

func NewDataplaneWatchdog(deps DataplaneWatchdogDependencies, proxyId *core_xds.ProxyId, streamId int64) *DataplaneWatchdog {
	return &DataplaneWatchdog{
		DataplaneWatchdogDependencies: deps,
		key:                           proxyId.ToResourceKey(),
		streamId:                      streamId,
		log:                           core.Log.WithValues("key", proxyId.ToResourceKey(), "streamID", streamId),
		proxyTypeSettled:              false,
	}
}

func (d *DataplaneWatchdog) Sync() error {
	ctx := context.Background()

	if d.dpType == "" {
		d.dpType = d.metadataTracker.Metadata(d.streamId).GetProxyType()
	}
	// backwards compatibility
	if d.dpType == mesh_proto.DataplaneProxyType && !d.proxyTypeSettled {
		dataplane := mesh_core.NewDataplaneResource()
		if err := d.dataplaneProxyBuilder.CachingResManager.Get(ctx, dataplane, store.GetBy(d.key)); err != nil {
			return err
		}
		if dataplane.Spec.IsIngress() {
			d.dpType = mesh_proto.IngressProxyType
		}
		d.proxyTypeSettled = true
	}
	switch d.dpType {
	case mesh_proto.DataplaneProxyType, mesh_proto.GatewayProxyType:
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
	case mesh_proto.DataplaneProxyType, mesh_proto.GatewayProxyType:
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

	envoyCtx, err := d.xdsContextBuilder.buildMeshedContext(d.streamId, d.key.Mesh, d.lastHash)
	if err != nil {
		return err
	}
	proxy, err := d.dataplaneProxyBuilder.build(d.key, d.streamId, &envoyCtx.Mesh)
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
	envoyCtx := d.xdsContextBuilder.buildContext(d.streamId)
	proxy, err := d.ingressProxyBuilder.build(d.key, d.streamId)
	if err != nil {
		return err
	}
	return d.ingressReconciler.Reconcile(*envoyCtx, proxy)
}
