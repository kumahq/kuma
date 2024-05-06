package chain

import (
	. "github.com/kumahq/kuma/pkg/transparentproxy/iptables/parameters"
	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/rules"
)

type Chain struct {
	table    string
	name     string
	commands []*rules.Rule
}

func (b *Chain) Name() string {
	return b.name
}

func (b *Chain) AddRule(parameters ...*Parameter) *Chain {
	b.commands = append(b.commands, rules.NewRule(b.table, b.name, 0, parameters))

	return b
}

func (b *Chain) AddRuleAtPosition(position uint, parameters ...*Parameter) *Chain {
	b.commands = append(b.commands, rules.NewRule(b.table, b.name, position, parameters))

	return b
}

func (b *Chain) AddRuleIf(predicate func() bool, parameters ...*Parameter) *Chain {
	if predicate() {
		return b.AddRule(parameters...)
	}

	return b
}

func (b *Chain) Build(verbose bool) []string {
	var cmds []string

	for _, cmd := range b.commands {
		cmds = append(cmds, cmd.Build(verbose))
	}

	return cmds
}

func NewChain(table, chain string) *Chain {
	return &Chain{
		table: table,
		name:  chain,
	}
}
