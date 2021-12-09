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
				MatchedPolicies: []types.MatchedPolicy{
					{
						ResourceType: core_mesh.TimeoutType,
						Items: []model.ResourceSpec{
							samples.Timeout,
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
			for i, mp := range unmarshaledInspectEntry.MatchedPolicies {
				Expect(mp.ResourceType).To(Equal(given.output.MatchedPolicies[i].ResourceType))
				Expect(mp.Items).To(HaveLen(len(given.output.MatchedPolicies[i].Items)))
				for j, item := range mp.Items {
					Expect(item).To(MatchProto(given.output.MatchedPolicies[i].Items[j]))
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
				MatchedPolicies: []types.MatchedPolicy{
					{
						ResourceType: core_mesh.TimeoutType,
						Items: []model.ResourceSpec{
							samples.Timeout,
						},
					},
				},
			},
		}),
		Entry("empty items", testCase{
			inputFile: "empty_items.json",
			output: &types.InspectEntry{
				Type: "inbound",
				Name: "192.168.0.1:80",
				MatchedPolicies: []types.MatchedPolicy{
					{
						ResourceType: core_mesh.TimeoutType,
						Items:        []model.ResourceSpec{},
					},
				},
			},
		}),
		Entry("empty items", testCase{
			inputFile: "empty_resource_type.json",
			output: &types.InspectEntry{
				Type: "inbound",
				Name: "192.168.0.1:80",
				MatchedPolicies: []types.MatchedPolicy{
					{
						Items: []model.ResourceSpec{},
					},
				},
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
		Entry("matchedPolicies is not list", testCase{
			inputFile: "error.matched_policies_is_not_list.json",
			errMsg:    "MatchedPolicies is not a list",
		}),
		Entry("matchedPolicies's element is not a map", testCase{
			inputFile: "error.matched_policies_item_is_not_map.json",
			errMsg:    "MatchedPolicies[0] is not a map[string]interface{}",
		}),
		Entry("matchedPolicies[i].items is not list", testCase{
			inputFile: "error.matched_policies_items_is_not_list.json",
			errMsg:    "MatchedPolicies[0].Items is not a list",
		}),
	)
})
