package mesh

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("ExternalServiceResource", func() {
	Describe("MarshalLog", func() {
		It("should mask the sensitive information when marshaling", func() {
			// given
			meshResourceList := ExternalServiceResourceList{
				Items: []*ExternalServiceResource{
					{
						Spec: &mesh_proto.ExternalService{
							Networking: &mesh_proto.ExternalService_Networking{
								Tls: &mesh_proto.ExternalService_Networking_TLS{
									CaCert:     &system_proto.DataSource{Type: &system_proto.DataSource_Inline{Inline: util_proto.Bytes([]byte("secret1"))}},
									ClientCert: &system_proto.DataSource{Type: &system_proto.DataSource_Inline{Inline: util_proto.Bytes([]byte("secret2"))}},
									ClientKey:  &system_proto.DataSource{Type: &system_proto.DataSource_Inline{Inline: util_proto.Bytes([]byte("secret3"))}},
								},
							},
						},
					},
				},
			}

			// when
			masked := meshResourceList.MarshalLog().(ExternalServiceResourceList)

			// then
			Expect(masked.Items[0].Spec.Networking.Tls.CaCert.String()).To(Equal(`inline:{value:"***"}`))
			Expect(masked.Items[0].Spec.Networking.Tls.ClientCert.String()).To(Equal(`inline:{value:"***"}`))
			Expect(masked.Items[0].Spec.Networking.Tls.ClientKey.String()).To(Equal(`inline:{value:"***"}`))
		})
	})
})
