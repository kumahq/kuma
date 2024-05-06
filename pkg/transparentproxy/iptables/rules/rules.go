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

func (r *Rule) Build(verbose bool) string {
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
