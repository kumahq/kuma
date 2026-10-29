package sync_test

import (
	"context"
	"net"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"k8s.io/client-go/util/cert"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	meshidentity_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshidentity/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshidentity/providers"
	"github.com/kumahq/kuma/v3/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	"github.com/kumahq/kuma/v3/pkg/core/xds/issuer"
	envoy_admin_tls "github.com/kumahq/kuma/v3/pkg/envoy/admin/tls"
	"github.com/kumahq/kuma/v3/pkg/events"
	"github.com/kumahq/kuma/v3/pkg/metrics"
	"github.com/kumahq/kuma/v3/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
	"github.com/kumahq/kuma/v3/pkg/test/resources/samples"
	"github.com/kumahq/kuma/v3/pkg/xds/cache/mesh"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
	"github.com/kumahq/kuma/v3/pkg/xds/envoy"
	xds_metrics "github.com/kumahq/kuma/v3/pkg/xds/metrics"
	"github.com/kumahq/kuma/v3/pkg/xds/server"
	"github.com/kumahq/kuma/v3/pkg/xds/sync"
)

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

type staticIdentityProvider struct {
	_ *core_xds.Proxy
}

func (s *staticIdentityProvider) Validate(_ context.Context, _ *meshidentity_api.MeshIdentityResource) error {
	return nil
}

func (s *staticIdentityProvider) Initialize(_ context.Context, _ *meshidentity_api.MeshIdentityResource) error {
	return nil
}

func (s *staticIdentityProvider) CreateIdentity(_ context.Context, _ *meshidentity_api.MeshIdentityResource, _ *core_xds.Proxy) (*core_xds.WorkloadIdentity, error) {
	return nil, nil
}

func (s *staticIdentityProvider) GetMeshTrustCA(_ context.Context, _ *meshidentity_api.MeshIdentityResource) ([]byte, error) {
	return nil, nil
}

var _ sync.SnapshotReconciler = &staticSnapshotReconciler{}

var _ = Describe("Dataplane Watchdog", func() {
	const zone = ""
	const cacheExpirationTime = time.Millisecond

	var resManager manager.ResourceManager
	var snapshotReconciler *staticSnapshotReconciler
	var deps sync.DataplaneWatchdogDependencies
	var xdsMetrics *xds_metrics.Metrics

	BeforeEach(func() {
		snapshotReconciler = &staticSnapshotReconciler{}

		store := memory.NewStore()
		resManager = manager.NewResourceManager(store)
		meshContextBuilder := xds_context.NewMeshContextBuilder(
			resManager,
			server.MeshResourceTypes(),
			net.LookupIP,
			zone,
		)
		newMetrics, err := metrics.NewMetrics(zone)
		Expect(err).ToNot(HaveOccurred())
		xdsMetrics, err = xds_metrics.NewMetrics(newMetrics)
		Expect(err).ToNot(HaveOccurred())
		cache, err := mesh.NewCache(cacheExpirationTime, meshContextBuilder, newMetrics)
		Expect(err).ToNot(HaveOccurred())
		eventBus, err := events.NewEventBus(10, newMetrics)
		Expect(err).ToNot(HaveOccurred())

		plugins := map[string]providers.IdentityProvider{
			"static": &staticIdentityProvider{},
		}

		deps = sync.DataplaneWatchdogDependencies{
			DataplaneProxyBuilder: &sync.DataplaneProxyBuilder{
				APIVersion: envoy.APIV3,
				Zone:       zone,
			},
			DataplaneReconciler: snapshotReconciler,
			EnvoyCpCtx: &xds_context.ControlPlaneContext{
				Zone:            zone,
				IdentityManager: providers.NewIdentityProviderManager(plugins, eventBus, issuer.Unlimited()),
			},
			MeshCache:  cache,
			ResManager: resManager,
			XdsMetrics: xdsMetrics,
		}

		pair, err := envoy_admin_tls.GenerateCA()
		Expect(err).ToNot(HaveOccurred())
		Expect(envoy_admin_tls.CreateCA(context.Background(), *pair, resManager)).To(Succeed())
	})

	Context("Dataplane", func() {
		var resKey core_model.ResourceKey
		var watchdog *sync.DataplaneWatchdog
		var ctx context.Context
		var metadata *core_xds.DataplaneMetadata

		BeforeEach(func() {
			Expect(samples.MeshDefaultBuilder().Create(resManager)).To(Succeed())
			Expect(samples.DataplaneBackendBuilder().Create(resManager)).To(Succeed())
			resKey = core_model.MetaToResourceKey(samples.DataplaneBackend().GetMeta())

			metadata = &core_xds.DataplaneMetadata{
				ProxyType: mesh_proto.DataplaneProxyType,
			}
			watchdog = sync.NewDataplaneWatchdog(deps, metadata, resKey)
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

		It("should count successful config-hash-triggered reconciles only", func() {
			// when
			_, err := watchdog.Sync(ctx)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(testutil.ToFloat64(xdsMetrics.DataplaneConfigRegenerated.WithLabelValues(resKey.Mesh))).To(Equal(1.0))

			// when
			time.Sleep(cacheExpirationTime)
			_, err = watchdog.Sync(ctx)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(testutil.ToFloat64(xdsMetrics.DataplaneConfigRegenerated.WithLabelValues(resKey.Mesh))).To(Equal(1.0))
		})
	})

	// A stream is authenticated once, but the identity is reissued whenever a
	// MeshIdentity selects the dataplane, which may happen long after connect.
	// The watchdog re-checks the token's workload binding at issuance so a
	// MeshIdentity applied to an already-connected proxy can't mint the victim's
	// identity from a self-asserted kuma.io/workload label (MESH-15).
	Context("workload identity token binding", func() {
		var resKey core_model.ResourceKey
		var meta *core_xds.DataplaneMetadata
		var tracker core_xds.ConnectionInfoTracker
		var ctx context.Context

		BeforeEach(func() {
			Expect(samples.MeshDefaultBuilder().Create(resManager)).To(Succeed())
			// DataplaneBackendBuilder declares the kuma.io/workload: backend label.
			Expect(samples.DataplaneBackendBuilder().Create(resManager)).To(Succeed())
			Expect(builders.MeshIdentity().Create(resManager)).To(Succeed())
			resKey = core_model.MetaToResourceKey(samples.DataplaneBackend().GetMeta())
			meta = &core_xds.DataplaneMetadata{ProxyType: mesh_proto.DataplaneProxyType}
			tracker = core_xds.NewConnectionInfoTracker()
			deps.ConnectionInfoTracker = tracker
			ctx = context.Background()
		})

		DescribeTable("should issue a workload identity only for a matching token binding",
			func(store func(core_xds.ConnectionInfoTracker), expectIdentity bool) {
				// given
				store(tracker)
				watchdog := sync.NewDataplaneWatchdog(deps, meta, resKey)

				// when
				_, err := watchdog.Sync(ctx)

				// then
				Expect(err).ToNot(HaveOccurred())
				Expect(snapshotReconciler.proxy).ToNot(BeNil())
				Expect(snapshotReconciler.proxy.WorkloadIdentityRequired).To(Equal(expectIdentity))
			},
			Entry("denies an unbound token selected by a MeshIdentity after connect", func(t core_xds.ConnectionInfoTracker) {
				t.Store(resKey, core_xds.ConnectionInfo{TokenWorkload: ""})
			}, false),
			Entry("denies a token bound to a different workload", func(t core_xds.ConnectionInfoTracker) {
				t.Store(resKey, core_xds.ConnectionInfo{TokenWorkload: "attacker"})
			}, false),
			Entry("allows a token bound to the dataplane workload", func(t core_xds.ConnectionInfoTracker) {
				t.Store(resKey, core_xds.ConnectionInfo{TokenWorkload: "backend"})
			}, true),
			Entry("allows when no dataplane-token binding was recorded", func(t core_xds.ConnectionInfoTracker) {}, true),
		)
	})
})
