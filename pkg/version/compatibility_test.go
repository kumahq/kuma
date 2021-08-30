package version_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/version"
)

var _ = Describe("Compatibility", func() {
	type testCase struct {
		dpVersion               string
		expectedEnvoyConstraint string
	}
	DescribeTable("should return the supported versions",
		func(given testCase) {
			// when
			dpCompatibility, err := version.CompatibilityMatrix.DP(given.dpVersion)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(dpCompatibility.Envoy).To(Equal(given.expectedEnvoyConstraint))
		},
		Entry("1.0.0", testCase{
			dpVersion:               "1.0.0",
			expectedEnvoyConstraint: "1.16.0",
		}),
		Entry("1.0.1", testCase{
			dpVersion:               "1.0.1",
			expectedEnvoyConstraint: "1.16.0",
		}),
		Entry("1.0.2", testCase{
			dpVersion:               "1.0.2",
			expectedEnvoyConstraint: "1.16.1",
		}),
		Entry("1.0.3", testCase{
			dpVersion:               "1.0.3",
			expectedEnvoyConstraint: "1.16.1",
		}),
		Entry("1.0.4", testCase{
			dpVersion:               "1.0.4",
			expectedEnvoyConstraint: "1.16.1",
		}),
		Entry("1.0.5", testCase{
			dpVersion:               "1.0.5",
			expectedEnvoyConstraint: "1.16.2",
		}),
		Entry("1.0.6", testCase{
			dpVersion:               "1.0.6",
			expectedEnvoyConstraint: "1.16.2",
		}),
		Entry("1.0.7", testCase{
			dpVersion:               "1.0.7",
			expectedEnvoyConstraint: "1.16.2",
		}),
		Entry("1.1.0", testCase{
			dpVersion:               "1.1.0",
			expectedEnvoyConstraint: "~1.17.0",
		}),
		Entry("1.2.0", testCase{
			dpVersion:               "1.2.0",
			expectedEnvoyConstraint: "~1.18.0",
		}),
		Entry("1.3.0", testCase{
			dpVersion:               "1.3.0",
			expectedEnvoyConstraint: "~1.18.4",
		}),
	)
})
