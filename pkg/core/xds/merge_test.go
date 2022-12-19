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
		SubPtr        *subPolicy
	}

	It("should replace slices by default but append slices that start with append", func() {
		// given
		p1 := policy{
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
		}

		p2 := policy{
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
		}

		p3 := policy{
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
		}

		policies := []interface{}{p1, p2, p3}

		// when
		merged, err := xds.MergeConfs(policies)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(merged.(policy)).To(Equal(policy{
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
		}))
	})
})
