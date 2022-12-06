package mesh

import (
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	v1alpha1 "github.com/kumahq/kuma/api/system/v1alpha1"
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
								Tls: &mesh_proto.ExternalService_Networking_TLS{
									CaCert:     &v1alpha1.DataSource{Type: &v1alpha1.DataSource_Inline{Inline: util_proto.Bytes([]byte("secret1"))}},
									ClientCert: &v1alpha1.DataSource{Type: &v1alpha1.DataSource_Inline{Inline: util_proto.Bytes([]byte("secret2"))}},
									ClientKey:  &v1alpha1.DataSource{Type: &v1alpha1.DataSource_Inline{Inline: util_proto.Bytes([]byte("secret3"))}},
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
