package tables

import (
	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/chains"
	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/consts"
)

type NatTable struct {
	prerouting  *chains.Chain
	input       *chains.Chain
	output      *chains.Chain
	postrouting *chains.Chain

	// custom chains
	chains []*chains.Chain
}

func (t *NatTable) Prerouting() *chains.Chain {
	return t.prerouting
}

func (t *NatTable) Input() *chains.Chain {
	return t.input
}

func (t *NatTable) Output() *chains.Chain {
	return t.output
}

func (t *NatTable) Postrouting() *chains.Chain {
	return t.postrouting
}

func (t *NatTable) WithChain(chain *chains.Chain) *NatTable {
	t.chains = append(t.chains, chain)

	return t
}

func (t *NatTable) BuildForRestore(verbose bool) string {
	table := &TableBuilder{
		name:      string(consts.TableNat),
		newChains: t.chains,
		chains: []*chains.Chain{
			t.prerouting,
			t.input,
			t.output,
			t.postrouting,
		},
	}

	return table.BuildForRestore(verbose)
}

func Nat() *NatTable {
	prerouting, _ := chains.NewChain(consts.TableNat, consts.ChainPrerouting)
	input, _ := chains.NewChain(consts.TableNat, consts.ChainInput)
	output, _ := chains.NewChain(consts.TableNat, consts.ChainOutput)
	postrouting, _ := chains.NewChain(consts.TableNat, consts.ChainPostrouting)

	return &NatTable{
		prerouting:  prerouting,
		input:       input,
		output:      output,
		postrouting: postrouting,
	}
}
