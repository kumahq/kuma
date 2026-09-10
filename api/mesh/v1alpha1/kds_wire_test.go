package v1alpha1_test

import (
	"encoding/hex"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/anypb"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
)

// The type URL and bytes below are what a 2.14 control plane put on the KDS wire for a
// Mesh. A zone of any released version reads the Any by its type URL and then parses
// protobuf, and fails the whole DeltaDiscoveryResponse when either does not match, so a
// global that changes them cannot sync to that zone at all.
var _ = Describe("Mesh KDS wire compatibility", func() {
	DescribeTable("writes the mesh bytes 2.14 wrote",
		func(spec *mesh_proto.Mesh, encoded string) {
			any, err := core_model.ToAny(spec)
			Expect(err).ToNot(HaveOccurred())
			Expect(any.GetTypeUrl()).To(Equal("type.googleapis.com/kuma.mesh.v1alpha1.Mesh"))
			Expect(hex.EncodeToString(any.GetValue())).To(Equal(encoded))

			read := &mesh_proto.Mesh{}
			Expect(core_model.FromAny(any, read)).To(Succeed())
			Expect(read.GetSkipCreatingInitialPolicies()).To(Equal(spec.GetSkipCreatingInitialPolicies()))
		},
		Entry("never set", &mesh_proto.Mesh{}, ""),
		Entry("all policies skipped",
			&mesh_proto.Mesh{SkipCreatingInitialPolicies: []string{"*"}}, "42012a"),
		Entry("some policies skipped",
			&mesh_proto.Mesh{SkipCreatingInitialPolicies: []string{"TrafficPermission", "MeshRetry"}},
			"4211547261666669635065726d697373696f6e42094d6573685265747279"),
	)

	It("still reads the type URL less json a control plane between the two wrote", func() {
		jsonOnly := &anypb.Any{Value: []byte(`{"skipCreatingInitialPolicies":["*"]}`)}
		read := &mesh_proto.Mesh{}
		Expect(core_model.FromAny(jsonOnly, read)).To(Succeed())
		Expect(read.GetSkipCreatingInitialPolicies()).To(Equal([]string{"*"}))
	})
})
