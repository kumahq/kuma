package rules

import (
	"fmt"
	"strings"

	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/parameters"
)

const (
	appendLong  = "--append"
	appendShort = "-A"
	insertLong  = "--insert"
	insertShort = "-I"
)

type Rule struct {
	table      string
	chain      string
	position   uint
	parameters parameters.Parameters
}

func NewRule(table, chain string, position uint, parameters []*parameters.Parameter) *Rule {
	return &Rule{
		table:      table,
		chain:      chain,
		position:   position,
		parameters: parameters,
	}
}

// BuildForRestore function generates iptables rule in a format suitable for
// restoration using the `iptables-restore`
//
// The function takes a boolean argument `verbose` which controls the output
// format. When true, it uses longer flags (e.g., "--append"). Otherwise, it
// uses shorter flags (e.g., "-A").
//
// The position argument can be used to specify where to insert the rule
// within the chain. If position is 0 and verbose is true, the rule is appended
// using the "--append" flag. If position is 0 and verbose is false, the rule
// is appended using the "-A" flag. Otherwise, the rule is inserted at the
// specified position using the "--insert" or "-I" flag accordingly.
//
// The final command string is a space-separated concatenation of the flag,
// chain name, optional position (if not zero), and the parameters built using
// the `parameters.Build(verbose)` method.
func (r *Rule) BuildForRestore(verbose bool) string {
	var flag string

	switch {
	case r.position == 0 && verbose:
		flag = appendLong
	case r.position == 0 && !verbose:
		flag = appendShort
	case r.position != 0 && verbose:
		flag = insertLong
	case r.position != 0 && !verbose:
		flag = insertShort
	}

	cmd := []string{flag}

	if r.chain != "" {
		cmd = append(cmd, r.chain)
	}

	if r.position != 0 {
		cmd = append(cmd, fmt.Sprintf("%d", r.position))
	}

	return strings.Join(append(cmd, r.parameters.Build(verbose)...), " ")
}
