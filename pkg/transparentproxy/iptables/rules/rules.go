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
	tableName  string
	chainName  string
	position   uint
	parameters parameters.Parameters
}

func NewRule(tableName, chainName string, position uint, parameters []*parameters.Parameter) *Rule {
	return &Rule{
		tableName:  tableName,
		chainName:  chainName,
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

	if r.chainName != "" {
		cmd = append(cmd, r.chainName)
	}

	if r.position != 0 {
		cmd = append(cmd, fmt.Sprintf("%d", r.position))
	}

	return strings.Join(append(cmd, r.parameters.Build(verbose)...), " ")
}
