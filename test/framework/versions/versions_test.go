package versions_test

import (
	"github.com/Masterminds/semver/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v3/test/framework/versions"
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

	DescribeTable("should include the last minor of the previous major across a major bump", func(currentStr string, expectedVersions []string) {
		// given
		vers := []versions.Version{
			{SemVer: semver.MustParse("2.7.29"), Lts: true, EndOfLifeDate: "2100-01-01", ReleaseDate: "2024-04-19"},
			{SemVer: semver.MustParse("2.12.14"), ReleaseDate: "2025-09-09"},
			{SemVer: semver.MustParse("2.13.10"), ReleaseDate: "2025-12-22"},
			{SemVer: semver.MustParse("2.14.4"), ReleaseDate: "2026-06-12"},
			{SemVer: semver.MustParse("3.0.2"), ReleaseDate: "2026-10-01"},
			{SemVer: semver.MustParse("3.1.0"), ReleaseDate: "2027-01-01"},
		}
		// when
		current := semver.MustParse(currentStr)
		upgradable := versions.UpgradableVersions(vers, *current)
		// then
		Expect(upgradable).To(Equal(expectedVersions))
	},
		Entry(nil, "3.0.0", []string{"2.7.29", "2.14.4"}),
		Entry(nil, "3.0.2", []string{"2.7.29", "2.14.4"}),
		Entry(nil, "3.1.0", []string{"2.7.29", "2.14.4", "3.0.2"}),
		Entry(nil, "3.2.0", []string{"2.7.29", "3.0.2", "3.1.0"}),
		Entry(nil, "0.0.0-preview.v123456789", []string{"2.7.29", "3.0.2", "3.1.0"}),
	)

	It("should keep the 2.14 line upgradable onto a 3.0 build from versions.yml", func() {
		// given
		vers := versions.ParseFromFile("../../../versions.yml")
		// when
		upgradable := versions.UpgradableVersions(vers, *semver.MustParse("3.0.0"))
		// then
		Expect(upgradable).To(ContainElement(HavePrefix("2.14.")))
		Expect(upgradable).ToNot(ContainElement(HavePrefix("2.12.")))
	})
})
