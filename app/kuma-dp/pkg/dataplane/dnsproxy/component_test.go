package dnsproxy_test

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/miekg/dns"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus"
	io_prometheus "github.com/prometheus/client_model/go"

	"github.com/kumahq/kuma/v2/app/kuma-dp/pkg/dataplane/dnsproxy"
	"github.com/kumahq/kuma/v2/pkg/test"
)

var _ = Describe("components", func() {
	var done chan struct{}
	var wg sync.WaitGroup
	var server *dnsproxy.Server
	var address string
	var mock atomic.Pointer[func(*dns.Msg) (*dns.Msg, error)]
	var registry *prometheus.Registry

	BeforeEach(func() {
		port, err := test.FindFreePort("127.0.0.1")
		Expect(err).ToNot(HaveOccurred())

		done = make(chan struct{})
		wg = sync.WaitGroup{}
		wg.Add(1)
		registry = prometheus.NewRegistry()

		address = net.JoinHostPort("127.0.0.1", strconv.Itoa(int(port)))
		server = dnsproxy.NewServerWithCustomClient([]string{address}, func(msg *dns.Msg) (*dns.Msg, error) {
			defer GinkgoRecover()
			f := *mock.Load()
			return f(msg)
		}, dnsproxy.WithRegisterer(registry))

		go func() {
			defer GinkgoRecover()
			defer wg.Done()
			err := server.Start(done)
			Expect(err).ToNot(HaveOccurred())
		}()
		server.WaitForReady()
	})
	AfterEach(func() {
		close(done)
		wg.Wait()
	})
	It("works with no local table", func() {
		f := func(req *dns.Msg) (*dns.Msg, error) { //nolint:unparam
			response := new(dns.Msg)
			response.SetRcode(req, dns.RcodeSuccess)
			response.Authoritative = true
			response.Answer = []dns.RR{
				&dns.A{Hdr: dns.RR_Header{Name: req.Question[0].Name, Ttl: uint32(123), Rrtype: dns.TypeA, Class: dns.ClassINET}, A: net.ParseIP("17.0.0.1")},
			}
			return response, nil
		}
		mock.Store(&f)
		msg := &dns.Msg{}
		msg.SetQuestion("example.com.", dns.TypeA)
		msg.RecursionAvailable = true

		c := new(dns.Client)
		res, _, err := c.Exchange(msg, address)
		Expect(err).ToNot(HaveOccurred())
		Expect(res.Answer[0].String()).To(ContainSubstring("17.0.0.1"))
	})
	It("failing upstream", func() {
		f := func(req *dns.Msg) (*dns.Msg, error) { //nolint:unparam
			response := new(dns.Msg)
			response.SetRcode(req, dns.RcodeServerFailure)
			response.Authoritative = true
			response.Answer = []dns.RR{}
			return response, nil
		}
		mock.Store(&f)
		msg := &dns.Msg{}
		msg.SetQuestion("example.com.", dns.TypeA)
		msg.RecursionAvailable = true

		c := new(dns.Client)
		res, _, err := c.Exchange(msg, address)
		Expect(err).ToNot(HaveOccurred())
		Expect(res).To(HaveField("Rcode", Equal(dns.RcodeServerFailure)))
	})
	It("failing network", func() {
		f := func(req *dns.Msg) (*dns.Msg, error) {
			return nil, fmt.Errorf("some error")
		}
		mock.Store(&f)
		msg := &dns.Msg{}
		msg.SetQuestion("example.com.", dns.TypeA)
		msg.RecursionAvailable = true

		c := new(dns.Client)
		res, _, err := c.Exchange(msg, address)
		Expect(err).ToNot(HaveOccurred())
		Expect(res).To(HaveField("Rcode", Equal(dns.RcodeServerFailure)))
	})
	It("hit the map", func() {
		f := func(req *dns.Msg) (*dns.Msg, error) { //nolint:unparam
			if req.Question[0].Name == "example.com." {
				Fail("should never call upstream for something in the map")
			}
			response := new(dns.Msg)
			response.SetRcode(req, dns.RcodeSuccess)
			response.Authoritative = true
			response.Answer = []dns.RR{
				&dns.A{Hdr: dns.RR_Header{Name: req.Question[0].Name, Ttl: uint32(125), Rrtype: dns.TypeA, Class: dns.ClassINET}, A: net.ParseIP("17.0.0.1")},
			}
			return response, nil
		}
		mock.Store(&f)

		Expect(server.ReloadMap(context.Background(), bytes.NewBuffer([]byte(`{
"ttl": 123,
"records": [{"name": "example.com", "ips": ["240.0.0.1", "::2"]}],
"extraLabels": {"mesh": "default", "kuma_workload": "backend"}
}`)))).To(Succeed())
		c := new(dns.Client)

		By("in the map A")
		msg := &dns.Msg{}
		msg.SetQuestion("example.com.", dns.TypeA)
		msg.RecursionAvailable = true
		res, _, err := c.Exchange(msg, address)
		Expect(err).ToNot(HaveOccurred())
		Expect(res).To(HaveField("Rcode", Equal(dns.RcodeSuccess)))
		Expect(res.Answer[0].String()).To(Equal("example.com.\t123\tIN\tA\t240.0.0.1"))

		By("in the map AAAA")
		msg = &dns.Msg{}
		msg.SetQuestion("example.com.", dns.TypeAAAA)
		msg.RecursionAvailable = true
		res, _, err = c.Exchange(msg, address)
		Expect(err).ToNot(HaveOccurred())
		Expect(res).To(HaveField("Rcode", Equal(dns.RcodeSuccess)))
		Expect(res.Answer[0].String()).To(Equal("example.com.\t123\tIN\tAAAA\t::2"))

		By("not in the map")
		msg = &dns.Msg{}
		msg.SetQuestion("foo.com.", dns.TypeA)
		msg.RecursionAvailable = true
		res, _, err = c.Exchange(msg, address)
		Expect(err).ToNot(HaveOccurred())
		Expect(res).To(HaveField("Rcode", Equal(dns.RcodeSuccess)))
		Expect(res.Answer[0].String()).To(Equal("foo.com.\t125\tIN\tA\t17.0.0.1"))
	})
	It("metrics have extra labels after ReloadMap", func() {
		f := func(req *dns.Msg) (*dns.Msg, error) {
			return nil, fmt.Errorf("should not call upstream for local entry")
		}
		mock.Store(&f)

		Expect(server.ReloadMap(context.Background(), bytes.NewBuffer([]byte(`{
"ttl": 123,
"records": [{"name": "example.com", "ips": ["240.0.0.1"]}],
"extraLabels": {"mesh": "default", "kuma_workload": "backend", "k8s_kuma_io_namespace": "test-ns"}
}`)))).To(Succeed())

		msg := &dns.Msg{}
		msg.SetQuestion("example.com.", dns.TypeA)
		c := new(dns.Client)
		_, _, err := c.Exchange(msg, address)
		Expect(err).ToNot(HaveOccurred())

		families, err := registry.Gather()
		Expect(err).ToNot(HaveOccurred())

		findMetric := func(name string) *io_prometheus.MetricFamily {
			for _, mf := range families {
				if mf.GetName() == name {
					return mf
				}
			}
			return nil
		}

		reqDuration := findMetric("kuma_dp_dns_request_duration_seconds")
		Expect(reqDuration).ToNot(BeNil())
		labels := reqDuration.GetMetric()[0].GetLabel()
		labelMap := map[string]string{}
		for _, l := range labels {
			labelMap[l.GetName()] = l.GetValue()
		}
		Expect(labelMap).To(Equal(map[string]string{
			"mesh":                  "default",
			"kuma_workload":         "backend",
			"k8s_kuma_io_namespace": "test-ns",
		}))
	})
})
