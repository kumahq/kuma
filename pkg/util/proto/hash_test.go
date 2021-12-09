package proto_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/proto"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("Hash", func() {
	It("should produce stable hash", func() {
		// given
		dp1 := &mesh_proto.Dataplane{
			Networking: &mesh_proto.Dataplane_Networking{
				Address: "192.168.0.1",
			},
		}
		dp2 := proto.Clone(dp1).(*mesh_proto.Dataplane)

		// when
		hash1, err1 := util_proto.Hash(dp1)
		hash2, err2 := util_proto.Hash(dp2)

		// then
		Expect(err1).ToNot(HaveOccurred())
		Expect(err2).ToNot(HaveOccurred())
		Expect(hash1).To(Equal(hash2))
	})
})
