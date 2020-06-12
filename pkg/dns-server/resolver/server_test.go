package resolver_test

import (
	"fmt"
	"strconv"

	"github.com/Kong/kuma/pkg/test"

	"github.com/miekg/dns"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/Kong/kuma/pkg/dns-server/resolver"
)

var _ = Describe("DNS server", func() {

	It("DNS Server basic functionality", func() {
		// setup
		resolver, err := NewSimpleDNSResolver("mesh", "127.0.0.1", "5653", "240.0.0.0/4")
		Expect(err).ToNot(HaveOccurred())

		// given
		_, err = resolver.AddService("service")
		Expect(err).ToNot(HaveOccurred())

		ip, err := resolver.ForwardLookup("service.mesh")
		Expect(err).ToNot(HaveOccurred())

		// when
		service, err := resolver.ReverseLookup(ip)

		// then
		Expect(err).ToNot(HaveOccurred())
		// and
		Expect(service).To(Equal("service.mesh"))
	})

	It("DNS Server network functionality", func() {
		// setup
		p, err := test.GetFreePort()
		Expect(err).ToNot(HaveOccurred())
		port := strconv.Itoa(p)

		resolver, err := NewSimpleDNSResolver("mesh", "127.0.0.1", port, "240.0.0.0/4")
		Expect(err).ToNot(HaveOccurred())

		// given
		_, err = resolver.AddService("service")
		Expect(err).ToNot(HaveOccurred())

		ip, err := resolver.ForwardLookup("service.mesh")
		Expect(err).ToNot(HaveOccurred())

		stop := make(chan struct{})
		go func() {
			err := resolver.Start(stop)
			Expect(err).ToNot(HaveOccurred())
		}()

		// when
		client := new(dns.Client)
		message := new(dns.Msg)
		_ = message.SetQuestion("service.mesh.", dns.TypeA)
		var response *dns.Msg
		Eventually(func() error {
			response, _, err = client.Exchange(message, "127.0.0.1:"+port)
			return err
		}).ShouldNot(HaveOccurred())
		// then
		Expect(response.Answer[0].String()).To(Equal(fmt.Sprintf("service.mesh.\t60\tIN\tA\t%s", ip)))

		// when
		message = new(dns.Msg)
		_ = message.SetQuestion("backend.mesh.", dns.TypeA)
		response, _, err = client.Exchange(message, "127.0.0.1:"+port)
		// then
		Expect(err).ToNot(HaveOccurred())
		// and
		Expect(len(response.Answer)).To(Equal(0))

		// stop the resolver
		stop <- struct{}{}
	})

	It("DNS Server service operation", func() {
		// given
		resolver, err := NewSimpleDNSResolver("mesh", "127.0.0.1", "5653", "240.0.0.0/4")
		Expect(err).ToNot(HaveOccurred())

		// when
		_, err = resolver.AddService("service")
		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		_, err = resolver.AddService("backend")
		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		err = resolver.RemoveService("service")
		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		err = resolver.RemoveService("backend")
		// then
		Expect(err).ToNot(HaveOccurred())
	})

	It("DNS Server sync operation", func() {
		// setup
		resolver, err := NewSimpleDNSResolver("mesh", "127.0.0.1", "5653", "240.0.0.0/4")
		Expect(err).ToNot(HaveOccurred())

		services := map[string]bool{
			"example-one.kuma-test.svc:80":   true,
			"example-two.kuma-test.svc:80":   true,
			"example-three.kuma-test.svc:80": true,
			"example-four.kuma-test.svc:80":  true,
			"example-five.kuma-test.svc:80":  true,
		}

		// given
		err = resolver.SyncServices(services)
		Expect(err).ToNot(HaveOccurred())

		// when
		_, err = resolver.ForwardLookup("example-one.mesh")
		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		_, err = resolver.ForwardLookup("example-five.mesh")
		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		_, err = resolver.ForwardLookup("example-five.other")
		// then
		Expect(err).To(HaveOccurred())

		// given
		delete(services, "example-five.kuma-test.svc:80")

		// when
		err = resolver.SyncServices(services)
		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		_, err = resolver.ForwardLookup("example-five.mesh")
		// then
		Expect(err).To(HaveOccurred())

		// when
		err = resolver.SyncServices(map[string]bool{})
		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		_, err = resolver.ForwardLookup("example-five.mesh")
		// then
		Expect(err).To(HaveOccurred())
	})
})
