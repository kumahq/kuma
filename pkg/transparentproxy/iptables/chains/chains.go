package chains

import (
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/transparentproxy/config"
	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/consts"
	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/rules"
)

type Chain struct {
	table consts.TableName
	name  string
	rules []*rules.Rule
}

func (c *Chain) Name() string {
	return c.name
}

func (c *Chain) AddRules(rules ...*rules.RuleBuilder) *Chain {
	for _, rule := range rules {
		c.rules = append(c.rules, rule.Build(c.table, c.name))
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
func (c *Chain) BuildForRestore(cfg config.InitializedConfig) []string {
	var lines []string

	for _, rule := range c.rules {
		lines = append(lines, rule.BuildForRestore(cfg))
	}

	return lines
}

func NewChain(table consts.TableName, chain string) (*Chain, error) {
	switch table {
	// Only raw, nat and mangle are supported
	case consts.TableRaw, consts.TableNat, consts.TableMangle:
	case "":
		return nil, errors.New("table is required and cannot be empty")
	default:
		return nil, errors.Errorf(
			"unsupported table %q (valid: [%q, %q, %q])",
			table,
			consts.TableRaw,
			consts.TableNat,
			consts.TableMangle,
		)
	}

	if chain == "" {
		return nil, errors.New("chain is required and cannot be empty")
	}

	return &Chain{
		table: table,
		name:  chain,
	}, nil
}
