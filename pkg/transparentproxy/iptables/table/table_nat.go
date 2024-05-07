package table

import (
	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/chain"
	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/consts"
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

func (t *NatTable) BuildForRestore(verbose bool) string {
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

	return table.BuildForRestore(verbose)
}

func NewNatChain(name string) (*chain.Chain, error) {
	return chain.NewChain(consts.TableNat, name)
}

func Nat() (*NatTable, error) {
	prerouting, err := NewNatChain("PREROUTING")
	if err != nil {
		return nil, err
	}

	input, err := NewNatChain("INPUT")
	if err != nil {
		return nil, err
	}

	output, err := NewNatChain("OUTPUT")
	if err != nil {
		return nil, err
	}

	postrouting, err := NewNatChain("POSTROUTING")
	if err != nil {
		return nil, err
	}

	return &NatTable{
		prerouting:  prerouting,
		input:       input,
		output:      output,
		postrouting: postrouting,
	}, nil
}
