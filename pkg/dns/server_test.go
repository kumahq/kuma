package dns_test

import (
	"fmt"

	"github.com/miekg/dns"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/pkg/dns"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/test"
	test_metrics "github.com/kumahq/kuma/pkg/test/metrics"
)

var _ = Describe("DNS server", func() {

	Describe("Network Operation", func() {
		var ip string
		var port uint32
		stop := make(chan struct{})
		done := make(chan struct{})
		var metrics core_metrics.Metrics

		BeforeEach(func() {
			// setup
			p, err := test.GetFreePort()
			port = uint32(p)
			Expect(err).ToNot(HaveOccurred())

			resolver := NewDNSResolver("mesh")
			m, err := core_metrics.NewMetrics("Standalone")
			metrics = m
			Expect(err).ToNot(HaveOccurred())
			server, err := NewDNSServer(port, resolver, metrics)
			Expect(err).ToNot(HaveOccurred())
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

		It("should resolve", func() {
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

			// and metrics are published
			Expect(test_metrics.FindMetric(metrics, "dns_server")).ToNot(BeNil())
			Expect(test_metrics.FindMetric(metrics, "dns_server_resolution", "result", "resolved").Counter.GetValue()).To(Equal(1.0))
		})

		It("should resolve concurrent", func() {
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

		It("should not resolve", func() {
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

			// and metrics are published
			Expect(test_metrics.FindMetric(metrics, "dns_server_resolution", "result", "unresolved").Counter.GetValue()).To(Equal(1.0))
		})
	})
})
