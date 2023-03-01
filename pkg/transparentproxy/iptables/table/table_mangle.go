package table

import (
	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/chain"
)

type MangleTable struct {
	prerouting  *chain.Chain
	input       *chain.Chain
	forward     *chain.Chain
	output      *chain.Chain
	postrouting *chain.Chain
}

func (t *MangleTable) Prerouting() *chain.Chain {
	return t.prerouting
}

func (t *MangleTable) Input() *chain.Chain {
	return t.input
}

func (t *MangleTable) Forward() *chain.Chain {
	return t.forward
}

func (t *MangleTable) Output() *chain.Chain {
	return t.output
}

func (t *MangleTable) Postrouting() *chain.Chain {
	return t.postrouting
}

func (t *MangleTable) Build(verbose bool) string {
	table := &TableBuilder{
		name: "mangle",
		chains: []*chain.Chain{
			t.prerouting,
			t.input,
			t.forward,
			t.output,
			t.postrouting,
		},
	}

	return table.Build(verbose)
}

func Mangle() *MangleTable {
	return &MangleTable{
		prerouting:  chain.NewChain("PREROUTING"),
		input:       chain.NewChain("INPUT"),
		forward:     chain.NewChain("FORWARD"),
		output:      chain.NewChain("OUTPUT"),
		postrouting: chain.NewChain("POSTROUTING"),
	}
}
