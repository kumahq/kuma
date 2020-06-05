package resolver

import (
	"fmt"

	"github.com/miekg/dns"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DNS server", func() {

	It("DNS Server basic functionality", func(done Done) {
		resolver, err := NewSimpleDNSResolver("127.0.0.1", "5653", "240.0.0.0/4")
		Expect(err).ToNot(HaveOccurred())

		err = resolver.AddDomain(".kuma")
		Expect(err).ToNot(HaveOccurred())

		_, err = resolver.AddServiceToDomain("service", "kuma")
		Expect(err).ToNot(HaveOccurred())

		ip, err := resolver.ForwardLookup("service.kuma")
		Expect(err).ToNot(HaveOccurred())

		service, err := resolver.ReverseLookup(ip)
		Expect(err).ToNot(HaveOccurred())
		Expect(service).To(Equal("service.kuma"))

		// ready
		close(done)
	})

	It("DNS Server network functionality", func(done Done) {
		resolver, err := NewSimpleDNSResolver("127.0.0.1", "5653", "240.0.0.0/4")
		Expect(err).ToNot(HaveOccurred())

		err = resolver.AddDomain(".kuma")
		Expect(err).ToNot(HaveOccurred())

		_, err = resolver.AddServiceToDomain("service", "kuma")
		Expect(err).ToNot(HaveOccurred())

		ip, err := resolver.ForwardLookup("service.kuma")
		Expect(err).ToNot(HaveOccurred())

		stop := make(chan struct{})
		go func() {
			err := resolver.Start(stop)
			Expect(err).ToNot(HaveOccurred())
		}()

		client := new(dns.Client)
		message := new(dns.Msg)
		_ = message.SetQuestion("service.kuma.", dns.TypeA)
		var response *dns.Msg
		Eventually(func() error {
			response, _, err = client.Exchange(message, "127.0.0.1:5653")
			return err
		}).ShouldNot(HaveOccurred())
		Expect(response.Answer[0].String()).To(Equal(fmt.Sprintf("service.kuma.\t3600\tIN\tA\t%s", ip)))

		message = new(dns.Msg)
		_ = message.SetQuestion("backend.kuma.", dns.TypeA)
		response, _, err = client.Exchange(message, "127.0.0.1:5653")
		Expect(err).ToNot(HaveOccurred())
		Expect(len(response.Answer)).To(Equal(0))

		// stop the resolver
		stop <- struct{}{}

		// ready
		close(done)
	})

	It("DNS Server domain operation", func(done Done) {
		resolver, err := NewSimpleDNSResolver("127.0.0.1", "5653", "240.0.0.0/4")
		Expect(err).ToNot(HaveOccurred())

		err = resolver.AddDomain("")
		Expect(err).To(HaveOccurred())

		err = resolver.AddDomain("kuma")
		Expect(err).ToNot(HaveOccurred())

		err = resolver.AddDomain(".kuma")
		Expect(err).ToNot(HaveOccurred())

		err = resolver.AddDomain(".other")
		Expect(err).ToNot(HaveOccurred())

		err = resolver.AddDomain(".third")
		Expect(err).ToNot(HaveOccurred())

		err = resolver.AddDomain(".kuma")
		Expect(err).ToNot(HaveOccurred())

		err = resolver.RemoveDomain(".kuma")
		Expect(err).ToNot(HaveOccurred())

		err = resolver.RemoveDomain(".other")
		Expect(err).ToNot(HaveOccurred())

		err = resolver.RemoveDomain(".other")
		Expect(err).To(HaveOccurred())

		// ready
		close(done)
	})

	It("DNS Server service operation", func(done Done) {
		resolver, err := NewSimpleDNSResolver("127.0.0.1", "5653", "240.0.0.0/4")
		Expect(err).ToNot(HaveOccurred())

		err = resolver.AddDomain("kuma")
		Expect(err).ToNot(HaveOccurred())

		err = resolver.AddDomain("other")
		Expect(err).ToNot(HaveOccurred())

		_, err = resolver.AddServiceToDomain("service", "third")
		Expect(err).To(HaveOccurred())

		_, err = resolver.AddServiceToDomain("service", "kuma")
		Expect(err).ToNot(HaveOccurred())

		_, err = resolver.AddServiceToDomain("backend", "kuma")
		Expect(err).ToNot(HaveOccurred())

		_, err = resolver.AddServiceToDomain("service", "kuma")
		Expect(err).ToNot(HaveOccurred())

		err = resolver.RemoveServiceFromDomain("service", "kuma")
		Expect(err).ToNot(HaveOccurred())

		err = resolver.RemoveServiceFromDomain("service", "kuma")
		Expect(err).To(HaveOccurred())

		err = resolver.RemoveServiceFromDomain("service", "third")
		Expect(err).To(HaveOccurred())

		// ready
		close(done)
	})

	It("DNS Server sync operation", func(done Done) {
		resolver, err := NewSimpleDNSResolver("127.0.0.1", "5653", "240.0.0.0/4")
		Expect(err).ToNot(HaveOccurred())

		err = resolver.AddDomain("kuma")
		Expect(err).ToNot(HaveOccurred())

		err = resolver.AddDomain("other")
		Expect(err).ToNot(HaveOccurred())

		services := map[string]bool{
			"one":   true,
			"two":   true,
			"three": true,
			"four":  true,
			"five":  true,
		}

		err = resolver.SyncServicesForDomain(services, "kuma")
		Expect(err).ToNot(HaveOccurred())

		err = resolver.SyncServicesForDomain(services, "other")
		Expect(err).ToNot(HaveOccurred())

		err = resolver.SyncServicesForDomain(services, "third")
		Expect(err).To(HaveOccurred())

		_, err = resolver.ForwardLookup("one.kuma")
		Expect(err).ToNot(HaveOccurred())

		_, err = resolver.ForwardLookup("five.kuma")
		Expect(err).ToNot(HaveOccurred())

		_, err = resolver.ForwardLookup("five.other")
		Expect(err).ToNot(HaveOccurred())

		delete(services, "five")

		err = resolver.SyncServicesForDomain(services, "kuma")
		Expect(err).ToNot(HaveOccurred())

		_, err = resolver.ForwardLookup("five.kuma")
		Expect(err).To(HaveOccurred())

		_, err = resolver.ForwardLookup("five.other")
		Expect(err).ToNot(HaveOccurred())

		err = resolver.SyncServicesForDomain(map[string]bool{}, "other")
		Expect(err).ToNot(HaveOccurred())

		_, err = resolver.ForwardLookup("five.other")
		Expect(err).To(HaveOccurred())

		// ready
		close(done)
	})
})
