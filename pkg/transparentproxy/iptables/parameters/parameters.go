package parameters

import (
	"strings"
)

type ParameterBuilder interface {
	Build(verbose bool) string
	Negate() ParameterBuilder
}

var _ ParameterBuilder = &Parameter{}

func negateNestedParameters(parameter *Parameter) ParameterBuilder {
	for _, parameter := range parameter.parameters {
		parameter.Negate()
	}

	return parameter
}

func negateSelf(parameter *Parameter) ParameterBuilder {
	parameter.negative = !parameter.negative

	return parameter
}

type Parameter struct {
	long       string
	short      string
	parameters []ParameterBuilder
	negate     func(parameter *Parameter) ParameterBuilder
	negative   bool
}

func (p *Parameter) Build(verbose bool) string {
	if p == nil {
		return ""
	}

	flag := p.short

	if verbose {
		flag = p.long
	}

	result := []string{flag}

	if p.negative {
		result = append([]string{"!"}, result...)
	}

	for _, parameter := range p.parameters {
		if parameter != nil {
			result = append(result, parameter.Build(verbose))
		}
	}

	return strings.Join(result, " ")
}

func (p *Parameter) Negate() ParameterBuilder {
	if p == nil || p.negate == nil {
		return p
	}

	return p.negate(p)
}

type Parameters []*Parameter

func NewParameters(parameters ...*Parameter) Parameters {
	return parameters
}

func (p Parameters) Build(verbose bool, additionalParameters ...string) []string {
	var result []string

	for _, parameter := range p {
		builtParameter := parameter.Build(verbose)

		if builtParameter != "" {
			result = append(result, builtParameter)
		}
	}

	return append(result, additionalParameters...)
}
