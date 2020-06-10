package resolver_test

import (
	"fmt"

	"github.com/miekg/dns"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/Kong/kuma/pkg/dns-server/resolver"
)

var _ = Describe("DNS server", func() {

	It("DNS Server basic functionality", func(done Done) {
		resolver, err := NewSimpleDNSResolver("kuma", "127.0.0.1", "5653", "240.0.0.0/4")
		Expect(err).ToNot(HaveOccurred())

		_, err = resolver.AddService("service")
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
		resolver, err := NewSimpleDNSResolver("kuma", "127.0.0.1", "5653", "240.0.0.0/4")
		Expect(err).ToNot(HaveOccurred())

		_, err = resolver.AddService("service")
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

	It("DNS Server service operation", func(done Done) {
		resolver, err := NewSimpleDNSResolver("kuma", "127.0.0.1", "5653", "240.0.0.0/4")
		Expect(err).ToNot(HaveOccurred())

		_, err = resolver.AddService("service")
		Expect(err).ToNot(HaveOccurred())

		_, err = resolver.AddService("backend")
		Expect(err).ToNot(HaveOccurred())

		err = resolver.RemoveService("service")
		Expect(err).ToNot(HaveOccurred())

		err = resolver.RemoveService("backend")
		Expect(err).ToNot(HaveOccurred())

		// ready
		close(done)
	})

	It("DNS Server sync operation", func(done Done) {
		resolver, err := NewSimpleDNSResolver("kuma", "127.0.0.1", "5653", "240.0.0.0/4")
		Expect(err).ToNot(HaveOccurred())

		services := map[string]bool{
			"example-one.kuma-test.svc:80":   true,
			"example-two.kuma-test.svc:80":   true,
			"example-three.kuma-test.svc:80": true,
			"example-four.kuma-test.svc:80":  true,
			"example-five.kuma-test.svc:80":  true,
		}

		err = resolver.SyncServices(services)
		Expect(err).ToNot(HaveOccurred())

		_, err = resolver.ForwardLookup("example-one.kuma")
		Expect(err).ToNot(HaveOccurred())

		_, err = resolver.ForwardLookup("example-five.kuma")
		Expect(err).ToNot(HaveOccurred())

		_, err = resolver.ForwardLookup("example-five.other")
		Expect(err).To(HaveOccurred())

		delete(services, "example-five.kuma-test.svc:80")

		err = resolver.SyncServices(services)
		Expect(err).ToNot(HaveOccurred())

		_, err = resolver.ForwardLookup("example-five.kuma")
		Expect(err).To(HaveOccurred())

		err = resolver.SyncServices(map[string]bool{})
		Expect(err).ToNot(HaveOccurred())

		_, err = resolver.ForwardLookup("example-five.kuma")
		Expect(err).To(HaveOccurred())

		// ready
		close(done)
	})
})
