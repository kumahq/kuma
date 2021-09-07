package release_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/version"
)

var _ = Describe("Compatibility", func() {

	It("current version is defined in CompatibilityMatrix", func() {
		// given
		currentVersion := version.Build.Version

		// when
		dpComptibility, err := version.CompatibilityMatrix.DataplaneConstraints(currentVersion)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(dpComptibility).ToNot(BeNil(), "Update version.CompatibilityMatrix so there is information about current compatible Envoy for this Kuma release.")
	})
})
