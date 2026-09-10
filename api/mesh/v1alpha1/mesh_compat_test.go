package v1alpha1_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
)

// The bytes below are what a 2.14 control plane wrote. The Mesh spec is stored, so a
// drift here would rewrite every mesh and reach an older control plane on rollback. The
// insight is recomputed and the overview is assembled for the API, but both are read by
// clients that parse the names protobuf chose.
var _ = Describe("Mesh family storage compatibility", func() {
	DescribeTable("round trips the mesh bytes 2.14 wrote",
		func(stored string, expected []string) {
			mesh := &mesh_proto.Mesh{}
			Expect(json.Unmarshal([]byte(stored), mesh)).To(Succeed())
			Expect(mesh.SkipCreatingInitialPolicies).To(Equal(expected))

			rewritten, err := json.Marshal(mesh)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(rewritten)).To(Equal(stored))
		},
		Entry("no policies skipped", `{}`, []string(nil)),
		Entry("all policies skipped", `{"skipCreatingInitialPolicies":["*"]}`, []string{"*"}),
		Entry("some policies skipped",
			`{"skipCreatingInitialPolicies":["TrafficPermission","MeshRetry"]}`,
			[]string{"TrafficPermission", "MeshRetry"}),
	)

	It("round trips a fully populated insight", func() {
		stored := `{"dataplanes":{"total":7,"online":4,"offline":2,"partiallyDegraded":1},"dpVersions":{"kumaDp":{"2.14.0":{"total":7,"online":4}},"envoy":{"1.31.0":{"total":7}}},"mTLS":{"issuedBackends":{"ca-1":{"total":3}},"supportedBackends":{"ca-1":{"total":5}}},"dataplanesByType":{"standard":{"total":5},"gateway":{"total":2},"gatewayDelegated":{"total":1}},"resources":{"MeshTimeout":{"total":9}}}`

		insight := &mesh_proto.MeshInsight{}
		Expect(json.Unmarshal([]byte(stored), insight)).To(Succeed())

		Expect(insight.GetDataplanes().GetPartiallyDegraded()).To(Equal(uint32(1)))
		Expect(insight.GetDpVersions().GetKumaDp()).To(HaveKey("2.14.0"))
		Expect(insight.GetMTLS().GetIssuedBackends()).To(HaveKey("ca-1"))
		Expect(insight.GetDataplanesByType().GetGatewayDelegated().GetTotal()).To(Equal(uint32(1)))
		Expect(insight.GetResources()).To(HaveKey("MeshTimeout"))

		rewritten, err := json.Marshal(insight)
		Expect(err).ToNot(HaveOccurred())
		Expect(string(rewritten)).To(Equal(stored))
	})

	It("keeps the two names protobuf did not spell as the field", func() {
		degraded, err := json.Marshal(&mesh_proto.MeshInsight{
			Dataplanes: &mesh_proto.MeshInsight_DataplaneStat{PartiallyDegraded: 1},
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(string(degraded)).To(Equal(`{"dataplanes":{"partiallyDegraded":1}}`))

		overview, err := json.Marshal(&mesh_proto.MeshOverview{
			MeshInsight: &mesh_proto.MeshInsight{},
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(string(overview)).To(Equal(`{"meshInsight":{}}`))
	})

	It("drops the counters protobuf left out", func() {
		out, err := json.Marshal(&mesh_proto.MeshOverview{})
		Expect(err).ToNot(HaveOccurred())
		Expect(string(out)).To(Equal(`{}`))

		out, err = json.Marshal(&mesh_proto.MeshInsight_DataplaneStat{})
		Expect(err).ToNot(HaveOccurred())
		Expect(string(out)).To(Equal(`{}`))
	})

	It("keeps the getters usable on a nil receiver", func() {
		var overview *mesh_proto.MeshOverview
		Expect(overview.GetMesh().GetSkipCreatingInitialPolicies()).To(BeNil())
		Expect(overview.GetMeshInsight().GetDataplanes().GetTotal()).To(Equal(uint32(0)))
		Expect(overview.GetMeshInsight().GetMTLS().GetIssuedBackends()).To(BeNil())
	})
})
