package parameters

import (
	"fmt"
)

type ParameterBuilder interface {
	Build(verbose bool) []string
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

func (p *Parameter) Build(verbose bool) []string {
	if p == nil {
		return nil
	}

	flag := p.short

	// If verbose is true or the short version is not available, use the long version
	if verbose || p.short == "" {
		flag = p.long
	}

	result := []string{flag}

	if p.negative {
		result = append([]string{"!"}, result...)
	}

	var parameters []string
	for _, parameter := range p.parameters {
		if parameter != nil {
			parameters = append(parameters, parameter.Build(verbose)...)
		}
	}

	// Some parameters for flags like "--wait" or "--wait-interval" require an
	// equal sign to be set, so "--wait 5" is invalid and should be "--wait=5"
	// to work. If the `Parameter` object has a `connector` property and only
	// one value, we will use it when joining the flag with the value
	if p.connector != "" && len(parameters) == 1 {
		lastElem := result[len(result)-1]
		resultWithoutLastElem := result[: len(result)-1 : len(result)-1] // full slice expression to protect before append overwrites

		return append(
			resultWithoutLastElem,
			fmt.Sprintf("%s%s%s", lastElem, p.connector, parameters[0]),
		)
	}

	return append(result, parameters...)
}

func (p *Parameter) Negate() ParameterBuilder {
	if p == nil || p.negate == nil {
		return p
	}

	return p.negate(p)
}

type Parameters []*Parameter

func (p *Parameters) Build(verbose bool, additionalParameters ...string) []string {
	var result []string

	for _, parameter := range *p {
		result = append(result, parameter.Build(verbose)...)
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

var _ ParameterBuilder = &SimpleParameter{}

// SimpleParameter represents a straightforward iptables parameter that doesn't
// involve nested parameters or complex logic. It holds both the short and long
// versions of the parameter name, as well as its value
type SimpleParameter struct {
	short string // Short version of the parameter (e.g., "-j").
	long  string // Long version of the parameter (e.g., "--jump").
	value string // The value associated with the parameter (e.g., "ACCEPT").
}

// Build constructs the command-line representation of the SimpleParameter.
// It returns a slice of strings containing the flag (short or long based on
// verbosity) followed by the parameter value
func (p *SimpleParameter) Build(verbose bool) []string {
	flag := p.short
	// If verbose is true or the short version is not available, use the long version
	if verbose || p.short == "" {
		flag = p.long
	}

	return []string{flag, p.value}
}

// Negate is a no-op for SimpleParameter since negation does not apply to simple
// parameters. It simply returns the receiver
func (p *SimpleParameter) Negate() ParameterBuilder {
	return p
}

var _ ParameterBuilder = &WrappingParameter{}

// WrappingParameter is a utility type that encapsulates a slice of strings,
// allowing for flexible parameter handling in scenarios where only raw string
// arguments are needed
type WrappingParameter struct {
	parameters []string
}

func NewWrappingParameter(param string, params ...string) *WrappingParameter {
	return &WrappingParameter{parameters: append([]string{param}, params...)}
}

func (p *WrappingParameter) Build(bool) []string {
	return p.parameters
}

func (p *WrappingParameter) Negate() ParameterBuilder {
	return p
}
