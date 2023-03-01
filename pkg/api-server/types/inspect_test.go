package types_test

import (
	"encoding/json"
	"os"
	"path"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/api-server/types"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest/v1alpha1"
	meshaccesslog "github.com/kumahq/kuma/pkg/plugins/policies/meshaccesslog/api/v1alpha1"
	. "github.com/kumahq/kuma/pkg/test/matchers"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
)

var _ = Describe("DataplaneInspectEntry", func() {
	type testCase struct {
		input      *types.DataplaneInspectEntry
		goldenFile string
	}
	_ = DescribeTable("Marshal/Unmarshal",
		func(given testCase) {
			actual, err := json.MarshalIndent(given.input, "", "  ")
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(MatchGoldenJSON(path.Join("testdata", given.goldenFile)))

			// Check we can marshal and unmarshal and get back to what we had
			expected := &types.DataplaneInspectEntry{}
			err = json.Unmarshal(actual, expected)
			Expect(err).ToNot(HaveOccurred())
			Expect(expected).To(Equal(given.input))
		},
		Entry("empty", testCase{
			input:      &types.DataplaneInspectEntry{},
			goldenFile: "empty.golden.json",
		}),
		Entry("full example", testCase{
			input: &types.DataplaneInspectEntry{
				AttachmentEntry: types.AttachmentEntry{
					Type:    "inbound",
					Name:    "192.168.0.1:80",
					Service: "web",
				},
				MatchedPolicies: map[model.ResourceType][]v1alpha1.ResourceMeta{
					core_mesh.TimeoutType: {
						rest.From.Meta(&core_mesh.TimeoutResource{
							Meta: &test_model.ResourceMeta{Mesh: "mesh-1", Name: "t-1"},
						}),
					},
					meshaccesslog.MeshAccessLogType: {
						rest.From.Meta(&meshaccesslog.MeshAccessLogResource{
							Meta: &test_model.ResourceMeta{Mesh: "mesh-1", Name: "mal-1"},
						}),
					},
				},
			},
			goldenFile: "full_example.golden.json",
		}),
		Entry("empty matched policies", testCase{
			input: &types.DataplaneInspectEntry{
				AttachmentEntry: types.AttachmentEntry{
					Type: "inbound",
					Name: "192.168.0.1:80",
				},
				MatchedPolicies: map[model.ResourceType][]v1alpha1.ResourceMeta{},
			},
			goldenFile: "empty_matched_policies.golden.json",
		}),
	)

	It("should unmarshal DataplaneInspectEntryList", func() {
		// given
		input, err := os.ReadFile(path.Join("testdata", "dataplane_inspect_entry_list.json"))
		Expect(err).ToNot(HaveOccurred())

		// when
		entryList := &types.DataplaneInspectEntryList{}
		Expect(json.Unmarshal(input, entryList)).To(Succeed())

		// then
		Expect(entryList.Total).To(Equal(uint32(3)))
		Expect(entryList.Items).To(HaveLen(3))
	})
})
