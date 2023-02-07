package chain

import (
	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/commands"
	. "github.com/kumahq/kuma/pkg/transparentproxy/iptables/parameters"
)

type Chain struct {
	name     string
	commands []*commands.Command
}

func (b *Chain) Name() string {
	return b.name
}

func (b *Chain) Append(parameters ...*Parameter) *Chain {
	b.commands = append(b.commands, commands.Append(b.name, parameters))

	return b
}

func (b *Chain) Insert(position int, parameters ...*Parameter) *Chain {
	b.commands = append(b.commands, commands.Insert(b.name, position, parameters))

	return b
}

func (b *Chain) AppendIf(predicate func() bool, parameters ...*Parameter) *Chain {
	if predicate() {
		return b.Append(parameters...)
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
