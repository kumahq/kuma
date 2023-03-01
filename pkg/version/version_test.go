package version_test

import (
	"time"

	"github.com/Masterminds/semver/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/pkg/version"
)

var _ = Describe("Verify build flags are set", func() {
	It("Should have build info has semver version", func() {
		_, err := semver.NewVersion(version.Build.Version)
		Expect(err).ToNot(HaveOccurred())
	})

	It("Should have a valid builddate", func() {
		_, err := time.Parse(time.RFC3339, version.Build.BuildDate)
		if err != nil {
			Expect(version.Build.BuildDate).To(Equal("local-build"))
		}
	})
	It("Should have a valid gittag", func() {
		Expect(version.Build.GitTag).To(MatchRegexp("[a-f0-9]+"))
	})
	It("Should have a valid git commit", func() {
		Expect(version.Build.GitCommit).ToNot(And(BeEmpty(), Equal("unknown")))
	})
	It("Should set envoy version", func() {
		Expect(version.Envoy).ToNot(And(BeEmpty(), Equal("unknown")))
	})
}, test.LabelBuildCheck)
