package dnsproxy

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strings"
	"sync/atomic"
	"time"

	"github.com/miekg/dns"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/dns/dpapi"
)

var log = core.Log.WithName("dnsproxy")

type Server struct {
	// HostPort or unix domain socket for tests
	address       string
	componentDone chan struct{}
	dnsMap        atomic.Pointer[dnsMap]
	metrics       *metrics
	// upstreamClient is used for testing purposes
	upstreamClient func(msg *dns.Msg) (*dns.Msg, error)
	ready          chan struct{}
}

func NewServer(address string) (*Server, error) {
	config, err := dns.ClientConfigFromFile("/etc/resolv.conf")
	if err != nil {
		return nil, fmt.Errorf("failed to read DNS for /etc/resolv.conf %w", err)
	}
	if len(config.Servers) == 0 {
		return nil, fmt.Errorf("no server found in /etc/resolv.conf")
	}
	etcResolveHostPort := net.JoinHostPort(config.Servers[0], config.Port)
	core.Log.Info("TEST, show config from resolve", "resolv", config, "etcResolveHostPort", etcResolveHostPort)
	handler := func(msg *dns.Msg) (*dns.Msg, error) {
		client := new(dns.Client)
		response, _, err := client.Exchange(msg, etcResolveHostPort)
		if err != nil {
			return nil, fmt.Errorf("failed to write message to upstream %w", err)
		}
		return response, nil
	}
	return NewServerWithCustomClient(address, handler), nil
}

// NewServerWithCustomClient is used for testing purposes
func NewServerWithCustomClient(address string, upstreamClient func(msg *dns.Msg) (*dns.Msg, error)) *Server {
	s := Server{
		address:        address,
		componentDone:  make(chan struct{}),
		dnsMap:         atomic.Pointer[dnsMap]{},
		ready:          make(chan struct{}),
		metrics:        newMetrics(),
		upstreamClient: upstreamClient,
	}
	s.dnsMap.Store(&dnsMap{ARecords: make(map[string]*dnsEntry), AAAARecords: make(map[string]*dnsEntry)})
	return &s
}

type dnsMap struct {
	ARecords    map[string]*dnsEntry
	AAAARecords map[string]*dnsEntry
}

type dnsEntry struct {
	RCode uint16
	RR    []dns.RR
}

var _ component.GracefulComponent = &Server{}

func (s *Server) Handler(res dns.ResponseWriter, req *dns.Msg) {
	start := time.Now()
	defer func() {
		s.metrics.RequestDuration.Observe(time.Since(start).Seconds())
		if err := recover(); err != nil {
			log.Error(fmt.Errorf("handler panic %v", err), "panic in DNS handler")
			response := new(dns.Msg)
			response.SetRcode(req, dns.RcodeServerFailure)
			_ = res.WriteMsg(response)
		}
	}()
	var response *dns.Msg
	var dnsEntry *dnsEntry
	if len(req.Question) > 0 { // Apparently most DNS don't support multiple questions so let's just support the first one
		if len(req.Question) > 1 {
			log.Info("[WARNING] multiple questions in a single request, this is not supported", "questions", req.Question)
		}
		switch req.Question[0].Qtype {
		case dns.TypeA:
			// lookup in our DNS map
			dnsMap := s.dnsMap.Load()
			dnsEntry = dnsMap.ARecords[req.Question[0].Name]
		case dns.TypeAAAA:
			dnsMap := s.dnsMap.Load()
			dnsEntry = dnsMap.AAAARecords[req.Question[0].Name]
		}
		log.Info("got request", "type", req.Question[0].Qtype, "name", req.Question[0].Name, "entry", dnsEntry)
	}
	if dnsEntry != nil {
		response = new(dns.Msg)
		response.SetRcode(req, int(dnsEntry.RCode))
		response.Authoritative = true
		response.Answer = append(response.Answer, dnsEntry.RR...)
	} else {
		proxyStart := time.Now()
		resp, err := s.upstreamClient(req)
		if err != nil {
			s.metrics.UpstreamRequestFailureCount.Inc()
			log.Error(err, "failed to write message to upstream")
			response = new(dns.Msg)
			response.SetRcode(req, dns.RcodeServerFailure)
		} else {
			response = resp
		}
		s.metrics.UpstreamRequestDuration.Observe(time.Since(proxyStart).Seconds())
	}
	err := res.WriteMsg(response)
	if err != nil {
		log.Error(err, "failed to write upstreamResponse")
	}
}

func (s *Server) Start(stop <-chan struct{}) error {
	defer close(s.componentDone)
	srv := &dns.Server{Addr: s.address, Handler: dns.HandlerFunc(s.Handler), Net: "udp"}

	serverDone := make(chan error)
	srv.NotifyStartedFunc = func() {
		close(s.ready)
	}
	go func() {
		err := srv.ListenAndServe()
		serverDone <- err
	}()
	select {
	case <-stop:
		err := srv.Shutdown()
		if err != nil {
			log.Error(err, "server shutdown returned an error")
		}
	case err := <-serverDone:
		log.Info("[WARNING] server stopped with shutdown never called")
		if err != nil {
			return err
		}
	}
	err := <-serverDone
	log.Info("server shutdown complete")
	return err
}

func (s *Server) NeedLeaderElection() bool {
	return false
}

func (s *Server) WaitForDone() {
	<-s.componentDone
}

func (s *Server) WaitForReady() {
	<-s.ready
}

// ReloadMap replaces the current map in memory so that future calls to the proxy.
// Because this is called inside the configfetcher there's no need to check if it changed or not.
func (s *Server) ReloadMap(ctx context.Context, reader io.Reader) error {
	configuration := dpapi.DNSProxyConfig{}
	err := json.NewDecoder(reader).Decode(&configuration)
	if err != nil {
		return err
	}
	res := dnsMap{
		ARecords:    make(map[string]*dnsEntry),
		AAAARecords: make(map[string]*dnsEntry),
	}
	for _, record := range configuration.Records {
		if ctx.Err() != nil { // context was cancelled no need to keep reloading the map
			return ctx.Err()
		}
		n := record.Name + "."
		res.ARecords[n] = &dnsEntry{
			RCode: dns.RcodeSuccess,
			RR:    []dns.RR{},
		}
		res.AAAARecords[n] = &dnsEntry{
			RCode: dns.RcodeSuccess,
			RR:    []dns.RR{},
		}
		for _, ipStr := range record.IPs {
			ip := net.ParseIP(ipStr)
			if ip == nil {
				log.Info("[WARNING] invalid IP address", "ip", ipStr, "record", record)
				continue
			}
			switch {
			case strings.Contains(ipStr, "."):
				res.ARecords[n].RR = append(res.ARecords[n].RR,
					&dns.A{Hdr: dns.RR_Header{Name: n, Ttl: uint32(configuration.TTL), Rrtype: dns.TypeA, Class: dns.ClassINET}, A: ip},
				)
			case strings.Contains(ipStr, ":"):
				res.AAAARecords[n].RR = append(res.AAAARecords[n].RR,
					&dns.AAAA{Hdr: dns.RR_Header{Name: n, Ttl: uint32(configuration.TTL), Rrtype: dns.TypeAAAA, Class: dns.ClassINET}, AAAA: ip},
				)
			}
		}
	}
	log.V(1).Info("DNS proxy configured", "config", res)
	s.dnsMap.Store(&res)

	return nil
}
