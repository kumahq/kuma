package parameters

import (
	"net"
	"reflect"
)

var _ ParameterBuilder = &EndpointParameter{}

type EndpointParameter struct {
	address string
}

func (p *EndpointParameter) Build(bool) []string {
	return []string{p.address}
}

func (p *EndpointParameter) Negate() ParameterBuilder {
	return p
}

func endpoint[T ~string | net.IP | net.IPNet](long, short string, value T, negative bool) *Parameter {
	var address string
	if address = extractAddress(value); address == "" {
		return nil
	}

	return &Parameter{
		long:       long,
		short:      short,
		parameters: []ParameterBuilder{&EndpointParameter{address: address}},
		negate:     negateSelf,
		negative:   negative,
	}
}

func extractAddress[T ~string | net.IP | net.IPNet](value T) string {
	if reflect.ValueOf(&value).Elem().IsZero() {
		return ""
	}

	switch v := any(value).(type) {
	case string:
		return v
	case net.IP:
		return v.String()
	case net.IPNet:
		return v.String()
	}

	// handle the remaining type set of ~string
	if r := reflect.ValueOf(value); r.Kind() == reflect.String {
		return r.String()
	}

	return ""
}

func source[T ~string | net.IP | net.IPNet](value T, negative bool) *Parameter {
	return endpoint("--source", "-s", value, negative)
}

func destination[T ~string | net.IP | net.IPNet](value T, negative bool) *Parameter {
	return endpoint("--destination", "-d", value, negative)
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
func Source[T ~string | net.IP | net.IPNet](address T) *Parameter {
	return source(address, false)
}

// Destination will generate arguments for the "-d, --destination address[/mask]" flag
// See the description of the -s (source) flag for a detailed description of the syntax
//
// ref. iptables(8) > PARAMETERS
func Destination[T ~string | net.IP | net.IPNet](address T) *Parameter {
	return destination(address, false)
}

func NotDestination[T ~string | net.IP | net.IPNet](address T) *Parameter {
	return destination(address, true)
}
