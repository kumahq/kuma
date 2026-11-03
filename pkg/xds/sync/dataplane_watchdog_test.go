package sync_test

import (
	"context"
	"maps"
	"net"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"k8s.io/client-go/util/cert"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	config_core "github.com/kumahq/kuma/v3/pkg/config/core"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	meshidentity_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshidentity/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshidentity/providers"
	"github.com/kumahq/kuma/v3/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/v3/pkg/core/resources/store"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	"github.com/kumahq/kuma/v3/pkg/core/xds/issuer"
	envoy_admin_tls "github.com/kumahq/kuma/v3/pkg/envoy/admin/tls"
	"github.com/kumahq/kuma/v3/pkg/events"
	"github.com/kumahq/kuma/v3/pkg/metrics"
	"github.com/kumahq/kuma/v3/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/v3/pkg/plugins/runtime/k8s/metadata"
	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
	"github.com/kumahq/kuma/v3/pkg/test/resources/samples"
	xds_auth "github.com/kumahq/kuma/v3/pkg/xds/auth"
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

	// A stream authenticates once, but the identity is issued whenever a
	// MeshIdentity selects the dataplane, possibly long after connect, so the
	// watchdog re-checks the stream's auth binding before issuing it.
	Context("workload identity auth binding", func() {
		var resKey core_model.ResourceKey
		var meta *core_xds.DataplaneMetadata

		BeforeEach(func() {
			Expect(samples.MeshDefaultBuilder().Create(resManager)).To(Succeed())
			// DataplaneBackendBuilder declares the kuma.io/workload: backend label.
			Expect(samples.DataplaneBackendBuilder().Create(resManager)).To(Succeed())
			// No SpiffeID: the default Universal /workload/{{ .Workload }} path.
			Expect(builders.MeshIdentity().Create(resManager)).To(Succeed())
			resKey = core_model.MetaToResourceKey(samples.DataplaneBackend().GetMeta())
			meta = &core_xds.DataplaneMetadata{ProxyType: mesh_proto.DataplaneProxyType}
			deps.Environment = config_core.UniversalEnvironment
		})

		DescribeTable("should issue a workload identity only when the binding vouches for the labels it reads",
			func(ctx context.Context, expectIdentity bool) {
				// given
				watchdog := sync.NewDataplaneWatchdog(deps, meta, resKey)

				// when
				_, err := watchdog.Sync(ctx)

				// then
				Expect(err).ToNot(HaveOccurred())
				Expect(snapshotReconciler.proxy).ToNot(BeNil())
				Expect(snapshotReconciler.proxy.WorkloadIdentityRequired).To(Equal(expectIdentity))
			},
			Entry("denies a binding that leaves the workload unbound",
				xds_auth.ContextWithBinding(context.Background(), xds_auth.Binding{}), false),
			Entry("denies a binding to a different workload",
				xds_auth.ContextWithBinding(context.Background(), xds_auth.Binding{Workload: "attacker"}), false),
			Entry("denies a stream without a binding",
				context.Background(), false),
			Entry("allows a binding to the dataplane workload",
				xds_auth.ContextWithBinding(context.Background(), xds_auth.Binding{Workload: "backend"}), true),
			Entry("allows a tag binding of the workload label",
				xds_auth.ContextWithBinding(context.Background(), xds_auth.Binding{
					Tags: mesh_proto.MultiValueTagSet{metadata.KumaWorkload: {"backend": true}},
				}), true),
			Entry("allows a binding that vouches for all labels",
				xds_auth.ContextWithBinding(context.Background(), xds_auth.Binding{AllLabels: true}), true),
		)

		It("should withdraw an issued identity once the dataplane's labels outgrow the binding", func() {
			// given an identity issued for the backend workload
			ctx := xds_auth.ContextWithBinding(context.Background(), xds_auth.Binding{Workload: "backend"})
			watchdog := sync.NewDataplaneWatchdog(deps, meta, resKey)
			_, err := watchdog.Sync(ctx)
			Expect(err).ToNot(HaveOccurred())
			Expect(snapshotReconciler.proxy.WorkloadIdentityRequired).To(BeTrue())

			// when the workload label changes after connect
			dp := core_mesh.NewDataplaneResource()
			Expect(resManager.Get(ctx, dp, core_store.GetBy(resKey))).To(Succeed())
			labels := maps.Clone(dp.GetMeta().GetLabels())
			labels[metadata.KumaWorkload] = "frontend"
			Expect(resManager.Update(ctx, dp, core_store.UpdateWithLabels(labels))).To(Succeed())
			time.Sleep(cacheExpirationTime)
			_, err = watchdog.Sync(ctx)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(snapshotReconciler.proxy.WorkloadIdentityRequired).To(BeFalse())
		})
	})
})
