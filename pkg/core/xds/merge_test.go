package xds_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/core/xds"
)

var _ = Describe("MergeConfs", func() {

	type subPolicy struct {
		Ints       []int
		AppendInts []int
	}
	type policy struct {
		FieldString   string `json:"fieldString,omitempty"`
		Strings       []string
		AppendStrings []string
		Sub           subPolicy
		SubPtr        *subPolicy `json:"subPtr,omitempty"`
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
			merged, err := xds.MergeConfs(policies)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(merged.(policy)).To(Equal(given.expected))
		},
		Entry("should replace slices by default but append slices that start with append", testCase{
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
					Strings:       []string{"p2"},
					AppendStrings: []string{"p2"},
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
					Strings:       []string{"p3"},
					AppendStrings: []string{"p3"},
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
				FieldString:   "p2",
				Strings:       []string{"p3"},
				AppendStrings: []string{"p1", "p2", "p3"},
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
})
