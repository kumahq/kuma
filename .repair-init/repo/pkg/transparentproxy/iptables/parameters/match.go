package parameters

import (
	"fmt"
	"strings"
)

var _ ParameterBuilder = &MatchParameter{}

type MatchParameter struct {
	name       string
	parameters []ParameterBuilder
}

func (p *MatchParameter) Build(verbose bool) []string {
	result := []string{p.name}

	for _, parameter := range p.parameters {
		result = append(result, parameter.Build(verbose)...)
	}

	return result
}

func (p *MatchParameter) Negate() ParameterBuilder {
	for _, parameter := range p.parameters {
		parameter.Negate()
	}

	return p
}

func Multiport() *MatchParameter {
	return &MatchParameter{
		name:       "multiport",
		parameters: []ParameterBuilder{},
	}
}

// Comment allows you to add comments (up to 256 characters) to any rule.
//
//	Example:
//	  iptables -A INPUT -i eth1 -m comment --comment "my local LAN"
//
// ref. iptables-extensions(8) > comment
func Comment(comments ...string) *MatchParameter {
	comment := strings.Join(comments, "/")
	if len(comment) > 254 {
		comment = comment[:254]
	}

	return &MatchParameter{
		name: "comment",
		parameters: []ParameterBuilder{
			&SimpleParameter{
				long:  "--comment",
				value: fmt.Sprintf("%q", comment),
			},
		},
	}
}

func Match(matchParameters ...*MatchParameter) *Parameter {
	var parameters []ParameterBuilder

	for _, parameter := range matchParameters {
		parameters = append(parameters, parameter)
	}

	return &Parameter{
		long:       "--match",
		short:      "-m",
		parameters: parameters,
		negate:     negateNestedParameters,
	}
}

func MatchIf(predicate bool, matchParameters ...*MatchParameter) *Parameter {
	if !predicate {
		return nil
	}

	return Match(matchParameters...)
}
