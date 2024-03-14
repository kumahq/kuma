package metrics

import (
	"bytes"
	"os"
	"path"

	"github.com/kumahq/kuma/pkg/plugins/policies/meshmetric/api/v1alpha1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/plugins/policies/meshmetric/api/v1alpha1"
)

var _ = Describe("Profiles", func() {
	type testCase struct {
		input    string
		expected string
		profiles *v1alpha1.Profiles
	}
	DescribeTable("should filter according to profiles data",
		func(given testCase) {
			expected, err := os.Open(path.Join("testdata", "profiles", given.expected))
			Expect(err).ToNot(HaveOccurred())
			input, err := os.Open(path.Join("testdata", "profiles", given.input))
			Expect(err).ToNot(HaveOccurred())
			sidecar := &v1alpha1.Sidecar{
				Profiles: given.profiles,
			}

			actual := new(bytes.Buffer)
			err = AggregatedMetricsMutator(ProfileMutatorGenerator(sidecar))(input, actual)
			Expect(err).ToNot(HaveOccurred())

			Expect(toLines(actual)).To(ConsistOf(toLines(expected)))
		},
		Entry("should not filter on All profile", testCase{
			input:    "counter.in",
			expected: "counter.golden",
			profiles: &v1alpha1.Profiles{
				AppendProfiles: &[]v1alpha1.Profile{{Name: v1alpha1.AllProfileName}},
				Exclude:        nil,
				Include:        nil,
			},
		}),
	)
})
