package dns

import (
	"fmt"

	"github.com/miekg/dns"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/kumahq/kuma/pkg/dns/resolver"

	"github.com/kumahq/kuma/pkg/core"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
)

const dnsTTL = "60"

var serverLog = core.Log.WithName("dns-server")

type DNSServer interface {
	Start(<-chan struct{}) error
	NeedLeaderElection() bool
}

type NameModifier = func(qName string) (string, error)

type SimpleDNSServer struct {
	address  string
	resolver resolver.DNSResolver

	latencyMetric    prometheus.Summary
	resolutionMetric *prometheus.CounterVec
	nameModifier     NameModifier
}

func NewDNSServer(port uint32, resolver resolver.DNSResolver, metrics core_metrics.Metrics, modifier NameModifier) (DNSServer, error) {
	handler := &SimpleDNSServer{
		address:  fmt.Sprintf("0.0.0.0:%d", port),
		resolver: resolver,
		latencyMetric: prometheus.NewSummary(prometheus.SummaryOpts{
			Name:       "dns_server",
			Help:       "Summary of DNS Server responses",
			Objectives: core_metrics.DefaultObjectives,
		}),
		resolutionMetric: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "dns_server_resolution",
			Help: "Counter for DNS Server resolutions",
		}, []string{"result"}),
		nameModifier: modifier,
	}
	if err := metrics.Register(handler.latencyMetric); err != nil {
		return nil, err
	}
	if err := metrics.Register(handler.resolutionMetric); err != nil {
		return nil, err
	}
	handler.registerDNSHandler()
	return handler, nil
}

func (h *SimpleDNSServer) parseQuery(m *dns.Msg) {
	for _, q := range m.Question {
		switch q.Qtype {
		case dns.TypeA:
			serverLog.V(1).Info("received a query for " + q.Name)
			ip, err := h.lookup(q.Name)
			if err != nil {
				serverLog.V(1).Info("unable to resolve", "Name", q.Name, "error", err.Error())
				h.resolutionMetric.WithLabelValues("unresolved").Inc()
				return
			}
			h.resolutionMetric.WithLabelValues("resolved").Inc()

			rr, err := dns.NewRR(fmt.Sprintf("%s %s IN A %s", q.Name, dnsTTL, ip))
			if err != nil {
				serverLog.Error(err, "unable to create response for", "Name", q.Name)
				return
			}

			m.Answer = append(m.Answer, rr)
		}
	}
}

func (h *SimpleDNSServer) handleDNSRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = false

	switch r.Opcode {
	case dns.OpcodeQuery:
		h.parseQuery(m)
	}

	err := w.WriteMsg(m)
	if err != nil {
		serverLog.Error(err, "unable to write the DNS response.")
	}
}

func (d *SimpleDNSServer) NeedLeaderElection() bool {
	return false
}

func (d *SimpleDNSServer) Start(stop <-chan struct{}) error {
	server := &dns.Server{
		Addr: d.address,
		Net:  "udp",
	}

	errChan := make(chan error)
	go func() {
		defer close(errChan)
		err := server.ListenAndServe()
		if err != nil {
			serverLog.Error(err, "failed to start the DNS listener.")
			errChan <- err
		}
	}()

	serverLog.Info("starting", "address", d.address)
	select {
	case <-stop:
		serverLog.Info("shutting down the DNS Server")
		return server.Shutdown()
	case err := <-errChan:
		return err
	}
}

func (h *SimpleDNSServer) registerDNSHandler() {
	dns.HandleFunc(h.resolver.GetDomain(), func(writer dns.ResponseWriter, msg *dns.Msg) {
		start := core.Now()
		defer func() {
			h.latencyMetric.Observe(float64(core.Now().Sub(start).Milliseconds()))
		}()
		h.handleDNSRequest(writer, msg)
	})
}

func (h *SimpleDNSServer) lookup(qName string) (string, error) {
	ip, err := h.resolver.ForwardLookupFQDN(qName)
	if err != nil {
		if h.nameModifier == nil {
			return "", err
		}

		modifiedName, err := h.nameModifier(qName)
		if err != nil {
			return "", err
		}

		ip, err = h.resolver.ForwardLookupFQDN(modifiedName)
		if err != nil {
			return "", err
		}
	}

	return ip, nil
}
