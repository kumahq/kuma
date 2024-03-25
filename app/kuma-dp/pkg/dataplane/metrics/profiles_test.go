package metrics

import (
	"bytes"
	"os"
	"path"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
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
		Entry("for All profile should not filter anything", testCase{
			input:    "all.in",
			expected: "all.golden",
			profiles: "all.yaml",
		}),
		Entry("for Basic profile should not filter dashboard metrics", testCase{
			input:    "dashboards.in",
			expected: "dashboards.golden",
			profiles: "dashboards.yaml",
		}),
		Entry("for None profile should not show any metrics", testCase{
			input:    "none.in",
			expected: "none.golden",
			profiles: "none.yaml",
		}),
		Entry("for None profile with include should include the metrics", testCase{
			input:    "include.in",
			expected: "include.golden",
			profiles: "include.yaml",
		}),
		Entry("for All profile with exclude should exclude the metrics", testCase{
			input:    "exclude.in",
			expected: "exclude.golden",
			profiles: "exclude.yaml",
		}),
		Entry("for Basic profile with exclude and include should exclude the metrics", testCase{
			input:    "mixed.in",
			expected: "mixed.golden",
			profiles: "mixed.yaml",
		}),
	)
})
