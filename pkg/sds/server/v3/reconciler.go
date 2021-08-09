package v3

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"sync"
	"time"

	envoy_types "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/config/core/resources/store"
	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	sds_ca "github.com/kumahq/kuma/pkg/sds/ca"
	sds_identity "github.com/kumahq/kuma/pkg/sds/identity"
	sds_metrics "github.com/kumahq/kuma/pkg/sds/metrics"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_secrets "github.com/kumahq/kuma/pkg/xds/envoy/secrets/v3"
)

// DataplaneReconciler keeps the state of the Cache for SDS consistent
// When Dataplane connects to the Control Plane, the Watchdog (separate goroutine) is started which on the defined interval
// execute DataplaneReconciler#Reconcile. It will then check if certs needs to be regenerated because Mesh CA was changed
// This follows the same pattern as XDS.
type DataplaneReconciler struct {
	resManager         core_manager.ResourceManager
	readOnlyResManager core_manager.ReadOnlyResourceManager
	meshCaProvider     sds_ca.Provider
	identityProvider   sds_identity.Provider
	cache              envoy_cache.SnapshotCache
	upsertConfig       store.UpsertConfig
	sdsMetrics         *sds_metrics.Metrics

	sync.RWMutex
	// proxySnapshotInfo contains information about snapshot for every proxy
	// It is used to make a decision whether to regenerate certificate or not.
	// This can be kept in memory and not synced between instances of CP because the state of the stream is local to the control plane
	// When DP reconnects to the CP, snapshot will be regenerated anyways, because the stream is reinitialized.
	proxySnapshotInfo map[string]snapshotInfo
}

type snapshotInfo struct {
	tags       mesh_proto.MultiValueTagSet
	mtls       *mesh_proto.Mesh_Mtls
	expiration time.Time
	generation time.Time
}

func (d *DataplaneReconciler) Reconcile(proxyId *core_xds.ProxyId) error {
	dataplane := core_mesh.NewDataplaneResource()
	if err := d.readOnlyResManager.Get(context.Background(), dataplane, core_store.GetBy(proxyId.ToResourceKey())); err != nil {
		if core_store.IsResourceNotFound(err) {
			sdsServerLog.V(1).Info("Dataplane not found. Clearing the Snapshot.", "dataplaneId", proxyId.ToResourceKey())
			if err := d.Cleanup(proxyId); err != nil {
				return errors.Wrap(err, "could not cleanup snapshot")
			}
			return nil
		}
		return err
	}

	mesh := core_mesh.NewMeshResource()
	if err := d.readOnlyResManager.Get(context.Background(), mesh, core_store.GetByKey(dataplane.GetMeta().GetMesh(), core_model.NoMesh)); err != nil {
		return errors.Wrap(err, "could not retrieve a mesh")
	}

	if !mesh.MTLSEnabled() {
		sdsServerLog.V(1).Info("mTLS for Mesh disabled. Clearing the Snapshot.", "dataplaneId", proxyId.ToResourceKey())
		if err := d.Cleanup(proxyId); err != nil {
			return errors.Wrap(err, "could not cleanup snapshot")
		}
		return nil
	}

	generateSnapshot, reason, err := d.shouldGenerateSnapshot(proxyId.String(), mesh, dataplane)
	if err != nil {
		return err
	}

	if generateSnapshot {
		sdsServerLog.Info("Generating the Snapshot.", "dataplaneId", proxyId.ToResourceKey(), "reason", reason)
		snapshot, info, err := d.generateSnapshot(dataplane, mesh)
		if err != nil {
			return err
		}

		d.sdsMetrics.CertGenerations(envoy_common.APIV3).Inc()
		if err := d.cache.SetSnapshot(proxyId.String(), snapshot); err != nil {
			return err
		}
		d.setSnapshotInfo(proxyId.String(), info)

		if err := d.updateInsights(proxyId.ToResourceKey(), info); err != nil {
			// do not stop updating Envoy even if insights update fails
			sdsServerLog.Error(err, "Could not update Dataplane Insights", "dataplaneId", proxyId.ToResourceKey())
		}
	}
	return nil
}

func (d *DataplaneReconciler) Cleanup(proxyId *core_xds.ProxyId) error {
	d.cache.ClearSnapshot(proxyId.String())
	d.Lock()
	delete(d.proxySnapshotInfo, proxyId.String())
	d.Unlock()
	return nil
}

func (d *DataplaneReconciler) shouldGenerateSnapshot(proxyID string, mesh *core_mesh.MeshResource, dataplane *core_mesh.DataplaneResource) (bool, string, error) {
	_, err := d.cache.GetSnapshot(proxyID)
	if err != nil {
		return true, "Snapshot does not exist", nil
	}
	info := d.snapshotInfo(proxyID)
	if !proto.Equal(info.mtls, mesh.Spec.Mtls) {
		return true, "Mesh mTLS settings has changed", nil
	}
	if dataplane.Spec.TagSet().String() != info.tags.String() {
		return true, "Dataplane tags have changed", nil
	}

	// generate snapshot if cert expired
	lifetime := info.expiration.Sub(info.generation)
	if core.Now().After(info.generation.Add(lifetime / 5 * 4)) { // regenerate cert after 4/5 of its lifetime
		reason := fmt.Sprintf("Certificate generated at %s will expire in %s", info.generation, info.expiration.Sub(core.Now()))
		return true, reason, nil
	}
	return false, "", nil
}

func (d *DataplaneReconciler) snapshotInfo(proxyID string) snapshotInfo {
	d.RLock()
	defer d.RUnlock()
	return d.proxySnapshotInfo[proxyID]
}

func (d *DataplaneReconciler) setSnapshotInfo(proxyID string, info snapshotInfo) {
	d.Lock()
	defer d.Unlock()
	d.proxySnapshotInfo[proxyID] = info
}

func (d *DataplaneReconciler) generateSnapshot(dataplane *core_mesh.DataplaneResource, mesh *core_mesh.MeshResource) (envoy_cache.Snapshot, snapshotInfo, error) {
	requestor := sds_identity.Identity{
		Services: dataplane.Spec.TagSet(),
		Mesh:     dataplane.GetMeta().GetMesh(),
	}
	identitySecret, err := d.identityProvider.Get(context.Background(), requestor)
	if err != nil {
		return envoy_cache.Snapshot{}, snapshotInfo{}, errors.Wrap(err, "could not get Dataplane cert pair")
	}

	caSecret, err := d.meshCaProvider.Get(context.Background(), dataplane.GetMeta().GetMesh())
	if err != nil {
		return envoy_cache.Snapshot{}, snapshotInfo{}, errors.Wrap(err, "could not get mesh CA cert")
	}

	block, _ := pem.Decode(identitySecret.PemCerts[0])
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return envoy_cache.Snapshot{}, snapshotInfo{}, err
	}

	info := snapshotInfo{
		tags:       dataplane.Spec.TagSet(),
		mtls:       mesh.Spec.Mtls,
		expiration: cert.NotAfter,
		generation: core.Now(),
	}

	resources := envoy_cache.SnapshotResources{
		Secrets: []envoy_types.Resource{
			envoy_secrets.CreateIdentitySecret(identitySecret),
			envoy_secrets.CreateCaSecret(caSecret),
		},
	}
	snaphot := envoy_cache.NewSnapshotWithResources(core.NewUUID(), resources)
	return snaphot, info, nil
}

func (d *DataplaneReconciler) updateInsights(dataplaneId core_model.ResourceKey, info snapshotInfo) error {
	return core_manager.Upsert(d.resManager, dataplaneId, core_mesh.NewDataplaneInsightResource(), func(resource core_model.Resource) {
		insight := resource.(*core_mesh.DataplaneInsightResource)
		if err := insight.Spec.UpdateCert(core.Now(), info.expiration); err != nil {
			sdsServerLog.Error(err, "could not update the certificate", "dataplaneId", dataplaneId)
		}
	}, core_manager.WithConflictRetry(d.upsertConfig.ConflictRetryBaseBackoff, d.upsertConfig.ConflictRetryMaxTimes)) // retry because DataplaneInsight could be updated from other parts of the code
}
