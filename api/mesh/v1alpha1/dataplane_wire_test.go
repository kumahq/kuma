package v1alpha1_test

import (
	"encoding/hex"
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/anypb"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
)

// The bytes below are what a 2.14 control plane put on the KDS wire, taken from the
// protobuf definition as it stood before the conversion. A zone of any released version
// reads the Any by its type URL and then parses protobuf, and fails the whole
// DeltaDiscoveryResponse when either does not match, so a global that changes them
// cannot sync to that zone at all.
var _ = Describe("Dataplane KDS wire compatibility", func() {
	const (
		dataplaneJSON = `{"networking":{"address":"10.0.0.1","inbound":[{"port":80,"serviceProbe":{"interval":"0.100s","healthyThreshold":1},"state":"NotReady","name":"http"}]}}`
		dataplaneWire = "0a230a171850420b0a051080c2d72f2202080148015204687474702a0831302e302e302e31"

		insightJSON = `{"subscriptions":[{"id":"sub-1","connectTime":"2026-09-10T12:00:00Z","status":{"total":{"responsesSent":"9","responsesRejected":"1"}}}],"mTLS":{"certificateExpirationTime":"2026-09-10T12:00:00Z","certificateRegenerations":2}}`
		insightWire = "0a170a057375622d311a0608c0b78ad5062a06120408091801120a0a0608c0b78ad5061802"
	)

	It("writes the dataplane bytes 2.14 wrote", func() {
		spec := &mesh_proto.Dataplane{}
		Expect(json.Unmarshal([]byte(dataplaneJSON), spec)).To(Succeed())

		any, err := core_model.ToAny(spec)
		Expect(err).ToNot(HaveOccurred())
		Expect(any.GetTypeUrl()).To(Equal("type.googleapis.com/kuma.mesh.v1alpha1.Dataplane"))
		Expect(hex.EncodeToString(any.GetValue())).To(Equal(dataplaneWire))

		read := &mesh_proto.Dataplane{}
		Expect(core_model.FromAny(any, read)).To(Succeed())
		rewritten, err := json.Marshal(read)
		Expect(err).ToNot(HaveOccurred())
		Expect(string(rewritten)).To(Equal(dataplaneJSON))
	})

	It("writes the dataplane insight bytes 2.14 wrote", func() {
		spec := &mesh_proto.DataplaneInsight{}
		Expect(json.Unmarshal([]byte(insightJSON), spec)).To(Succeed())

		any, err := core_model.ToAny(spec)
		Expect(err).ToNot(HaveOccurred())
		Expect(any.GetTypeUrl()).To(Equal("type.googleapis.com/kuma.mesh.v1alpha1.DataplaneInsight"))
		Expect(hex.EncodeToString(any.GetValue())).To(Equal(insightWire))

		read := &mesh_proto.DataplaneInsight{}
		Expect(core_model.FromAny(any, read)).To(Succeed())
		rewritten, err := json.Marshal(read)
		Expect(err).ToNot(HaveOccurred())
		Expect(string(rewritten)).To(Equal(insightJSON))
	})

	It("still reads the type URL less json a control plane between the two wrote", func() {
		read := &mesh_proto.Dataplane{}
		Expect(core_model.FromAny(&anypb.Any{Value: []byte(dataplaneJSON)}, read)).To(Succeed())
		Expect(read.GetNetworking().GetAddress()).To(Equal("10.0.0.1"))
	})
})
