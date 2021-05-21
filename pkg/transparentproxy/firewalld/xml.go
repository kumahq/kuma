package firewalld

import (
	"encoding/xml"
	"reflect"
)

type Chain struct {
	// required, ip family: "ipv4", "ipv6", "eb"
	IPv string `xml:"ipv,attr"`

	// required, netfilter table: "nat", "mangle", etc.
	Table string `xml:"table,attr"`

	// required, netfilter chain: "FORWARD", custom chain names
	Chain string `xml:"chain,attr"`

	XMLName struct{} `xml:"chain"`
}

func (c *Chain) String() string {
	data, _ := xml.MarshalIndent(c, "", "  ")
	return string(data)
}

func NewIP4Chain(table, chain string) *Chain {
	c := &Chain{
		IPv:   "ipv4",
		Table: table,
		Chain: chain,
	}
	return c
}

func NewIP6Chain(table, chain string) *Chain {
	c := &Chain{
		IPv:   "ipv6",
		Table: table,
		Chain: chain,
	}
	return c
}

type Rule struct {
	// required, ip family: "ipv4", "ipv6", "eb"
	IPv string `xml:"ipv,attr"`

	// required, netfilter table: "nat", "mangle", etc.
	Table string `xml:"table,attr"`

	// required, netfilter chain: "FORWARD", custom chain names
	Chain string `xml:"chain,attr"`

	// required, smaller the number more front the rule in chain
	Priority int `xml:"priority,attr"`

	// match and action command line options for {ip,ip6,eb}tables
	Body string `xml:",chardata"`

	XMLName struct{} `xml:"rule"`
}

func (r *Rule) String() string {
	data, _ := xml.MarshalIndent(r, "", "  ")
	return string(data)
}

func NewIP4Rule(prio int, table, chain, body string) *Rule {
	r := &Rule{
		Priority: prio,
		IPv:      "ipv4",
		Table:    table,
		Chain:    chain,
		Body:     body,
	}
	return r
}

func NewIP6Rule(prio int, table, chain, body string) *Rule {
	r := &Rule{
		Priority: prio,
		IPv:      "ipv6",
		Table:    table,
		Chain:    chain,
		Body:     body,
	}
	return r
}

type Direct struct {
	Chains []*Chain
	Rules  []*Rule

	XMLName struct{} `xml:"direct"`
}

func (d *Direct) Bytes() []byte {
	data, _ := xml.MarshalIndent(d, "", "  ")
	return append([]byte(xml.Header), data...)
}

func (d *Direct) String() string {
	return string(d.Bytes())
}

func (d *Direct) AddChain(chain *Chain) {
	for _, c := range d.Chains {
		if reflect.DeepEqual(c, chain) {
			return
		}
	}
	d.Chains = append(d.Chains, chain)
}

func (d *Direct) AddRule(rule *Rule) {
	for _, r := range d.Rules {
		if reflect.DeepEqual(r, rule) {
			return
		}
	}
	d.Rules = append(d.Rules, rule)
}

func NewDirect(rules ...*Rule) *Direct {
	d := &Direct{
		Rules: rules,
	}
	return d
}
