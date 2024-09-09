package chains

import (
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/transparentproxy/config"
	"github.com/kumahq/kuma/pkg/transparentproxy/consts"
	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/rules"
)

type Chain struct {
	// The name of the iptables table (e.g., "nat", "filter") to which this
	// chain belongs.
	table consts.TableName
	// The name of the iptables chain (e.g., "PREROUTING", "OUTPUT").
	name string
	// A slice of rules contained within this chain.
	rules []*rules.Rule
	// position reflects the current position for "insert" rules, indicating
	// where new rules should be inserted within the chain. This is crucial for
	// maintaining the correct order of rules when specific positioning is
	// required.
	position uint
}

func (c *Chain) Name() string {
	return c.name
}

func (c *Chain) AddRules(rules ...*rules.RuleBuilder) *Chain {
	for _, r := range rules {
		rule, newPosition := r.Build(c.table, c.name, c.position)
		c.rules = append(c.rules, rule)
		c.position = newPosition
	}

	return c
}

// BuildForRestore generates all iptables rules within the chain in a format
// suitable for restoration using the `iptables-restore` command
func (c *Chain) BuildForRestore(cfg config.InitializedConfigIPvX) []string {
	var lines []string

	for _, rule := range c.rules {
		lines = append(lines, rule.BuildForRestore(cfg))
	}

	return lines
}

// NewChain creates a new Chain object for a specified iptables table and chain
// name. This function validates that the provided table is one of the supported
// tables (raw, nat, or mangle) and that the chain name is not empty. If these
// conditions are met, it returns a new Chain object initialized with the
// provided table and chain name. If the validation fails, it returns an error
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
// chain name, and panics if the provided table or chain is invalid
func MustNewChain(table consts.TableName, chain string) *Chain {
	newChain, err := NewChain(table, chain)
	if err != nil {
		panic(err)
	}

	return newChain
}
