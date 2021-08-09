package dns

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/miekg/dns"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/dns/resolver"
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
		address:  net.JoinHostPort("0.0.0.0", strconv.FormatUint(uint64(port), 10)),
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
		case dns.TypeA, dns.TypeAAAA:
			serverLog.V(1).Info("received a query for " + q.Name)
			ip, err := h.lookup(q.Name)
			if err != nil {
				serverLog.V(1).Info("unable to resolve", "Name", q.Name, "error", err.Error())
				h.resolutionMetric.WithLabelValues("unresolved").Inc()
				return
			}
			h.resolutionMetric.WithLabelValues("resolved").Inc()

			recordType := "A"
			if govalidator.IsIPv6(ip) {
				recordType = "AAAA"
			}

			rr, err := dns.NewRR(fmt.Sprintf("%s %s IN %s %s", q.Name, dnsTTL, recordType, ip))
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
			errString := "failed to start the DNS listener."
			if strings.Contains(err.Error(), "bind") {
				errString = bindError(d.address)
			}
			serverLog.Error(err, errString)
			errChan <- errors.Wrap(err, errString)
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

func bindError(address string) string {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return fmt.Sprintf("invalid DNS bind address %s", address)
	}
	return fmt.Sprintf(
		"unable to bind the DNS server to %s.\n\nPlease consider setting KUMA_DNS_SERVER_PORT=5653 (the default).\n"+
			"Then redirect the incoming UDP traffinc on port 53 to it. The `iptables` command for this would be:\n\n"+
			"iptables -t nat -A OUTPUT -p udp -d %s --dport 53 -j DNAT --to-destination %s:5653\n\n"+
			"On hosts which use firewalld, the command would be:\n\n"+
			"firewall-cmd --direct --add-rule ipv4 nat OUTPUT 1 -p udp -d %s --dport 53 -j DNAT --to-destination %s:5653\n\n",
		address,
		host, host,
		host, host)
}
