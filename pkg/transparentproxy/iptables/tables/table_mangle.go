package tables

import (
	"github.com/kumahq/kuma/pkg/transparentproxy/consts"
	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/chains"
)

var _ Table = &MangleTable{}

type MangleTable struct {
	prerouting  *chains.Chain
	input       *chains.Chain
	forward     *chains.Chain
	output      *chains.Chain
	postrouting *chains.Chain
}

func (t *MangleTable) Name() consts.TableName {
	return consts.TableMangle
}

func (t *MangleTable) Prerouting() *chains.Chain {
	return t.prerouting
}

func (t *MangleTable) Input() *chains.Chain {
	return t.input
}

func (t *MangleTable) Forward() *chains.Chain {
	return t.forward
}

func (t *MangleTable) Output() *chains.Chain {
	return t.output
}

func (t *MangleTable) Postrouting() *chains.Chain {
	return t.postrouting
}

func (t *MangleTable) Chains() []*chains.Chain {
	return []*chains.Chain{t.prerouting, t.input, t.forward, t.output, t.postrouting}
}

func (t *MangleTable) CustomChains() []*chains.Chain {
	return nil
}

func Mangle() *MangleTable {
	return &MangleTable{
		prerouting:  chains.MustNewChain(consts.TableMangle, consts.ChainPrerouting),
		input:       chains.MustNewChain(consts.TableMangle, consts.ChainInput),
		forward:     chains.MustNewChain(consts.TableMangle, consts.ChainForward),
		output:      chains.MustNewChain(consts.TableMangle, consts.ChainOutput),
		postrouting: chains.MustNewChain(consts.TableMangle, consts.ChainPostrouting),
	}
}
