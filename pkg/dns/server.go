package dns

import (
	"fmt"

	"github.com/Kong/kuma/pkg/core"

	"github.com/miekg/dns"
)

const dnsTTL = "60"

var serverLog = core.Log.WithName("dns-server")

type DNSServer interface {
	Start(<-chan struct{}) error
	NeedLeaderElection() bool
}

type SimpleDNSServer struct {
	address  string
	resolver DNSResolver
}

func NewDNSServer(port uint32, resolver DNSResolver) DNSServer {
	handler := &SimpleDNSServer{
		address:  fmt.Sprintf("0.0.0.0:%d", port),
		resolver: resolver,
	}
	handler.registerDNSHandler()
	return handler
}

func (h *SimpleDNSServer) parseQuery(m *dns.Msg) {
	for _, q := range m.Question {
		switch q.Qtype {
		case dns.TypeA:
			serverLog.Info("Query for " + q.Name)
			ip, err := h.resolver.ForwardLookupFQDN(q.Name)
			if err != nil {
				serverLog.Error(err, "unable to resolve", "Name", q.Name)
				return
			}

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
	dns.HandleFunc(h.resolver.GetDomain(), h.handleDNSRequest)
}
