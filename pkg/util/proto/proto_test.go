package proto_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("ToJson", func() {
	It("test ordering", func() {
		spec := &mesh_proto.Dataplane{
			Networking: &mesh_proto.Dataplane_Networking{
				Address: "123.11.11.11",
				Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
					{
						Port: 9090,
					},
				},
				Admin: &mesh_proto.EnvoyAdmin{
					Port: 9901,
				},
			},
		}
		messageSorted, err := util_proto.ToJSONSorted(spec)
		Expect(err).ToNot(HaveOccurred())
		Expect(messageSorted).To(Equal([]byte(`{"networking":{"address":"123.11.11.11","admin":{"port":9901},"inbound":[{"port":9090}]}}`)))
		messageNotSorted, err := util_proto.ToJSON(spec)
		Expect(err).ToNot(HaveOccurred())
		Expect(messageNotSorted).To(Equal([]byte(`{"networking":{"address":"123.11.11.11","inbound":[{"port":9090}],"admin":{"port":9901}}}`)))
	})
})
