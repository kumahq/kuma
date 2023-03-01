package parameters

type OutInterfaceParameter struct {
	name string
}

func (p *OutInterfaceParameter) Build(bool) string {
	return p.name
}

func (p *OutInterfaceParameter) Negate() ParameterBuilder {
	return p
}

// OutInterface will generate arguments for the "-o, --out-interface name" flag
// Name of an interface via which a packet is going to be sent (for packets entering the FORWARD,
// OUTPUT and POSTROUTING chains). If the interface name ends in a "+", then any interface
// which begins with this name will match
//
// ref. iptables(8) > PARAMETERS
func OutInterface(name string) *Parameter {
	return &Parameter{
		long:       "--out-interface",
		short:      "-o",
		parameters: []ParameterBuilder{&OutInterfaceParameter{name: name}},
		negate:     negateSelf,
	}
}
