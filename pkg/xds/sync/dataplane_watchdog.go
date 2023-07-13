package sync

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
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
	MetadataTracker       DataplaneMetadataTracker
	ResManager            core_manager.ReadOnlyResourceManager
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

func NewDataplaneWatchdog(deps DataplaneWatchdogDependencies, dpKey core_model.ResourceKey) *DataplaneWatchdog {
	return &DataplaneWatchdog{
		DataplaneWatchdogDependencies: deps,
		key:                           dpKey,
		log:                           core.Log.WithValues("key", dpKey),
		proxyTypeSettled:              false,
	}
}

func (d *DataplaneWatchdog) Sync(ctx context.Context) error {
	metadata := d.MetadataTracker.Metadata(d.key)
	if metadata == nil {
		return errors.New("metadata cannot be nil")
	}

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
		return nil
	}
}

func (d *DataplaneWatchdog) Cleanup() error {
	proxyID := core_xds.FromResourceKey(d.key)
	switch d.dpType {
	case mesh_proto.DataplaneProxyType:
		d.EnvoyCpCtx.Secrets.Cleanup(d.key)
		return d.DataplaneReconciler.Clear(&proxyID)
	case mesh_proto.IngressProxyType:
		return d.IngressReconciler.Clear(&proxyID)
	case mesh_proto.EgressProxyType:
		return d.EgressReconciler.Clear(&proxyID)
	default:
		return nil
	}
}

// syncDataplane syncs state of the Dataplane.
// It uses Mesh Hash to decide if we need to regenerate configuration or not.
func (d *DataplaneWatchdog) syncDataplane(ctx context.Context, metadata *core_xds.DataplaneMetadata) error {
	meshCtx, err := d.MeshCache.GetMeshContext(ctx, d.key.Mesh)
	if err != nil {
		return err
	}

	certInfo := d.EnvoyCpCtx.Secrets.Info(d.key)
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
		ControlPlane: d.EnvoyCpCtx,
		Mesh:         meshCtx,
	}
	proxy, err := d.DataplaneProxyBuilder.Build(ctx, d.key, meshCtx)
	if err != nil {
		return err
	}
	networking := proxy.Dataplane.Spec.Networking
	envoyAdminMTLS, err := d.getEnvoyAdminMTLS(ctx, networking.Address, networking.AdvertisedAddress)
	if err != nil {
		return errors.Wrap(err, "could not get Envoy Admin mTLS certs")
	}
	proxy.EnvoyAdminMTLSCerts = envoyAdminMTLS
	if !envoyCtx.Mesh.Resource.MTLSEnabled() {
		d.EnvoyCpCtx.Secrets.Cleanup(d.key) // we need to cleanup secrets if mtls is disabled
	}
	proxy.Metadata = metadata
	err = d.DataplaneReconciler.Reconcile(ctx, *envoyCtx, proxy)
	if err != nil {
		return err
	}
	d.lastHash = meshCtx.Hash
	return nil
}

// syncIngress synces state of Ingress Dataplane. Notice that it does not use Mesh Hash yet because Ingress supports many Meshes.
func (d *DataplaneWatchdog) syncIngress(ctx context.Context, metadata *core_xds.DataplaneMetadata) error {
	envoyCtx := &xds_context.Context{
		ControlPlane: d.EnvoyCpCtx,
		Mesh:         xds_context.MeshContext{}, // ZoneIngress does not have a mesh!
	}
	proxy, err := d.IngressProxyBuilder.Build(ctx, d.key)
	if err != nil {
		return err
	}
	networking := proxy.ZoneIngressProxy.ZoneIngressResource.Spec.GetNetworking()
	envoyAdminMTLS, err := d.getEnvoyAdminMTLS(ctx, networking.GetAddress(), networking.GetAdvertisedAddress())
	if err != nil {
		return errors.Wrap(err, "could not get Envoy Admin mTLS certs")
	}
	proxy.EnvoyAdminMTLSCerts = envoyAdminMTLS
	proxy.Metadata = metadata
	return d.IngressReconciler.Reconcile(ctx, *envoyCtx, proxy)
}

// syncEgress syncs state of Egress Dataplane. Notice that it does not use
// Mesh Hash yet because Egress supports many Meshes.
func (d *DataplaneWatchdog) syncEgress(ctx context.Context, metadata *core_xds.DataplaneMetadata) error {
	envoyCtx := &xds_context.Context{
		ControlPlane: d.EnvoyCpCtx,
		Mesh:         xds_context.MeshContext{}, // ZoneEgress does not have a mesh!
	}

	proxy, err := d.EgressProxyBuilder.Build(ctx, d.key)
	if err != nil {
		return err
	}
	networking := proxy.ZoneEgressProxy.ZoneEgressResource.Spec.Networking
	envoyAdminMTLS, err := d.getEnvoyAdminMTLS(ctx, networking.Address, "")
	if err != nil {
		return errors.Wrap(err, "could not get Envoy Admin mTLS certs")
	}
	proxy.EnvoyAdminMTLSCerts = envoyAdminMTLS
	proxy.Metadata = metadata
	return d.EgressReconciler.Reconcile(ctx, *envoyCtx, proxy)
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
