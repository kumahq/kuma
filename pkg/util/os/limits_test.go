package os

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("File limits", func() {
	It("should query the open file limit", func() {
		Expect(CurrentFileLimit()).Should(BeNumerically(">", 0))
	})

	It("should raise the open file limit", func() {
		current, err := CurrentFileLimit()
		Expect(err).Should(Succeed())

		Expect(RaiseFileLimit()).Should(Succeed())

		raised, err := CurrentFileLimit()
		Expect(err).Should(Succeed())

		Expect(raised).Should(BeNumerically(">", current))

		// Restore the original limit.
		Expect(setFileLimit(current)).Should(Succeed())
	})

	It("should fail to exceed the hard file limit", func() {
		Expect(setFileLimit(uint64(1) << 63)).Should(HaveOccurred())
	})
})
