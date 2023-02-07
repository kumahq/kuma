package parameters

type InInterfaceParameter struct {
	name string
}

func (p *InInterfaceParameter) Build(bool) string {
	return p.name
}

func (p *InInterfaceParameter) Negate() ParameterBuilder {
	return p
}

// InInterface will generate arguments for the "-i, --in-interface name" flag
// Name of an interface via which a packet was received (only for packets
// entering the INPUT, FORWARD and PREROUTING chains). If the interface name
// ends in a "+", then any interface which begins with this name will match
//
// ref. iptables(8) > PARAMETERS
func InInterface(name string) *Parameter {
	return &Parameter{
		long:       "--in-interface",
		short:      "-i",
		parameters: []ParameterBuilder{&InInterfaceParameter{name: name}},
		negate:     negateSelf,
	}
}
