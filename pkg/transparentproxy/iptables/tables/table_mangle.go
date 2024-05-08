package tables

import (
	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/chains"
	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/consts"
)

type MangleTable struct {
	prerouting  *chains.Chain
	input       *chains.Chain
	forward     *chains.Chain
	output      *chains.Chain
	postrouting *chains.Chain
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

func (t *MangleTable) BuildForRestore(verbose bool) string {
	table := &TableBuilder{
		name: string(consts.TableMangle),
		chains: []*chains.Chain{
			t.prerouting,
			t.input,
			t.forward,
			t.output,
			t.postrouting,
		},
	}

	return table.BuildForRestore(verbose)
}

func Mangle() *MangleTable {
	prerouting, _ := chains.NewChain(consts.TableMangle, consts.ChainPrerouting)
	input, _ := chains.NewChain(consts.TableMangle, consts.ChainInput)
	forward, _ := chains.NewChain(consts.TableMangle, consts.ChainForward)
	output, _ := chains.NewChain(consts.TableMangle, consts.ChainOutput)
	postrouting, _ := chains.NewChain(consts.TableMangle, consts.ChainPostrouting)

	return &MangleTable{
		prerouting:  prerouting,
		input:       input,
		forward:     forward,
		output:      output,
		postrouting: postrouting,
	}
}
