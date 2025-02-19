package sync

import (
	"context"
	std_errors "errors"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	envoy_admin_tls "github.com/kumahq/kuma/pkg/envoy/admin/tls"
	util_tls "github.com/kumahq/kuma/pkg/tls"
	"github.com/kumahq/kuma/pkg/xds/cache/mesh"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
)

type DataplaneWatchdogDependencies struct {
	DataplaneProxyBuilder *DataplaneProxyBuilder
	DataplaneReconciler   SnapshotReconciler
	IngressProxyBuilder   *IngressProxyBuilder
	IngressReconciler     SnapshotReconciler
	EgressProxyBuilder    *EgressProxyBuilder
	EgressReconciler      SnapshotReconciler
	EnvoyCpCtx            *xds_context.ControlPlaneContext
	MeshCache             *mesh.Cache
	ResManager            core_manager.ReadOnlyResourceManager
}

type Status string

var (
	SkipStatus      Status = "skip"
	GeneratedStatus Status = "generated"
	ChangedStatus   Status = "changed"
)

type SyncResult struct {
	ProxyType mesh_proto.ProxyType
	Status    Status
}

type DataplaneWatchdog struct {
	DataplaneWatchdogDependencies
	key core_model.ResourceKey
	log logr.Logger

	// state of watchdog
	lastHash         string // last Mesh hash that was used to **successfully** generate Reconcile Envoy config
	dpType           mesh_proto.ProxyType
	proxyTypeSettled bool
	envoyAdminMTLS   *core_xds.ServerSideMTLSCerts
	dpAddress        string
}

func NewDataplaneWatchdog(l logr.Logger, deps DataplaneWatchdogDependencies, dpKey core_model.ResourceKey) *DataplaneWatchdog {
	return &DataplaneWatchdog{
		DataplaneWatchdogDependencies: deps,
		key:                           dpKey,
		log:                           l.WithName("xds"),
		proxyTypeSettled:              false,
	}
}

func (d *DataplaneWatchdog) Sync(ctx context.Context, metadata *core_xds.DataplaneMetadata) (SyncResult, error) {
	if d.dpType == "" {
		d.dpType = metadata.GetProxyType()
	}
	switch d.dpType {
	case mesh_proto.DataplaneProxyType:
		return d.syncDataplane(ctx, metadata)
	case mesh_proto.IngressProxyType:
		return d.syncIngress(ctx, metadata)
	case mesh_proto.EgressProxyType:
		return d.syncEgress(ctx, metadata)
	default:
		// It might be a case that dp type is not yet inferred because there is no Dataplane definition yet.
		return SyncResult{}, nil
	}
}

func (d *DataplaneWatchdog) Cleanup() error {
	proxyID := core_xds.FromResourceKey(d.key)
	switch d.dpType {
	case mesh_proto.DataplaneProxyType:
		d.EnvoyCpCtx.Secrets.Cleanup(mesh_proto.DataplaneProxyType, d.key)
		return d.DataplaneReconciler.Clear(&proxyID)
	case mesh_proto.IngressProxyType:
		return d.IngressReconciler.Clear(&proxyID)
	case mesh_proto.EgressProxyType:
		aggregatedMeshCtxs, aggregateMeshContextsErr := xds_context.AggregateMeshContexts(
			context.TODO(),
			d.ResManager,
			d.MeshCache.GetMeshContext,
		)
		for _, mesh := range aggregatedMeshCtxs.Meshes {
			d.EnvoyCpCtx.Secrets.Cleanup(
				mesh_proto.EgressProxyType,
				core_model.ResourceKey{Mesh: mesh.GetMeta().GetName(), Name: d.key.Name},
			)
		}
		return std_errors.Join(aggregateMeshContextsErr, d.EgressReconciler.Clear(&proxyID))
	default:
		return nil
	}
}

// syncDataplane syncs state of the Dataplane.
// It uses Mesh Hash to decide if we need to regenerate configuration or not.
func (d *DataplaneWatchdog) syncDataplane(ctx context.Context, metadata *core_xds.DataplaneMetadata) (SyncResult, error) {
	meshCtx, err := d.MeshCache.GetMeshContext(ctx, d.key.Mesh)
	if err != nil {
		return SyncResult{}, errors.Wrap(err, "could not get mesh context")
	}

	certInfo := d.EnvoyCpCtx.Secrets.Info(mesh_proto.DataplaneProxyType, d.key)
	syncForCert := certInfo != nil && certInfo.ExpiringSoon() // check if we need to regenerate config because identity cert is expiring soon.
	syncForConfig := meshCtx.Hash != d.lastHash               // check if we need to regenerate config because Kuma policies has changed.
	result := SyncResult{
		ProxyType: mesh_proto.DataplaneProxyType,
	}
	if !syncForCert && !syncForConfig {
		result.Status = SkipStatus
		return result, nil
	}
	if syncForConfig {
		d.log.V(1).Info("snapshot hash updated, reconcile", "prev", d.lastHash, "current", meshCtx.Hash)
	}
	if syncForCert {
		d.log.V(1).Info("certs expiring soon, reconcile")
	}

	envoyCtx := &xds_context.Context{
		ControlPlane: d.EnvoyCpCtx,
		Mesh:         meshCtx,
	}
	if _, found := meshCtx.DataplanesByName[d.key.Name]; !found {
		d.log.Info("Dataplane object not found. Can't regenerate XDS configuration. It's expected during Kubernetes namespace termination. " +
			"If it persists it's a bug.")
		result.Status = SkipStatus
		return result, nil
	}
	proxy, err := d.DataplaneProxyBuilder.Build(ctx, d.key, meshCtx)
	if err != nil {
		return SyncResult{}, errors.Wrap(err, "could not build dataplane proxy")
	}
	networking := proxy.Dataplane.Spec.Networking
	envoyAdminMTLS, err := d.getEnvoyAdminMTLS(ctx, networking.Address, networking.AdvertisedAddress)
	if err != nil {
		return SyncResult{}, errors.Wrap(err, "could not get Envoy Admin mTLS certs")
	}
	proxy.EnvoyAdminMTLSCerts = envoyAdminMTLS
	if !envoyCtx.Mesh.Resource.MTLSEnabled() {
		d.EnvoyCpCtx.Secrets.Cleanup(mesh_proto.DataplaneProxyType, d.key) // we need to cleanup secrets if mtls is disabled
	}
	proxy.Metadata = metadata
	changed, err := d.DataplaneReconciler.Reconcile(ctx, *envoyCtx, proxy)
	if err != nil {
		return SyncResult{}, errors.Wrap(err, "could not reconcile")
	}
	d.lastHash = meshCtx.Hash

	if changed {
		result.Status = ChangedStatus
	} else {
		result.Status = GeneratedStatus
	}
	return result, nil
}

// syncIngress synces state of Ingress Dataplane.
// It uses Mesh Hash to decide if we need to regenerate configuration or not.
func (d *DataplaneWatchdog) syncIngress(ctx context.Context, metadata *core_xds.DataplaneMetadata) (SyncResult, error) {
	envoyCtx := &xds_context.Context{
		ControlPlane: d.EnvoyCpCtx,
		Mesh:         xds_context.MeshContext{}, // ZoneIngress does not have a mesh!
	}

	aggregatedMeshCtxs, err := xds_context.AggregateMeshContexts(ctx, d.ResManager, d.MeshCache.GetMeshContext)
	if err != nil {
		return SyncResult{}, errors.Wrap(err, "could not aggregate mesh contexts")
	}

	result := SyncResult{
		ProxyType: mesh_proto.IngressProxyType,
	}
	syncForConfig := aggregatedMeshCtxs.Hash != d.lastHash
	var syncForCert bool
	for _, mesh := range aggregatedMeshCtxs.Meshes {
		certInfo := d.EnvoyCpCtx.Secrets.Info(
			mesh_proto.IngressProxyType,
			core_model.ResourceKey{Mesh: mesh.GetMeta().GetName(), Name: d.key.Name},
		)
		syncForCert = syncForCert || (certInfo != nil && certInfo.ExpiringSoon()) // check if we need to regenerate config because identity cert is expiring soon.
	}
	if !syncForConfig && !syncForCert {
		result.Status = SkipStatus
		return result, nil
	}

	d.lastHash = aggregatedMeshCtxs.Hash
	if syncForConfig {
		d.log.V(1).Info("snapshot hash updated, reconcile", "prev", d.lastHash, "current", aggregatedMeshCtxs.Hash)
	}
	if syncForCert {
		d.log.V(1).Info("certs expiring soon, reconcile")
	}

	proxy, err := d.IngressProxyBuilder.Build(ctx, d.key, aggregatedMeshCtxs)
	if err != nil {
		return SyncResult{}, errors.Wrap(err, "could not build ingress proxy")
	}
	networking := proxy.ZoneIngressProxy.ZoneIngressResource.Spec.GetNetworking()
	envoyAdminMTLS, err := d.getEnvoyAdminMTLS(ctx, networking.GetAddress(), networking.GetAdvertisedAddress())
	if err != nil {
		return SyncResult{}, errors.Wrap(err, "could not get Envoy Admin mTLS certs")
	}
	proxy.EnvoyAdminMTLSCerts = envoyAdminMTLS
	proxy.Metadata = metadata
	changed, err := d.IngressReconciler.Reconcile(ctx, *envoyCtx, proxy)
	if err != nil {
		return SyncResult{}, errors.Wrap(err, "could not reconcile")
	}
	if changed {
		result.Status = ChangedStatus
	} else {
		result.Status = GeneratedStatus
	}
	return result, nil
}

// syncEgress syncs state of Egress Dataplane.
// It uses Mesh Hash to decide if we need to regenerate configuration or not.
func (d *DataplaneWatchdog) syncEgress(ctx context.Context, metadata *core_xds.DataplaneMetadata) (SyncResult, error) {
	envoyCtx := &xds_context.Context{
		ControlPlane: d.EnvoyCpCtx,
		Mesh:         xds_context.MeshContext{}, // ZoneEgress does not have a mesh!
	}

	aggregatedMeshCtxs, err := xds_context.AggregateMeshContexts(ctx, d.ResManager, d.MeshCache.GetMeshContext)
	if err != nil {
		return SyncResult{}, errors.Wrap(err, "could not aggregate mesh contexts")
	}

	result := SyncResult{
		ProxyType: mesh_proto.EgressProxyType,
	}
	syncForConfig := aggregatedMeshCtxs.Hash != d.lastHash
	var syncForCert bool
	for _, mesh := range aggregatedMeshCtxs.Meshes {
		certInfo := d.EnvoyCpCtx.Secrets.Info(
			mesh_proto.EgressProxyType,
			core_model.ResourceKey{Mesh: mesh.GetMeta().GetName(), Name: d.key.Name},
		)
		syncForCert = syncForCert || (certInfo != nil && certInfo.ExpiringSoon()) // check if we need to regenerate config because identity cert is expiring soon.
	}
	if !syncForConfig && !syncForCert {
		result.Status = SkipStatus
		return result, nil
	}

	d.lastHash = aggregatedMeshCtxs.Hash
	if syncForConfig {
		d.log.V(1).Info("snapshot hash updated, reconcile", "prev", d.lastHash, "current", aggregatedMeshCtxs.Hash)
	}
	if syncForCert {
		d.log.V(1).Info("certs expiring soon, reconcile")
	}

	proxy, err := d.EgressProxyBuilder.Build(ctx, d.key, aggregatedMeshCtxs)
	if err != nil {
		return SyncResult{}, errors.Wrap(err, "could not build egress proxy")
	}
	networking := proxy.ZoneEgressProxy.ZoneEgressResource.Spec.Networking
	envoyAdminMTLS, err := d.getEnvoyAdminMTLS(ctx, networking.Address, "")
	if err != nil {
		return SyncResult{}, errors.Wrap(err, "could not get Envoy Admin mTLS certs")
	}
	proxy.EnvoyAdminMTLSCerts = envoyAdminMTLS
	proxy.Metadata = metadata
	changed, err := d.EgressReconciler.Reconcile(ctx, *envoyCtx, proxy)
	if err != nil {
		return SyncResult{}, errors.Wrap(err, "could not reconcile")
	}
	if changed {
		result.Status = ChangedStatus
	} else {
		result.Status = GeneratedStatus
	}
	return result, nil
}

func (d *DataplaneWatchdog) getEnvoyAdminMTLS(ctx context.Context, address string, advertisedAddress string) (core_xds.ServerSideMTLSCerts, error) {
	if d.envoyAdminMTLS == nil || d.dpAddress != address {
		ca, err := envoy_admin_tls.LoadCA(ctx, d.ResManager)
		if err != nil {
			return core_xds.ServerSideMTLSCerts{}, errors.Wrap(err, "could not load the CA")
		}
		caPair, err := util_tls.ToKeyPair(ca.PrivateKey, ca.Certificate[0])
		if err != nil {
			return core_xds.ServerSideMTLSCerts{}, err
		}
		ips := []string{address}
		if advertisedAddress != "" && advertisedAddress != address {
			ips = append(ips, advertisedAddress)
		}
		serverPair, err := envoy_admin_tls.GenerateServerCert(ca, ips...)
		if err != nil {
			return core_xds.ServerSideMTLSCerts{}, errors.Wrap(err, "could not generate server certificate")
		}

		envoyAdminMTLS := core_xds.ServerSideMTLSCerts{
			CaPEM:      caPair.CertPEM,
			ServerPair: serverPair,
		}
		// cache the Envoy Admin MTLS and dp address, so we
		// 1) don't have to do I/O on every sync
		// 2) have a stable certs = stable Envoy config
		// This means that if we want to change Envoy Admin CA, we need to restart all CP instances.
		// Additionally, we need to trigger cert generation when DP address has changed without DP reconnection.
		d.envoyAdminMTLS = &envoyAdminMTLS
		d.dpAddress = address
	}
	return *d.envoyAdminMTLS, nil
}
