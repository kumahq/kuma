package mesh

import (
	. "github.com/onsi/ginkgo/v2"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
)

var _ = Describe("ExternalServiceResource", func() {
	Describe("MarshalLog", func() {
		It("should marshal log without panicking", func() {
			// given
			meshResourceList := ExternalServiceResourceList{
				Items: []*ExternalServiceResource{
					{
						Spec: &mesh_proto.ExternalService{
							Networking: &mesh_proto.ExternalService_Networking{
								Tls: &mesh_proto.ExternalService_Networking_TLS{},
							},
						},
					},
				},
			}

			// when
			meshResourceList.MarshalLog()

			// then
			// expect no panic
		})
	})
})
