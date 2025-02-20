package sync_test

import (
	"context"
	"net"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/util/cert"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	config_manager "github.com/kumahq/kuma/pkg/core/config/manager"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/dns/vips"
	envoy_admin_tls "github.com/kumahq/kuma/pkg/envoy/admin/tls"
	"github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
	"github.com/kumahq/kuma/pkg/xds/cache/mesh"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/secrets"
	"github.com/kumahq/kuma/pkg/xds/server"
	"github.com/kumahq/kuma/pkg/xds/sync"
)

type staticMetadataTracker struct {
	metadata *core_xds.DataplaneMetadata
}

var _ sync.DataplaneMetadataTracker = &staticMetadataTracker{}

func (s *staticMetadataTracker) Metadata(dpKey core_model.ResourceKey) *core_xds.DataplaneMetadata {
	return s.metadata
}

type staticSnapshotReconciler struct {
	proxy *core_xds.Proxy
}

func (s *staticSnapshotReconciler) Reconcile(_ context.Context, _ xds_context.Context, proxy *core_xds.Proxy) (bool, error) {
	s.proxy = proxy
	return true, nil
}

func (s *staticSnapshotReconciler) Clear(proxyId *core_xds.ProxyId) error {
	return nil
}

var _ sync.SnapshotReconciler = &staticSnapshotReconciler{}

var _ = Describe("Dataplane Watchdog", func() {
	const zone = ""
	const cacheExpirationTime = time.Millisecond

	var resManager manager.ResourceManager
	var snapshotReconciler *staticSnapshotReconciler
	var metadataTracker *staticMetadataTracker
	var deps sync.DataplaneWatchdogDependencies

	BeforeEach(func() {
		snapshotReconciler = &staticSnapshotReconciler{}
		metadataTracker = &staticMetadataTracker{}

		store := memory.NewStore()
		resManager = manager.NewResourceManager(store)
		meshContextBuilder := xds_context.NewMeshContextBuilder(
			resManager,
			server.MeshResourceTypes(),
			net.LookupIP,
			zone,
			vips.NewPersistence(resManager, config_manager.NewConfigManager(store), false),
			".mesh",
			80,
			xds_context.AnyToAnyReachableServicesGraphBuilder,
		)
		newMetrics, err := metrics.NewMetrics(zone)
		Expect(err).ToNot(HaveOccurred())
		cache, err := mesh.NewCache(cacheExpirationTime, meshContextBuilder, newMetrics)
		Expect(err).ToNot(HaveOccurred())

		secrets, err := secrets.NewSecrets(nil, nil, newMetrics) // nil is ok for now, because we don't use it
		Expect(err).ToNot(HaveOccurred())

		deps = sync.DataplaneWatchdogDependencies{
			DataplaneProxyBuilder: &sync.DataplaneProxyBuilder{
				APIVersion: envoy.APIV3,
				Zone:       zone,
			},
			DataplaneReconciler: snapshotReconciler,
			EnvoyCpCtx: &xds_context.ControlPlaneContext{
				Secrets: secrets,
				Zone:    zone,
			},
			MeshCache:       cache,
			MetadataTracker: metadataTracker,
			ResManager:      resManager,
		}

		pair, err := envoy_admin_tls.GenerateCA()
		Expect(err).ToNot(HaveOccurred())
		Expect(envoy_admin_tls.CreateCA(context.Background(), *pair, resManager)).To(Succeed())
	})

	Context("Dataplane", func() {
		var resKey core_model.ResourceKey
		var watchdog *sync.DataplaneWatchdog
		var ctx context.Context

		BeforeEach(func() {
			Expect(samples.MeshDefaultBuilder().Create(resManager)).To(Succeed())
			Expect(samples.DataplaneBackendBuilder().Create(resManager)).To(Succeed())
			resKey = core_model.MetaToResourceKey(samples.DataplaneBackend().GetMeta())

			metadataTracker.metadata = &core_xds.DataplaneMetadata{
				ProxyType: mesh_proto.DataplaneProxyType,
			}
			watchdog = sync.NewDataplaneWatchdog(deps, resKey)
			ctx = context.Background()
		})

		It("should reissue admin tls certificate when address has changed", func() {
			// when
			_, err := watchdog.Sync(ctx)

			// then
			Expect(err).ToNot(HaveOccurred())

			certs, err := cert.ParseCertsPEM(snapshotReconciler.proxy.EnvoyAdminMTLSCerts.ServerPair.CertPEM)
			Expect(err).ToNot(HaveOccurred())
			Expect(certs[0].IPAddresses).To(HaveLen(1))
			Expect(certs[0].IPAddresses[0].String()).To(Equal(samples.DataplaneBackend().Spec.Networking.Address))

			// when address has changed
			newAddress := "192.168.1.100"
			err = manager.Upsert(ctx, resManager, resKey, core_mesh.NewDataplaneResource(), func(resource core_model.Resource) error {
				resource.(*core_mesh.DataplaneResource).Spec.Networking.Address = newAddress
				return nil
			})
			Expect(err).ToNot(HaveOccurred())

			// and
			time.Sleep(cacheExpirationTime)
			_, err = watchdog.Sync(ctx)

			// then cert is reissued with a new address
			Expect(err).ToNot(HaveOccurred())

			certs, err = cert.ParseCertsPEM(snapshotReconciler.proxy.EnvoyAdminMTLSCerts.ServerPair.CertPEM)
			Expect(err).ToNot(HaveOccurred())
			Expect(certs[0].IPAddresses).To(HaveLen(1))
			Expect(certs[0].IPAddresses[0].String()).To(Equal(newAddress))
		})

		It("should not reconcile if mesh hash is the same", func() {
			// when
			_, err := watchdog.Sync(ctx)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(snapshotReconciler.proxy).ToNot(BeNil())

			// when
			snapshotReconciler.proxy = nil // set to nil so we can check if it was not called again
			time.Sleep(cacheExpirationTime)
			_, err = watchdog.Sync(ctx)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(snapshotReconciler.proxy).To(BeNil())
		})
	})
})
