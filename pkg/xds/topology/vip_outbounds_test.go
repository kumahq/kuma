package topology_test

import (
	"fmt"
	"strconv"

	"github.com/Kong/kuma/pkg/dns"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	test_model "github.com/Kong/kuma/pkg/test/resources/model"
	"github.com/Kong/kuma/pkg/xds/topology"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PatchDataplaneWithVIPOutbounds", func() {

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
								"service": "backend",
							},
						},
					},
				},
			},
		}

		// setup
		resolver := dns.NewDNSResolver("mesh")

		// given
		dataplanes := core_mesh.DataplaneResourceList{}
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
									"service": service,
								},
							},
						},
					},
				},
			})
		}
		resolver.SetVIPs(vipList)

		// when
		err := topology.PatchDataplaneWithVIPOutbounds(dataplane, &dataplanes, resolver)
		// then
		Expect(err).ToNot(HaveOccurred())
		// and
		Expect(len(dataplane.Spec.Networking.Outbound)).To(Equal(4))
		// and
		Expect(dataplane.Spec.Networking.Outbound[3].GetService()).To(Equal("service-5"))
		// and
		Expect(dataplane.Spec.Networking.Outbound[3].GetTags()[mesh_proto.ServiceTag]).To(Equal("service-5"))
		// and
		Expect(dataplane.Spec.Networking.Outbound[3].Port).To(Equal(topology.VIPListenPort))
	})

})
