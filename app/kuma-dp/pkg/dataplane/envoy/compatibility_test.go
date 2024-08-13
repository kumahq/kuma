package envoy_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/app/kuma-dp/pkg/dataplane/envoy"
	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/pkg/version"
)

var _ = Describe("Compatibility", func() {
	It("Should be ok by default", func() {
		r, err := envoy.VersionCompatible("1.21.1", "1.21.1")
		Expect(err).ToNot(HaveOccurred())
		Expect(r).To(BeTrue())
	})
	It("Should fail with old version", func() {
		r, err := envoy.VersionCompatible("1.21.1", "1.17.1")
		Expect(err).ToNot(HaveOccurred())
		Expect(r).To(BeFalse())
	})
	It("Should fail with too recent", func() {
		r, err := envoy.VersionCompatible("1.21.1", "10.17.1") // Let's guess envoy 10 will never happen
		Expect(err).ToNot(HaveOccurred())
		Expect(r).To(BeFalse())
	})
	It("Should be ok with higher patch", func() {
		r, err := envoy.VersionCompatible("1.21.1", "1.21.2")
		Expect(err).ToNot(HaveOccurred())
		Expect(r).To(BeTrue())
	})
	It("Should fail with lower patch", func() {
		r, err := envoy.VersionCompatible("1.21.1", "1.21.0")
		Expect(err).ToNot(HaveOccurred())
		Expect(r).To(BeFalse())
	})
	It("Should fail with higher minor", func() {
		r, err := envoy.VersionCompatible("1.21.1", "1.22.1")
		Expect(err).ToNot(HaveOccurred())
		Expect(r).To(BeFalse())
	})
	It("Should fail with lower minor", func() {
		r, err := envoy.VersionCompatible("1.21.1", "1.20.1")
		Expect(err).ToNot(HaveOccurred())
		Expect(r).To(BeFalse())
	})
	It("Should err with bad version", func() {
		r, err := envoy.VersionCompatible("1.21.1", "funky")
		Expect(err).To(HaveOccurred())
		Expect(r).To(BeFalse())
	})
	It("Should err with bad expected version", func() {
		r, err := envoy.VersionCompatible("xyz", "1.20.1")
		Expect(err).To(HaveOccurred())
		Expect(r).To(BeFalse())
	})
	It("Should work on exact versions that aren't semver", func() {
		r, err := envoy.VersionCompatible("xyz", "xyz")
		Expect(err).ToNot(HaveOccurred())
		Expect(r).To(BeTrue())
	})
	It("Works with actual build version", func() {
		r, err := envoy.VersionCompatible(version.Envoy, version.Envoy)
		Expect(err).ToNot(HaveOccurred())
		Expect(r).To(BeTrue())
	}, test.LabelBuildCheck)
})
