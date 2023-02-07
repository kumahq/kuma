package parameters

import (
	"strings"
)

type MatchParameter struct {
	name       string
	parameters []ParameterBuilder
}

func (p *MatchParameter) Build(verbose bool) string {
	result := []string{p.name}

	for _, parameter := range p.parameters {
		result = append(result, parameter.Build(verbose))
	}

	return strings.Join(result, " ")
}

func (p *MatchParameter) Negate() ParameterBuilder {
	for _, parameter := range p.parameters {
		parameter.Negate()
	}

	return p
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
