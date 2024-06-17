package versions_test

import (
	"github.com/Masterminds/semver/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/framework/versions"
)

var _ = Describe("versions", func() {
	DescribeTable("should return the oldest version that can be upgraded to the latest", func(currentStr string, expectedVersion string) {
		// given
		vers := []*semver.Version{
			semver.MustParse("1.2.3"),
			semver.MustParse("1.3.1"),
			semver.MustParse("1.4.2"),
			semver.MustParse("1.5.8"),
		}
		// when
		current := semver.MustParse(currentStr)
		oldest := versions.OldestUpgradableToVersion(vers, *current)
		// then
		Expect(oldest).To(Equal(expectedVersion))
	},
		Entry("preview", "0.0.0-preview.v123456789", "1.4.2"),
		Entry("new version before being added to list", "1.6.0", "1.4.2"),
		Entry("version in list", "1.5.8", "1.3.1"),
	)
})
