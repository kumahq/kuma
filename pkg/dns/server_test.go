package dns_test

import (
	"fmt"

	"github.com/miekg/dns"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/Kong/kuma/pkg/dns"
	"github.com/Kong/kuma/pkg/test"
)

var _ = Describe("DNS server", func() {

	Describe("Network Operation", func() {
		var ip string
		var port uint32
		stop := make(chan struct{})
		done := make(chan struct{})
		BeforeEach(func() {
			// setup
			p, err := test.GetFreePort()
			port = uint32(p)
			Expect(err).ToNot(HaveOccurred())

			resolver := NewDNSResolver("mesh")
			server := NewDNSServer(port, resolver)
			resolver.SetVIPs(map[string]string{
				"service": "240.0.0.1",
			})

			// given
			ip, err = resolver.ForwardLookupFQDN("service.mesh")
			Expect(err).ToNot(HaveOccurred())

			go func() {
				err := server.Start(stop)
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
				response, _, err = client.Exchange(message, fmt.Sprintf("127.0.0.1:%d", port))
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
						response, _, err = client.Exchange(message, fmt.Sprintf("127.0.0.1:%d", port))
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
				response, _, err = client.Exchange(message, fmt.Sprintf("127.0.0.1:%d", port))
				return err
			}).ShouldNot(HaveOccurred())
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(len(response.Answer)).To(Equal(0))
		})
	})
})
