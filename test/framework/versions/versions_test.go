package versions_test

import (
	"github.com/Masterminds/semver/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/framework/versions"
)

var _ = Describe("versions", func() {
	DescribeTable("should return the list of versions that can be upgraded to the latest", func(currentStr string, expectedVersions []string) {
		// given
		vers := []versions.Version{
			{SemVer: semver.MustParse("1.1.1"), ReleaseDate: "2024-02-01"},
			{SemVer: semver.MustParse("1.2.3"), Lts: true, EndOfLifeDate: "2100-01-01", ReleaseDate: "2024-03-01"},
			{SemVer: semver.MustParse("1.3.1"), ReleaseDate: "2024-04-01"},
			{SemVer: semver.MustParse("1.4.2"), ReleaseDate: "2024-05-01"},
			{SemVer: semver.MustParse("1.5.8"), ReleaseDate: "2024-06-01"},
			{SemVer: semver.MustParse("1.5.9")},
		}
		// when
		current := semver.MustParse(currentStr)
		oldest := versions.UpgradableVersions(vers, *current)
		// then
		Expect(oldest).To(Equal(expectedVersions))
	},
		Entry(nil, "0.0.0-preview.v123456789", []string{"1.2.3", "1.4.2", "1.5.8"}),
		Entry(nil, "1.6.0", []string{"1.2.3", "1.4.2", "1.5.8"}),
		Entry(nil, "1.5.8", []string{"1.2.3", "1.3.1", "1.4.2"}),
	)
})
