package resolver_test

import (
	"math"

	. "github.com/Kong/kuma/pkg/dns-server/resolver"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DNS ip", func() {

	It("IP allocate and free operations", func(done Done) {
		resolver := NewSimpleIPAM("192.168.0.1/32")
		Expect(resolver).ToNot(BeNil())

		ip1, err := resolver.AllocateIP()
		Expect(err).ToNot(HaveOccurred())

		ip2, err := resolver.AllocateIP()
		Expect(err).To(HaveOccurred())

		err = resolver.FreeIP(ip1)
		Expect(err).ToNot(HaveOccurred())

		err = resolver.FreeIP(ip2)
		Expect(err).To(HaveOccurred())

		// ready
		close(done)
	})

	It("IP allocate 2^16 addresses", func(done Done) {
		resolver := NewSimpleIPAM("240.0.0.0/4")
		Expect(resolver).ToNot(BeNil())

		for i := 0; i < math.MaxInt16; i++ {
			_, err := resolver.AllocateIP()
			Expect(err).ToNot(HaveOccurred())
		}

		// ready
		close(done)
	})
})
