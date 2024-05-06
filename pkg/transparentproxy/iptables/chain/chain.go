package chain

import (
	. "github.com/kumahq/kuma/pkg/transparentproxy/iptables/parameters"
	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/rules"
)

type Chain struct {
	table string
	name  string
	rules []*rules.Rule
}

func (c *Chain) Name() string {
	return c.name
}

func (c *Chain) AddRule(parameters ...*Parameter) *Chain {
	c.rules = append(c.rules, rules.NewRule(c.table, c.name, 0, parameters))

	return c
}

func (c *Chain) AddRuleAtPosition(position uint, parameters ...*Parameter) *Chain {
	c.rules = append(c.rules, rules.NewRule(c.table, c.name, position, parameters))

	return c
}

func (c *Chain) AddRuleIf(predicate func() bool, parameters ...*Parameter) *Chain {
	if predicate() {
		return c.AddRule(parameters...)
	}

	return c
}

func (c *Chain) Build(verbose bool) []string {
	var cmds []string

	for _, rule := range c.rules {
		cmds = append(cmds, rule.Build(verbose))
	}

	return cmds
}

func NewChain(table, chain string) *Chain {
	return &Chain{
		table: table,
		name:  chain,
	}
}
