package dns_test

import (
	"fmt"
	"strconv"

	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/util/proto"

	. "github.com/onsi/ginkgo"
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
			vipList[vips.NewServiceEntry(service)] = vip

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
		outbounds := dns.VIPOutbounds(model.MetaToResourceKey(dataplane.Meta), dataplanes.Items, nil, vipList, externalServices.Items)
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
			vips.NewServiceEntry("service-a"): "240.0.0.1",
			vips.NewServiceEntry("service-b"): "240.0.0.2",
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
		outbounds := dns.VIPOutbounds(model.MetaToResourceKey(dataplane.Meta), dataplanes.Items, nil, vipList, externalServices.Items)
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
			vipList[vips.NewServiceEntry(service)] = vip

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
		vipList[vips.NewServiceEntry("first-external-service")] = "240.0.0.6"
		vipList[vips.NewServiceEntry("second-external-service")] = "240.0.0.7"
		vipList[vips.NewServiceEntry("third-external-service")] = "240.0.0.8"

		actual := &mesh_proto.Dataplane_Networking{}
		actual.Outbound = dns.VIPOutbounds(model.MetaToResourceKey(dataplane.Meta), otherDataplanes, nil, vipList, externalServices)

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

	It("should take ingresses into account", func() {
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

		vipList := vips.List{
			vips.NewServiceEntry("old-ingress-svc-1"): "240.0.0.0",
			vips.NewServiceEntry("old-ingress-svc-2"): "240.0.0.1",
			vips.NewServiceEntry("new-ingress-svc-1"): "240.0.0.2",
			vips.NewServiceEntry("new-ingress-svc-2"): "240.0.0.3",
		}

		otherDataplanes := []*core_mesh.DataplaneResource{{
			Meta: &test_model.ResourceMeta{Name: "old-ingress", Mesh: "default"},
			Spec: &mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					Inbound: []*mesh_proto.Dataplane_Networking_Inbound{{
						Port:    10001,
						Address: "192.168.0.2",
						Tags: map[string]string{
							"kuma.io/service": "ingress",
						},
					}},
					Ingress: &mesh_proto.Dataplane_Networking_Ingress{
						AvailableServices: []*mesh_proto.Dataplane_Networking_Ingress_AvailableService{
							{Mesh: "default", Tags: map[string]string{mesh_proto.ServiceTag: "old-ingress-svc-1"}},
							{Mesh: "default", Tags: map[string]string{mesh_proto.ServiceTag: "old-ingress-svc-2"}},
						},
					},
				},
			},
		}}

		zoneIngresses := []*core_mesh.ZoneIngressResource{{
			Meta: &test_model.ResourceMeta{Name: "new-ingress", Mesh: model.NoMesh},
			Spec: &mesh_proto.ZoneIngress{
				Networking: &mesh_proto.ZoneIngress_Networking{
					Address: "192.168.0.3",
					Port:    10001,
				},
				AvailableServices: []*mesh_proto.ZoneIngress_AvailableService{
					{Mesh: "default", Tags: map[string]string{mesh_proto.ServiceTag: "new-ingress-svc-1"}},
					{Mesh: "default", Tags: map[string]string{mesh_proto.ServiceTag: "new-ingress-svc-2"}},
				},
			},
		}}

		actual := &mesh_proto.Dataplane_Networking{}
		actual.Outbound = dns.VIPOutbounds(model.MetaToResourceKey(dataplane.Meta), otherDataplanes, zoneIngresses, vipList, nil)

		expected := `
     outbound:
      - address: 240.0.0.2
        port: 80
        tags:
          kuma.io/service: new-ingress-svc-1
      - address: 240.0.0.3
        port: 80
        tags:
          kuma.io/service: new-ingress-svc-2
      - address: 240.0.0.0
        port: 80
        tags:
          kuma.io/service: old-ingress-svc-1
      - address: 240.0.0.1
        port: 80
        tags:
          kuma.io/service: old-ingress-svc-2
`
		Expect(proto.ToYAML(actual)).To(MatchYAML(expected))
	})
})
