package dns_test

import (
	"github.com/kumahq/kuma/pkg/core/dns"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net"
	"time"
)

var _ = Describe("DNS with cache", func() {
	var counter int
	var table map[string][]net.IP
	lookupFunc := func(host string) ([]net.IP, error) {
		counter++
		return table[host], nil
	}
	var cachingLookupFunc dns.LookupIPFunc

	BeforeEach(func() {
		cachingLookupFunc = dns.MakeCaching(lookupFunc, 1*time.Second)
		table = map[string][]net.IP{}
		counter = 0
	})

	It("should use cache on the second call", func() {
		_, _ = cachingLookupFunc("example.com")
		_, _ = cachingLookupFunc("example.com")
		Expect(counter).To(Equal(1))
	})

	It("should avoid cache after TTL", func() {
		table["example.com"] = []net.IP{net.ParseIP("192.168.0.1")}

		ip, _ := cachingLookupFunc("example.com")
		Expect(ip[0]).To(Equal(net.ParseIP("192.168.0.1")))

		ip, _ = cachingLookupFunc("example.com")
		Expect(ip[0]).To(Equal(net.ParseIP("192.168.0.1")))

		table["example.com"] = []net.IP{net.ParseIP("10.20.0.1")}
		<-time.After(2 * time.Second)
		ip, _ = cachingLookupFunc("example.com")
		Expect(ip[0]).To(Equal(net.ParseIP("10.20.0.1")))
		Expect(counter).To(Equal(2))
	})
})
