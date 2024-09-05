package tables

import (
	"github.com/kumahq/kuma/pkg/transparentproxy/consts"
	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/chains"
)

var _ Table = &NatTable{}

type NatTable struct {
	prerouting  *chains.Chain
	input       *chains.Chain
	output      *chains.Chain
	postrouting *chains.Chain

	// custom chains
	customChains []*chains.Chain
}

func (t *NatTable) Name() consts.TableName {
	return consts.TableNat
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

func (t *NatTable) Chains() []*chains.Chain {
	return []*chains.Chain{t.prerouting, t.input, t.output, t.postrouting}
}

func (t *NatTable) CustomChains() []*chains.Chain {
	return t.customChains
}

func (t *NatTable) WithCustomChain(chain *chains.Chain) *NatTable {
	t.customChains = append(t.customChains, chain)

	return t
}

func Nat() *NatTable {
	return &NatTable{
		prerouting:  chains.MustNewChain(consts.TableNat, consts.ChainPrerouting),
		input:       chains.MustNewChain(consts.TableNat, consts.ChainInput),
		output:      chains.MustNewChain(consts.TableNat, consts.ChainOutput),
		postrouting: chains.MustNewChain(consts.TableNat, consts.ChainPostrouting),
	}
}
