package versions_test

import (
	"github.com/Masterminds/semver/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/util/versions"
)

var _ = Describe("OldestUpgradableToLatest", func() {
	It("should return the oldest version that can be upgraded to the latest", func() {
		// given
		vers := []*semver.Version{
			semver.MustParse("1.2.3"),
			semver.MustParse("1.3.1"),
			semver.MustParse("1.4.2"),
			semver.MustParse("1.5.8"),
		}
		// when
		oldest := versions.OldestUpgradableToLatest(vers)
		// then
		Expect(oldest).To(Equal("1.3.1"))
	})
})
