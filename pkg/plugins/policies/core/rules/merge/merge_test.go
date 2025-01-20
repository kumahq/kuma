package merge_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/merge"
)

var _ = Describe("Confs", func() {
	type subPolicy struct {
		Ints       []int
		AppendInts []int
	}
	type policy struct {
		FieldString            string `json:"fieldString,omitempty"`
		Strings                []string
		AppendStrings          []string
		AppendPointerToStrings *[]string
		Sub                    subPolicy
		SubPtr                 *subPolicy `json:"subPtr,omitempty"`
	}

	type testCase struct {
		policies []policy
		expected policy
	}

	DescribeTable("merge confs",
		func(given testCase) {
			// given
			var policies []interface{}
			for _, policy := range given.policies {
				policies = append(policies, policy)
			}

			// when
			merged, err := merge.Confs(policies)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(merged).To(Equal([]interface{}{given.expected}))
		},
		Entry("should replace slices by default but append slices that start with append", testCase{
			policies: []policy{
				{
					FieldString:            "p1",
					Strings:                []string{"p1"},
					AppendStrings:          []string{"p1"},
					AppendPointerToStrings: &[]string{"p1"},
					Sub: subPolicy{
						Ints:       []int{1},
						AppendInts: []int{1},
					},
					SubPtr: &subPolicy{
						Ints:       []int{1},
						AppendInts: []int{1},
					},
				},
				{
					FieldString:            "p2",
					Strings:                []string{"p2"},
					AppendStrings:          []string{"p2"},
					AppendPointerToStrings: &[]string{"p2"},
					Sub: subPolicy{
						Ints:       []int{2},
						AppendInts: []int{2},
					},
					SubPtr: &subPolicy{
						Ints:       []int{2},
						AppendInts: []int{2},
					},
				},
				{
					Strings:                []string{"p3"},
					AppendStrings:          []string{"p3"},
					AppendPointerToStrings: &[]string{"p3"},
					Sub: subPolicy{
						Ints:       []int{3},
						AppendInts: []int{3},
					},
					SubPtr: &subPolicy{
						Ints:       []int{3},
						AppendInts: []int{3},
					},
				},
			},
			expected: policy{
				FieldString:            "p2",
				Strings:                []string{"p3"},
				AppendStrings:          []string{"p1", "p2", "p3"},
				AppendPointerToStrings: &[]string{"p1", "p2", "p3"},
				Sub: subPolicy{
					Ints:       []int{3},
					AppendInts: []int{1, 2, 3},
				},
				SubPtr: &subPolicy{
					Ints:       []int{3},
					AppendInts: []int{1, 2, 3},
				},
			},
		}),
		Entry("should handle nils on destination side", testCase{
			policies: []policy{
				{
					FieldString:   "p1",
					Strings:       []string{"p1"},
					AppendStrings: []string{"p1"},
					Sub: subPolicy{
						Ints:       []int{1},
						AppendInts: []int{1},
					},
					SubPtr: &subPolicy{
						Ints:       []int{1},
						AppendInts: []int{1},
					},
				},
				{
					FieldString:   "p2",
					Strings:       nil,
					AppendStrings: nil,
					Sub: subPolicy{
						Ints:       nil,
						AppendInts: nil,
					},
					SubPtr: nil,
				},
			},
			expected: policy{
				FieldString:   "p2",
				Strings:       nil,
				AppendStrings: []string{"p1"},
				Sub: subPolicy{
					Ints:       nil,
					AppendInts: []int{1},
				},
				SubPtr: &subPolicy{
					Ints:       []int{1},
					AppendInts: []int{1},
				},
			},
		}),
		Entry("should handle nils on source side", testCase{
			policies: []policy{
				{
					FieldString:   "p1",
					Strings:       nil,
					AppendStrings: nil,
					Sub: subPolicy{
						Ints:       nil,
						AppendInts: nil,
					},
					SubPtr: nil,
				},
				{
					FieldString: "p2",
					Sub: subPolicy{
						AppendInts: []int{2},
					},
					SubPtr: &subPolicy{
						AppendInts: []int{2},
					},
				},
				{
					FieldString:   "p3",
					Strings:       nil,
					AppendStrings: nil,
					Sub: subPolicy{
						Ints:       nil,
						AppendInts: nil,
					},
					SubPtr: &subPolicy{
						Ints:       []int{3},
						AppendInts: nil,
					},
				},
				{
					FieldString:   "p4",
					Strings:       []string{"4"},
					AppendStrings: []string{"4"},
					Sub: subPolicy{
						Ints:       []int{4},
						AppendInts: []int{4},
					},
					SubPtr: &subPolicy{
						Ints:       []int{4},
						AppendInts: []int{4},
					},
				},
			},
			expected: policy{
				FieldString:   "p4",
				Strings:       []string{"4"},
				AppendStrings: []string{"4"},
				Sub: subPolicy{
					Ints:       []int{4},
					AppendInts: []int{2, 4},
				},
				SubPtr: &subPolicy{
					Ints:       []int{4},
					AppendInts: []int{2, 4},
				},
			},
		}),
	)

	type mergeKey struct {
		Left  bool
		Right bool
	}
	type mergeConf struct {
		A *bool `json:"a,omitempty"`
		B *bool `json:"b,omitempty"`
	}
	type mergeEntry struct {
		Key     mergeKey  `json:"key" policyMerge:"mergeKey"`
		Default mergeConf `json:"default"`
	}
	type nonMergeEntry struct {
		Key     mergeKey  `json:"key"`
		Default mergeConf `json:"default"`
	}
	type testPolicy struct {
		MergeValues       []mergeEntry     `policyMerge:"mergeValuesByKey"`
		DirectMergeValues []string         `policyMerge:"mergeValues"`
		OtherValues       *[]nonMergeEntry `json:"otherValues,omitempty"`
	}

	type mergeValuesByKeyCase struct {
		policies []testPolicy
		expected []testPolicy
	}

	t := true
	f := false
	DescribeTable("mergeValuesByKey",
		func(given mergeValuesByKeyCase) {
			var givens []interface{}
			for _, p := range given.policies {
				givens = append(givens, p)
			}

			merged, err := merge.Confs(givens)
			var values []testPolicy
			for _, mergedConf := range merged {
				values = append(values, mergedConf.(testPolicy))
			}
			Expect(err).ToNot(HaveOccurred())
			Expect(values).To(Equal(given.expected))
		},
		Entry("should work with a basic policy", mergeValuesByKeyCase{
			policies: []testPolicy{{
				MergeValues: []mergeEntry{{
					Key: mergeKey{Left: true}, Default: mergeConf{A: &t},
				}, {
					Key: mergeKey{Left: true}, Default: mergeConf{B: &t},
				}},
				OtherValues: &[]nonMergeEntry{{
					Key: mergeKey{Left: true}, Default: mergeConf{A: &t},
				}, {
					Key: mergeKey{Left: true}, Default: mergeConf{B: &t},
				}, {
					Key: mergeKey{Right: true}, Default: mergeConf{B: &t},
				}},
			}, {
				MergeValues: []mergeEntry{{
					Key: mergeKey{Right: true}, Default: mergeConf{B: &t},
				}},
			}, {
				MergeValues: []mergeEntry{{
					Key: mergeKey{Left: true}, Default: mergeConf{B: &f},
				}},
			}},
			expected: []testPolicy{{
				MergeValues: []mergeEntry{{
					Key: mergeKey{Right: true}, Default: mergeConf{B: &t},
				}, {
					Key: mergeKey{Left: true}, Default: mergeConf{A: &t, B: &f},
				}},
				OtherValues: &[]nonMergeEntry{{
					Key: mergeKey{Left: true}, Default: mergeConf{A: &t},
				}, {
					Key: mergeKey{Left: true}, Default: mergeConf{B: &t},
				}, {
					Key: mergeKey{Right: true}, Default: mergeConf{B: &t},
				}},
			}},
		}), Entry("should only merge based on set of merge values", mergeValuesByKeyCase{
			policies: []testPolicy{{
				DirectMergeValues: []string{"a", "b"},
				MergeValues: []mergeEntry{{
					Key: mergeKey{Left: true}, Default: mergeConf{A: &t},
				}, {
					Key: mergeKey{Left: true}, Default: mergeConf{B: &t},
				}},
				OtherValues: &[]nonMergeEntry{{
					Key: mergeKey{Left: true}, Default: mergeConf{A: &t},
				}, {
					Key: mergeKey{Left: true}, Default: mergeConf{B: &t},
				}, {
					Key: mergeKey{Right: true}, Default: mergeConf{B: &t},
				}},
			}, {
				DirectMergeValues: []string{"b"},
				MergeValues: []mergeEntry{{
					Key: mergeKey{Right: true}, Default: mergeConf{B: &t},
				}},
			}, {
				MergeValues: []mergeEntry{{
					Key: mergeKey{Right: true, Left: true}, Default: mergeConf{B: &t},
				}},
			}},
			expected: []testPolicy{{
				DirectMergeValues: []string{"a"},
				MergeValues: []mergeEntry{{
					Key: mergeKey{Left: true}, Default: mergeConf{A: &t, B: &t},
				}},
				OtherValues: &[]nonMergeEntry{{
					Key: mergeKey{Left: true}, Default: mergeConf{A: &t},
				}, {
					Key: mergeKey{Left: true}, Default: mergeConf{B: &t},
				}, {
					Key: mergeKey{Right: true}, Default: mergeConf{B: &t},
				}},
			}, {
				DirectMergeValues: []string{"b"},
				MergeValues: []mergeEntry{{
					Key: mergeKey{Left: true}, Default: mergeConf{A: &t, B: &t},
				}, {
					Key: mergeKey{Right: true}, Default: mergeConf{B: &t},
				}},
				OtherValues: &[]nonMergeEntry{{
					Key: mergeKey{Left: true}, Default: mergeConf{A: &t},
				}, {
					Key: mergeKey{Left: true}, Default: mergeConf{B: &t},
				}, {
					Key: mergeKey{Right: true}, Default: mergeConf{B: &t},
				}},
			}, {
				MergeValues: []mergeEntry{{
					Key: mergeKey{Right: true, Left: true}, Default: mergeConf{B: &t},
				}},
			}},
		}), Entry("should work with an empty list of key values", mergeValuesByKeyCase{
			policies: []testPolicy{{}},
			expected: []testPolicy{{}},
		}),
	)
})
