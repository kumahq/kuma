package dns_test

import (
	"fmt"
	"os"
	"runtime"

	"github.com/miekg/dns"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/pkg/dns"
	"github.com/kumahq/kuma/pkg/dns/resolver"
	"github.com/kumahq/kuma/pkg/dns/vips"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/test"
	test_metrics "github.com/kumahq/kuma/pkg/test/metrics"
)

var _ = Describe("DNS server", func() {

	Describe("Network Operation", func() {
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

		type dnsTestCase struct {
			givenVips map[vips.HostnameEntry]string
			whenQuery string
			whenType  dns.Type
			// If ip is empty this will be consider as a dns miss
			thenIp string
		}
		DescribeTable("simple point resolve",
			func(tc dnsTestCase) {
				// given
				var err error
				dnsResolver.SetVIPs(tc.givenVips)

				// when
				client := new(dns.Client)
				message := new(dns.Msg)
				q := tc.whenQuery + "."
				_ = message.SetQuestion(q, uint16(tc.whenType))
				var response *dns.Msg
				Eventually(func() error {
					response, _, err = client.Exchange(message, fmt.Sprintf("127.0.0.1:%d", port))
					return err
				}).ShouldNot(HaveOccurred())

				// then
				Expect(test_metrics.FindMetric(metrics, "dns_server")).ToNot(BeNil())
				if tc.thenIp != "" {
					Expect(response.Answer).To(HaveLen(1))
					Expect(response.Answer[0].String()).To(Or(
						Equal(fmt.Sprintf("%s\t60\tIN\tA\t%s", q, tc.thenIp)),
						Equal(fmt.Sprintf("%s\t60\tIN\tAAAA\t%s", q, tc.thenIp)),
					))
					// and metrics are published
					Expect(test_metrics.FindMetric(metrics, "dns_server_resolution", "result", "resolved").Counter.GetValue()).To(Equal(1.0))
				} else {
					Expect(response.Answer).To(HaveLen(0))
					// and metrics are published
					Expect(test_metrics.FindMetric(metrics, "dns_server_resolution", "result", "unresolved").Counter.GetValue()).To(Equal(1.0))
				}
			},
			Entry("A should resolve", dnsTestCase{
				givenVips: map[vips.HostnameEntry]string{vips.NewServiceEntry("service"): "240.0.0.1"},
				whenQuery: "service.mesh",
				whenType:  dns.Type(dns.TypeA),
				thenIp:    "240.0.0.1",
			}),
			Entry("AAAA with v4 should add prefix", dnsTestCase{
				givenVips: map[vips.HostnameEntry]string{vips.NewServiceEntry("service"): "240.0.0.1"},
				whenQuery: "service.mesh",
				whenType:  dns.Type(dns.TypeAAAA),
				thenIp:    "240.0.0.1",
			}),
			Entry("AAAA with v6 should succeed", dnsTestCase{
				givenVips: map[vips.HostnameEntry]string{vips.NewServiceEntry("service"): "2001:db8::ff00:42:8329"},
				whenQuery: "service.mesh",
				whenType:  dns.Type(dns.TypeAAAA),
				thenIp:    "2001:db8::ff00:42:8329",
			}),
			Entry("should not resolve with no vips", dnsTestCase{
				givenVips: map[vips.HostnameEntry]string{},
				whenQuery: "service.mesh",
				whenType:  dns.Type(dns.TypeA),
				thenIp:    "",
			}),
			Entry("should not resolve", dnsTestCase{
				givenVips: map[vips.HostnameEntry]string{vips.NewServiceEntry("service"): "240.0.0.1"},
				whenQuery: "not-service.mesh",
				whenType:  dns.Type(dns.TypeA),
				thenIp:    "",
			}),
			Entry("should resolve services with '.'", dnsTestCase{
				givenVips: map[vips.HostnameEntry]string{vips.NewServiceEntry("my.service"): "240.0.0.1"},
				whenQuery: "my.service.mesh",
				whenType:  dns.Type(dns.TypeA),
				thenIp:    "240.0.0.1",
			}),
			Entry("should resolve converted services with '.'", dnsTestCase{
				givenVips: map[vips.HostnameEntry]string{vips.NewServiceEntry("my-service_test-namespace_svc_80"): "240.0.0.1"},
				whenQuery: "my-service.test-namespace.svc.80.mesh",
				whenType:  dns.Type(dns.TypeA),
				thenIp:    "240.0.0.1",
			}),
			Entry("should resolve fqdn service with .mesh", dnsTestCase{
				givenVips: map[vips.HostnameEntry]string{vips.NewFqdnEntry("my.service.foo.mesh"): "240.0.0.1"},
				whenQuery: "my.service.foo.mesh",
				whenType:  dns.Type(dns.TypeA),
				thenIp:    "240.0.0.1",
			}),
			Entry("should resolve, service entry has priority over fqdn entry", dnsTestCase{
				givenVips: map[vips.HostnameEntry]string{
					vips.NewFqdnEntry("my.service.foo.mesh"): "240.0.0.2",
					vips.NewServiceEntry("my.service.foo"):   "240.0.0.1",
				},
				whenQuery: "my.service.foo.mesh",
				whenType:  dns.Type(dns.TypeA),
				thenIp:    "240.0.0.1",
			}),
			Entry("should resolve, simple fqdn entry", dnsTestCase{
				givenVips: map[vips.HostnameEntry]string{
					vips.NewFqdnEntry("service.com.mesh"): "240.0.0.2",
				},
				whenQuery: "service.com.mesh",
				whenType:  dns.Type(dns.TypeA),
				thenIp:    "240.0.0.2",
			}),
		)

		It("should resolve concurrent", func() {
			// given
			dnsResolver.SetVIPs(map[vips.HostnameEntry]string{
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
			dnsResolver.SetVIPs(map[vips.HostnameEntry]string{
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
		})
	})

})
