package sync

import (
	"context"
	"encoding/base64"
	std_errors "errors"
	"hash/fnv"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	meshidentity_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshidentity/api/v1alpha1"
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
	xdsMeta          *core_xds.DataplaneMetadata
	// used by MeshIdentity
	workloadIdentity *core_xds.WorkloadIdentity
	lastIdentityHash string // last Hash of MeshIdentities
}

func NewDataplaneWatchdog(deps DataplaneWatchdogDependencies, meta *core_xds.DataplaneMetadata, dpKey core_model.ResourceKey) *DataplaneWatchdog {
	return &DataplaneWatchdog{
		DataplaneWatchdogDependencies: deps,
		key:                           dpKey,
		log:                           core.Log.WithName("xds").WithValues("key", dpKey),
		proxyTypeSettled:              false,
		xdsMeta:                       meta,
	}
}

func (d *DataplaneWatchdog) Sync(ctx context.Context) (SyncResult, error) {
	if d.dpType == "" {
		d.dpType = d.xdsMeta.GetProxyType()
	}
	switch d.dpType {
	case mesh_proto.DataplaneProxyType:
		return d.syncDataplane(ctx)
	case mesh_proto.IngressProxyType:
		return d.syncIngress(ctx)
	case mesh_proto.EgressProxyType:
		return d.syncEgress(ctx)
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
func (d *DataplaneWatchdog) syncDataplane(ctx context.Context) (SyncResult, error) {
	meshCtx, err := d.MeshCache.GetMeshContext(ctx, d.key.Mesh)
	if err != nil {
		return SyncResult{}, errors.Wrap(err, "could not get mesh context")
	}
	result := SyncResult{
		ProxyType: mesh_proto.DataplaneProxyType,
	}

	var dpp *core_mesh.DataplaneResource
	var found bool
	if dpp, found = meshCtx.DataplanesByName[d.key.Name]; !found {
		d.log.Info("Dataplane object not found. Can't regenerate XDS configuration. It's expected during Kubernetes namespace termination. " +
			"If it persists it's a bug.")
		result.Status = SkipStatus
		return result, nil
	}

	certInfo := d.EnvoyCpCtx.Secrets.Info(mesh_proto.DataplaneProxyType, d.key)
	syncForCert := certInfo != nil && certInfo.ExpiringSoon() // check if we need to regenerate config because identity cert is expiring soon.
	syncForConfig := meshCtx.Hash != d.lastHash               // check if we need to regenerate config because Kuma policies has changed.
	identity := d.EnvoyCpCtx.IdentityManager.SelectedIdentity(dpp, meshCtx.Resources.MeshIdentities().Items)
	identityHash := base64.StdEncoding.EncodeToString(hashMeshIdentity(identity))
	syncIdentity := identityHash != d.lastIdentityHash ||
		// check if is expired
		(d.workloadIdentity != nil && d.workloadIdentity.ManagementMode == core_xds.KumaManagementMode && d.workloadIdentity.ExpiringSoon()) ||
		(d.workloadIdentity != nil && identity == nil) // check if someone changed identity and it doesn't target the dpp anymore
	if !syncForCert && !syncForConfig && !syncIdentity {
		result.Status = SkipStatus
		return result, nil
	}
	if syncForConfig {
		d.log.V(1).Info("snapshot hash updated, reconcile", "prev", d.lastHash, "current", meshCtx.Hash)
	}
	if syncForCert {
		d.log.V(1).Info("certs expiring soon, reconcile")
	}
	if syncIdentity {
		d.log.V(1).Info("config generation based on identity change, reconcile")
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
	proxy, err := d.DataplaneProxyBuilder.Build(ctx, d.key, d.xdsMeta, meshCtx)
	if err != nil {
		return SyncResult{}, errors.Wrap(err, "could not build dataplane proxy")
	}
	if syncIdentity {
		identity, err := d.EnvoyCpCtx.IdentityManager.GetWorkloadIdentity(ctx, proxy, identity)
		if err != nil {
			return SyncResult{}, errors.Wrap(err, "could not get identity")
		}
		d.workloadIdentity = identity
	}
	if d.workloadIdentity != nil {
		proxy.WorkloadIdentity = d.workloadIdentity
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
	changed, err := d.DataplaneReconciler.Reconcile(ctx, *envoyCtx, proxy)
	if err != nil {
		return SyncResult{}, errors.Wrap(err, "could not reconcile")
	}
	d.lastHash = meshCtx.Hash
	d.lastIdentityHash = identityHash

	if changed {
		result.Status = ChangedStatus
	} else {
		result.Status = GeneratedStatus
	}
	return result, nil
}

func hashMeshIdentity(identity *meshidentity_api.MeshIdentityResource) []byte {
	hasher := fnv.New128a()
	if identity != nil {
		_, _ = hasher.Write(core_model.Hash(identity))
	}
	return hasher.Sum(nil)
}

// syncIngress synces state of Ingress Dataplane.
// It uses Mesh Hash to decide if we need to regenerate configuration or not.
func (d *DataplaneWatchdog) syncIngress(ctx context.Context) (SyncResult, error) {
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

	proxy, err := d.IngressProxyBuilder.Build(ctx, d.key, d.xdsMeta, aggregatedMeshCtxs)
	if err != nil {
		return SyncResult{}, errors.Wrap(err, "could not build ingress proxy")
	}
	networking := proxy.ZoneIngressProxy.ZoneIngressResource.Spec.GetNetworking()
	envoyAdminMTLS, err := d.getEnvoyAdminMTLS(ctx, networking.GetAddress(), networking.GetAdvertisedAddress())
	if err != nil {
		return SyncResult{}, errors.Wrap(err, "could not get Envoy Admin mTLS certs")
	}
	proxy.EnvoyAdminMTLSCerts = envoyAdminMTLS
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
func (d *DataplaneWatchdog) syncEgress(ctx context.Context) (SyncResult, error) {
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

	proxy, err := d.EgressProxyBuilder.Build(ctx, d.key, d.xdsMeta, aggregatedMeshCtxs)
	if err != nil {
		return SyncResult{}, errors.Wrap(err, "could not build egress proxy")
	}
	networking := proxy.ZoneEgressProxy.ZoneEgressResource.Spec.Networking
	envoyAdminMTLS, err := d.getEnvoyAdminMTLS(ctx, networking.Address, "")
	if err != nil {
		return SyncResult{}, errors.Wrap(err, "could not get Envoy Admin mTLS certs")
	}
	proxy.EnvoyAdminMTLSCerts = envoyAdminMTLS
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
