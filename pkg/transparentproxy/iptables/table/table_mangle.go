package table

import (
	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/chain"
	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/consts"
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

func (t *MangleTable) BuildForRestore(verbose bool) string {
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

	return table.BuildForRestore(verbose)
}

func Mangle() *MangleTable {
	prerouting, _ := chain.NewChain(consts.TableMangle, "PREROUTING")
	input, _ := chain.NewChain(consts.TableMangle, "INPUT")
	forward, _ := chain.NewChain(consts.TableMangle, "FORWARD")
	output, _ := chain.NewChain(consts.TableMangle, "OUTPUT")
	postrouting, _ := chain.NewChain(consts.TableMangle, "POSTROUTING")

	return &MangleTable{
		prerouting:  prerouting,
		input:       input,
		forward:     forward,
		output:      output,
		postrouting: postrouting,
	}
}
