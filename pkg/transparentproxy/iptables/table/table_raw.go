package table

import (
	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/chain"
)

type RawTable struct {
	prerouting *chain.Chain
	output     *chain.Chain
}

func (t *RawTable) Prerouting() *chain.Chain {
	return t.prerouting
}

func (t *RawTable) Output() *chain.Chain {
	return t.output
}

func (t *RawTable) Build(verbose bool) string {
	table := &TableBuilder{
		name: "raw",
		chains: []*chain.Chain{
			t.prerouting,
			t.output,
		},
	}

	return table.Build(verbose)
}

func Raw() *RawTable {
	return &RawTable{
		prerouting: chain.NewChain("PREROUTING"),
		output:     chain.NewChain("OUTPUT"),
	}
}
