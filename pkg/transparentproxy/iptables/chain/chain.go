package chain

import (
	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/rules"
	. "github.com/kumahq/kuma/pkg/transparentproxy/iptables/parameters"
)

type Chain struct {
	name     string
	commands []*rules.Rule
}

func (b *Chain) Name() string {
	return b.name
}

func (b *Chain) AddRule(parameters ...*Parameter) *Chain {
	b.commands = append(b.commands, rules.NewRule(b.name, 0, parameters))

	return b
}

func (b *Chain) AddRuleAtPosition(position uint, parameters ...*Parameter) *Chain {
	b.commands = append(b.commands, rules.NewRule(b.name, position, parameters))

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

func NewChain(name string) *Chain {
	return &Chain{
		name: name,
	}
}
