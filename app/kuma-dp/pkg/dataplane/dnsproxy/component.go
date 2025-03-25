package dnsproxy

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/miekg/dns"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/dns/dpapi"
)

var log = core.Log.WithName("embeededdns")

type Server struct {
	address          string
	port             uint32
	done             chan struct{}
	dnsMap           atomic.Pointer[dnsMap]
	metrics          *metrics
	upstreamHostPort string
}

func NewServer(address string, port uint32) *Server {
	s := Server{
		address: address,
		port:    port,
		done:    make(chan struct{}),
		dnsMap:  atomic.Pointer[dnsMap]{},
		metrics: newMetrics(),
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
	}()
	var response *dns.Msg
	// In case it was never loaded
	s.dnsMap.CompareAndSwap(nil, &dnsMap{
		ARecords:    make(map[string]*dnsEntry),
		AAAARecords: make(map[string]*dnsEntry),
	})
	var dnsEntry *dnsEntry
	if len(req.Question) > 0 { // Apparently most DNS doesn't support multiple questions so let's just support the first one
		switch req.Question[0].Qtype {
		case dns.TypeA:
			// lookup in our DNS map
			dnsMap := s.dnsMap.Load()
			dnsEntry = dnsMap.ARecords[req.Question[0].Name]
		case dns.TypeAAAA:
			dnsMap := s.dnsMap.Load()
			dnsEntry = dnsMap.AAAARecords[req.Question[0].Name]
		}
		log.V(1).Info("Got request", "type", req.Question[0].Qtype, "name", req.Question[0].Name, "entry", dnsEntry)
	}
	if dnsEntry != nil {
		response = new(dns.Msg)
		response.SetRcode(req, int(dnsEntry.RCode))
		response.Authoritative = true
		response.Answer = append(response.Answer, dnsEntry.RR...)
	} else {
		proxyStart := time.Now()
		c := new(dns.Client)
		resp, _, err := c.Exchange(req, s.upstreamHostPort)
		if err != nil {
			s.metrics.UpstreamRequestFailureCount.Inc()
			log.Error(err, "Failed to write message to upstream")
			response = new(dns.Msg)
			response.SetRcode(req, dns.RcodeServerFailure)
		} else {
			response = resp
		}
		s.metrics.UpstreamRequestDuration.Observe(time.Since(proxyStart).Seconds())
	}
	err := res.WriteMsg(response)
	if err != nil {
		log.Error(err, "Failed to write upstreamResponse")
	}
}

func (s *Server) Start(stop <-chan struct{}) error {
	defer close(s.done)
	srv := &dns.Server{Addr: net.JoinHostPort(s.address, strconv.Itoa(int(s.port))), Handler: dns.HandlerFunc(s.Handler), Net: "udp"}
	config, err := dns.ClientConfigFromFile("/etc/resolv.conf")
	if err != nil {
		return fmt.Errorf("failed to read DNS for /etc/resolv.conf %w", err)
	}
	s.upstreamHostPort = net.JoinHostPort(config.Servers[0], config.Port)

	done := make(chan error, 1)
	go func() {
		defer close(done)
		err := srv.ListenAndServe()
		done <- err
	}()
	select {
	case <-stop:
		err := srv.Shutdown()
		if err != nil {
			log.Error(err, "server shutdown returned an error")
		}
	case err = <-done:
		log.Info("[WARNING] server stopped with shutdown ever being called")
		if err != nil {
			return err
		}
	}
	err = <-done
	log.Info("server shutdown complete")
	return err
}

func (s *Server) NeedLeaderElection() bool {
	return false
}

func (s *Server) WaitForDone() {
	<-s.done
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
