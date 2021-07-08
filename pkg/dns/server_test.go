package dns_test

import (
	"fmt"
	"os"
	"runtime"

	"github.com/kumahq/kuma/pkg/dns/vips"

	"github.com/miekg/dns"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/dns/resolver"

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
		var dnsResolver resolver.DNSResolver

		BeforeEach(func() {
			// setup
			p, err := test.GetFreePort()
			port = uint32(p)
			Expect(err).ToNot(HaveOccurred())

			dnsResolver = resolver.NewDNSResolver("mesh")
			m, err := core_metrics.NewMetrics("Standalone")
			metrics = m
			Expect(err).ToNot(HaveOccurred())
			server, err := NewDNSServer(port, dnsResolver, metrics, DnsNameToKumaCompliant)
			Expect(err).ToNot(HaveOccurred())

			// given

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
			// given
			var err error
			dnsResolver.SetVIPs(vips.List{
				vips.NewServiceEntry("service"): "240.0.0.1",
			})
			ip, err = dnsResolver.ForwardLookupFQDN("service.mesh")
			Expect(err).ToNot(HaveOccurred())

			// when
			client := new(dns.Client)
			message := new(dns.Msg)
			_ = message.SetQuestion("service.mesh.", dns.TypeA)
			var response *dns.Msg
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
			// given
			dnsResolver.SetVIPs(vips.List{
				vips.NewServiceEntry("service"): "240.0.0.1",
			})
			ip, err := dnsResolver.ForwardLookupFQDN("service.mesh")
			Expect(err).ToNot(HaveOccurred())

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

		It("should resolve IPv6 concurrent", func() {
			// given
			dnsResolver.SetVIPs(vips.List{
				vips.NewServiceEntry("service"): "fd00::1",
			})
			ip, err := dnsResolver.ForwardLookupFQDN("service.mesh")
			Expect(err).ToNot(HaveOccurred())

			resolved := make(chan struct{})
			for i := 0; i < 100; i++ {
				go func() {
					// when
					client := new(dns.Client)
					message := new(dns.Msg)
					_ = message.SetQuestion("service.mesh.", dns.TypeAAAA)
					var response *dns.Msg
					var err error
					Eventually(func() error {
						response, _, err = client.Exchange(message, fmt.Sprintf("127.0.0.1:%d", port))
						return err
					}).ShouldNot(HaveOccurred())
					// then
					Expect(response.Answer[0].String()).To(Equal(fmt.Sprintf("service.mesh.\t60\tIN\tAAAA\t%s", ip)))
					resolved <- struct{}{}
				}()
			}

			for i := 0; i < 100; i++ {
				<-resolved
			}
		})

		It("should not resolve", func() {
			// given
			var err error
			dnsResolver.SetVIPs(vips.List{
				vips.NewServiceEntry("service"): "240.0.0.1",
			})
			ip, err = dnsResolver.ForwardLookupFQDN("service.mesh")
			Expect(err).ToNot(HaveOccurred())

			// when
			client := new(dns.Client)
			message := new(dns.Msg)
			_ = message.SetQuestion("backend.mesh.", dns.TypeA)
			var response *dns.Msg
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

		It("should not resolve when no vips", func() {
			// given
			dnsResolver.SetVIPs(vips.List{})

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
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(len(response.Answer)).To(Equal(0))

			// and metrics are published
			Expect(test_metrics.FindMetric(metrics, "dns_server_resolution", "result", "unresolved").Counter.GetValue()).To(Equal(1.0))
		})

		It("should resolve services with '.'", func() {
			// given
			var err error
			dnsResolver.SetVIPs(vips.List{
				vips.NewServiceEntry("my.service"): "240.0.0.1",
			})
			ip, err = dnsResolver.ForwardLookupFQDN("my.service.mesh")
			Expect(err).ToNot(HaveOccurred())

			// when
			client := new(dns.Client)
			message := new(dns.Msg)
			_ = message.SetQuestion("my.service.mesh.", dns.TypeA)
			var response *dns.Msg
			Eventually(func() error {
				response, _, err = client.Exchange(message, fmt.Sprintf("127.0.0.1:%d", port))
				return err
			}).ShouldNot(HaveOccurred())

			// then
			Expect(response.Answer[0].String()).To(Equal(fmt.Sprintf("my.service.mesh.\t60\tIN\tA\t%s", ip)))

			// and metrics are published
			Expect(test_metrics.FindMetric(metrics, "dns_server")).ToNot(BeNil())
			Expect(test_metrics.FindMetric(metrics, "dns_server_resolution", "result", "resolved").Counter.GetValue()).To(Equal(1.0))
		})

		It("should resolve converted services with '.'", func() {
			// given
			var err error
			dnsResolver.SetVIPs(vips.List{
				vips.NewServiceEntry("my-service_test-namespace_svc_80"): "240.0.0.1",
			})
			ip, err = dnsResolver.ForwardLookupFQDN("my-service_test-namespace_svc_80.mesh")
			Expect(err).ToNot(HaveOccurred())

			// when
			client := new(dns.Client)
			message := new(dns.Msg)
			_ = message.SetQuestion("my-service.test-namespace.svc.80.mesh.", dns.TypeA)
			var response *dns.Msg
			Eventually(func() error {
				response, _, err = client.Exchange(message, fmt.Sprintf("127.0.0.1:%d", port))
				return err
			}).ShouldNot(HaveOccurred())

			// then
			Expect(response.Answer[0].String()).To(Equal(fmt.Sprintf("my-service.test-namespace.svc.80.mesh.\t60\tIN\tA\t%s", ip)))

			// and metrics are published
			Expect(test_metrics.FindMetric(metrics, "dns_server")).ToNot(BeNil())
			Expect(test_metrics.FindMetric(metrics, "dns_server_resolution", "result", "resolved").Counter.GetValue()).To(Equal(1.0))
		})
	})

	Describe("host operation", func() {
		It("should fail to bind to a privileged port", func() {

			if runtime.GOOS != "linux" || os.Geteuid() == 0 {
				// this test will pass only on Linux when not run as root
				return
			}

			// setup
			port := uint32(53)
			stop := make(chan struct{})
			defer close(stop)

			// given
			dnsResolver := resolver.NewDNSResolver("mesh")
			metrics, err := core_metrics.NewMetrics("Standalone")
			Expect(err).ToNot(HaveOccurred())
			server, err := NewDNSServer(port, dnsResolver, metrics, DnsNameToKumaCompliant)
			Expect(err).ToNot(HaveOccurred())

			err = server.Start(stop)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unable to bind the DNS server to 0.0.0.0:53"))
		}, 1)
	})

})
