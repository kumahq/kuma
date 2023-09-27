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
	connector  string
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

	var parameters []string
	for _, parameter := range p.parameters {
		if parameter != nil {
			parameters = append(parameters, parameter.Build(verbose))
		}
	}

	// Some parameters for flags like "--wait" or "--wait-interval" require an
	// equal sign to be set, so "--wait 5" is invalid and should be "--wait=5"
	// to work. If the `Parameter` object has a `connector` property and only
	// one value, we will use it when joining the flag with the value.
	connector := " "
	if p.connector != "" && len(parameters) == 1 {
		connector = p.connector
	}

	return strings.Join(append(result, parameters...), connector)
}

func (p *Parameter) Negate() ParameterBuilder {
	if p == nil || p.negate == nil {
		return p
	}

	return p.negate(p)
}

type Parameters []*Parameter

func NewParameters(parameters ...*Parameter) *Parameters {
	var result Parameters
	result = append(result, parameters...)
	return &result
}

func (p *Parameters) Build(verbose bool, additionalParameters ...string) []string {
	var result []string

	for _, parameter := range *p {
		builtParameter := parameter.Build(verbose)

		if builtParameter != "" {
			result = append(result, builtParameter)
		}
	}

	return append(result, additionalParameters...)
}

func (p *Parameters) Append(parameters ...*Parameter) *Parameters {
	return p.AppendIf(true, parameters...)
}

func (p *Parameters) AppendIf(predicate bool, parameters ...*Parameter) *Parameters {
	if predicate {
		*p = append(*p, parameters...)
	}

	return p
}
