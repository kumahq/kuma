package resolver

import (
	"fmt"

	"github.com/miekg/dns"
)

func (d *SimpleDNSResolver) parseQuery(m *dns.Msg) {
	for _, q := range m.Question {
		switch q.Qtype {
		case dns.TypeA:
			simpleDNSLog.Info("Query for %s\n", q.Name)
			ip, err := d.ForwardLookup(q.Name)
			if err != nil {
				simpleDNSLog.Error(err, "Unable to resolve ", "q.Name", q.Name)
				return
			}
			rr, err := dns.NewRR(fmt.Sprintf("%s A %s", q.Name, ip))
			if err != nil {
				simpleDNSLog.Error(err, "Unable to create response for ", "q.Name", q.Name)
				return
			}
			m.Answer = append(m.Answer, rr)
		}
	}
}

func (d *SimpleDNSResolver) handleDnsRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = false

	switch r.Opcode {
	case dns.OpcodeQuery:
		d.parseQuery(m)
	}

	err := w.WriteMsg(m)
	if err != nil {
		simpleDNSLog.Error(err, "Unable to write the DNS response.")
	}
}

func (d *SimpleDNSResolver) registerDNSHandler(domain string) {
	dns.HandleFunc(domain, d.handleDnsRequest)
}

func (d *SimpleDNSResolver) registerDNSHandlers() {
	for domain := range d.domains {
		d.registerDNSHandler(domain)
	}
}
