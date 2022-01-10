package types_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/api-server/types"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/test/kds/samples"
	. "github.com/kumahq/kuma/pkg/test/matchers"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
)

var _ = Describe("Marshal InspectEntry", func() {

	type testCase struct {
		input      *types.InspectEntry
		goldenFile string
	}

	DescribeTable("should marshal InspectEntry with jsonpb",
		func(given testCase) {
			actual, err := json.MarshalIndent(given.input, "", "  ")
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(MatchGoldenJSON(path.Join("testdata", given.goldenFile)))
		},
		Entry("empty", testCase{
			input:      &types.InspectEntry{},
			goldenFile: "empty.json",
		}),
		Entry("full example", testCase{
			input: &types.InspectEntry{
				Type: "inbound",
				Name: "192.168.0.1:80",
				MatchedPolicies: map[model.ResourceType][]model.Resource{
					core_mesh.TimeoutType: {
						&core_mesh.TimeoutResource{
							Meta: &test_model.ResourceMeta{Mesh: "mesh-1", Name: "t-1"},
							Spec: samples.Timeout,
						},
					},
				},
			},
			goldenFile: "full_example.json",
		}),
	)
})

var _ = Describe("Unmarshal InspectEntry", func() {

	type testCase struct {
		inputFile string
		output    *types.InspectEntry
		errMsg    string
	}

	DescribeTable("should unmarshal InspectEntry with jsonpb",
		func(given testCase) {
			inputFile, err := os.Open(path.Join("testdata", given.inputFile))
			Expect(err).ToNot(HaveOccurred())
			bytes, err := ioutil.ReadAll(inputFile)
			Expect(err).ToNot(HaveOccurred())

			unmarshaledInspectEntry := &types.InspectEntry{}
			err = json.Unmarshal(bytes, unmarshaledInspectEntry)
			Expect(err).ToNot(HaveOccurred())

			Expect(unmarshaledInspectEntry.Name).To(Equal(given.output.Name))
			Expect(unmarshaledInspectEntry.Type).To(Equal(given.output.Type))
			Expect(unmarshaledInspectEntry.MatchedPolicies).To(HaveLen(len(given.output.MatchedPolicies)))
			for resType, mp := range unmarshaledInspectEntry.MatchedPolicies {
				Expect(given.output.MatchedPolicies[resType]).ToNot(BeNil())
				Expect(mp).To(HaveLen(len(given.output.MatchedPolicies[resType])))
				for i, item := range mp {
					Expect(item.GetSpec()).To(MatchProto(given.output.MatchedPolicies[resType][i].GetSpec()))
				}
			}
		},
		Entry("empty", testCase{
			inputFile: "empty.json",
			output:    &types.InspectEntry{},
		}),
		Entry("full example", testCase{
			inputFile: "full_example.json",
			output: &types.InspectEntry{
				Type: "inbound",
				Name: "192.168.0.1:80",
				MatchedPolicies: map[model.ResourceType][]model.Resource{
					core_mesh.TimeoutType: {
						&core_mesh.TimeoutResource{
							Meta: &test_model.ResourceMeta{Mesh: "mesh-1", Name: "t-1"},
							Spec: samples.Timeout,
						},
					},
				},
			},
		}),
		Entry("empty matched policies", testCase{
			inputFile: "empty_matched_policies.json",
			output: &types.InspectEntry{
				Type:            "inbound",
				Name:            "192.168.0.1:80",
				MatchedPolicies: map[model.ResourceType][]model.Resource{},
			},
		}),
	)

	DescribeTable("should return error",
		func(given testCase) {
			inputFile, err := os.Open(path.Join("testdata", given.inputFile))
			Expect(err).ToNot(HaveOccurred())
			bytes, err := ioutil.ReadAll(inputFile)
			Expect(err).ToNot(HaveOccurred())

			unmarshalledInspectEntry := &types.InspectEntry{}
			err = json.Unmarshal(bytes, unmarshalledInspectEntry)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(given.errMsg))
		},
		Entry("matchedPolicies has wrong type", testCase{
			inputFile: "error.matched_policies_wrong_type.json",
			errMsg:    "json: cannot unmarshal array into Go struct field intermediateInspectEntry.matchedPolicies of type map[string][]*types.intermediateResource",
		}),
		Entry("matchedPolicies key is empty", testCase{
			inputFile: "error.matched_policies_empty_key.json",
			errMsg:    "MatchedPolicies key is empty",
		}),
		Entry("matchedPolicies value is not a list", testCase{
			inputFile: "error.matched_policies_value_not_list.json",
			errMsg:    "json: cannot unmarshal object into Go struct field intermediateInspectEntry.matchedPolicies of type []*types.intermediateResource",
		}),
	)
})
