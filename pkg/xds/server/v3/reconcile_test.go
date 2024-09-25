package v3

import (
	"context"
	"sync"

	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	envoy_auth "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	envoy_types "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/anypb"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	xds_model "github.com/kumahq/kuma/pkg/core/xds"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	"github.com/kumahq/kuma/pkg/util/proto"
	util_xds "github.com/kumahq/kuma/pkg/util/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
)

func hcmForRoute(routeName string) *anypb.Any {
	hcm := envoy_hcm.HttpConnectionManager{
		RouteSpecifier: &envoy_hcm.HttpConnectionManager_Rds{
			Rds: &envoy_hcm.Rds{
				RouteConfigName: routeName,
			},
		},
	}

	a, err := proto.MarshalAnyDeterministic(&hcm)
	Expect(err).To(Succeed())

	return a
}

var _ = Describe("Reconcile", func() {
	Describe("reconciler", func() {
		var xdsContext XdsContext

		BeforeEach(func() {
			xdsContext = NewXdsContext()
		})

		snapshot := envoy_cache.Snapshot{
			Resources: [envoy_types.UnknownType]envoy_cache.Resources{
				envoy_types.Listener: {
					Items: map[string]envoy_types.ResourceWithTTL{
						"listener": {
							Resource: &envoy_listener.Listener{
								Address: &envoy_core.Address{
									Address: &envoy_core.Address_SocketAddress{
										SocketAddress: &envoy_core.SocketAddress{
											Address: "127.0.0.1",
											PortSpecifier: &envoy_core.SocketAddress_PortValue{
												PortValue: 99,
											},
										},
									},
								},
								FilterChains: []*envoy_listener.FilterChain{{
									Filters: []*envoy_listener.Filter{{
										Name: "envoy.filters.network.http_connection_manager",
										ConfigType: &envoy_listener.Filter_TypedConfig{
											TypedConfig: hcmForRoute("route"),
										},
									}},
								}},
							},
						},
					},
				},
				envoy_types.Route: {
					Items: map[string]envoy_types.ResourceWithTTL{
						"route": {
							Resource: &envoy_route.RouteConfiguration{},
						},
					},
				},
				envoy_types.Cluster: {
					Items: map[string]envoy_types.ResourceWithTTL{
						"cluster": {
							Resource: &envoy_cluster.Cluster{
								Name:                 "cluster",
								ClusterDiscoveryType: &envoy_cluster.Cluster_Type{Type: envoy_cluster.Cluster_EDS},
								EdsClusterConfig: &envoy_cluster.Cluster_EdsClusterConfig{
									EdsConfig: &envoy_core.ConfigSource{
										ResourceApiVersion: envoy_core.ApiVersion_V3,
										ConfigSourceSpecifier: &envoy_core.ConfigSource_Ads{
											Ads: &envoy_core.AggregatedConfigSource{},
										},
									},
								},
							},
						},
					},
				},
				envoy_types.Endpoint: {
					Items: map[string]envoy_types.ResourceWithTTL{
						"cluster": {
							Resource: &envoy_endpoint.ClusterLoadAssignment{ClusterName: "cluster"},
						},
					},
				},
				envoy_types.Secret: {
					Items: map[string]envoy_types.ResourceWithTTL{
						"secret": {
							Resource: &envoy_auth.Secret{},
						},
					},
				},
			},
		}

		It("should generate a Snapshot per Envoy Node", func() {
			// given
			snapshots := make(chan envoy_cache.Snapshot, 3)
			snapshots <- snapshot               // initial Dataplane configuration
			snapshots <- snapshot               // same Dataplane configuration
			snapshots <- envoy_cache.Snapshot{} // new Dataplane configuration

			metrics, err := core_metrics.NewMetrics("Zone")
			Expect(err).ToNot(HaveOccurred())
			statsCallbacks, err := util_xds.NewStatsCallbacks(metrics, "xds", util_xds.NoopVersionExtractor)
			Expect(err).ToNot(HaveOccurred())

			// setup
			r := &reconciler{
				snapshotGeneratorFunc(func(_ context.Context, ctx xds_context.Context, proxy *xds_model.Proxy) (*envoy_cache.Snapshot, error) {
					snap := <-snapshots
					return &snap, nil
				}),
				&simpleSnapshotCacher{xdsContext.Hasher(), xdsContext.Cache()},
				statsCallbacks,
				&sync.Mutex{},
			}

			// given
			dataplane := &core_mesh.DataplaneResource{
				Meta: &test_model.ResourceMeta{
					Mesh:    "demo",
					Name:    "example",
					Version: "abcdefg",
				},
				Spec: &mesh_proto.Dataplane{},
			}

			By("simulating discovery event")
			// when
			proxy := &xds_model.Proxy{
				Id:        *xds_model.BuildProxyId("demo", "example"),
				Dataplane: dataplane,
			}
			changed, err := r.Reconcile(context.Background(), xds_context.Context{}, proxy)
			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(changed).To(BeTrue())
			Expect(snapshot.Resources[envoy_types.Listener].Version).To(BeEmpty())
			Expect(snapshot.Resources[envoy_types.Route].Version).To(BeEmpty())
			Expect(snapshot.Resources[envoy_types.Cluster].Version).To(BeEmpty())
			Expect(snapshot.Resources[envoy_types.Endpoint].Version).To(BeEmpty())
			Expect(snapshot.Resources[envoy_types.Secret].Version).To(BeEmpty())

			By("verifying that snapshot versions were auto-generated")
			// when
			snapshot, err := xdsContext.Cache().GetSnapshot("demo.example")
			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(snapshot).ToNot(BeZero())
			// and
			listenerV1 := snapshot.GetVersion(resource.ListenerType)
			routeV1 := snapshot.GetVersion(resource.RouteType)
			clusterV1 := snapshot.GetVersion(resource.ClusterType)
			endpointV1 := snapshot.GetVersion(resource.EndpointType)
			secretV1 := snapshot.GetVersion(resource.SecretType)
			Expect(listenerV1).ToNot(BeEmpty())
			Expect(routeV1).ToNot(BeEmpty())
			Expect(clusterV1).ToNot(BeEmpty())
			Expect(endpointV1).ToNot(BeEmpty())
			Expect(secretV1).ToNot(BeEmpty())

			By("simulating discovery event (Dataplane watchdog triggers refresh)")
			// when
			changed, err = r.Reconcile(context.Background(), xds_context.Context{}, proxy)
			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(changed).To(BeFalse())

			By("verifying that snapshot versions remain the same")
			// when
			snapshot, err = xdsContext.Cache().GetSnapshot("demo.example")
			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(snapshot).ToNot(BeZero())
			// and
			Expect(snapshot.GetVersion(resource.ListenerType)).To(Equal(listenerV1))
			Expect(snapshot.GetVersion(resource.RouteType)).To(Equal(routeV1))
			Expect(snapshot.GetVersion(resource.ClusterType)).To(Equal(clusterV1))
			Expect(snapshot.GetVersion(resource.EndpointType)).To(Equal(endpointV1))
			Expect(snapshot.GetVersion(resource.SecretType)).To(Equal(secretV1))

			By("simulating discovery event (Dataplane gets changed)")
			// when
			_, err = r.Reconcile(context.Background(), xds_context.Context{}, proxy)
			// then
			Expect(err).ToNot(HaveOccurred())

			By("verifying that snapshot versions are new")
			// when
			snapshot, err = xdsContext.Cache().GetSnapshot("demo.example")
			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(snapshot).ToNot(BeZero())
			// and
			Expect(snapshot.GetVersion(resource.ListenerType)).To(SatisfyAll(
				Not(Equal(listenerV1)),
				Not(BeEmpty()),
			))
			Expect(snapshot.GetVersion(resource.RouteType)).To(SatisfyAll(
				Not(Equal(routeV1)),
				Not(BeEmpty()),
			))
			Expect(snapshot.GetVersion(resource.ClusterType)).To(Equal(clusterV1))
			Expect(snapshot.GetVersion(resource.EndpointType)).To(Equal(endpointV1))
			Expect(snapshot.GetVersion(resource.SecretType)).To(SatisfyAll(
				Not(Equal(secretV1)),
				Not(BeEmpty()),
			))

			By("simulating clear")
			// when
			err = r.Clear(&proxy.Id)
			Expect(err).ToNot(HaveOccurred())
			snapshot, err = xdsContext.Cache().GetSnapshot("demo.example")

			// then
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no snapshot found"))
			Expect(snapshot).To(BeNil())
		})
	})
})

type snapshotGeneratorFunc func(context.Context, xds_context.Context, *xds_model.Proxy) (*envoy_cache.Snapshot, error)

func (f snapshotGeneratorFunc) GenerateSnapshot(ctx context.Context, xdsCtx xds_context.Context, proxy *xds_model.Proxy) (*envoy_cache.Snapshot, error) {
	return f(ctx, xdsCtx, proxy)
}
