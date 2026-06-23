package v3

import (
	"context"
	"strconv"
	"testing"

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

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	xds_model "github.com/kumahq/kuma/v3/pkg/core/xds"
	core_metrics "github.com/kumahq/kuma/v3/pkg/metrics"
	test_model "github.com/kumahq/kuma/v3/pkg/test/resources/model"
	"github.com/kumahq/kuma/v3/pkg/util/proto"
	util_xds "github.com/kumahq/kuma/v3/pkg/util/xds"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
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
				generator: snapshotGeneratorFunc(func(_ context.Context, ctx xds_context.Context, proxy *xds_model.Proxy) (*envoy_cache.Snapshot, error) {
					snap := <-snapshots
					return &snap, nil
				}),
				cacher:         &simpleSnapshotCacher{xdsContext.Hasher(), xdsContext.Cache()},
				statsCallbacks: statsCallbacks,
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

			By("verifying that snapshot versions are empty for empty snapshot")
			// when
			snapshot, err = xdsContext.Cache().GetSnapshot("demo.example")
			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(snapshot).ToNot(BeZero())
			// and: empty snapshot produces empty ("") versions, not new UUIDs
			Expect(snapshot.GetVersion(resource.ListenerType)).To(BeEmpty())
			Expect(snapshot.GetVersion(resource.RouteType)).To(BeEmpty())
			Expect(snapshot.GetVersion(resource.ClusterType)).To(BeEmpty())
			Expect(snapshot.GetVersion(resource.EndpointType)).To(BeEmpty())
			Expect(snapshot.GetVersion(resource.SecretType)).To(BeEmpty())

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

		It("should force EDS re-push when cluster changes", func() {
			// Both clusters share the same name so endpoints stay consistent across
			// reconciles. AltStatName differs so the cluster hash changes without
			// affecting EDS reference resolution (which uses cluster Name).
			makeCluster := func(altStatName string) *envoy_cluster.Cluster {
				return &envoy_cluster.Cluster{
					Name:                 "cluster",
					AltStatName:          altStatName,
					ClusterDiscoveryType: &envoy_cluster.Cluster_Type{Type: envoy_cluster.Cluster_EDS},
					EdsClusterConfig: &envoy_cluster.Cluster_EdsClusterConfig{
						EdsConfig: &envoy_core.ConfigSource{
							ResourceApiVersion: envoy_core.ApiVersion_V3,
							ConfigSourceSpecifier: &envoy_core.ConfigSource_Ads{
								Ads: &envoy_core.AggregatedConfigSource{},
							},
						},
					},
				}
			}
			cluster1 := makeCluster("v1")
			cluster2 := makeCluster("v2")
			endpoint := &envoy_endpoint.ClusterLoadAssignment{ClusterName: "cluster"}

			makeSnap := func(cluster *envoy_cluster.Cluster) envoy_cache.Snapshot {
				snap := envoy_cache.Snapshot{}
				snap.Resources[envoy_types.Cluster] = envoy_cache.Resources{
					Items: map[string]envoy_types.ResourceWithTTL{
						cluster.Name: {Resource: cluster},
					},
				}
				snap.Resources[envoy_types.Endpoint] = envoy_cache.Resources{
					Items: map[string]envoy_types.ResourceWithTTL{
						endpoint.ClusterName: {Resource: endpoint},
					},
				}
				return snap
			}

			snapshots := make(chan envoy_cache.Snapshot, 3)
			snapshots <- makeSnap(cluster1)
			snapshots <- makeSnap(cluster2)
			snapshots <- makeSnap(cluster1)

			metrics, err := core_metrics.NewMetrics("Zone")
			Expect(err).ToNot(HaveOccurred())
			statsCallbacks, err := util_xds.NewStatsCallbacks(metrics, "xds", util_xds.NoopVersionExtractor)
			Expect(err).ToNot(HaveOccurred())

			r := &reconciler{
				generator: snapshotGeneratorFunc(func(_ context.Context, ctx xds_context.Context, proxy *xds_model.Proxy) (*envoy_cache.Snapshot, error) {
					snap := <-snapshots
					return &snap, nil
				}),
				cacher:         &simpleSnapshotCacher{xdsContext.Hasher(), xdsContext.Cache()},
				statsCallbacks: statsCallbacks,
			}

			proxy := &xds_model.Proxy{
				Id: *xds_model.BuildProxyId("demo", "eds-fold-test"),
				Dataplane: &core_mesh.DataplaneResource{
					Meta: &test_model.ResourceMeta{Mesh: "demo", Name: "eds-fold-test"},
					Spec: &mesh_proto.Dataplane{},
				},
			}

			By("first reconcile — populate initial versions")
			changed, err := r.Reconcile(context.Background(), xds_context.Context{}, proxy)
			Expect(err).ToNot(HaveOccurred())
			Expect(changed).To(BeTrue())
			cached, err := xdsContext.Cache().GetSnapshot("demo.eds-fold-test")
			Expect(err).ToNot(HaveOccurred())
			clusterV1 := cached.GetVersion(resource.ClusterType)
			endpointV1 := cached.GetVersion(resource.EndpointType)
			Expect(clusterV1).ToNot(BeEmpty())
			Expect(endpointV1).ToNot(BeEmpty())

			By("second reconcile — different cluster, identical endpoint content")
			changed, err = r.Reconcile(context.Background(), xds_context.Context{}, proxy)
			Expect(err).ToNot(HaveOccurred())
			Expect(changed).To(BeTrue())
			cached, err = xdsContext.Cache().GetSnapshot("demo.eds-fold-test")
			Expect(err).ToNot(HaveOccurred())
			clusterV2 := cached.GetVersion(resource.ClusterType)
			endpointV2 := cached.GetVersion(resource.EndpointType)
			// Cluster changed → endpoint version must change too (EDS warming fold).
			Expect(clusterV2).ToNot(Equal(clusterV1))
			Expect(endpointV2).ToNot(Equal(endpointV1))

			By("third reconcile — back to original cluster")
			changed, err = r.Reconcile(context.Background(), xds_context.Context{}, proxy)
			Expect(err).ToNot(HaveOccurred())
			Expect(changed).To(BeTrue())
			cached, err = xdsContext.Cache().GetSnapshot("demo.eds-fold-test")
			Expect(err).ToNot(HaveOccurred())
			// Deterministic: same content as first reconcile → same hashes.
			Expect(cached.GetVersion(resource.ClusterType)).To(Equal(clusterV1))
			Expect(cached.GetVersion(resource.EndpointType)).To(Equal(endpointV1))
		})

		It("should leave empty resource types unversioned", func() {
			// Only Secret is populated; all other types are empty.
			// Secret has no cross-reference requirements so the snapshot is consistent.
			onlySecrets := envoy_cache.Snapshot{}
			onlySecrets.Resources[envoy_types.Secret] = envoy_cache.Resources{
				Items: map[string]envoy_types.ResourceWithTTL{
					"secret": {Resource: &envoy_auth.Secret{Name: "secret"}},
				},
			}

			snapshots := make(chan envoy_cache.Snapshot, 2)
			snapshots <- onlySecrets
			snapshots <- onlySecrets

			metrics, err := core_metrics.NewMetrics("Zone")
			Expect(err).ToNot(HaveOccurred())
			statsCallbacks, err := util_xds.NewStatsCallbacks(metrics, "xds", util_xds.NoopVersionExtractor)
			Expect(err).ToNot(HaveOccurred())

			r := &reconciler{
				generator: snapshotGeneratorFunc(func(_ context.Context, ctx xds_context.Context, proxy *xds_model.Proxy) (*envoy_cache.Snapshot, error) {
					snap := <-snapshots
					return &snap, nil
				}),
				cacher:         &simpleSnapshotCacher{xdsContext.Hasher(), xdsContext.Cache()},
				statsCallbacks: statsCallbacks,
			}

			proxy := &xds_model.Proxy{
				Id: *xds_model.BuildProxyId("demo", "empty-slot-test"),
				Dataplane: &core_mesh.DataplaneResource{
					Meta: &test_model.ResourceMeta{Mesh: "demo", Name: "empty-slot-test"},
					Spec: &mesh_proto.Dataplane{},
				},
			}

			By("first reconcile")
			changed, err := r.Reconcile(context.Background(), xds_context.Context{}, proxy)
			Expect(err).ToNot(HaveOccurred())
			Expect(changed).To(BeTrue())

			cached, err := xdsContext.Cache().GetSnapshot("demo.empty-slot-test")
			Expect(err).ToNot(HaveOccurred())
			// Populated type has a content hash.
			Expect(cached.GetVersion(resource.SecretType)).ToNot(BeEmpty())
			// Empty types produce "" version — not versioned.
			Expect(cached.GetVersion(resource.ListenerType)).To(BeEmpty())
			Expect(cached.GetVersion(resource.RouteType)).To(BeEmpty())
			Expect(cached.GetVersion(resource.ClusterType)).To(BeEmpty())
			Expect(cached.GetVersion(resource.EndpointType)).To(BeEmpty())

			By("second identical reconcile — no changes")
			changed, err = r.Reconcile(context.Background(), xds_context.Context{}, proxy)
			Expect(err).ToNot(HaveOccurred())
			// Unchanged: same content hash for Route, "" == "" for empty types.
			Expect(changed).To(BeFalse())
		})

		It("should not pass empty version to ConfigReadyForDelivery when types clear", func() {
			// First snapshot has a Secret; second is fully empty.
			// When Secret transitions secretHash→"", the "" must NOT reach
			// ConfigReadyForDelivery: stats_callbacks.OnStreamRequest skips
			// requests with VersionInfo=="" so the key would never be consumed.
			withSecret := envoy_cache.Snapshot{}
			withSecret.Resources[envoy_types.Secret] = envoy_cache.Resources{
				Items: map[string]envoy_types.ResourceWithTTL{
					"secret": {Resource: &envoy_auth.Secret{Name: "secret"}},
				},
			}
			emptySnap := envoy_cache.Snapshot{}

			snapshots := make(chan envoy_cache.Snapshot, 2)
			snapshots <- withSecret
			snapshots <- emptySnap

			spy := &spyStatsCallbacks{}
			r := &reconciler{
				generator: snapshotGeneratorFunc(func(_ context.Context, ctx xds_context.Context, proxy *xds_model.Proxy) (*envoy_cache.Snapshot, error) {
					snap := <-snapshots
					return &snap, nil
				}),
				cacher:         &simpleSnapshotCacher{xdsContext.Hasher(), xdsContext.Cache()},
				statsCallbacks: spy,
			}

			proxy := &xds_model.Proxy{
				Id: *xds_model.BuildProxyId("demo", "empty-delivery-test"),
				Dataplane: &core_mesh.DataplaneResource{
					Meta: &test_model.ResourceMeta{Mesh: "demo", Name: "empty-delivery-test"},
					Spec: &mesh_proto.Dataplane{},
				},
			}

			By("first reconcile — Secret gets a real content hash")
			_, err := r.Reconcile(context.Background(), xds_context.Context{}, proxy)
			Expect(err).ToNot(HaveOccurred())

			By("second reconcile — Secret clears, transitioning secretHash→\"\"")
			_, err = r.Reconcile(context.Background(), xds_context.Context{}, proxy)
			Expect(err).ToNot(HaveOccurred())

			Expect(spy.deliveredVersions).ToNot(ContainElement(""))
		})
	})
})

// spyStatsCallbacks records versions passed to ConfigReadyForDelivery.
type spyStatsCallbacks struct {
	util_xds.NoopCallbacks
	deliveredVersions []string
}

func (s *spyStatsCallbacks) ConfigReadyForDelivery(v string) {
	s.deliveredVersions = append(s.deliveredVersions, v)
}

func (s *spyStatsCallbacks) DiscardConfig(string) {}

type snapshotGeneratorFunc func(context.Context, xds_context.Context, *xds_model.Proxy) (*envoy_cache.Snapshot, error)

func (f snapshotGeneratorFunc) GenerateSnapshot(ctx context.Context, xdsCtx xds_context.Context, proxy *xds_model.Proxy) (*envoy_cache.Snapshot, error) {
	return f(ctx, xdsCtx, proxy)
}

// buildRealisticSnapshot returns a snapshot with 20 clusters, 20 endpoints,
// 5 listeners, and 5 routes — representative of a non-trivial data plane.
func buildRealisticSnapshot() *envoy_cache.Snapshot {
	snap := &envoy_cache.Snapshot{}

	clusterItems := make(map[string]envoy_types.ResourceWithTTL, 20)
	endpointItems := make(map[string]envoy_types.ResourceWithTTL, 20)
	for i := range 20 {
		name := "cluster-" + strconv.Itoa(i)
		clusterItems[name] = envoy_types.ResourceWithTTL{
			Resource: &envoy_cluster.Cluster{
				Name:                 name,
				ClusterDiscoveryType: &envoy_cluster.Cluster_Type{Type: envoy_cluster.Cluster_EDS},
			},
		}
		endpointItems[name] = envoy_types.ResourceWithTTL{
			Resource: &envoy_endpoint.ClusterLoadAssignment{ClusterName: name},
		}
	}
	snap.Resources[envoy_types.Cluster] = envoy_cache.Resources{Items: clusterItems}
	snap.Resources[envoy_types.Endpoint] = envoy_cache.Resources{Items: endpointItems}

	// Listeners with static (non-RDS) route configs so no Route resources are
	// required for Consistent() to pass.
	listenerItems := make(map[string]envoy_types.ResourceWithTTL, 5)
	for i := range 5 {
		lName := "listener-" + strconv.Itoa(i)
		listenerItems[lName] = envoy_types.ResourceWithTTL{
			Resource: &envoy_listener.Listener{Name: lName},
		}
	}
	snap.Resources[envoy_types.Listener] = envoy_cache.Resources{Items: listenerItems}

	return snap
}

func BenchmarkAutoVersion(b *testing.B) {
	nodeId := "demo.benchmark"
	snap := buildRealisticSnapshot()

	// Populate an "old" snapshot with content hashes.
	populated, _, err := autoVersion(nodeId, &envoy_cache.Snapshot{}, snap)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Each iteration mirrors the real hot path: generator allocates a new
		// snapshot, autoVersion hashes it and compares to the cached version.
		n := buildRealisticSnapshot()
		if _, _, err := autoVersion(nodeId, populated, n); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkReconcileUnchanged(b *testing.B) {
	metrics, err := core_metrics.NewMetrics("Zone")
	if err != nil {
		b.Fatal(err)
	}
	statsCallbacks, err := util_xds.NewStatsCallbacks(metrics, "xds", util_xds.NoopVersionExtractor)
	if err != nil {
		b.Fatal(err)
	}

	xdsCtx := NewXdsContext()
	r := &reconciler{
		generator: snapshotGeneratorFunc(func(_ context.Context, ctx xds_context.Context, proxy *xds_model.Proxy) (*envoy_cache.Snapshot, error) {
			return buildRealisticSnapshot(), nil
		}),
		cacher:         &simpleSnapshotCacher{xdsCtx.Hasher(), xdsCtx.Cache()},
		statsCallbacks: statsCallbacks,
	}

	proxy := &xds_model.Proxy{
		Id: *xds_model.BuildProxyId("demo", "benchmark"),
		Dataplane: &core_mesh.DataplaneResource{
			Meta: &test_model.ResourceMeta{Mesh: "demo", Name: "benchmark"},
			Spec: &mesh_proto.Dataplane{},
		},
	}

	// Populate cache so the benchmark measures the no-change steady-state path.
	if _, err := r.Reconcile(context.Background(), xds_context.Context{}, proxy); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := r.Reconcile(context.Background(), xds_context.Context{}, proxy); err != nil {
			b.Fatal(err)
		}
	}
}
