package parameters

import (
	"strconv"
)

type WaitIntervalParameter struct {
	microseconds uint
}

func (p *WaitIntervalParameter) Build(bool) string {
	return strconv.Itoa(int(p.microseconds))
}

func (p *WaitIntervalParameter) Negate() ParameterBuilder {
	return p
}

// WaitInterval will generate arguments for the "-W, --wait-interval
// microseconds" flag
// Interval to wait per each iteration. When running latency sensitive
// applications, waiting for the xtables lock for extended durations may not be
// acceptable. This option will make each iteration take the amount of time
// specified. The default interval is 1 second. This option only works with -w.
//
// ref. iptables(8) > OTHER OPTIONS
// ref. iptables-restore(8) > DESCRIPTION
func WaitInterval(microseconds uint) *Parameter {
	if microseconds == 0 {
		return nil
	}

	return &Parameter{
		long:      "--wait-interval",
		short:     "--wait-interval",
		connector: "=",
		parameters: []ParameterBuilder{
			&WaitIntervalParameter{microseconds: microseconds},
		},
		negate: nil, // no negation allowed
	}
}
