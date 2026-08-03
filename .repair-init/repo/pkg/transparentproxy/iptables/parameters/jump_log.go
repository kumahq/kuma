package parameters

import (
	"fmt"
)

type LogParameter WrappingParameter

func newLogParameter(param string, params ...string) *LogParameter {
	return (*LogParameter)(NewWrappingParameter(param, params...))
}

// Turn on kernel logging of matching packets. When this option is set for
// a rule, the Linux kernel will print some information on all matching packets
// (like most IP header fields) via the kernel log (where it can be read with
// dmesg or syslogd(8)). This is a "non-terminating target", i.e. rule traversal
// continues at the next rule. So if you want to LOG the packets you refuse, use
// two separate rules with the same matching criteria, first using target LOG
// then DROP (or REJECT).
// ref. iptables(8) > LOG
func Log(params ...*LogParameter) *JumpParameter {
	var parameters []string
	for _, param := range params {
		parameters = append(parameters, param.parameters...)
	}

	return newJumpParameter("LOG", parameters...)
}

// Valid levels:
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
func LogLevel(level uint16) *LogParameter {
	return newLogParameter("--log-level", fmt.Sprintf("%d", level))
}

// LogPrefix returns a --log-prefix parameter with prefix ending with ":"
func LogPrefix(prefix string) *LogParameter {
	return newLogParameter("--log-prefix", prefix+":")
}
