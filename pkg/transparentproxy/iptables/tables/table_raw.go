package tables

import (
	"github.com/kumahq/kuma/pkg/transparentproxy/consts"
	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/chains"
)

var _ Table = &RawTable{}

type RawTable struct {
	prerouting *chains.Chain
	output     *chains.Chain
}

func (t *RawTable) Name() consts.TableName {
	return consts.TableRaw
}

func (t *RawTable) Prerouting() *chains.Chain {
	return t.prerouting
}

func (t *RawTable) Output() *chains.Chain {
	return t.output
}

func (t *RawTable) Chains() []*chains.Chain {
	return []*chains.Chain{t.prerouting, t.output}
}

func (t *RawTable) CustomChains() []*chains.Chain {
	return nil
}

func Raw() *RawTable {
	return &RawTable{
		prerouting: chains.MustNewChain(consts.TableRaw, consts.ChainPrerouting),
		output:     chains.MustNewChain(consts.TableRaw, consts.ChainOutput),
	}
}
