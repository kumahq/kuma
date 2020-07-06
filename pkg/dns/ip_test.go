package dns_test

import (
	"math"

	. "github.com/Kong/kuma/pkg/dns"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DNS ip", func() {

	It("should allocate and free IP", func() {
		// given
		ipam, err := NewSimpleIPAM("192.168.0.1/32")
		Expect(err).ToNot(HaveOccurred())
		Expect(ipam).ToNot(BeNil())

		// when
		ip1, err := ipam.AllocateIP()
		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		ip2, err := ipam.AllocateIP()
		// then
		Expect(err).To(HaveOccurred())

		// when
		err = ipam.FreeIP(ip1)
		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		err = ipam.FreeIP(ip2)
		// then
		Expect(err).To(HaveOccurred())
	})

	It("should allocate 2^16 IP addresses", func() {
		// given
		ipam, err := NewSimpleIPAM("240.0.0.0/4")
		Expect(err).ToNot(HaveOccurred())
		Expect(ipam).ToNot(BeNil())

		for i := 0; i < math.MaxInt16; i++ {
			// when
			_, err := ipam.AllocateIP()
			// then
			Expect(err).ToNot(HaveOccurred())
		}
	})
})
