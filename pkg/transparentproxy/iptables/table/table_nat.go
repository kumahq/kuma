package table

import (
	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/chain"
)

type NatTable struct {
	prerouting  *chain.Chain
	input       *chain.Chain
	output      *chain.Chain
	postrouting *chain.Chain

	// custom chains
	chains []*chain.Chain
}

func (t *NatTable) Prerouting() *chain.Chain {
	return t.prerouting
}

func (t *NatTable) Input() *chain.Chain {
	return t.input
}

func (t *NatTable) Output() *chain.Chain {
	return t.output
}

func (t *NatTable) Postrouting() *chain.Chain {
	return t.postrouting
}

func (t *NatTable) WithChain(chain *chain.Chain) *NatTable {
	t.chains = append(t.chains, chain)

	return t
}

func (t *NatTable) Build(verbose bool) string {
	table := &TableBuilder{
		name:      "nat",
		newChains: t.chains,
		chains: []*chain.Chain{
			t.prerouting,
			t.input,
			t.output,
			t.postrouting,
		},
	}

	return table.Build(verbose)
}

func Nat() *NatTable {
	return &NatTable{
		prerouting:  chain.NewChain("PREROUTING"),
		input:       chain.NewChain("INPUT"),
		output:      chain.NewChain("OUTPUT"),
		postrouting: chain.NewChain("POSTROUTING"),
	}
}
