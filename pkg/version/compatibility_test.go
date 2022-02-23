package version_test

import (
	. "github.com/onsi/ginkgo/v2"
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
			dpCompatibility, err := version.CompatibilityMatrix.DataplaneConstraints(given.dpVersion)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(dpCompatibility.Envoy).To(Equal(given.expectedEnvoyConstraint))
		},
		Entry("1.2.0", testCase{
			dpVersion:               "1.2.0",
			expectedEnvoyConstraint: "~1.18.0",
		}),
		Entry("1.3.0", testCase{
			dpVersion:               "1.3.0",
			expectedEnvoyConstraint: "~1.18.4",
		}),
	)

	It("should return error when there is no compatibility information for given version", func() {
		// when
		_, err := version.CompatibilityMatrix.DataplaneConstraints("100.0.0")

		// then
		Expect(err).To(MatchError("no constraints for version: 100.0.0 found"))
	})

	It("should return error when version is invalid", func() {
		// when
		_, err := version.CompatibilityMatrix.DataplaneConstraints("!@#")

		// then
		Expect(err).To(MatchError(`could not build a constraint for Kuma version !@#: Invalid Semantic Version`))
	})

	It("should throw an error when there are multiple matching constraints", func() {
		// given
		compatibility := version.Compatibility{
			KumaDP: map[string]version.DataplaneCompatibility{
				"1.0.0": {
					Envoy: "1.16.0",
				},
				"~1.0.0": {
					Envoy: "1.16.0",
				},
			},
		}

		// when
		_, err := compatibility.DataplaneConstraints("1.0.0")

		// then
		Expect(err).To(MatchError("more than one constraint for version 1.0.0: 1.16.0, 1.16.0"))
	})

	It("should include an entry that matches the master version after a release is made", func() {
		// given
		currentVersion := version.Build.Version

		// when
		dpComptibility, err := version.CompatibilityMatrix.DataplaneConstraints(currentVersion)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(dpComptibility).ToNot(BeNil(), "Update version.CompatibilityMatrix with a tag including prereleases so there is information about the master branch.")
	})
})
