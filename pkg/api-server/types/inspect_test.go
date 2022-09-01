package types_test

import (
	"encoding/json"
	"io"
	"os"
	"path"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/api-server/types"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	rest_unversioned "github.com/kumahq/kuma/pkg/core/resources/model/rest/unversioned"
	"github.com/kumahq/kuma/pkg/test/kds/samples"
	. "github.com/kumahq/kuma/pkg/test/matchers"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
)

var _ = Describe("Marshal DataplaneInspectEntry", func() {

	type testCase struct {
		input      *types.DataplaneInspectEntry
		goldenFile string
	}

	DescribeTable("should marshal DataplaneInspectEntry with jsonpb",
		func(given testCase) {
			actual, err := json.MarshalIndent(given.input, "", "  ")
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(MatchGoldenJSON(path.Join("testdata", given.goldenFile)))
		},
		Entry("empty", testCase{
			input:      &types.DataplaneInspectEntry{},
			goldenFile: "empty.json",
		}),
		Entry("full example", testCase{
			input: &types.DataplaneInspectEntry{
				AttachmentEntry: types.AttachmentEntry{
					Type:    "inbound",
					Name:    "192.168.0.1:80",
					Service: "web",
				},
				MatchedPolicies: map[model.ResourceType][]*rest_unversioned.Resource{
					core_mesh.TimeoutType: {
						rest_unversioned.From.Resource(&core_mesh.TimeoutResource{
							Meta: &test_model.ResourceMeta{Mesh: "mesh-1", Name: "t-1"},
							Spec: samples.Timeout,
						}),
					},
				},
			},
			goldenFile: "full_example.json",
		}),
	)
})

var _ = Describe("Unmarshal DataplaneInspectEntry", func() {

	type testCase struct {
		inputFile string
		output    *types.DataplaneInspectEntry
		errMsg    string
	}

	DescribeTable("should unmarshal DataplaneInspectEntry with jsonpb",
		func(given testCase) {
			inputFile, err := os.Open(path.Join("testdata", given.inputFile))
			Expect(err).ToNot(HaveOccurred())
			bytes, err := io.ReadAll(inputFile)
			Expect(err).ToNot(HaveOccurred())

			entry := &types.DataplaneInspectEntry{}
			err = json.Unmarshal(bytes, entry)
			Expect(err).ToNot(HaveOccurred())

			Expect(entry.Name).To(Equal(given.output.Name))
			Expect(entry.Type).To(Equal(given.output.Type))
			Expect(entry.MatchedPolicies).To(HaveLen(len(given.output.MatchedPolicies)))
			for resType, mp := range entry.MatchedPolicies {
				Expect(given.output.MatchedPolicies[resType]).ToNot(BeNil())
				Expect(mp).To(HaveLen(len(given.output.MatchedPolicies[resType])))
				for i, item := range mp {
					Expect(item.GetSpec()).To(MatchProto(given.output.MatchedPolicies[resType][i].GetSpec()))
				}
			}
		},
		Entry("empty", testCase{
			inputFile: "empty.json",
			output:    &types.DataplaneInspectEntry{},
		}),
		Entry("full example", testCase{
			inputFile: "full_example.json",
			output: &types.DataplaneInspectEntry{
				AttachmentEntry: types.AttachmentEntry{
					Type:    "inbound",
					Name:    "192.168.0.1:80",
					Service: "web",
				},
				MatchedPolicies: map[model.ResourceType][]*rest_unversioned.Resource{
					core_mesh.TimeoutType: {
						rest_unversioned.From.Resource(&core_mesh.TimeoutResource{
							Meta: &test_model.ResourceMeta{Mesh: "mesh-1", Name: "t-1"},
							Spec: samples.Timeout,
						}),
					},
				},
			},
		}),
		Entry("empty matched policies", testCase{
			inputFile: "empty_matched_policies.json",
			output: &types.DataplaneInspectEntry{
				AttachmentEntry: types.AttachmentEntry{
					Type: "inbound",
					Name: "192.168.0.1:80",
				},
				MatchedPolicies: map[model.ResourceType][]*rest_unversioned.Resource{},
			},
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

	DescribeTable("should return error",
		func(given testCase) {
			inputFile, err := os.Open(path.Join("testdata", given.inputFile))
			Expect(err).ToNot(HaveOccurred())
			bytes, err := io.ReadAll(inputFile)
			Expect(err).ToNot(HaveOccurred())

			entry := &types.DataplaneInspectEntry{}
			err = json.Unmarshal(bytes, entry)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(given.errMsg))
		},
		Entry("matchedPolicies key is empty", testCase{
			inputFile: "error.matched_policies_empty_key.json",
			errMsg:    `invalid resource type ""`,
		}),
	)
})
