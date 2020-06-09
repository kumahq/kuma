package resolver

import (
	"fmt"

	"github.com/miekg/dns"
)

type DNSHandler interface {
}

type SimpleDNSHandler struct {
	resolver DNSResolver
}

func NewSimpleDNSHandler(resolver DNSResolver) DNSHandler {
	handler := &SimpleDNSHandler{
		resolver: resolver,
	}

	handler.registerDNSHandler()

	return handler
}

func (h *SimpleDNSHandler) parseQuery(m *dns.Msg) {
	for _, q := range m.Question {
		switch q.Qtype {
		case dns.TypeA:
			simpleDNSLog.Info("Query for " + q.Name)
			ip, err := h.resolver.ForwardLookup(q.Name)
			if err != nil {
				simpleDNSLog.Error(err, "Unable to resolve", "q.Name", q.Name)
				return
			}

			rr, err := dns.NewRR(fmt.Sprintf("%s A %s", q.Name, ip))
			if err != nil {
				simpleDNSLog.Error(err, "Unable to create response for", "q.Name", q.Name)
				return
			}

			m.Answer = append(m.Answer, rr)
		}
	}
}

func (h *SimpleDNSHandler) handleDNSRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = false

	switch r.Opcode {
	case dns.OpcodeQuery:
		h.parseQuery(m)
	}

	err := w.WriteMsg(m)
	if err != nil {
		simpleDNSLog.Error(err, "Unable to write the DNS response.")
	}
}

func (h *SimpleDNSHandler) registerDNSHandler() {
	dns.HandleFunc(h.resolver.GetDomain(), h.handleDNSRequest)
}
