package sync_test

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	rest_types "github.com/kumahq/kuma/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/dns/vips"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/test/matchers"
	test_runtime "github.com/kumahq/kuma/pkg/test/runtime"
	"github.com/kumahq/kuma/pkg/xds/cache/mesh"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/server"
	"github.com/kumahq/kuma/pkg/xds/sync"
)

func initializeStore(ctx context.Context, resourceManager core_manager.ResourceManager, fileWithResourcesName string) {
	resourcePath := filepath.Join(
		"testdata", "input", fileWithResourcesName,
	)

	resourceBytes, err := os.ReadFile(resourcePath)
	Expect(err).ToNot(HaveOccurred())

	rawResources := strings.Split(string(resourceBytes), "---")
	for _, rawResource := range rawResources {
		bytes := []byte(rawResource)

		res, err := rest_types.YAML.UnmarshalCore(bytes)
		Expect(err).ToNot(HaveOccurred())

		switch res.Descriptor().Name {
		case core_mesh.ZoneEgressType:
			zoneEgress := res.(*core_mesh.ZoneEgressResource)
			Expect(resourceManager.Create(ctx, zoneEgress, store.CreateBy(core_model.MetaToResourceKey(zoneEgress.GetMeta())))).To(Succeed())
		case core_mesh.ZoneIngressType:
			zoneIngress := res.(*core_mesh.ZoneIngressResource)
			Expect(resourceManager.Create(ctx, zoneIngress, store.CreateBy(core_model.MetaToResourceKey(zoneIngress.GetMeta())))).To(Succeed())
		case core_mesh.DataplaneType:
			dataplane := res.(*core_mesh.DataplaneResource)
			Expect(resourceManager.Create(ctx, dataplane, store.CreateBy(core_model.MetaToResourceKey(dataplane.GetMeta())))).To(Succeed())
		case core_mesh.MeshType:
			meshResource := res.(*core_mesh.MeshResource)
			Expect(resourceManager.Create(ctx, meshResource, store.CreateBy(core_model.MetaToResourceKey(meshResource.GetMeta())))).To(Succeed())
		case core_mesh.ExternalServiceType:
			externalService := res.(*core_mesh.ExternalServiceResource)
			Expect(resourceManager.Create(ctx, externalService, store.CreateBy(core_model.MetaToResourceKey(externalService.GetMeta())))).To(Succeed())
		case core_mesh.MeshGatewayType:
			meshGateway := res.(*core_mesh.MeshGatewayResource)
			Expect(resourceManager.Create(ctx, meshGateway, store.CreateBy(core_model.MetaToResourceKey(meshGateway.GetMeta())))).To(Succeed())
		}
	}
}

var _ = Describe("Proxy Builder", func() {
	localZone := "zone-1"

	ctx := context.Background()
	config := kuma_cp.DefaultConfig()
	config.Multizone.Zone.Name = localZone
	builder, err := test_runtime.BuilderFor(ctx, config)
	Expect(err).ToNot(HaveOccurred())
	builder.WithLookupIP(func(s string) ([]net.IP, error) {
		if s == "one.one.one.one" {
			return []net.IP{net.ParseIP("1.1.1.1")}, nil
		}
		return nil, errors.New("No such host to resolve:" + s)
	})
	meshCtxBuilder := xds_context.NewMeshContextBuilder(
		builder.ReadOnlyResourceManager(),
		server.MeshResourceTypes(),
		builder.LookupIP(),
		builder.Config().Multizone.Zone.Name,
		vips.NewPersistence(builder.ReadOnlyResourceManager(), builder.ConfigManager(), false),
		builder.Config().DNSServer.Domain,
		builder.Config().DNSServer.ServiceVipPort,
		xds_context.AnyToAnyReachableServicesGraphBuilder,
	)
	metrics, err := core_metrics.NewMetrics("cache")
	Expect(err).ToNot(HaveOccurred())
	meshCache, err := mesh.NewCache(builder.Config().Store.Cache.ExpirationTime.Duration, meshCtxBuilder, metrics)
	Expect(err).ToNot(HaveOccurred())
	builder.WithMeshCache(meshCache)

	rt, err := builder.Build()
	Expect(err).ToNot(HaveOccurred())
	initializeStore(ctx, rt.ResourceManager(), "default_resources.yaml")

	Describe("Build() zone egress", func() {
		egressProxyBuilder := sync.DefaultEgressProxyBuilder(rt, envoy_common.APIV3)

		It("should build proxy object for egress", func() {
			// given
			rk := core_model.ResourceKey{Name: "zone-egress-1"}
			meshContexts, err := xds_context.AggregateMeshContexts(ctx, rt.ReadOnlyResourceManager(), meshCache.GetMeshContext)
			Expect(err).ToNot(HaveOccurred())

			// when
			proxy, err := egressProxyBuilder.Build(ctx, rk, meshContexts)
			Expect(err).ToNot(HaveOccurred())

			// then
			Expect(proxy.ZoneEgressProxy.ZoneEgressResource.Spec).To(
				matchers.MatchProto(&mesh_proto.ZoneEgress{
					Zone: "zone-1",
					Networking: &mesh_proto.ZoneEgress_Networking{
						Address: "1.1.1.1",
						Port:    10002,
					},
				}))
			Expect(proxy.ZoneEgressProxy.ZoneIngresses[0].Spec).To(
				matchers.MatchProto(&mesh_proto.ZoneIngress{
					AvailableServices: []*mesh_proto.ZoneIngress_AvailableService{
						{
							Tags: map[string]string{
								mesh_proto.ServiceTag: "service-in-zone-2",
							},
							Instances: 1,
							Mesh:      "default",
						},
						{
							Tags: map[string]string{
								mesh_proto.ServiceTag: "external-service-in-zone-2",
								mesh_proto.ZoneTag:    "zone-2",
							},
							Instances:       1,
							Mesh:            "default",
							ExternalService: true,
						},
					},
					Zone: "zone-2",
					Networking: &mesh_proto.ZoneIngress_Networking{
						Address:           "6.6.6.6",
						AdvertisedAddress: "7.7.7.7",
						AdvertisedPort:    20003,
						Port:              20002,
					},
				}))
			Expect(proxy.ZoneEgressProxy.ZoneIngresses[1].Spec).To(
				matchers.MatchProto(&mesh_proto.ZoneIngress{
					AvailableServices: []*mesh_proto.ZoneIngress_AvailableService{
						{
							Tags: map[string]string{
								mesh_proto.ServiceTag: "service-in-zone-2",
							},
							Instances: 1,
							Mesh:      "default",
						},
						{
							Tags: map[string]string{
								mesh_proto.ServiceTag: "external-service-in-zone-2",
								mesh_proto.ZoneTag:    "zone-2",
							},
							Instances:       1,
							Mesh:            "default",
							ExternalService: true,
						},
					},
					Zone: "zone-2",
					Networking: &mesh_proto.ZoneIngress_Networking{
						Address:           "6.6.6.7",
						AdvertisedAddress: "1.1.1.1",
						AdvertisedPort:    20003,
						Port:              20002,
					},
				}))
		})
	})

	Describe("Build() zone ingress", func() {
		ingressProxyBuilder := sync.DefaultIngressProxyBuilder(
			rt,
			envoy_common.APIV3,
		)

		It("should build proxy object for ingress", func() {
			// given
			rk := core_model.ResourceKey{Name: "zone-ingress-zone-1"}
			meshContexts, err := xds_context.AggregateMeshContexts(ctx, rt.ReadOnlyResourceManager(), meshCache.GetMeshContext)
			Expect(err).ToNot(HaveOccurred())

			// when
			proxy, err := ingressProxyBuilder.Build(ctx, rk, meshContexts)
			Expect(err).ToNot(HaveOccurred())

			// then
			Expect(proxy.ZoneIngressProxy.MeshResourceList).To(HaveLen(1))
			Expect(proxy.ZoneIngressProxy.MeshResourceList[0].EndpointMap).To(HaveKeyWithValue("cross-mesh-gateway", []core_xds.Endpoint{{
				Target: "192.168.0.3",
				Port:   8080,
				Tags: map[string]string{
					"kuma.io/service": "cross-mesh-gateway",
				},
				Weight: 1,
			}}))
			Expect(proxy.ZoneIngressProxy.ZoneIngressResource.Spec).To(PointTo(MatchFields(IgnoreExtras, Fields{
				"Zone": Equal("zone-1"),
				"Networking": BeComparableTo(&mesh_proto.ZoneIngress_Networking{
					Address:           "3.3.3.3",
					AdvertisedAddress: "4.4.4.4",
					AdvertisedPort:    30004,
					Port:              30003,
				}, cmp.Comparer(proto.Equal)),
			})))
		})
	})
})
