package parameters

import (
	"strconv"

	"github.com/kumahq/kuma/pkg/transparentproxy/config"
)

var _ ParameterBuilder = &ProtocolParameter{}

type ProtocolParameter struct {
	name       string
	parameters []ParameterBuilder
}

func (p *ProtocolParameter) Build(verbose bool) []string {
	result := []string{p.name}

	// If the -p or --protocol was specified and if and only if an unknown option is encountered,
	// iptables will try load a match module of the same name as the protocol, to try making
	// the option available, so we don't have to add --match tcp or -m tcp parameters to the rule
	//
	// ref. iptables-extensions(8) > MATCH EXTENSIONS
	for _, parameter := range p.parameters {
		if parameter != nil {
			result = append(result, parameter.Build(verbose)...)
		}
	}

	return result
}

func (p *ProtocolParameter) Negate() ParameterBuilder {
	for _, parameter := range p.parameters {
		parameter.Negate()
	}

	return p
}

var _ ParameterBuilder = &TcpUdpParameter{}

type TcpUdpParameter struct {
	long     string
	short    string
	value    string
	negative bool
}

func (p *TcpUdpParameter) Build(verbose bool) []string {
	if p.value == "" {
		return nil
	}

	flag := p.short

	if verbose {
		flag = p.long
	}

	result := []string{flag}

	if p.negative {
		result = append([]string{"!"}, result...)
	}

	return append(result, p.value)
}

func (p *TcpUdpParameter) Negate() ParameterBuilder {
	p.negative = !p.negative

	return p
}

func destinationPort(port uint16, negative bool) *TcpUdpParameter {
	return &TcpUdpParameter{
		long:     "--destination-port",
		short:    "--dport",
		value:    strconv.Itoa(int(port)),
		negative: negative,
	}
}

func DestinationPort(port uint16) *TcpUdpParameter {
	return destinationPort(port, false)
}

func DestinationPortRangeOrValue(exclusion config.Exclusion) *TcpUdpParameter {
	if exclusion.Ports == "" {
		return nil
	}

	return &TcpUdpParameter{
		long:  "--destination-port",
		short: "--dport",
		value: string(exclusion.Ports),
	}
}

func NotDestinationPort(port uint16) *TcpUdpParameter {
	return destinationPort(port, true)
}

func NotDestinationPortIf(predicate func() bool, port uint16) *TcpUdpParameter {
	return NotDestinationPortIfBool(predicate(), port)
}

func NotDestinationPortIfBool(condition bool, port uint16) *TcpUdpParameter {
	if condition {
		return destinationPort(port, true)
	}

	return nil
}

func sourcePort(port uint16, negative bool) *TcpUdpParameter {
	return &TcpUdpParameter{
		long:     "--source-port",
		short:    "--sport",
		value:    strconv.Itoa(int(port)),
		negative: negative,
	}
}

func SourcePort(port uint16) *TcpUdpParameter {
	return sourcePort(port, false)
}

func tcpUdp(proto string, params []*TcpUdpParameter) *ProtocolParameter {
	var parameters []ParameterBuilder

	for _, parameter := range params {
		if parameter != nil {
			parameters = append(parameters, parameter)
		}
	}

	return &ProtocolParameter{
		name:       proto,
		parameters: parameters,
	}
}

func Udp(udpParameters ...*TcpUdpParameter) *ProtocolParameter {
	return tcpUdp("udp", udpParameters)
}

func UdpIf(predicate bool, udpParameters ...*TcpUdpParameter) *ProtocolParameter {
	if !predicate {
		return nil
	}

	return tcpUdp("udp", udpParameters)
}

func Tcp(tcpParameters ...*TcpUdpParameter) *ProtocolParameter {
	return tcpUdp("tcp", tcpParameters)
}

func TcpIf(predicate bool, tcpParameters ...*TcpUdpParameter) *ProtocolParameter {
	if !predicate {
		return nil
	}

	return tcpUdp("tcp", tcpParameters)
}

func Protocol(p ...*ProtocolParameter) *Parameter {
	var parameters []ParameterBuilder
	for _, parameter := range p {
		if parameter != nil {
			parameters = append(parameters, parameter)
		}
	}

	if parameters == nil {
		return nil
	}

	return &Parameter{
		long:       "--protocol",
		short:      "-p",
		parameters: parameters,
		negate:     negateSelf,
	}
}
