package dns_server

import (
	"math"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DNS ip", func() {

	It("IP allocate and free operations", func(done Done) {
		resolver, err := newSimpleDNSResolver("127.0.0.1", "5653", "192.168.0.1/32")
		Expect(err).ToNot(HaveOccurred())

		ip1, err := resolver.allocateIP()
		Expect(err).ToNot(HaveOccurred())

		ip2, err := resolver.allocateIP()
		Expect(err).To(HaveOccurred())

		err = resolver.freeIP(ip1)
		Expect(err).ToNot(HaveOccurred())

		err = resolver.freeIP(ip2)
		Expect(err).To(HaveOccurred())

		// ready
		close(done)
	})

	It("IP allocate 2^16 addresses", func(done Done) {
		resolver, err := newSimpleDNSResolver("127.0.0.1", "5653", "240.0.0.0/4")
		Expect(err).ToNot(HaveOccurred())

		for i := 0; i < math.MaxInt16; i++ {
			_, err := resolver.allocateIP()
			Expect(err).ToNot(HaveOccurred())
		}

		// ready
		close(done)
	})
})
