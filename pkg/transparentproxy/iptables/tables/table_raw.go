package tables

import (
	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/chains"
	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/consts"
)

type RawTable struct {
	prerouting *chains.Chain
	output     *chains.Chain
}

func (t *RawTable) Prerouting() *chains.Chain {
	return t.prerouting
}

func (t *RawTable) Output() *chains.Chain {
	return t.output
}

func (t *RawTable) BuildForRestore(verbose bool) string {
	table := &TableBuilder{
		name: string(consts.TableRaw),
		chains: []*chains.Chain{
			t.prerouting,
			t.output,
		},
	}

	return table.BuildForRestore(verbose)
}

func Raw() *RawTable {
	prerouting, _ := chains.NewChain(consts.TableRaw, consts.ChainPrerouting)
	output, _ := chains.NewChain(consts.TableRaw, consts.ChainOutput)

	return &RawTable{
		prerouting: prerouting,
		output:     output,
	}
}
