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

func NewMangleChain(name string) (*chain.Chain, error) {
	return chain.NewChain(consts.TableMangle, name)
}

func Mangle() (*MangleTable, error) {
	prerouting, err := NewMangleChain("PREROUTING")
	if err != nil {
		return nil, err
	}

	input, err := NewMangleChain("INPUT")
	if err != nil {
		return nil, err
	}

	forward, err := NewMangleChain("FORWARD")
	if err != nil {
		return nil, err
	}

	output, err := NewMangleChain("OUTPUT")
	if err != nil {
		return nil, err
	}

	postrouting, err := NewMangleChain("POSTROUTING")
	if err != nil {
		return nil, err
	}

	return &MangleTable{
		prerouting:  prerouting,
		input:       input,
		forward:     forward,
		output:      output,
		postrouting: postrouting,
	}, nil
}
