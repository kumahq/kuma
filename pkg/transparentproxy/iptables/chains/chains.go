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
func (c *Chain) BuildForRestore(
	cfg config.InitializedConfig,
	ipv6 bool,
) []string {
	var lines []string

	for _, rule := range c.rules {
		lines = append(lines, rule.BuildForRestore(cfg, ipv6))
	}

	return lines
}

// NewChain creates a new Chain object for a specified iptables table and chain
// name.
//
// This function validates that the provided table is one of the supported
// tables (raw, nat, or mangle) and that the chain name is not empty. If these
// conditions are met, it returns a new Chain object initialized with the
// provided table and chain name. If the validation fails, it returns an error.
//
// Args:
//
//   - table (consts.TableName): The name of the iptables table. Supported
//     values are "raw", "nat", and "mangle".
//   - chain (string): The name of the chain. This cannot be an empty string.
//
// Returns:
//
//   - *Chain: A pointer to a new Chain object if the inputs are valid.
//   - error: An error if the table is unsupported or if the chain name is
//     empty.
//
// Supported Tables:
//   - consts.TableRaw    // "raw"
//   - consts.TableNat    // "nat
//   - consts.TableMangle // "mangle"
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

// MustNewChain creates a new Chain object for a specified iptables table and
// chain name, and panics if the provided table or chain is invalid.
//
// This function should be used when you are certain that the provided table is
// one of the supported tables (raw, nat, or mangle) and that the chain name is
// not empty. If these conditions are not met, the function will panic. This is
// useful for scenarios where validation has been performed elsewhere, and you
// want to ensure that the creation of the Chain object does not fail.
//
// Args:
//   - table (consts.TableName): The name of the iptables table. Supported
//     values are "raw", "nat", and "mangle".
//   - chain (string): The name of the chain. This cannot be an empty string.
//
// Returns:
//   - *Chain: A pointer to a new Chain object.
//
// Panics:
//
//	If the table is unsupported or the chain name is empty.
//
// Supported Tables:
//   - consts.TableRaw    // "raw"
//   - consts.TableNat    // "nat
//   - consts.TableMangle // "mangle"
func MustNewChain(table consts.TableName, chain string) *Chain {
	newChain, err := NewChain(table, chain)
	if err != nil {
		panic(err)
	}

	return newChain
}
