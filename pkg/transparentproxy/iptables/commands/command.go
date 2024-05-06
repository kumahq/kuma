package commands

import (
	"strconv"
	"strings"

	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/parameters"
)

type Command struct {
	chainName  string
	position   int
	parameters parameters.Parameters
}

func (c *Command) Build(verbose bool) string {
	var flag string

	switch {
	case c.position == 0 && verbose:
		flag = "--append"
	case c.position == 0 && !verbose:
		flag = "-A"
	case c.position != 0 && verbose:
		flag = "--insert"
	case c.position != 0 && !verbose:
		flag = "-I"
	}

	cmd := []string{flag}

	if c.chainName != "" {
		cmd = append(cmd, c.chainName)
	}

	if c.position != 0 {
		cmd = append(cmd, strconv.Itoa(c.position))
	}

	return strings.Join(append(cmd, c.parameters.Build(verbose)...), " ")
}

func Append(chainName string, parameters []*parameters.Parameter) *Command {
	return &Command{
		position:   0,
		chainName:  chainName,
		parameters: parameters,
	}
}

func Insert(chainName string, position int, parameters []*parameters.Parameter) *Command {
	return &Command{
		chainName:  chainName,
		position:   position,
		parameters: parameters,
	}
}
