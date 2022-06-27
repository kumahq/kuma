package dns_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	dns_server "github.com/kumahq/kuma/pkg/config/dns-server"
	config_manager "github.com/kumahq/kuma/pkg/core/config/manager"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	config_model "github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/dns"
	"github.com/kumahq/kuma/pkg/dns/vips"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
)

func dpWithTags(tags ...map[string]string) *mesh_proto.Dataplane {
	inbound := []*mesh_proto.Dataplane_Networking_Inbound{}
	for _, t := range tags {
		inbound = append(inbound, &mesh_proto.Dataplane_Networking_Inbound{
			Port: 8080,
			Tags: t,
		})
	}
	return &mesh_proto.Dataplane{
		Networking: &mesh_proto.Dataplane_Networking{
			Address: "127.0.0.1",
			Inbound: inbound,
		},
	}
}

func dp(services ...string) *mesh_proto.Dataplane {
	var tags []map[string]string
	for _, s := range services {
		tags = append(tags, map[string]string{mesh_proto.ServiceTag: s})
	}
	return dpWithTags(tags...)
}

type errConfigManager struct {
	config_manager.ConfigManager
}

func (e *errConfigManager) Update(ctx context.Context, r *config_model.ConfigResource, opts ...store.UpdateOptionsFunc) error {
	meshName, _ := vips.MeshFromConfigKey(r.GetMeta().GetName())
	return errors.Errorf("error during update, mesh = %s", meshName)
}

var testConfig = dns_server.Config{
	ServiceVipEnabled: true,
	CIDR:              "240.0.0.0/24",
	Domain:            "mesh",
}

var _ = Describe("VIP Allocator", func() {
	var rm manager.ResourceManager
	var cm config_manager.ConfigManager
	var allocator *dns.VIPsAllocator

	NoModifications := func(view *vips.VirtualOutboundMeshView) error {
		return nil
	}

	BeforeEach(func() {
		s := memory.NewStore()
		rm = manager.NewResourceManager(s)
		cm = config_manager.NewConfigManager(s)

		err := rm.Create(context.Background(), mesh.NewMeshResource(), store.CreateByKey("mesh-1", model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		err = rm.Create(context.Background(), mesh.NewMeshResource(), store.CreateByKey("mesh-2", model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		err = rm.Create(context.Background(), &mesh.DataplaneResource{Spec: dp("backend")}, store.CreateByKey("dp-1", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		err = rm.Create(context.Background(), &mesh.DataplaneResource{Spec: dp("frontend")}, store.CreateByKey("dp-2", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		err = rm.Create(context.Background(), &mesh.DataplaneResource{Spec: dp("web")}, store.CreateByKey("dp-3", "mesh-2"))
		Expect(err).ToNot(HaveOccurred())

		allocator, err = dns.NewVIPsAllocator(rm, cm, testConfig, "")
		Expect(err).ToNot(HaveOccurred())
	})

	It("should create VIP config for each mesh", func() {
		// when
		ctx := context.Background()
		err := allocator.CreateOrUpdateVIPConfigs(ctx)
		Expect(err).ToNot(HaveOccurred())

		persistence := vips.NewPersistence(rm, cm)

		// then
		vipList, err := persistence.GetByMesh(ctx, "mesh-1")
		Expect(err).ToNot(HaveOccurred())
		Expect(vipList.HostnameEntries()).To(HaveLen(2))

		vipList, err = persistence.GetByMesh(ctx, "mesh-2")
		Expect(err).ToNot(HaveOccurred())

		Expect(vipList.HostnameEntries()).To(HaveLen(1))
	})

	It("should respect already allocated VIPs in case of IPAM restarts", func() {
		// setup
		ctx := context.Background()
		persistence := vips.NewPersistence(rm, cm)
		vobv, err := vips.NewVirtualOutboundView(map[vips.HostnameEntry]vips.VirtualOutbound{
			vips.NewServiceEntry("frontend"): {Address: "240.0.0.0", Outbounds: []vips.OutboundEntry{{TagSet: map[string]string{mesh_proto.ServiceTag: "frontend"}}}},
			vips.NewServiceEntry("backend"):  {Address: "240.0.0.1", Outbounds: []vips.OutboundEntry{{TagSet: map[string]string{mesh_proto.ServiceTag: "backend"}}}},
		})
		Expect(err).ToNot(HaveOccurred())
		// we add VIPs directly to the 'persistence' object
		// that emulates situation when IPAM is fresh and doesn't aware of allocated VIPs
		err = persistence.Set(ctx, "mesh-1", vobv)
		Expect(err).ToNot(HaveOccurred())

		err = rm.Create(context.Background(), &mesh.DataplaneResource{Spec: dp("database")}, store.CreateByKey("dp-3", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		// when
		err = allocator.CreateOrUpdateVIPConfig(ctx, "mesh-1", NoModifications)
		Expect(err).ToNot(HaveOccurred())

		vipList, err := persistence.GetByMesh(ctx, "mesh-1")
		Expect(err).ToNot(HaveOccurred())
		// then
		expected, err := vips.NewVirtualOutboundView(map[vips.HostnameEntry]vips.VirtualOutbound{
			vips.NewServiceEntry("backend"):  {Address: "240.0.0.1", Outbounds: []vips.OutboundEntry{{TagSet: map[string]string{mesh_proto.ServiceTag: "backend"}}}},
			vips.NewServiceEntry("database"): {Address: "240.0.0.2", Outbounds: []vips.OutboundEntry{{TagSet: map[string]string{mesh_proto.ServiceTag: "database"}}}},
			vips.NewServiceEntry("frontend"): {Address: "240.0.0.0", Outbounds: []vips.OutboundEntry{{TagSet: map[string]string{mesh_proto.ServiceTag: "frontend"}}}},
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(vipList.HostnameEntries()).To(Equal(expected.HostnameEntries()))
		for _, k := range vipList.HostnameEntries() {
			Expect(vipList.Get(k).Address).To(Equal(expected.Get(k).Address))
		}
	})

	It("should return error if failed to update VIP config", func() {
		errConfigManager := &errConfigManager{ConfigManager: cm}
		ctx := context.Background()
		errAllocator, err := dns.NewVIPsAllocator(rm, errConfigManager, testConfig, "")
		Expect(err).ToNot(HaveOccurred())

		err = errAllocator.CreateOrUpdateVIPConfig(ctx, "mesh-1", NoModifications)
		Expect(err).ToNot(HaveOccurred())

		err = rm.Create(context.Background(), &mesh.DataplaneResource{Spec: dp("database")}, store.CreateByKey("dp-3", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		err = errAllocator.CreateOrUpdateVIPConfig(ctx, "mesh-1", NoModifications)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("error during update, mesh = mesh-1"))
	})

	It("should try to update all meshes and return combined error", func() {
		errConfigManager := &errConfigManager{ConfigManager: cm}
		errAllocator, err := dns.NewVIPsAllocator(rm, errConfigManager, testConfig, "")
		Expect(err).ToNot(HaveOccurred())

		err = errAllocator.CreateOrUpdateVIPConfigs(context.Background())
		Expect(err).ToNot(HaveOccurred())

		err = rm.Create(context.Background(), &mesh.DataplaneResource{Spec: dp("database")}, store.CreateByKey("dp-3", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		err = rm.Create(context.Background(), &mesh.DataplaneResource{Spec: dp("payment")}, store.CreateByKey("dp-4", "mesh-2"))
		Expect(err).ToNot(HaveOccurred())

		err = errAllocator.CreateOrUpdateVIPConfigs(context.Background())
		Expect(err).To(HaveOccurred())
		Expect(err).To(MatchError("error during update, mesh = mesh-1; error during update, mesh = mesh-2"))
	})

	It("will allocate the same VIPs for different services in multiple meshes", func() {
		ctx := context.Background()
		// given VIPs in mesh-1
		err := allocator.CreateOrUpdateVIPConfig(ctx, "mesh-1", NoModifications)
		Expect(err).ToNot(HaveOccurred())

		// when VIPs from other meshes are created
		err = allocator.CreateOrUpdateVIPConfig(ctx, "mesh-2", NoModifications)

		// then the addresses should not overlap
		Expect(err).ToNot(HaveOccurred())

		mesh1View, err := vips.NewPersistence(rm, cm).GetByMesh(ctx, "mesh-1")
		Expect(err).ToNot(HaveOccurred())
		mesh2View, err := vips.NewPersistence(rm, cm).GetByMesh(ctx, "mesh-2")
		Expect(err).ToNot(HaveOccurred())

		mesh1Address := mesh1View.Get(mesh1View.HostnameEntries()[0]).Address
		mesh2Address := mesh2View.Get(mesh2View.HostnameEntries()[0]).Address
		Expect(mesh1Address).To(Equal(mesh2Address))
	})
})

type outboundViewTestCase struct {
	givenResources      map[model.ResourceKey]model.Resource
	whenZone            string
	whenMesh            string
	whenSkipServiceVips bool
	thenHostnameEntries []vips.HostnameEntry
	thenOutbounds       map[vips.HostnameEntry][]vips.OutboundEntry
}

var _ = DescribeTable("outboundView",
	func(tc outboundViewTestCase) {
		// Given
		rm := manager.NewResourceManager(memory.NewStore())
		meshes := map[string]bool{}

		ctx := context.Background()
		for k, res := range tc.givenResources {
			if exists := meshes[k.Mesh]; !exists {
				Expect(rm.Create(ctx, mesh.NewMeshResource(), store.CreateBy(model.WithoutMesh(k.Mesh)))).ToNot(HaveOccurred())
				meshes[k.Mesh] = true
			}
			Expect(rm.Create(ctx, res, store.CreateBy(k))).ToNot(HaveOccurred())
		}

		cfg := dns_server.Config{
			Domain:            "mesh",
			CIDR:              "240.0.0.0/24",
			ServiceVipEnabled: !tc.whenSkipServiceVips,
		}
		// When
		allocator, err := dns.NewVIPsAllocator(rm, nil, cfg, tc.whenZone)
		Expect(err).NotTo(HaveOccurred())
		serviceSet, err := allocator.BuildVirtualOutboundMeshView(ctx, tc.whenMesh)

		// Then
		Expect(err).ToNot(HaveOccurred())
		Expect(serviceSet.HostnameEntries()).To(Equal(tc.thenHostnameEntries))
		for k, entries := range tc.thenOutbounds {
			entry := serviceSet.Get(k)
			Expect(entry).ToNot(BeNil(), "key:"+k.String())
			Expect(entry.Outbounds).To(Equal(entries), "key:"+k.String())
		}
	},
	Entry("no resource", outboundViewTestCase{whenMesh: "mesh", thenHostnameEntries: []vips.HostnameEntry{}}),
	Entry("dp with multiple services", outboundViewTestCase{
		givenResources: map[model.ResourceKey]model.Resource{
			model.WithMesh("mesh", "dp1"): &mesh.DataplaneResource{Spec: dp("service1", "service2")},
		},
		whenMesh:            "mesh",
		thenHostnameEntries: []vips.HostnameEntry{vips.NewServiceEntry("service1"), vips.NewServiceEntry("service2")},
		thenOutbounds: map[vips.HostnameEntry][]vips.OutboundEntry{
			vips.NewServiceEntry("service1"): {
				{TagSet: map[string]string{mesh_proto.ServiceTag: "service1"}, Origin: "service"},
			},
		},
	}),
	Entry("meshgateways of all meshes", outboundViewTestCase{
		givenResources: map[model.ResourceKey]model.Resource{
			model.WithMesh("mesh1", "gateway"): &mesh.MeshGatewayResource{
				Spec: &mesh_proto.MeshGateway{
					Conf: &mesh_proto.MeshGateway_Conf{
						Listeners: []*mesh_proto.MeshGateway_Listener{{
							Hostname: "gateway1.mesh",
							Port:     80,
							Protocol: mesh_proto.MeshGateway_Listener_HTTP,
							Tags: map[string]string{
								"listener": "internal",
							},
						}, {
							Hostname: "*",
							Port:     80,
							Protocol: mesh_proto.MeshGateway_Listener_HTTP,
							Tags: map[string]string{
								"listener": "wildcard",
							},
						}},
					},
					Selectors: []*mesh_proto.Selector{{
						Match: map[string]string{
							mesh_proto.ServiceTag: "gateway",
						},
					}},
					Tags: map[string]string{
						"gateway": "prod",
					},
				},
			},
			model.WithMesh("mesh2", "gateway"): &mesh.MeshGatewayResource{
				Spec: &mesh_proto.MeshGateway{
					Conf: &mesh_proto.MeshGateway_Conf{
						Listeners: []*mesh_proto.MeshGateway_Listener{{
							Hostname:  "gateway2.mesh",
							Port:      80,
							CrossMesh: true,
							Protocol:  mesh_proto.MeshGateway_Listener_HTTP,
							Tags: map[string]string{
								"listener": "internal",
							},
						}, {
							Port:      81,
							CrossMesh: true,
							Protocol:  mesh_proto.MeshGateway_Listener_HTTP,
							Tags: map[string]string{
								"listener": "internal2",
							},
						}, {
							Hostname: "*",
							Port:     80,
							Protocol: mesh_proto.MeshGateway_Listener_HTTP,
							Tags: map[string]string{
								"listener": "wildcard",
							},
						}},
					},
					Selectors: []*mesh_proto.Selector{{
						Match: map[string]string{
							mesh_proto.ServiceTag: "gateway",
						},
					}},
					Tags: map[string]string{
						"gateway": "prod",
					},
				},
			},
		},
		whenMesh:            "mesh1",
		thenHostnameEntries: []vips.HostnameEntry{vips.NewFqdnEntry("gateway2.mesh"), vips.NewFqdnEntry("internal.gateway.mesh2.mesh")},
		thenOutbounds: map[vips.HostnameEntry][]vips.OutboundEntry{
			vips.NewFqdnEntry("gateway2.mesh"): {{
				TagSet: map[string]string{
					"listener":            "internal",
					"gateway":             "prod",
					mesh_proto.ServiceTag: "gateway",
					"kuma.io/mesh":        "mesh2",
				},
				Origin: "mesh-gateway:mesh2:gateway:gateway2.mesh",
				Port:   80,
			}},
			vips.NewFqdnEntry("internal.gateway.mesh2.mesh"): {{
				TagSet: map[string]string{
					"listener":            "internal2",
					"gateway":             "prod",
					mesh_proto.ServiceTag: "gateway",
					"kuma.io/mesh":        "mesh2",
				},
				Origin: "mesh-gateway:mesh2:gateway:internal.gateway.mesh2.mesh",
				Port:   81,
			}},
		},
	}),
	Entry("external service", outboundViewTestCase{
		givenResources: map[model.ResourceKey]model.Resource{
			model.WithMesh("mesh", "es-1"): &mesh.ExternalServiceResource{
				Spec: &mesh_proto.ExternalService{
					Networking: &mesh_proto.ExternalService_Networking{
						Address: "external.service.com:8080",
					},
					Tags: map[string]string{
						mesh_proto.ServiceTag: "my-external-service-1",
					},
				},
			},
		},
		whenMesh:            "mesh",
		thenHostnameEntries: []vips.HostnameEntry{vips.NewServiceEntry("my-external-service-1"), vips.NewHostEntry("external.service.com")},
		thenOutbounds: map[vips.HostnameEntry][]vips.OutboundEntry{
			vips.NewServiceEntry("my-external-service-1"): {
				{TagSet: map[string]string{mesh_proto.ServiceTag: "my-external-service-1"}, Origin: "service", Port: 8080},
			},
			vips.NewHostEntry("external.service.com"): {
				{TagSet: map[string]string{mesh_proto.ServiceTag: "my-external-service-1"}, Origin: "host", Port: 8080},
			},
		},
	}),
	Entry("zone ingress", outboundViewTestCase{
		givenResources: map[model.ResourceKey]model.Resource{
			model.WithMesh("default", "ingress-1"): &mesh.ZoneIngressResource{
				Spec: &mesh_proto.ZoneIngress{
					Zone:       "zone2",
					Networking: &mesh_proto.ZoneIngress_Networking{Port: 1000, AdvertisedPort: 1000, AdvertisedAddress: "127.0.0.1", Address: "127.0.0.1"},
					AvailableServices: []*mesh_proto.ZoneIngress_AvailableService{
						{
							Mesh: "other-mesh",
							Tags: map[string]string{
								mesh_proto.ServiceTag: "srv1",
							},
							Instances: 2,
						},
						{
							Mesh: "mesh",
							Tags: map[string]string{
								mesh_proto.ServiceTag: "srv1",
							},
							Instances: 2,
						},
						{
							Mesh: "mesh",
							Tags: map[string]string{
								mesh_proto.ServiceTag: "srv2",
							},
							Instances: 2,
						},
					},
				},
			},
		},
		whenMesh:            "mesh",
		whenZone:            "zone1",
		thenHostnameEntries: []vips.HostnameEntry{vips.NewServiceEntry("srv1"), vips.NewServiceEntry("srv2")},
		thenOutbounds: map[vips.HostnameEntry][]vips.OutboundEntry{
			vips.NewServiceEntry("srv1"): {
				{TagSet: map[string]string{mesh_proto.ServiceTag: "srv1"}, Origin: "service"},
			},
		},
	}),
	Entry("zone ingress from own zone is ignored", outboundViewTestCase{
		givenResources: map[model.ResourceKey]model.Resource{
			model.WithMesh("mesh", "dp1"): &mesh.DataplaneResource{Spec: dp("service1", "service2")},
			model.WithMesh("default", "ingress-1"): &mesh.ZoneIngressResource{
				Spec: &mesh_proto.ZoneIngress{
					Zone:       "zone1",
					Networking: &mesh_proto.ZoneIngress_Networking{Port: 1000, AdvertisedPort: 1000, AdvertisedAddress: "127.0.0.1", Address: "127.0.0.1"},
					AvailableServices: []*mesh_proto.ZoneIngress_AvailableService{
						{
							Mesh: "other-mesh",
							Tags: map[string]string{
								mesh_proto.ServiceTag: "srv1",
							},
							Instances: 2,
						},
						{
							Mesh: "mesh",
							Tags: map[string]string{
								mesh_proto.ServiceTag: "srv1",
							},
							Instances: 2,
						},
						{
							Mesh: "mesh",
							Tags: map[string]string{
								mesh_proto.ServiceTag: "srv2",
							},
							Instances: 2,
						},
					},
				},
			},
		},
		whenMesh:            "mesh",
		whenZone:            "zone1",
		thenHostnameEntries: []vips.HostnameEntry{vips.NewServiceEntry("service1"), vips.NewServiceEntry("service2")},
		thenOutbounds: map[vips.HostnameEntry][]vips.OutboundEntry{
			vips.NewServiceEntry("service1"): {
				{TagSet: map[string]string{mesh_proto.ServiceTag: "service1"}, Origin: "service"},
			},
		},
	}),
	Entry("virtual outbound simple", outboundViewTestCase{
		givenResources: map[model.ResourceKey]model.Resource{
			model.WithMesh("mesh", "dp1-a"): &mesh.DataplaneResource{Spec: dpWithTags(map[string]string{mesh_proto.ServiceTag: "service1", "instance": "a", "port": "9000"})},
			model.WithMesh("mesh", "dp1-b"): &mesh.DataplaneResource{Spec: dpWithTags(map[string]string{mesh_proto.ServiceTag: "service1", "instance": "b"})},
			model.WithMesh("mesh", "dp2"):   &mesh.DataplaneResource{Spec: dp("service2")},
			model.WithMesh("mesh", "vob-1"): &mesh.VirtualOutboundResource{
				Spec: &mesh_proto.VirtualOutbound{
					Selectors: []*mesh_proto.Selector{
						{Match: map[string]string{mesh_proto.ServiceTag: "*", "instance": "*"}},
					},
					Conf: &mesh_proto.VirtualOutbound_Conf{
						Host: "{{.srv}}.{{.instance}}.mesh",
						Port: "{{if .port}}{{.port}}{{else}}8080{{end}}",
						Parameters: []*mesh_proto.VirtualOutbound_Conf_TemplateParameter{
							{Name: "srv", TagKey: mesh_proto.ServiceTag},
							{Name: "instance"},
							{Name: "port"},
						},
					},
				},
			},
		},
		whenMesh: "mesh",
		thenHostnameEntries: []vips.HostnameEntry{
			vips.NewServiceEntry("service1"),
			vips.NewServiceEntry("service2"),
			vips.NewFqdnEntry("service1.a.mesh"),
			vips.NewFqdnEntry("service1.b.mesh"),
		},
		thenOutbounds: map[vips.HostnameEntry][]vips.OutboundEntry{
			vips.NewFqdnEntry("service1.a.mesh"): {
				{Port: 9000, TagSet: map[string]string{mesh_proto.ServiceTag: "service1", "instance": "a", "port": "9000"}, Origin: "virtual-outbound:vob-1"},
			},
			vips.NewFqdnEntry("service1.b.mesh"): {
				{Port: 8080, TagSet: map[string]string{mesh_proto.ServiceTag: "service1", "instance": "b"}, Origin: "virtual-outbound:vob-1"},
			},
		},
	}),
	Entry("virtual outbound same hostname different ports", outboundViewTestCase{
		givenResources: map[model.ResourceKey]model.Resource{
			model.WithMesh("mesh", "dp1-a"): &mesh.DataplaneResource{Spec: dpWithTags(map[string]string{mesh_proto.ServiceTag: "service1", "port": "9000"})},
			model.WithMesh("mesh", "dp1-b"): &mesh.DataplaneResource{Spec: dpWithTags(map[string]string{mesh_proto.ServiceTag: "service1", "port": "8000"})},
			model.WithMesh("mesh", "dp2"):   &mesh.DataplaneResource{Spec: dpWithTags(map[string]string{mesh_proto.ServiceTag: "service2"})},
			model.WithMesh("mesh", "vob-1"): &mesh.VirtualOutboundResource{
				Spec: &mesh_proto.VirtualOutbound{
					Selectors: []*mesh_proto.Selector{
						{Match: map[string]string{mesh_proto.ServiceTag: "*"}},
					},
					Conf: &mesh_proto.VirtualOutbound_Conf{
						Host: "{{.srv}}.mesh",
						Port: "{{if .port}}{{.port}}{{else}}8080{{end}}",
						Parameters: []*mesh_proto.VirtualOutbound_Conf_TemplateParameter{
							{Name: "srv", TagKey: mesh_proto.ServiceTag},
							{Name: "port"},
						},
					},
				},
			},
		},
		whenMesh: "mesh",
		thenHostnameEntries: []vips.HostnameEntry{
			vips.NewServiceEntry("service1"),
			vips.NewServiceEntry("service2"),
			vips.NewFqdnEntry("service1.mesh"),
			vips.NewFqdnEntry("service2.mesh"),
		},
		thenOutbounds: map[vips.HostnameEntry][]vips.OutboundEntry{
			vips.NewFqdnEntry("service1.mesh"): {
				{Port: 8000, TagSet: map[string]string{mesh_proto.ServiceTag: "service1", "port": "8000"}, Origin: "virtual-outbound:vob-1"},
				{Port: 9000, TagSet: map[string]string{mesh_proto.ServiceTag: "service1", "port": "9000"}, Origin: "virtual-outbound:vob-1"},
			},
			vips.NewFqdnEntry("service2.mesh"): {
				{Port: 8080, TagSet: map[string]string{mesh_proto.ServiceTag: "service2"}, Origin: "virtual-outbound:vob-1"},
			},
		},
	}),
	Entry("virtual outbound collision, picks the most specific", outboundViewTestCase{
		givenResources: map[model.ResourceKey]model.Resource{
			model.WithMesh("mesh", "dp1"): &mesh.DataplaneResource{Spec: dpWithTags(map[string]string{mesh_proto.ServiceTag: "service1", "instance": "1"})},
			model.WithMesh("mesh", "vob-1"): &mesh.VirtualOutboundResource{
				Spec: &mesh_proto.VirtualOutbound{
					Selectors: []*mesh_proto.Selector{
						{Match: map[string]string{mesh_proto.ServiceTag: "*"}},
					},
					Conf: &mesh_proto.VirtualOutbound_Conf{
						Host: "{{.srv}}.mesh",
						Port: "8080",
						Parameters: []*mesh_proto.VirtualOutbound_Conf_TemplateParameter{
							{Name: "srv", TagKey: mesh_proto.ServiceTag},
						},
					},
				},
			},
			model.WithMesh("mesh", "vob-2"): &mesh.VirtualOutboundResource{
				Spec: &mesh_proto.VirtualOutbound{
					Selectors: []*mesh_proto.Selector{
						// High weight for this vob
						{Match: map[string]string{mesh_proto.ServiceTag: "*", "instance": "*"}},
					},
					Conf: &mesh_proto.VirtualOutbound_Conf{
						Host: "{{.srv}}.mesh",
						Port: "8080",
						Parameters: []*mesh_proto.VirtualOutbound_Conf_TemplateParameter{
							{Name: "srv", TagKey: mesh_proto.ServiceTag},
						},
					},
				},
			},
		},
		whenMesh: "mesh",
		thenHostnameEntries: []vips.HostnameEntry{
			vips.NewServiceEntry("service1"),
			vips.NewFqdnEntry("service1.mesh"),
		},
		thenOutbounds: map[vips.HostnameEntry][]vips.OutboundEntry{
			vips.NewFqdnEntry("service1.mesh"): {
				{Port: 8080, TagSet: map[string]string{mesh_proto.ServiceTag: "service1"}, Origin: "virtual-outbound:vob-2"},
			},
		},
	}),
	Entry("dp skip service vips", outboundViewTestCase{
		givenResources: map[model.ResourceKey]model.Resource{
			model.WithMesh("mesh", "dp1"): &mesh.DataplaneResource{Spec: dp("service1")},
			model.WithMesh("mesh", "es-1"): &mesh.ExternalServiceResource{
				Spec: &mesh_proto.ExternalService{
					Networking: &mesh_proto.ExternalService_Networking{
						Address: "external.service.com:8080",
					},
					Tags: map[string]string{
						mesh_proto.ServiceTag: "my-external-service-1",
					},
				},
			},
		},
		whenSkipServiceVips: true,
		whenMesh:            "mesh",
		thenHostnameEntries: []vips.HostnameEntry{vips.NewHostEntry("external.service.com")},
		thenOutbounds: map[vips.HostnameEntry][]vips.OutboundEntry{
			vips.NewHostEntry("external.service.com"): {
				{TagSet: map[string]string{mesh_proto.ServiceTag: "my-external-service-1"}, Origin: "host", Port: 8080},
			},
		},
	}),
)

var _ = Describe("AllocateVIPs", func() {
	It("should allocate new VIPs", func() {
		// setup
		gv, err := vips.NewGlobalView("240.0.0.0/4")
		Expect(err).ToNot(HaveOccurred())
		serviceSet := vips.NewEmptyVirtualOutboundView()
		Expect(serviceSet.Add(vips.NewServiceEntry("backend"), vips.OutboundEntry{TagSet: map[string]string{mesh_proto.ServiceTag: "backend"}})).ToNot(HaveOccurred())
		Expect(serviceSet.Add(vips.NewServiceEntry("frontend"), vips.OutboundEntry{TagSet: map[string]string{mesh_proto.ServiceTag: "frontend"}})).ToNot(HaveOccurred())
		// when
		err = dns.AllocateVIPs(gv, serviceSet)
		Expect(err).ToNot(HaveOccurred())
		// then
		Expect(serviceSet.HostnameEntries()).To(Equal([]vips.HostnameEntry{vips.NewServiceEntry("backend"), vips.NewServiceEntry("frontend")}))
		Expect(serviceSet.Get(vips.NewServiceEntry("backend")).Address).ToNot(BeEmpty())
		Expect(serviceSet.Get(vips.NewServiceEntry("frontend")).Address).ToNot(BeEmpty())
	})

	It("should generate the same VIP for services across meshes", func() {
		// setup
		gv, err := vips.NewGlobalView("240.0.0.0/4")
		Expect(err).ToNot(HaveOccurred())
		Expect(gv.Reserve(vips.NewServiceEntry("backend"), "240.0.0.0")).ToNot(HaveOccurred())
		Expect(gv.Reserve(vips.NewServiceEntry("frontend"), "240.0.0.1")).ToNot(HaveOccurred())
		Expect(gv.Reserve(vips.NewServiceEntry("database"), "240.0.0.10")).ToNot(HaveOccurred())
		Expect(err).ToNot(HaveOccurred())
		serviceSet := vips.NewEmptyVirtualOutboundView()
		Expect(serviceSet.Add(vips.NewServiceEntry("backend"), vips.OutboundEntry{Origin: "default", TagSet: map[string]string{mesh_proto.ServiceTag: "backend"}})).ToNot(HaveOccurred())
		Expect(serviceSet.Add(vips.NewServiceEntry("frontend"), vips.OutboundEntry{Origin: "default", TagSet: map[string]string{mesh_proto.ServiceTag: "frontend"}})).ToNot(HaveOccurred())
		Expect(serviceSet.Add(vips.NewServiceEntry("database"), vips.OutboundEntry{Origin: "default", TagSet: map[string]string{mesh_proto.ServiceTag: "database"}})).ToNot(HaveOccurred())
		// when
		err = dns.AllocateVIPs(gv, serviceSet)
		Expect(err).ToNot(HaveOccurred())
		// then
		Expect(serviceSet.HostnameEntries()).To(Equal([]vips.HostnameEntry{vips.NewServiceEntry("backend"), vips.NewServiceEntry("database"), vips.NewServiceEntry("frontend")}))
		Expect(serviceSet.Get(vips.NewServiceEntry("backend")).Address).To(Equal("240.0.0.0"))
		Expect(serviceSet.Get(vips.NewServiceEntry("frontend")).Address).To(Equal("240.0.0.1"))
		Expect(serviceSet.Get(vips.NewServiceEntry("database")).Address).To(Equal("240.0.0.10"))
	})
})
