package dns_test

import (
	"fmt"
	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

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
			Spec: mesh_proto.Dataplane{
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
		vipList := dns.VIPList{}
		for i := 1; i <= 5; i++ {
			service := "service-" + strconv.Itoa(i)
			vip := fmt.Sprintf("240.0.0.%d", i)
			vipList[service] = vip

			dataplanes.Items = append(dataplanes.Items, &core_mesh.DataplaneResource{
				Meta: &test_model.ResourceMeta{
					Name: "dp" + strconv.Itoa(i),
					Mesh: "default",
				},
				Spec: mesh_proto.Dataplane{
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
		outbounds := dns.VIPOutbounds(dataplane.Meta.GetName(), dataplanes.Items, vipList, externalServices.Items)
		// and
		Expect(outbounds).To(HaveLen(4))
		// and
		Expect(outbounds[3].GetTags()[mesh_proto.ServiceTag]).To(Equal("service-5"))
		// and
		Expect(outbounds[3].Port).To(Equal(dns.VIPListenPort))
	})
})
