package chain

import (
	"errors"

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

// BuildForRestore function generates all iptables rules within chain in
// a format suitable for restoration using the `iptables-restore` command.
//
// It iterates over each rule in the `rules` slice and calls the rule's
// `BuildForRestore(verbose)` method to generate the individual command string
// for each rule. The `verbose` flag is passed along to maintain consistent
// output formatting throughout the chain.
func (c *Chain) BuildForRestore(verbose bool) []string {
	var lines []string

	for _, rule := range c.rules {
		lines = append(lines, rule.BuildForRestore(verbose))
	}

	return lines
}

func NewChain(table, chain string) (*Chain, error) {
	if table == "" {
		return nil, errors.New("table is required and cannot be empty")
	}

	if chain == "" {
		return nil, errors.New("chain is required and cannot be empty")
	}

	return &Chain{
		table: table,
		name:  chain,
	}, nil
}
