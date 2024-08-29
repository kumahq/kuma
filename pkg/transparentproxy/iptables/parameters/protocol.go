package parameters

import (
	"strconv"

	"github.com/kumahq/kuma/pkg/transparentproxy/config"
	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/consts"
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

func destinationPort[T ~uint16](port T, negative bool) *TcpUdpParameter {
	return &TcpUdpParameter{
		long:     "--destination-port",
		short:    "--dport",
		value:    strconv.Itoa(int(port)),
		negative: negative,
	}
}

func DestinationPort[T ~uint16](port T) *TcpUdpParameter {
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

func NotDestinationPort[T ~uint16](port T) *TcpUdpParameter {
	return destinationPort(port, true)
}

func NotDestinationPortIf[T ~uint16](predicate func() bool, port T) *TcpUdpParameter {
	return NotDestinationPortIfBool(predicate(), port)
}

func NotDestinationPortIfBool[T ~uint16](condition bool, port T) *TcpUdpParameter {
	if condition {
		return destinationPort(port, true)
	}

	return nil
}

func sourcePort[T ~uint16](port T, negative bool) *TcpUdpParameter {
	return &TcpUdpParameter{
		long:     "--source-port",
		short:    "--sport",
		value:    strconv.Itoa(int(port)),
		negative: negative,
	}
}

func SourcePort[T ~uint16](port T) *TcpUdpParameter {
	return sourcePort(port, false)
}

func tcpUdp(proto consts.ProtocolL4, params []*TcpUdpParameter) *ProtocolParameter {
	var parameters []ParameterBuilder

	for _, parameter := range params {
		if parameter != nil {
			parameters = append(parameters, parameter)
		}
	}

	return &ProtocolParameter{
		name:       string(proto),
		parameters: parameters,
	}
}

func Udp(udpParameters ...*TcpUdpParameter) *ProtocolParameter {
	return tcpUdp(consts.ProtocolUDP, udpParameters)
}

func UdpIf(predicate bool, udpParameters ...*TcpUdpParameter) *ProtocolParameter {
	if !predicate {
		return nil
	}

	return tcpUdp(consts.ProtocolUDP, udpParameters)
}

func Tcp(tcpParameters ...*TcpUdpParameter) *ProtocolParameter {
	return tcpUdp(consts.ProtocolTCP, tcpParameters)
}

func TcpIf(predicate bool, tcpParameters ...*TcpUdpParameter) *ProtocolParameter {
	if !predicate {
		return nil
	}

	return tcpUdp(consts.ProtocolTCP, tcpParameters)
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
