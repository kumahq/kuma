package resolver_test

import (
	"math"

	. "github.com/Kong/kuma/pkg/dns-server/resolver"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DNS ip", func() {

	It("should allocate and free IP", func(done Done) {
		// given
		ipam := NewSimpleIPAM("192.168.0.1/32")
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

		// ready
		close(done)
	})

	It("should allocate 2^16 IP addresses", func(done Done) {
		// given
		ipam := NewSimpleIPAM("240.0.0.0/4")
		Expect(ipam).ToNot(BeNil())

		for i := 0; i < math.MaxInt16; i++ {
			// when
			_, err := ipam.AllocateIP()
			// then
			Expect(err).ToNot(HaveOccurred())
		}

		// ready
		close(done)
	})
})
