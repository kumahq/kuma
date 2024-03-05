package parameters

import (
	"strconv"
)

type WaitParameter struct {
	seconds uint
}

func (p *WaitParameter) Build(bool) string {
	return strconv.Itoa(int(p.seconds))
}

func (p *WaitParameter) Negate() ParameterBuilder {
	return p
}

// Wait will generate arguments for the "-w, --wait [seconds]" flag
// Wait for the xtables lock. To prevent multiple instances of the program from
// running concurrently, an attempt will be made to obtain an exclusive lock
// at launch. By default, the program will exit if the lock cannot be obtained.
// This option will make the program wait (indefinitely  or  for  optional
// seconds) until the exclusive lock can be obtained.
//
// ref. iptables(8) > OTHER OPTIONS
// ref. iptables-restore(8) > DESCRIPTION
func Wait(seconds uint) *Parameter {
	if seconds == 0 {
		return nil
	}

	return &Parameter{
		long:       "--wait",
		short:      "--wait",
		connector:  "=",
		parameters: []ParameterBuilder{&WaitParameter{seconds: seconds}},
		negate:     nil, // no negation allowed
	}
}
