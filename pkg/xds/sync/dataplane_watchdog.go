package sync

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/xds/cache/mesh"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
)

type DataplaneWatchdogDependencies struct {
	dataplaneProxyBuilder *DataplaneProxyBuilder
	dataplaneReconciler   SnapshotReconciler
	ingressProxyBuilder   *IngressProxyBuilder
	ingressReconciler     SnapshotReconciler
	envoyCpCtx            *xds_context.ControlPlaneContext
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
	metadata := d.metadataTracker.Metadata(d.key)
	if metadata == nil {
		return errors.New("metadata cannot be nil")
	}

	if d.dpType == "" {
		d.dpType = metadata.GetProxyType()
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
		d.envoyCpCtx.Secrets.Cleanup(d.key)
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
	meshCtx, err := d.meshCache.GetMeshContext(context.Background(), syncLog, d.key.Mesh)
	if err != nil {
		return err
	}

	certInfo := d.envoyCpCtx.Secrets.Info(d.key)
	syncForCert := certInfo != nil && certInfo.ExpiringSoon() // check if we need to regenerate config because identity cert is expiring soon.
	syncForConfig := meshCtx.Hash != d.lastHash               // check if we need to regenerate config because Kuma policies has changed.
	if !syncForCert && !syncForConfig {
		return nil
	}
	if syncForConfig {
		d.log.V(1).Info("snapshot hash updated, reconcile", "prev", d.lastHash, "current", meshCtx.Hash)
	}
	if syncForCert {
		d.log.V(1).Info("certs expiring soon, reconcile")
	}

	envoyCtx := &xds_context.Context{
		ControlPlane: d.envoyCpCtx,
		Mesh:         meshCtx,
	}
	proxy, err := d.dataplaneProxyBuilder.Build(d.key, envoyCtx)
	if err != nil {
		return err
	}
	if !envoyCtx.Mesh.Resource.MTLSEnabled() {
		d.envoyCpCtx.Secrets.Cleanup(d.key) // we need to cleanup secrets if mtls is disabled
	}
	if err := d.dataplaneReconciler.Reconcile(*envoyCtx, proxy); err != nil {
		return err
	}
	d.lastHash = meshCtx.Hash
	return nil
}

// syncIngress synces state of Ingress Dataplane. Notice that it does not use Mesh Hash yet because Ingress supports many Meshes.
func (d *DataplaneWatchdog) syncIngress() error {
	envoyCtx := &xds_context.Context{
		ControlPlane: d.envoyCpCtx,
		Mesh:         xds_context.MeshContext{}, // ZoneIngress does not need MeshContext
	}
	proxy, err := d.ingressProxyBuilder.build(d.key)
	if err != nil {
		return err
	}
	return d.ingressReconciler.Reconcile(*envoyCtx, proxy)
}
