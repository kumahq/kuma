package parameters

import (
	"strconv"
	"strings"
)

type JumpParameter struct {
	parameters []string
}

func (p *JumpParameter) Build(bool) string {
	return strings.Join(p.parameters, " ")
}

func (p *JumpParameter) Negate() ParameterBuilder {
	return p
}

func Jump(parameter *JumpParameter) *Parameter {
	return &Parameter{
		long:       "--jump",
		short:      "-j",
		parameters: []ParameterBuilder{parameter},
	}
}

func ToUserDefinedChain(chainName string) *JumpParameter {
	return &JumpParameter{parameters: []string{chainName}}
}

func ToPort(port uint16) *JumpParameter {
	return &JumpParameter{parameters: []string{
		"REDIRECT",
		"--to-ports",
		strconv.Itoa(int(port)),
	}}
}

func Return() *JumpParameter {
	return &JumpParameter{parameters: []string{"RETURN"}}
}

func Drop() *JumpParameter {
	return &JumpParameter{parameters: []string{"DROP"}}
}

func Log(prefix string, level uint16) *JumpParameter {
	return &JumpParameter{
		parameters: []string{
			"LOG",
			"--log-prefix", prefix,
			"--log-level", strconv.Itoa(int(level)),
		},
	}
}
