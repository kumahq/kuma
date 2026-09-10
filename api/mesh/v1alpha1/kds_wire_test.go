package v1alpha1_test

import (
	"encoding/hex"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/anypb"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
)

// The type URL below is what a 2.14 control plane put on the KDS wire for a Mesh. A zone
// of any released version reads the Any by its type URL and then parses protobuf, and
// fails the whole DeltaDiscoveryResponse when either does not match, so a global that
// changes them cannot sync to that zone at all. The Mesh spec has no fields since 3.0,
// so the bytes are empty and everything a control plane before it wrote is unknown
// fields, which protobuf drops instead of failing on.
var _ = Describe("Mesh KDS wire compatibility", func() {
	It("writes the type URL 2.14 wrote and no bytes", func() {
		any, err := core_model.ToAny(&mesh_proto.Mesh{})
		Expect(err).ToNot(HaveOccurred())
		Expect(any.GetTypeUrl()).To(Equal("type.googleapis.com/kuma.mesh.v1alpha1.Mesh"))
		Expect(hex.EncodeToString(any.GetValue())).To(Equal(""))
	})

	DescribeTable("still reads the mesh bytes 2.14 wrote",
		func(encoded string) {
			value, err := hex.DecodeString(encoded)
			Expect(err).ToNot(HaveOccurred())

			read := &mesh_proto.Mesh{}
			Expect(core_model.FromAny(&anypb.Any{
				TypeUrl: "type.googleapis.com/kuma.mesh.v1alpha1.Mesh",
				Value:   value,
			}, read)).To(Succeed())
			Expect(read).To(Equal(&mesh_proto.Mesh{}))
		},
		Entry("never set", ""),
		Entry("all policies skipped", "42012a"),
		Entry("some policies skipped", "4211547261666669635065726d697373696f6e42094d6573685265747279"),
	)

	It("still reads the type URL less json a control plane between the two wrote", func() {
		jsonOnly := &anypb.Any{Value: []byte(`{"skipCreatingInitialPolicies":["*"]}`)}
		read := &mesh_proto.Mesh{}
		Expect(core_model.FromAny(jsonOnly, read)).To(Succeed())
		Expect(read).To(Equal(&mesh_proto.Mesh{}))
	})
})
