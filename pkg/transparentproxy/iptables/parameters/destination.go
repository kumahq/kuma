package parameters

type DestinationParameter struct {
	address string
}

func (p *DestinationParameter) Build(bool) string {
	return p.address
}

func (p *DestinationParameter) Negate() ParameterBuilder {
	return p
}

func destination(address string, negative bool) *Parameter {
	return &Parameter{
		long:       "--destination",
		short:      "-d",
		parameters: []ParameterBuilder{&DestinationParameter{address: address}},
		negate:     negateSelf,
		negative:   negative,
	}
}

// Destination will generate arguments for the "-d, --destination address[/mask]" flag
// See the description of the -s (source) flag for a detailed description of the syntax
//
// ref. iptables(8) > PARAMETERS
func Destination(address string) *Parameter {
	return destination(address, false)
}

func NotDestination(address string) *Parameter {
	return destination(address, true)
}
