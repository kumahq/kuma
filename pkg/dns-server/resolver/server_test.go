package resolver_test

import (
	"fmt"
	"strconv"

	"github.com/miekg/dns"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/Kong/kuma/pkg/dns-server/resolver"
	"github.com/Kong/kuma/pkg/test"
)

var _ = Describe("DNS server", func() {

	Describe("Network Operation", func() {

		var ip, port string
		stop := make(chan struct{})
		done := make(chan struct{})
		BeforeEach(func() {
			// setup
			p, err := test.GetFreePort()
			Expect(err).ToNot(HaveOccurred())
			port = strconv.Itoa(p)

			resolver, err := NewSimpleDNSResolver("mesh", "127.0.0.1", port, "240.0.0.0/4")
			Expect(err).ToNot(HaveOccurred())

			// given
			_, err = resolver.AddService("service")
			Expect(err).ToNot(HaveOccurred())

			ip, err = resolver.ForwardLookupFQDN("service.mesh")
			Expect(err).ToNot(HaveOccurred())

			go func() {
				err := resolver.Start(stop)
				Expect(err).ToNot(HaveOccurred())
				done <- struct{}{}
			}()
		})

		AfterEach(func() {
			// stop the resolver
			stop <- struct{}{}
			<-done
		})

		It("Should resolve", func() {
			// when
			client := new(dns.Client)
			message := new(dns.Msg)
			_ = message.SetQuestion("service.mesh.", dns.TypeA)
			var response *dns.Msg
			var err error
			Eventually(func() error {
				response, _, err = client.Exchange(message, "127.0.0.1:"+port)
				return err
			}).ShouldNot(HaveOccurred())
			// then
			Expect(response.Answer[0].String()).To(Equal(fmt.Sprintf("service.mesh.\t60\tIN\tA\t%s", ip)))
		})

		It("Should resolve concurrent", func() {
			resolved := make(chan struct{})
			for i := 0; i < 100; i++ {
				go func() {
					// when
					client := new(dns.Client)
					message := new(dns.Msg)
					_ = message.SetQuestion("service.mesh.", dns.TypeA)
					var response *dns.Msg
					var err error
					Eventually(func() error {
						response, _, err = client.Exchange(message, "127.0.0.1:"+port)
						return err
					}).ShouldNot(HaveOccurred())
					// then
					Expect(response.Answer[0].String()).To(Equal(fmt.Sprintf("service.mesh.\t60\tIN\tA\t%s", ip)))
					resolved <- struct{}{}
				}()
			}

			for i := 0; i < 100; i++ {
				<-resolved
			}
		})

		It("Should not resolve", func() {
			// when
			client := new(dns.Client)
			message := new(dns.Msg)
			_ = message.SetQuestion("backend.mesh.", dns.TypeA)
			var response *dns.Msg
			var err error
			Eventually(func() error {
				response, _, err = client.Exchange(message, "127.0.0.1:"+port)
				return err
			}).ShouldNot(HaveOccurred())
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(len(response.Answer)).To(Equal(0))
		})
	})

	It("DNS Server basic functionality", func() {
		// setup
		resolver, err := NewSimpleDNSResolver("mesh", "127.0.0.1", "5653", "240.0.0.0/4")
		Expect(err).ToNot(HaveOccurred())

		// given
		_, err = resolver.AddService("service")
		Expect(err).ToNot(HaveOccurred())

		ipService, err := resolver.ForwardLookup("service")
		Expect(err).ToNot(HaveOccurred())

		ipFQDN, err := resolver.ForwardLookupFQDN("service.mesh")
		Expect(err).ToNot(HaveOccurred())

		// when
		service, err := resolver.ReverseLookup(ipFQDN)

		// then
		Expect(err).ToNot(HaveOccurred())
		// and
		Expect(service).To(Equal("service.mesh"))
		// and
		Expect(ipService).To(Equal(ipFQDN))
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
			"example-one_kuma-test_svc_80":   true,
			"example-two_kuma-test_svc_80":   true,
			"example-three_kuma-test_svc_80": true,
			"example-four_kuma-test_svc_80":  true,
			"example-five_kuma-test_svc_80":  true,
		}

		// given
		err = resolver.SyncServices(services)
		Expect(err).ToNot(HaveOccurred())

		// when
		_, err = resolver.ForwardLookupFQDN("example-one_kuma-test_svc_80.mesh")
		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		_, err = resolver.ForwardLookupFQDN("example-five_kuma-test_svc_80.mesh")
		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		_, err = resolver.ForwardLookupFQDN("example-five.other")
		// then
		Expect(err).To(HaveOccurred())

		// given
		delete(services, "example-five_kuma-test_svc_80")

		// when
		err = resolver.SyncServices(services)
		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		_, err = resolver.ForwardLookupFQDN("example-five_kuma-test_svc_80.mesh")
		// then
		Expect(err).To(HaveOccurred())

		// when
		err = resolver.SyncServices(map[string]bool{})
		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		_, err = resolver.ForwardLookupFQDN("example-five_kuma-test_svc_80.mesh")
		// then
		Expect(err).To(HaveOccurred())
	})
})
