package dnsproxy_test

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"strconv"
	"sync"

	"github.com/miekg/dns"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/app/kuma-dp/pkg/dataplane/dnsproxy"
	"github.com/kumahq/kuma/pkg/test"
)

var _ = Describe("components", func() {
	var done chan struct{}
	var wg sync.WaitGroup
	var server *dnsproxy.Server
	var address string
	var mock func(*dns.Msg) (*dns.Msg, error)

	BeforeEach(func() {
		port, err := test.FindFreePort("127.0.0.1")
		Expect(err).ToNot(HaveOccurred())

		done = make(chan struct{})
		wg = sync.WaitGroup{}
		wg.Add(1)

		address = net.JoinHostPort("127.0.0.1", strconv.Itoa(int(port)))
		server = dnsproxy.NewServer(address)
		server.MockDNS(func(msg *dns.Msg) (*dns.Msg, error) {
			defer GinkgoRecover()
			return mock(msg)
		})

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
		mock = func(req *dns.Msg) (*dns.Msg, error) { // nolint:unparam
			response := new(dns.Msg)
			response.SetRcode(req, dns.RcodeSuccess)
			response.Authoritative = true
			response.Answer = []dns.RR{
				&dns.A{Hdr: dns.RR_Header{Name: req.Question[0].Name, Ttl: uint32(123), Rrtype: dns.TypeA, Class: dns.ClassINET}, A: net.ParseIP("17.0.0.1")},
			}
			return response, nil
		}
		msg := &dns.Msg{}
		msg.SetQuestion("example.com.", dns.TypeA)
		msg.RecursionAvailable = true

		c := new(dns.Client)
		res, _, err := c.Exchange(msg, address)
		Expect(err).ToNot(HaveOccurred())
		Expect(res.Answer[0].String()).To(ContainSubstring("17.0.0.1"))
	})
	It("failing upstream", func() {
		mock = func(req *dns.Msg) (*dns.Msg, error) { // nolint:unparam
			response := new(dns.Msg)
			response.SetRcode(req, dns.RcodeServerFailure)
			response.Authoritative = true
			response.Answer = []dns.RR{}
			return response, nil
		}
		msg := &dns.Msg{}
		msg.SetQuestion("example.com.", dns.TypeA)
		msg.RecursionAvailable = true

		c := new(dns.Client)
		res, _, err := c.Exchange(msg, address)
		Expect(err).ToNot(HaveOccurred())
		Expect(res).To(HaveField("Rcode", Equal(dns.RcodeServerFailure)))
	})
	It("failing network", func() {
		mock = func(req *dns.Msg) (*dns.Msg, error) {
			return nil, fmt.Errorf("some error")
		}
		msg := &dns.Msg{}
		msg.SetQuestion("example.com.", dns.TypeA)
		msg.RecursionAvailable = true

		c := new(dns.Client)
		res, _, err := c.Exchange(msg, address)
		Expect(err).ToNot(HaveOccurred())
		Expect(res).To(HaveField("Rcode", Equal(dns.RcodeServerFailure)))
	})
	It("hit the map", func() {
		mock = func(req *dns.Msg) (*dns.Msg, error) { // nolint:unparam
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

		_ = server.ReloadMap(context.Background(), bytes.NewBuffer([]byte(`{
"ttl": 123,
"records": [{"name": "example.com", "ips": ["240.0.0.1", "::2"]}]
}`)))
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
})
