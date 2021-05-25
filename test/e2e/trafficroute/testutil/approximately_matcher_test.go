package testutil_test

import (
	"github.com/kumahq/kuma/test/e2e/trafficroute/testutil"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ApproximatelyEqual Matcher", func() {

	It("should match if error rate is less than diff", func() {
		matcher := testutil.ApproximatelyEqual(80.0, 5.0)
		success, err := matcher.Match(76.0)
		Expect(err).ToNot(HaveOccurred())
		Expect(success).To(BeTrue())

		matcher = testutil.ApproximatelyEqual(80, 5.0)
		success, err = matcher.Match(84.0)
		Expect(err).ToNot(HaveOccurred())
		Expect(success).To(BeTrue())

		matcher = testutil.ApproximatelyEqual(80.0, 5)
		success, err = matcher.Match(80)
		Expect(err).ToNot(HaveOccurred())
		Expect(success).To(BeTrue())
	})

	It("should not match if error rate is less than diff", func() {
		matcher := testutil.ApproximatelyEqual(80, 5)
		success, err := matcher.Match(86)
		Expect(err).ToNot(HaveOccurred())
		Expect(success).To(BeFalse())
		Expect(matcher.NegatedFailureMessage(86)).To(Equal("86 is not approximately equal to 80"))
	})
})
