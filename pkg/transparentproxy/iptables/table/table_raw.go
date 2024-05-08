package table

import (
	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/chain"
	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/consts"
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

func (t *RawTable) BuildForRestore(verbose bool) string {
	table := &TableBuilder{
		name: "raw",
		chains: []*chain.Chain{
			t.prerouting,
			t.output,
		},
	}

	return table.BuildForRestore(verbose)
}

func Raw() *RawTable {
	prerouting, _ := chain.NewChain(consts.TableRaw, "PREROUTING")
	output, _ := chain.NewChain(consts.TableRaw, "OUTPUT")

	return &RawTable{
		prerouting: prerouting,
		output:     output,
	}
}
