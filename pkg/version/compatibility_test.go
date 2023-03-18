package version_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/pkg/version"
)

var _ = Describe("Version Compatibility", func() {
	It("should accept <= two prior minor versions", func() {
		result := DeploymentVersionCompatible("1.4.0", "1.4.1")
		Expect(result).To(BeTrue())
		result = DeploymentVersionCompatible("1.4.1", "1.4.0")
		Expect(result).To(BeTrue())
		result = DeploymentVersionCompatible("1.4.1", "1.3.5")
		Expect(result).To(BeTrue())
		result = DeploymentVersionCompatible("1.4.1", "1.2.0")
		Expect(result).To(BeTrue())
	})

	It("should accept <= two latter minor versions", func() {
		result := DeploymentVersionCompatible("1.4.0", "1.4.1")
		Expect(result).To(BeTrue())
		result = DeploymentVersionCompatible("1.4.1", "1.4.0")
		Expect(result).To(BeTrue())
		result = DeploymentVersionCompatible("1.3.1", "1.4.5")
		Expect(result).To(BeTrue())
		result = DeploymentVersionCompatible("1.2.1", "1.4.0")
		Expect(result).To(BeTrue())
	})

	It("should reject > two prior minor versions", func() {
		result := DeploymentVersionCompatible("1.4.0", "1.1.1")
		Expect(result).To(BeFalse())
		result = DeploymentVersionCompatible("1.4.1", "1.1.0")
		Expect(result).To(BeFalse())
		result = DeploymentVersionCompatible("1.4.1", "1.0.5")
		Expect(result).To(BeFalse())
		result = DeploymentVersionCompatible("1.4.1", "1.0.0")
		Expect(result).To(BeFalse())
	})

	It("should reject > two latter minor versions", func() {
		result := DeploymentVersionCompatible("1.4.0", "1.7.1")
		Expect(result).To(BeFalse())
		result = DeploymentVersionCompatible("1.4.1", "1.8.0")
		Expect(result).To(BeFalse())
		result = DeploymentVersionCompatible("1.3.1", "1.6.5")
		Expect(result).To(BeFalse())
		result = DeploymentVersionCompatible("1.2.1", "1.5.0")
		Expect(result).To(BeFalse())
	})

	It("should reject disparate major versions", func() {
		result := DeploymentVersionCompatible("1.4.0", "2.4.0")
		Expect(result).To(BeFalse())
		result = DeploymentVersionCompatible("2.4.1", "1.4.1")
		Expect(result).To(BeFalse())
		result = DeploymentVersionCompatible("1.0.0", "2.1.1")
		Expect(result).To(BeFalse())
		result = DeploymentVersionCompatible("2.2.2", "1.0.0")
		Expect(result).To(BeFalse())
	})

	It("should always accept dev versions", func() {
		result := DeploymentVersionCompatible("1.4.0", "dev-1234")
		Expect(result).To(BeTrue())
		result = DeploymentVersionCompatible("abc-dev-1234", "1.4.0")
		Expect(result).To(BeTrue())
	})

	It("should not underflow while decrementing acceptable versions", func() {
		result := DeploymentVersionCompatible("1.0.0", "1.1.0")
		Expect(result).To(BeTrue())
		result = DeploymentVersionCompatible("1.1.0", "1.0.0")
		Expect(result).To(BeTrue())
		result = DeploymentVersionCompatible("dev-abcd", "1.0.0")
		Expect(result).To(BeTrue())
		result = DeploymentVersionCompatible("1.0.0", "dev-abcd")
		Expect(result).To(BeTrue())
	})
})
