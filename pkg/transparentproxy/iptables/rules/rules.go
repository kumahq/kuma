package rules

import (
	"strconv"
	"strings"

	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/parameters"
)

type Rule struct {
	chainName  string
	position   int
	parameters parameters.Parameters
}

func (r *Rule) Build(verbose bool) string {
	var flag string

	switch {
	case r.position == 0 && verbose:
		flag = "--append"
	case r.position == 0 && !verbose:
		flag = "-A"
	case r.position != 0 && verbose:
		flag = "--insert"
	case r.position != 0 && !verbose:
		flag = "-I"
	}

	cmd := []string{flag}

	if r.chainName != "" {
		cmd = append(cmd, r.chainName)
	}

	if r.position != 0 {
		cmd = append(cmd, strconv.Itoa(r.position))
	}

	return strings.Join(append(cmd, r.parameters.Build(verbose)...), " ")
}

func Append(chainName string, parameters []*parameters.Parameter) *Rule {
	return &Rule{
		position:   0,
		chainName:  chainName,
		parameters: parameters,
	}
}

func Insert(chainName string, position int, parameters []*parameters.Parameter) *Rule {
	return &Rule{
		chainName:  chainName,
		position:   position,
		parameters: parameters,
	}
}
