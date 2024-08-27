package parameters

import (
	"strconv"
)

var _ ParameterBuilder = &JumpParameter{}

type JumpParameter struct {
	parameters []string
}

func (p *JumpParameter) Build(bool) []string {
	return p.parameters
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

func JumpConditional(
	condition bool,
	parameterTrue *JumpParameter,
	parameterFalse *JumpParameter,
) *Parameter {
	if !condition {
		return Jump(parameterFalse)
	}

	return Jump(parameterTrue)
}

func ToUserDefinedChain(chainName string) *JumpParameter {
	return &JumpParameter{parameters: []string{chainName}}
}

func ToPort[T ~uint16](port T) *JumpParameter {
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

// Turn on kernel logging of matching packets. When this option is set for
// a rule, the Linux kernel will print some information on all matching packets
// (like most IP header fields) via the kernel log (where it can be read with
// dmesg or syslogd(8)). This is a "non-terminating target", i.e. rule traversal
// continues at the next rule. So if you want to LOG the packets you refuse, use
// two separate rules with the same matching criteria, first using target LOG
// then DROP (or REJECT).
// ref. iptables(8) > LOG
// levels:
//
//	0 - EMERGENCY	- system is unusable
//	1 - ALERT		- action must be taken immediately
//	2 - CRITICAL	- critical conditions
//	3 - ERR			- error conditions
//	4 - WARNING		- warning conditions
//	5 - NOTICE		- normal but significant condition
//	6 - INFO		- informational
//	7 - DEBUG		- debug-level messages
//
// ref. https://git.netfilter.org/iptables/tree/extensions/libebt_log.c#n27
func Log(prefix string, level uint16) *JumpParameter {
	return &JumpParameter{
		parameters: []string{
			"LOG",
			"--log-prefix", prefix,
			"--log-level", strconv.Itoa(int(level)),
		},
	}
}
