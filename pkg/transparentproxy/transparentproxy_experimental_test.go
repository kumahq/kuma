package transparentproxy_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/transparentproxy"
)

var _ = Describe("ShouldEnableIPv6", func() {
	It("should return that IPv6 is disabled when port is 0", func() {
		// when
		enabled, err := transparentproxy.ShouldEnableIPv6(uint16(0))

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(enabled).To(BeFalse())
	})
})
