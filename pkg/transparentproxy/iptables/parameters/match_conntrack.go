package parameters

import (
	"strings"

	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/parameters/match/conntrack"
)

// Conntrack
//       This module, when combined with connection tracking, allows access
//       to the connection tracking state for this packet/connection.
//
//       [!]  --ctstate statelist
//              statelist is a comma separated list of the connection states to match.
//              Possible states are listed in ./parameters/match/conntrack
//
// ref. iptables-extensions(8) > conntrack

var _ ParameterBuilder = &ConntrackParameter{}

type ConntrackParameter struct {
	flag     string
	values   []string
	negative bool
}

func (p *ConntrackParameter) Negate() ParameterBuilder {
	p.negative = !p.negative

	return p
}

func (p *ConntrackParameter) Build(bool) []string {
	value := strings.Join(p.values, ",")

	if p.negative {
		return []string{"!", p.flag, value}
	}

	return []string{p.flag, value}
}

// Ctstate expects at least one state is necessary, so that's the reason for split
// of parameters
func Ctstate(state conntrack.State, states ...conntrack.State) *ConntrackParameter {
	values := []string{string(state)}

	for _, s := range states {
		values = append(values, string(s))
	}

	return &ConntrackParameter{
		flag:     "--ctstate",
		values:   values,
		negative: false,
	}
}

// Conntrack when combined with connection tracking, allows access to the connection
// tracking state for this packet/connection
func Conntrack(conntrackParameters ...*ConntrackParameter) *MatchParameter {
	var parameters []ParameterBuilder

	for _, parameter := range conntrackParameters {
		parameters = append(parameters, parameter)
	}

	return &MatchParameter{
		name:       "conntrack",
		parameters: parameters,
	}
}
