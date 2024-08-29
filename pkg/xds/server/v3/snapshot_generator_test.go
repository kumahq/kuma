package v3_test

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"path/filepath"
	"time"

	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	config_manager "github.com/kumahq/kuma/pkg/core/config/manager"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	model "github.com/kumahq/kuma/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/pkg/core/xds/types"
	"github.com/kumahq/kuma/pkg/dns/vips"
	"github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
	"github.com/kumahq/kuma/pkg/test/xds"
	util_cache_v3 "github.com/kumahq/kuma/pkg/util/cache/v3"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/cache/cla"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/generator"
	xds_hooks "github.com/kumahq/kuma/pkg/xds/hooks"
	"github.com/kumahq/kuma/pkg/xds/server"
	v3 "github.com/kumahq/kuma/pkg/xds/server/v3"
	"github.com/kumahq/kuma/pkg/xds/sync"
	"github.com/kumahq/kuma/pkg/xds/template"
)

type staticClusterAddHook struct {
	name string
}

func (s *staticClusterAddHook) Modify(resourceSet *model.ResourceSet, ctx xds_context.Context, proxy *model.Proxy) error {
	resourceSet.Add(&model.Resource{
		Name: s.name,
		Resource: &envoy_cluster.Cluster{
			Name: s.name,
		},
	})
	return nil
}

type shuffleStore struct {
	r *rand.Rand
	core_store.ResourceStore
}

func (s *shuffleStore) List(ctx context.Context, rl core_model.ResourceList, opts ...core_store.ListOptionsFunc) error {
	newList, err := registry.Global().NewList(rl.GetItemType())
	if err != nil {
		return err
	}
	if err := s.ResourceStore.List(ctx, newList, opts...); err != nil {
		return err
	}
	resources := newList.GetItems()
	s.r.Shuffle(len(resources), func(i, j int) {
		resources[i], resources[j] = resources[j], resources[i]
	})
	for i := range resources {
		_ = rl.AddItem(resources[i])
	}
	return nil
}

var _ xds_hooks.ResourceSetHook = &staticClusterAddHook{}

var _ = Describe("GenerateSnapshot", func() {
	var store core_store.ResourceStore
	var gen *v3.TemplateSnapshotGenerator
	var proxyBuilder *sync.DataplaneProxyBuilder
	var mCtxBuilder xds_context.MeshContextBuilder
	// #nosec G404 -- used just for tests
	r := rand.New(rand.NewSource(GinkgoRandomSeed()))

	BeforeEach(func() {
		store = &shuffleStore{
			ResourceStore: memory.NewStore(), r: r,
		}
		store = core_store.NewPaginationStore(store)

		rm := manager.NewResourceManager(store)

		gen = &v3.TemplateSnapshotGenerator{
			ProxyTemplateResolver: template.SequentialResolver(
				&template.SimpleProxyTemplateResolver{
					ReadOnlyResourceManager: rm,
				},
				generator.DefaultTemplateResolver,
			),
		}

		cfg := kuma_cp.DefaultConfig()
		cfg.Multizone.Zone.Name = "eun-blue"
		mCtxBuilder = xds_context.NewMeshContextBuilder(
			manager.NewResourceManager(store),
			server.MeshResourceTypes(),
			net.LookupIP,
			cfg.Multizone.Zone.Name,
			vips.NewPersistence(rm, config_manager.NewConfigManager(store), false),
			cfg.DNSServer.Domain,
			cfg.DNSServer.ServiceVipPort,
			xds_context.AnyToAnyReachableServicesGraphBuilder,
			cfg.Experimental.SkipPersistedVIPs,
		)

		proxyBuilder = sync.DefaultDataplaneProxyBuilder(cfg, envoy_common.APIV3)
	})

	create := func(r core_model.Resource) {
		err := store.Create(context.Background(), r, core_store.CreateBy(core_model.MetaToResourceKey(r.GetMeta())))
		Expect(err).ToNot(HaveOccurred())
	}

	generateSnapshot := func(name, mesh string, outbounds xds_types.Outbounds) []byte {
		mCtx, err := mCtxBuilder.Build(context.Background(), mesh)
		Expect(err).ToNot(HaveOccurred())

		mCtx.VIPOutbounds = outbounds

		proxy, err := proxyBuilder.Build(context.Background(), core_model.ResourceKey{Name: name, Mesh: mesh}, mCtx)
		Expect(err).ToNot(HaveOccurred())

		metrics, err := metrics.NewMetrics("")
		Expect(err).ToNot(HaveOccurred())

		claCache, err := cla.NewCache(1*time.Second, metrics)
		Expect(err).ToNot(HaveOccurred())

		s, err := gen.GenerateSnapshot(
			context.Background(),
			xds_context.Context{
				ControlPlane: &xds_context.ControlPlaneContext{
					Secrets:  &xds.TestSecrets{},
					CLACache: claCache,
				},
				Mesh: mCtx,
			},
			proxy,
		)
		Expect(err).ToNot(HaveOccurred())

		resp, err := util_cache_v3.ToDeltaDiscoveryResponse(*s)
		Expect(err).ToNot(HaveOccurred())
		actual, err := util_proto.ToYAML(resp)
		Expect(err).ToNot(HaveOccurred())

		return actual
	}

	It("should execute hooks before proxy template modifications", func() {
		// given
		create(
			samples.MeshMTLSBuilder().
				WithName("demo").
				Build(),
		)

		create(
			builders.Dataplane().
				WithName("web1").
				WithMesh("demo").
				WithVersion("1").
				WithAddress("192.168.0.1").
				AddInbound(
					builders.Inbound().
						WithPort(80).
						WithServicePort(8080).
						WithService("backend-1"),
				).
				AddInbound(
					builders.Inbound().
						WithPort(443).
						WithServicePort(8443).
						WithService("backend-2"),
				).
				AddInbound(
					builders.Inbound().
						WithAddress("192.168.0.2").
						WithPort(80).
						WithServicePort(8080).
						WithService("backend-3"),
				).
				AddInbound(
					builders.Inbound().
						WithAddress("192.168.0.2").
						WithPort(443).
						WithServicePort(8443).
						WithService("backend-4"),
				).
				WithTransparentProxying(15001, 15006, "ipv4").
				Build(),
		)

		create(
			&core_mesh.ProxyTemplateResource{
				Meta: &test_model.ResourceMeta{Name: "pt", Mesh: "demo"},
				Spec: &mesh_proto.ProxyTemplate{
					Selectors: []*mesh_proto.Selector{
						{
							Match: map[string]string{
								mesh_proto.ServiceTag: mesh_proto.MatchAllTag,
							},
						},
					},
					Conf: &mesh_proto.ProxyTemplate_Conf{
						Imports: []string{core_mesh.ProfileDefaultProxy},
						Modifications: []*mesh_proto.ProxyTemplate_Modifications{
							{
								Type: &mesh_proto.ProxyTemplate_Modifications_Cluster_{
									Cluster: &mesh_proto.ProxyTemplate_Modifications_Cluster{
										Operation: mesh_proto.OpRemove,
										Match: &mesh_proto.ProxyTemplate_Modifications_Cluster_Match{
											Name: "to-be-removed",
										},
									},
								},
							},
						},
					},
				},
			},
		)

		gen.ResourceSetHooks = []xds_hooks.ResourceSetHook{
			&staticClusterAddHook{
				name: "to-be-removed",
			},
		}

		// when
		snapshot := generateSnapshot("web1", "demo", nil)

		// then
		Expect(snapshot).To(matchers.MatchGoldenYAML(filepath.Join("testdata", "hook-before-pt.golden.yaml")))
	})

	It("should generate stable envoy config for external services", func() {
		// given
		create(
			samples.MeshMTLSBuilder().
				WithName("demo").
				With(func(resource *core_mesh.MeshResource) {
					resource.Spec.Routing = &mesh_proto.Routing{
						LocalityAwareLoadBalancing: true,
					}
				}).
				Build(),
		)

		create(
			builders.Dataplane().
				WithName("web1").
				WithMesh("demo").
				WithVersion("1").
				WithAddress("192.168.0.1").
				AddInbound(
					builders.Inbound().
						WithPort(80).
						WithServicePort(8080).
						WithService("backend-1"),
				).
				AddInbound(
					builders.Inbound().
						WithPort(443).
						WithServicePort(8443).
						WithService("backend-2"),
				).
				AddInbound(
					builders.Inbound().
						WithAddress("192.168.0.2").
						WithPort(80).
						WithServicePort(8080).
						WithService("backend-3"),
				).
				AddInbound(
					builders.Inbound().
						WithAddress("192.168.0.2").
						WithPort(443).
						WithServicePort(8443).
						WithService("backend-4"),
				).
				WithTransparentProxying(15001, 15006, "ipv4").
				Build(),
		)

		create(
			&core_mesh.TrafficRouteResource{
				Meta: &test_model.ResourceMeta{Name: "tr", Mesh: "demo"},
				Spec: &mesh_proto.TrafficRoute{
					Sources: []*mesh_proto.Selector{{
						Match: mesh_proto.MatchAnyService(),
					}},
					Destinations: []*mesh_proto.Selector{{
						Match: mesh_proto.MatchAnyService(),
					}},
					Conf: &mesh_proto.TrafficRoute_Conf{
						Destination: mesh_proto.MatchAnyService(),
						LoadBalancer: &mesh_proto.TrafficRoute_LoadBalancer{
							LbType: &mesh_proto.TrafficRoute_LoadBalancer_RoundRobin_{},
						},
					},
				},
			},
		)

		create(
			&core_mesh.TrafficPermissionResource{
				Meta: &test_model.ResourceMeta{Name: "tp", Mesh: "demo"},
				Spec: &mesh_proto.TrafficPermission{
					Sources: []*mesh_proto.Selector{{
						Match: mesh_proto.MatchAnyService(),
					}},
					Destinations: []*mesh_proto.Selector{{
						Match: mesh_proto.MatchAnyService(),
					}},
				},
			},
		)

		const numOfExtSrvs = 4

		for i := 0; i < numOfExtSrvs; i++ {
			create(
				&core_mesh.ExternalServiceResource{
					Meta: &test_model.ResourceMeta{Name: fmt.Sprintf("es-%d", i), Mesh: "demo"},
					Spec: &mesh_proto.ExternalService{
						Networking: &mesh_proto.ExternalService_Networking{
							Address: fmt.Sprintf("hostname-%d.com:443", i),
							Tls:     &mesh_proto.ExternalService_Networking_TLS{Enabled: true},
						},
						Tags: map[string]string{
							mesh_proto.ServiceTag:  "es-with-tls",
							mesh_proto.ZoneTag:     fmt.Sprintf("zone-%d", numOfExtSrvs-i),
							mesh_proto.ProtocolTag: "http",
						},
					},
				},
			)
		}

		// when
		snapshot := generateSnapshot("web1", "demo", xds_types.Outbounds{
			{
				LegacyOutbound: &mesh_proto.Dataplane_Networking_Outbound{
					Port: builders.FirstOutboundPort,
					Tags: map[string]string{
						mesh_proto.ServiceTag: "es-with-tls",
					},
				},
			},
		})

		// then
		Expect(snapshot).To(matchers.MatchGoldenYAML(filepath.Join("testdata", "stable-es.golden.yaml")))
	})
})
