package dns_test

import (
	"fmt"
	"strconv"

	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/util/proto"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/dns/vips"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/dns"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
)

var _ = Describe("VIPOutbounds", func() {

	It("should update outbounds", func() {
		dataplane := &core_mesh.DataplaneResource{
			Meta: &test_model.ResourceMeta{
				Name: "dp1",
				Mesh: "default",
			},
			Spec: &mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					Address: "192.168.0.1",
					Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
						{
							Port: 8080,
							Tags: map[string]string{
								"kuma.io/service": "backend",
							},
						},
					},
				},
			},
		}

		// given
		dataplanes := core_mesh.DataplaneResourceList{}
		externalServices := &core_mesh.ExternalServiceResourceList{}
		vipList := vips.List{}
		for i := 1; i <= 5; i++ {
			service := "service-" + strconv.Itoa(i)
			vip := fmt.Sprintf("240.0.0.%d", i)
			vipList[service] = vip

			dataplanes.Items = append(dataplanes.Items, &core_mesh.DataplaneResource{
				Meta: &test_model.ResourceMeta{
					Name: "dp" + strconv.Itoa(i),
					Mesh: "default",
				},
				Spec: &mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
							{
								Port:    uint32(1234 + i),
								Address: vip,
								Tags: map[string]string{
									"kuma.io/service": service,
								},
							},
						},
					},
				},
			})
		}

		// when
		outbounds := dns.VIPOutbounds(model.MetaToResourceKey(dataplane.Meta), dataplanes.Items, vipList, externalServices.Items)
		// and
		Expect(outbounds).To(HaveLen(5))
		// and
		Expect(outbounds[4].GetTags()[mesh_proto.ServiceTag]).To(Equal("service-5"))
		// and
		Expect(outbounds[4].Port).To(Equal(dns.VIPListenPort))
	})

	It("shouldn't add outbounds from other meshes", func() {
		dataplane := &core_mesh.DataplaneResource{
			Meta: &test_model.ResourceMeta{
				Name: "dp1",
				Mesh: "default",
			},
			Spec: &mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					Address: "192.168.0.1",
					Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
						{
							Port: 8080,
							Tags: map[string]string{
								"kuma.io/service": "backend",
							},
						},
					},
				},
			},
		}

		// given
		vipList := vips.List{
			"service-a": "240.0.0.1",
			"service-b": "240.0.0.2",
		}
		services := []*mesh_proto.Dataplane_Networking_Ingress_AvailableService{
			{
				Mesh: "default",
				Tags: map[string]string{
					"kuma.io/service": "service-a",
				},
			},
			{
				Mesh: "other",
				Tags: map[string]string{
					"kuma.io/service": "service-b",
				},
			},
		}
		externalServices := &core_mesh.ExternalServiceResourceList{}
		dataplanes := core_mesh.DataplaneResourceList{
			Items: []*core_mesh.DataplaneResource{
				{
					Meta: &test_model.ResourceMeta{
						Name: "dp-ingress",
						Mesh: "default",
					},
					Spec: &mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Ingress: &mesh_proto.Dataplane_Networking_Ingress{
								AvailableServices: services,
							},
						},
					},
				},
			},
		}

		// when
		outbounds := dns.VIPOutbounds(model.MetaToResourceKey(dataplane.Meta), dataplanes.Items, vipList, externalServices.Items)
		// and
		Expect(outbounds).To(HaveLen(1))
		// and
		Expect(outbounds[0].GetTags()[mesh_proto.ServiceTag]).To(Equal("service-a"))
		// and
		Expect(outbounds[0].Port).To(Equal(dns.VIPListenPort))
	})

	It("should preserve ExternalService port", func() {
		dataplane := &core_mesh.DataplaneResource{
			Meta: &test_model.ResourceMeta{Name: "dp1", Mesh: "default"},
			Spec: &mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					Address: "192.168.0.1",
					Inbound: []*mesh_proto.Dataplane_Networking_Inbound{{
						Port: 8080,
						Tags: map[string]string{
							"kuma.io/service": "backend",
						}},
					},
				},
			},
		}

		otherDataplanes := []*core_mesh.DataplaneResource{}
		vipList := vips.List{}
		for i := 1; i <= 5; i++ {
			service := "service-" + strconv.Itoa(i)
			vip := fmt.Sprintf("240.0.0.%d", i)
			vipList[service] = vip

			otherDataplanes = append(otherDataplanes, &core_mesh.DataplaneResource{
				Meta: &test_model.ResourceMeta{
					Name: "dp" + strconv.Itoa(i),
					Mesh: "default",
				},
				Spec: &mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
							{
								Port:    uint32(1234 + i),
								Address: vip,
								Tags: map[string]string{
									"kuma.io/service": service,
								},
							},
						},
					},
				},
			})
		}

		externalServices := []*core_mesh.ExternalServiceResource{
			{
				Meta: &test_model.ResourceMeta{Name: "es1", Mesh: "default"},
				Spec: &mesh_proto.ExternalService{
					Networking: &mesh_proto.ExternalService_Networking{
						Address: "1.1.1.1:1234",
					},
					Tags: map[string]string{
						mesh_proto.ServiceTag: "first-external-service",
					},
				},
			},
			{
				Meta: &test_model.ResourceMeta{Name: "es2", Mesh: "default"},
				Spec: &mesh_proto.ExternalService{
					Networking: &mesh_proto.ExternalService_Networking{
						Address: "2.2.2.2:4321",
					},
					Tags: map[string]string{
						mesh_proto.ServiceTag: "second-external-service",
					},
				},
			},
			{
				Meta: &test_model.ResourceMeta{Name: "es2", Mesh: "default"},
				Spec: &mesh_proto.ExternalService{
					Networking: &mesh_proto.ExternalService_Networking{
						Address: "kuma.io",
					},
					Tags: map[string]string{
						mesh_proto.ServiceTag: "third-external-service",
					},
				},
			},
		}
		vipList["first-external-service"] = "240.0.0.6"
		vipList["second-external-service"] = "240.0.0.7"
		vipList["third-external-service"] = "240.0.0.8"

		actual := &mesh_proto.Dataplane_Networking{}
		actual.Outbound = dns.VIPOutbounds(model.MetaToResourceKey(dataplane.Meta), otherDataplanes, vipList, externalServices)

		expected := `
     outbound:
      - address: 240.0.0.6
        port: 1234
        tags:
          kuma.io/service: first-external-service
      - address: 240.0.0.6
        port: 80
        tags:
          kuma.io/service: first-external-service
      - address: 240.0.0.7
        port: 4321
        tags:
          kuma.io/service: second-external-service
      - address: 240.0.0.7
        port: 80
        tags:
          kuma.io/service: second-external-service
      - address: 240.0.0.1
        port: 80
        tags:
          kuma.io/service: service-1
      - address: 240.0.0.2
        port: 80
        tags:
          kuma.io/service: service-2
      - address: 240.0.0.3
        port: 80
        tags:
          kuma.io/service: service-3
      - address: 240.0.0.4
        port: 80
        tags:
          kuma.io/service: service-4
      - address: 240.0.0.5
        port: 80
        tags:
          kuma.io/service: service-5
      - address: 240.0.0.8
        port: 80
        tags:
          kuma.io/service: third-external-service
`
		Expect(proto.ToYAML(actual)).To(MatchYAML(expected))
	})
})

var _ = Describe("VirtualOutbounds", func() {
	type testCase struct {
		givenDpName           string
		givenDps              []*core_mesh.DataplaneResource
		givenExternalServices []*core_mesh.ExternalServiceResource
		givenCidr             string
		givenVirtualOutbounds []*core_mesh.VirtualOutboundResource
		expectedOutbounds     []*mesh_proto.Dataplane_Networking_Outbound
	}
	DescribeTable("SuccessfullyGenerates",
		func(tc testCase) {
			givenResource := model.ResourceKey{Mesh: "mesh", Name: tc.givenDpName}
			res, err := dns.VirtualOutbounds(givenResource, tc.givenDps, tc.givenExternalServices, tc.givenVirtualOutbounds, tc.givenCidr)

			Expect(err).To(BeNil())
			Expect(res).To(Equal(tc.expectedOutbounds))
		},
		Entry("nothing", testCase{expectedOutbounds: nil, givenCidr: "240.0.0.0/9"}),
		Entry("no virtual outbounds return nothing", testCase{
			givenCidr:   "240.0.0.0/9",
			givenDpName: "me",
			givenDps: []*core_mesh.DataplaneResource{
				{
					Meta: &test_model.ResourceMeta{Mesh: "mesh", Name: "me"},
					Spec: dp("a", "b"),
				},
				{
					Meta: &test_model.ResourceMeta{Mesh: "mesh", Name: "you"},
					Spec: dp("a", "b"),
				},
			},
			expectedOutbounds: nil,
		}),
		Entry("only me adds me", testCase{
			givenCidr:   "240.0.0.0/9",
			givenDpName: "me",
			givenDps: []*core_mesh.DataplaneResource{
				simpleDp("me", []map[string]string{
					{mesh_proto.ServiceTag: "a"},
					{mesh_proto.ServiceTag: "b"},
				}),
			},
			givenVirtualOutbounds: []*core_mesh.VirtualOutboundResource{
				virtualOutbound("vo", nil, "80", "{{.service}}.mesh", map[string]string{"service": mesh_proto.ServiceTag}),
			},
			expectedOutbounds: []*mesh_proto.Dataplane_Networking_Outbound{
				outbound("240.0.0.0", "a.mesh", 80, map[string]string{mesh_proto.ServiceTag: "a"}),
				outbound("::ffff:f000:0", "a.mesh", 80, map[string]string{mesh_proto.ServiceTag: "a"}),
				outbound("240.0.0.1", "b.mesh", 80, map[string]string{mesh_proto.ServiceTag: "b"}),
				outbound("::ffff:f000:1", "b.mesh", 80, map[string]string{mesh_proto.ServiceTag: "b"}),
			},
		}),
		Entry("port template", testCase{
			givenCidr:   "240.0.0.0/9",
			givenDpName: "me",
			givenDps: []*core_mesh.DataplaneResource{
				simpleDp("me", []map[string]string{
					{mesh_proto.ServiceTag: "a", "port": "80"},
					{mesh_proto.ServiceTag: "b", "port": "81"},
				}),
			},
			givenVirtualOutbounds: []*core_mesh.VirtualOutboundResource{
				virtualOutbound("vo", nil, "{{.port}}", "{{.service}}.mesh", map[string]string{"service": mesh_proto.ServiceTag, "port": "port"}),
			},
			expectedOutbounds: []*mesh_proto.Dataplane_Networking_Outbound{
				outbound("240.0.0.0", "a.mesh", 80, map[string]string{mesh_proto.ServiceTag: "a", "port": "80"}),
				outbound("::ffff:f000:0", "a.mesh", 80, map[string]string{mesh_proto.ServiceTag: "a", "port": "80"}),
				outbound("240.0.0.1", "b.mesh", 81, map[string]string{mesh_proto.ServiceTag: "b", "port": "81"}),
				outbound("::ffff:f000:1", "b.mesh", 81, map[string]string{mesh_proto.ServiceTag: "b", "port": "81"}),
			},
		}),
		Entry("multiple dps", testCase{
			givenCidr:   "240.0.0.0/9",
			givenDpName: "me",
			givenDps: []*core_mesh.DataplaneResource{
				simpleDp("me", []map[string]string{{mesh_proto.ServiceTag: "a"}}),
				simpleDp("other", []map[string]string{{mesh_proto.ServiceTag: "b"}}),
			},
			givenVirtualOutbounds: []*core_mesh.VirtualOutboundResource{
				virtualOutbound("vo", nil, "80", "{{.service}}.mesh", map[string]string{"service": mesh_proto.ServiceTag}),
			},
			expectedOutbounds: []*mesh_proto.Dataplane_Networking_Outbound{
				outbound("240.0.0.0", "a.mesh", 80, map[string]string{mesh_proto.ServiceTag: "a"}),
				outbound("::ffff:f000:0", "a.mesh", 80, map[string]string{mesh_proto.ServiceTag: "a"}),
				outbound("240.0.0.1", "b.mesh", 80, map[string]string{mesh_proto.ServiceTag: "b"}),
				outbound("::ffff:f000:1", "b.mesh", 80, map[string]string{mesh_proto.ServiceTag: "b"}),
			},
		}),
		Entry("overlap of hostnames dedupes and pick the most matching", testCase{
			givenCidr:   "240.0.0.0/9",
			givenDpName: "me",
			givenDps: []*core_mesh.DataplaneResource{
				simpleDp("me", []map[string]string{{mesh_proto.ServiceTag: "a", "version": "v1", "info": "hello"}}),
				simpleDp("other", []map[string]string{{mesh_proto.ServiceTag: "b", "version": "v1"}}),
			},
			givenVirtualOutbounds: []*core_mesh.VirtualOutboundResource{
				virtualOutbound("vo", []*mesh_proto.Selector{
					{Match: map[string]string{"version": "v1"}},
				}, "80", "{{.version}}.mesh", map[string]string{"version": "version"}),
				virtualOutbound("vi", []*mesh_proto.Selector{
					{Match: map[string]string{"version": "v1", "info": "*"}},
				}, "80", "{{.version}}.mesh", map[string]string{"version": "version", "info": "info"}),
			},
			expectedOutbounds: []*mesh_proto.Dataplane_Networking_Outbound{
				outbound("240.0.0.0", "v1.mesh", 80, map[string]string{mesh_proto.ServiceTag: "a", "version": "v1", "info": "hello"}),
				outbound("::ffff:f000:0", "v1.mesh", 80, map[string]string{mesh_proto.ServiceTag: "a", "version": "v1", "info": "hello"}),
			},
		}),
		Entry("ingress", testCase{
			givenCidr:   "240.0.0.0/9",
			givenDpName: "me",
			givenDps: []*core_mesh.DataplaneResource{
				simpleDp("me", []map[string]string{{mesh_proto.ServiceTag: "a"}}),
				{
					Meta: &test_model.ResourceMeta{Mesh: "mesh", Name: "ingress"},
					Spec: &mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Ingress: &mesh_proto.Dataplane_Networking_Ingress{
								AvailableServices: []*mesh_proto.Dataplane_Networking_Ingress_AvailableService{
									{
										Mesh: "mesh",
										Tags: map[string]string{mesh_proto.ServiceTag: "foo"},
									},
									{
										Mesh: "mesh2",
										Tags: map[string]string{mesh_proto.ServiceTag: "foo2"},
									},
								},
							},
						},
					},
				},
			},
			givenVirtualOutbounds: []*core_mesh.VirtualOutboundResource{
				virtualOutbound("vo", nil, "80", "{{.service}}.mesh", map[string]string{"service": mesh_proto.ServiceTag}),
			},
			expectedOutbounds: []*mesh_proto.Dataplane_Networking_Outbound{
				outbound("240.0.0.0", "a.mesh", 80, map[string]string{mesh_proto.ServiceTag: "a"}),
				outbound("::ffff:f000:0", "a.mesh", 80, map[string]string{mesh_proto.ServiceTag: "a"}),
				outbound("240.0.0.1", "foo.mesh", 80, map[string]string{mesh_proto.ServiceTag: "foo"}),
				outbound("::ffff:f000:1", "foo.mesh", 80, map[string]string{mesh_proto.ServiceTag: "foo"}),
			},
		}),
		Entry("external_services", testCase{
			givenCidr:   "240.0.0.0/9",
			givenDpName: "me",
			givenDps: []*core_mesh.DataplaneResource{
				simpleDp("me", []map[string]string{{mesh_proto.ServiceTag: "a"}}),
			},
			givenExternalServices: []*core_mesh.ExternalServiceResource{
				{
					Meta: &test_model.ResourceMeta{Mesh: "mesh", Name: "external"},
					Spec: &mesh_proto.ExternalService{Tags: map[string]string{mesh_proto.ServiceTag: "bom"}},
				},
			},
			givenVirtualOutbounds: []*core_mesh.VirtualOutboundResource{
				virtualOutbound("vo", nil, "80", "{{.service}}.mesh", map[string]string{"service": mesh_proto.ServiceTag}),
			},
			expectedOutbounds: []*mesh_proto.Dataplane_Networking_Outbound{
				outbound("240.0.0.0", "a.mesh", 80, map[string]string{mesh_proto.ServiceTag: "a"}),
				outbound("::ffff:f000:0", "a.mesh", 80, map[string]string{mesh_proto.ServiceTag: "a"}),
				outbound("240.0.0.1", "bom.mesh", 80, map[string]string{mesh_proto.ServiceTag: "bom"}),
				outbound("::ffff:f000:1", "bom.mesh", 80, map[string]string{mesh_proto.ServiceTag: "bom"}),
			},
		}),
		Entry("multiple dps same tags generate hostnames depending on templates", testCase{
			givenCidr:   "240.0.0.0/9",
			givenDpName: "me",
			givenDps: []*core_mesh.DataplaneResource{
				simpleDp("me", []map[string]string{{mesh_proto.ServiceTag: "a", "version": "v1"}}),
				simpleDp("other", []map[string]string{{mesh_proto.ServiceTag: "a", "version": "v2"}}),
			},
			givenVirtualOutbounds: []*core_mesh.VirtualOutboundResource{
				virtualOutbound("va", nil, "80", "{{.service}}.mesh", map[string]string{"service": mesh_proto.ServiceTag}),
				virtualOutbound("vo", nil, "80", "{{.service}}.{{.version}}.mesh", map[string]string{"service": mesh_proto.ServiceTag, "version": "version"}),
			},
			expectedOutbounds: []*mesh_proto.Dataplane_Networking_Outbound{
				outbound("240.0.0.0", "a.mesh", 80, map[string]string{mesh_proto.ServiceTag: "a"}),
				outbound("::ffff:f000:0", "a.mesh", 80, map[string]string{mesh_proto.ServiceTag: "a"}),
				outbound("240.0.0.1", "a.v1.mesh", 80, map[string]string{mesh_proto.ServiceTag: "a", "version": "v1"}),
				outbound("::ffff:f000:1", "a.v1.mesh", 80, map[string]string{mesh_proto.ServiceTag: "a", "version": "v1"}),
				outbound("240.0.0.2", "a.v2.mesh", 80, map[string]string{mesh_proto.ServiceTag: "a", "version": "v2"}),
				outbound("::ffff:f000:2", "a.v2.mesh", 80, map[string]string{mesh_proto.ServiceTag: "a", "version": "v2"}),
			},
		}),
		Entry("preserve previously defined outbounds", testCase{
			givenCidr:   "240.0.0.0/9",
			givenDpName: "me",
			givenDps: []*core_mesh.DataplaneResource{
				simpleDp("me", []map[string]string{{mesh_proto.ServiceTag: "a", "version": "v1"}}, outbound("240.0.3.0", "a.mesh", 80, map[string]string{mesh_proto.ServiceTag: "a"}), outbound("::ffff:f000:300", "a.mesh", 80, map[string]string{mesh_proto.ServiceTag: "a"})),
				simpleDp("other", []map[string]string{{mesh_proto.ServiceTag: "a", "version": "v2"}}),
			},
			givenVirtualOutbounds: []*core_mesh.VirtualOutboundResource{
				virtualOutbound("va", nil, "80", "{{.service}}.mesh", map[string]string{"service": mesh_proto.ServiceTag}),
				virtualOutbound("vo", nil, "80", "{{.service}}.{{.version}}.mesh", map[string]string{"service": mesh_proto.ServiceTag, "version": "version"}),
			},
			expectedOutbounds: []*mesh_proto.Dataplane_Networking_Outbound{
				outbound("240.0.3.0", "a.mesh", 80, map[string]string{mesh_proto.ServiceTag: "a"}),
				outbound("::ffff:f000:300", "a.mesh", 80, map[string]string{mesh_proto.ServiceTag: "a"}),
				outbound("240.0.0.0", "a.v1.mesh", 80, map[string]string{mesh_proto.ServiceTag: "a", "version": "v1"}),
				outbound("::ffff:f000:0", "a.v1.mesh", 80, map[string]string{mesh_proto.ServiceTag: "a", "version": "v1"}),
				outbound("240.0.0.1", "a.v2.mesh", 80, map[string]string{mesh_proto.ServiceTag: "a", "version": "v2"}),
				outbound("::ffff:f000:1", "a.v2.mesh", 80, map[string]string{mesh_proto.ServiceTag: "a", "version": "v2"}),
			},
		}),
		Entry("ignore invalid hostnames", testCase{
			givenCidr:   "240.0.0.0/9",
			givenDpName: "me",
			givenDps: []*core_mesh.DataplaneResource{
				simpleDp("me", []map[string]string{{mesh_proto.ServiceTag: "a", "version": "v1"}}),
				simpleDp("other", []map[string]string{{mesh_proto.ServiceTag: "a"}}),
			},
			givenVirtualOutbounds: []*core_mesh.VirtualOutboundResource{
				virtualOutbound("vo", nil, "80", "{{.service}}.{{.version}}.mesh", map[string]string{"service": mesh_proto.ServiceTag, "version": "version"}),
			},
			expectedOutbounds: []*mesh_proto.Dataplane_Networking_Outbound{
				outbound("240.0.0.0", "a.v1.mesh", 80, map[string]string{mesh_proto.ServiceTag: "a", "version": "v1"}),
				outbound("::ffff:f000:0", "a.v1.mesh", 80, map[string]string{mesh_proto.ServiceTag: "a", "version": "v1"}),
			},
		}),
		Entry("sorts picks by best matching selector if overlap", testCase{
			givenCidr:   "240.0.0.0/9",
			givenDpName: "me",
			givenDps: []*core_mesh.DataplaneResource{
				simpleDp("me", []map[string]string{{mesh_proto.ServiceTag: "a", "version": "v1"}}),
				simpleDp("other", []map[string]string{{mesh_proto.ServiceTag: "a", "version": "v2"}}),
				simpleDp("other", []map[string]string{{mesh_proto.ServiceTag: "b", "version": "v2"}}),
			},
			givenVirtualOutbounds: []*core_mesh.VirtualOutboundResource{
				virtualOutbound("vo", []*mesh_proto.Selector{{Match: map[string]string{mesh_proto.ServiceTag: "a"}}}, "80", "{{.service}}.mesh", map[string]string{"service": mesh_proto.ServiceTag}),
				virtualOutbound("va", []*mesh_proto.Selector{{Match: map[string]string{mesh_proto.ServiceTag: "*"}}}, "80", "{{.service}}.{{.version}}.mesh", map[string]string{"service": mesh_proto.ServiceTag, "version": "version"}),
			},
			expectedOutbounds: []*mesh_proto.Dataplane_Networking_Outbound{
				outbound("240.0.0.0", "a.v1.mesh", 80, map[string]string{mesh_proto.ServiceTag: "a", "version": "v1"}),
				outbound("::ffff:f000:0", "a.v1.mesh", 80, map[string]string{mesh_proto.ServiceTag: "a", "version": "v1"}),
				outbound("240.0.0.1", "a.v2.mesh", 80, map[string]string{mesh_proto.ServiceTag: "a", "version": "v2"}),
				outbound("::ffff:f000:1", "a.v2.mesh", 80, map[string]string{mesh_proto.ServiceTag: "a", "version": "v2"}),
				outbound("240.0.0.2", "b.v2.mesh", 80, map[string]string{mesh_proto.ServiceTag: "b", "version": "v2"}),
				outbound("::ffff:f000:2", "b.v2.mesh", 80, map[string]string{mesh_proto.ServiceTag: "b", "version": "v2"}),
				outbound("240.0.0.3", "a.mesh", 80, map[string]string{mesh_proto.ServiceTag: "a"}),
				outbound("::ffff:f000:3", "a.mesh", 80, map[string]string{mesh_proto.ServiceTag: "a"}),
			},
		}),
		Entry("v6 cidr doesn't generate v4 addresses", testCase{
			givenCidr:   "2001:DB8::/96",
			givenDpName: "a",
			givenDps: []*core_mesh.DataplaneResource{
				simpleDp("a", []map[string]string{{mesh_proto.ServiceTag: "a"}}, &mesh_proto.Dataplane_Networking_Outbound{Address: "2001:db8::1", Hostname: "c.mesh", Port: 80, Tags: map[string]string{mesh_proto.ServiceTag: "c"}}),
				simpleDp("b", []map[string]string{{mesh_proto.ServiceTag: "b"}}),
				simpleDp("c", []map[string]string{{mesh_proto.ServiceTag: "c"}}),
			},
			givenVirtualOutbounds: []*core_mesh.VirtualOutboundResource{
				virtualOutbound("vo", []*mesh_proto.Selector{{Match: map[string]string{mesh_proto.ServiceTag: "*"}}}, "80", "{{.service}}.mesh", map[string]string{"service": mesh_proto.ServiceTag}),
			},
			expectedOutbounds: []*mesh_proto.Dataplane_Networking_Outbound{
				outbound("2001:db8::", "a.mesh", 80, map[string]string{mesh_proto.ServiceTag: "a"}),
				outbound("2001:db8::2", "b.mesh", 80, map[string]string{mesh_proto.ServiceTag: "b"}),
				outbound("2001:db8::1", "c.mesh", 80, map[string]string{mesh_proto.ServiceTag: "c"}),
			},
		}),
	)
})

func outbound(address string, host string, port uint32, tags map[string]string) *mesh_proto.Dataplane_Networking_Outbound {
	return &mesh_proto.Dataplane_Networking_Outbound{
		Address:  address,
		Hostname: host,
		Port:     port,
		Tags:     tags,
	}
}

func virtualOutbound(name string, selectors []*mesh_proto.Selector, port string, host string, parameters map[string]string) *core_mesh.VirtualOutboundResource {
	return &core_mesh.VirtualOutboundResource{
		Meta: &test_model.ResourceMeta{Mesh: "mesh", Name: name},
		Spec: &mesh_proto.VirtualOutbound{
			Selectors: selectors,
			Conf: &mesh_proto.VirtualOutbound_Conf{
				Port:       port,
				Host:       host,
				Parameters: parameters,
			},
		},
	}
}

func simpleDp(name string, tags []map[string]string, outbounds ...*mesh_proto.Dataplane_Networking_Outbound) *core_mesh.DataplaneResource {
	var inbounds []*mesh_proto.Dataplane_Networking_Inbound
	for _, inbound := range tags {
		inbounds = append(inbounds, &mesh_proto.Dataplane_Networking_Inbound{
			Tags: inbound,
		})
	}
	return &core_mesh.DataplaneResource{
		Meta: &test_model.ResourceMeta{Mesh: "mesh", Name: name},
		Spec: &mesh_proto.Dataplane{
			Networking: &mesh_proto.Dataplane_Networking{
				Outbound: outbounds,
				Inbound:  inbounds,
			},
		},
	}
}
