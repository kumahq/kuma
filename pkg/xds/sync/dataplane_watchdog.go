package sync

import (
	"context"

	"github.com/go-logr/logr"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/xds/cache/mesh"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
)

type DataplaneWatchdogDependencies struct {
	resManager            manager.ResourceManager
	dataplaneProxyBuilder *dataplaneProxyBuilder
	dataplaneReconciler   SnapshotReconciler
	ingressProxyBuilder   *ingressProxyBuilder
	ingressReconciler     SnapshotReconciler
	connectionInfoTracker ConnectionInfoTracker
	envoyCpCtx            *xds_context.ControlPlaneContext
	meshCache             *mesh.Cache
}

type DataplaneWatchdog struct {
	DataplaneWatchdogDependencies
	key      core_model.ResourceKey
	streamId int64
	log      logr.Logger

	// state of watchdog
	prevHash string
	dpType   mesh_proto.DpType
}

func NewDataplaneWatchdog(deps DataplaneWatchdogDependencies, key core_model.ResourceKey, streamId int64) *DataplaneWatchdog {
	return &DataplaneWatchdog{
		DataplaneWatchdogDependencies: deps,
		key:                           key,
		streamId:                      streamId,
		log:                           core.Log.WithValues("key", "key", "streamID", streamId),
	}
}

func (d *DataplaneWatchdog) Sync() error {
	if d.dpType == "" {
		if err := d.determineDpType(); err != nil {
			return err
		}
	}
	switch d.dpType {
	case mesh_proto.IngressDpType:
		return d.syncIngress()
	case mesh_proto.RegularDpType:
		return d.syncDataplane()
	case mesh_proto.GatewayDpType:
		return d.syncDataplane()
	default:
		return nil
	}
}

func (d *DataplaneWatchdog) syncIngress() error {
	envoyCtx := xds_context.Context{
		ControlPlane:   d.envoyCpCtx,
		ConnectionInfo: d.connectionInfoTracker.ConnectionInfo(d.streamId),
	}
	proxy, err := d.ingressProxyBuilder.Build(d.key, d.streamId)
	if err != nil {
		return err
	}
	envoyCtx.Mesh = xds_context.MeshContext{}
	return d.ingressReconciler.Reconcile(envoyCtx, proxy)
}

func (d *DataplaneWatchdog) syncDataplane() error {
	snapshotHash, err := d.meshCache.GetHash(context.Background(), d.key.Mesh)
	if err != nil {
		return err
	}
	if d.prevHash != "" && snapshotHash == d.prevHash {
		return nil
	}
	d.log.Info("snapshot hash updated, reconcile", "prev", d.prevHash, "current", snapshotHash)

	envoyCtx := xds_context.Context{
		ControlPlane:   d.envoyCpCtx,
		ConnectionInfo: d.connectionInfoTracker.ConnectionInfo(d.streamId),
	}
	proxy, meshCtx, err := d.dataplaneProxyBuilder.Build(d.key, d.streamId)
	if err != nil {
		return err
	}
	envoyCtx.Mesh = *meshCtx
	if err := d.dataplaneReconciler.Reconcile(envoyCtx, proxy); err != nil {
		return err
	}
	d.prevHash = snapshotHash
	return nil
}

func (d *DataplaneWatchdog) determineDpType() error {
	dataplane := core_mesh.NewDataplaneResource()
	if err := d.resManager.Get(context.Background(), dataplane, core_store.GetBy(d.key)); err != nil {
		if core_store.IsResourceNotFound(err) {
			return nil
		}
		return err
	}
	d.dpType = dataplane.Spec.DpType()
	return nil
}
