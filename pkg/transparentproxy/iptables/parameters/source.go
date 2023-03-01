package parameters

type SourceParameter struct {
	address string
}

func (p *SourceParameter) Build(bool) string {
	return p.address
}

func (p *SourceParameter) Negate() ParameterBuilder {
	return p
}

func Address(address string) *SourceParameter {
	return &SourceParameter{address: address}
}

// Source will generate arguments for the "-s, --source address[/mask]" flag
// Address can be either a network name, a hostname, a network IP address (with /mask),
// or a plain IP address. Hostnames will be resolved once only, before the rule is submitted
// to the kernel. Please note that specifying any name to be resolved with a remote query such as
// DNS is a horrible idea. The mask can be either an ipv4 network mask (for iptables) or
// a plain number, specifying the number of 1's on the left side of the network mask.
// Thus, an iptables mask of 24 is equivalent to 255.255.255.0
//
// ref. iptables(8) > PARAMETERS
func Source(parameter *SourceParameter) *Parameter {
	return &Parameter{
		long:       "--source",
		short:      "-s",
		parameters: []ParameterBuilder{parameter},
		negate:     negateSelf,
	}
}
