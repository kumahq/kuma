package metrics

import (
	"bytes"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"os"
	"path"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/plugins/policies/meshmetric/api/v1alpha1"
)

var _ = Describe("Profiles", func() {
	type testCase struct {
		input    string
		expected string
		profiles string
	}
	DescribeTable("should filter according to profiles data",
		func(given testCase) {
			// setup
			mm := v1alpha1.NewMeshMetricResource()
			policy, err := os.ReadFile(path.Join("testdata", "profiles", given.profiles))
			Expect(err).ToNot(HaveOccurred())
			err = core_model.FromYAML(policy, &mm.Spec)
			Expect(err).ToNot(HaveOccurred())

			expected, err := os.Open(path.Join("testdata", "profiles", given.expected))
			Expect(err).ToNot(HaveOccurred())
			input, err := os.Open(path.Join("testdata", "profiles", given.input))
			Expect(err).ToNot(HaveOccurred())

			actual := new(bytes.Buffer)
			err = AggregatedMetricsMutator(ProfileMutatorGenerator(mm.Spec.Default.Sidecar))(input, actual)
			Expect(err).ToNot(HaveOccurred())

			Expect(toLines(actual)).To(ConsistOf(toLines(expected)))
		},
		Entry("should not filter on All profile", testCase{
			input:    "all.in",
			expected: "all.golden",
			profiles: "all.yaml",
		}),
	)
})
