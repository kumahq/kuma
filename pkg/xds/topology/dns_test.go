package topology_test

import (
	"fmt"
	"strconv"

	"github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/xds/topology"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/core/resources/model"

	"github.com/kumahq/kuma/pkg/dns/vips"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
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
		domains, outbounds := topology.VIPOutbounds(model.MetaToResourceKey(dataplane.Meta), dataplanes.Items, nil, vipList, "mesh", externalServices.Items)
		// and
		Expect(outbounds).To(HaveLen(5))
		// and
		Expect(outbounds[4].GetTags()[mesh_proto.ServiceTag]).To(Equal("service-5"))
		// and
		Expect(outbounds[4].Port).To(Equal(topology.VIPListenPort))
		// and
		Expect(domains).To(HaveLen(5))
		Expect(domains[4].Domains).To(Equal([]string{"service-5.mesh"}))
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
		domains, outbounds := topology.VIPOutbounds(model.MetaToResourceKey(dataplane.Meta), dataplanes.Items, nil, vipList, "mesh", externalServices.Items)
		// and
		Expect(outbounds).To(HaveLen(1))
		// and
		Expect(outbounds[0].GetTags()[mesh_proto.ServiceTag]).To(Equal("service-a"))
		// and
		Expect(outbounds[0].Port).To(Equal(topology.VIPListenPort))
		// and
		Expect(domains).To(HaveLen(1))
		Expect(domains[0].Domains).To(Equal([]string{"service-a.mesh"}))
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

		actualVips, actualOutbounds := topology.VIPOutbounds(model.MetaToResourceKey(dataplane.Meta), otherDataplanes, nil, vipList, "mesh", externalServices)

		expectedVips := []xds.VIPDomains{
			{Address: "240.0.0.6", Domains: []string{"first-external-service.mesh"}},
			{Address: "240.0.0.7", Domains: []string{"second-external-service.mesh"}},
			{Address: "240.0.0.1", Domains: []string{"service-1.mesh"}},
			{Address: "240.0.0.2", Domains: []string{"service-2.mesh"}},
			{Address: "240.0.0.3", Domains: []string{"service-3.mesh"}},
			{Address: "240.0.0.4", Domains: []string{"service-4.mesh"}},
			{Address: "240.0.0.5", Domains: []string{"service-5.mesh"}},
			{Address: "240.0.0.8", Domains: []string{"third-external-service.mesh"}},
		}
		Expect(actualVips).To(Equal(expectedVips))
		expectedOutbounds := []*mesh_proto.Dataplane_Networking_Outbound{
			{Address: "240.0.0.6", Port: 1234, Tags: map[string]string{mesh_proto.ServiceTag: "first-external-service"}},
			{Address: "240.0.0.6", Port: 80, Tags: map[string]string{mesh_proto.ServiceTag: "first-external-service"}},
			{Address: "240.0.0.7", Port: 4321, Tags: map[string]string{mesh_proto.ServiceTag: "second-external-service"}},
			{Address: "240.0.0.7", Port: 80, Tags: map[string]string{mesh_proto.ServiceTag: "second-external-service"}},
			{Address: "240.0.0.1", Port: 80, Tags: map[string]string{mesh_proto.ServiceTag: "service-1"}},
			{Address: "240.0.0.2", Port: 80, Tags: map[string]string{mesh_proto.ServiceTag: "service-2"}},
			{Address: "240.0.0.3", Port: 80, Tags: map[string]string{mesh_proto.ServiceTag: "service-3"}},
			{Address: "240.0.0.4", Port: 80, Tags: map[string]string{mesh_proto.ServiceTag: "service-4"}},
			{Address: "240.0.0.5", Port: 80, Tags: map[string]string{mesh_proto.ServiceTag: "service-5"}},
			{Address: "240.0.0.8", Port: 80, Tags: map[string]string{mesh_proto.ServiceTag: "third-external-service"}},
		}
		Expect(actualOutbounds).To(Equal(expectedOutbounds))
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

		actualVips, actualOutbounds := topology.VIPOutbounds(model.MetaToResourceKey(dataplane.Meta), otherDataplanes, zoneIngresses, vipList, "mesh", nil)

		expectedVips := []xds.VIPDomains{
			{Address: "240.0.0.2", Domains: []string{"new-ingress-svc-1.mesh"}},
			{Address: "240.0.0.3", Domains: []string{"new-ingress-svc-2.mesh"}},
			{Address: "240.0.0.0", Domains: []string{"old-ingress-svc-1.mesh"}},
			{Address: "240.0.0.1", Domains: []string{"old-ingress-svc-2.mesh"}},
		}
		Expect(actualVips).To(Equal(expectedVips))
		expectedOutbounds := []*mesh_proto.Dataplane_Networking_Outbound{
			{Address: "240.0.0.2", Port: 80, Tags: map[string]string{mesh_proto.ServiceTag: "new-ingress-svc-1"}},
			{Address: "240.0.0.3", Port: 80, Tags: map[string]string{mesh_proto.ServiceTag: "new-ingress-svc-2"}},
			{Address: "240.0.0.0", Port: 80, Tags: map[string]string{mesh_proto.ServiceTag: "old-ingress-svc-1"}},
			{Address: "240.0.0.1", Port: 80, Tags: map[string]string{mesh_proto.ServiceTag: "old-ingress-svc-2"}},
		}
		Expect(actualOutbounds).To(Equal(expectedOutbounds))
	})
})
