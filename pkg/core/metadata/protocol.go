package metadata

import (
	"slices"
	"strings"

	util_slices "github.com/kumahq/kuma/pkg/util/slices"
)

// Protocol identifies a protocol supported by a service.
type Protocol string

const (
	ProtocolUnknown   Protocol = "<unknown>"
	ProtocolTCP       Protocol = "tcp"
	ProtocolTLS       Protocol = "tls"
	ProtocolHTTP      Protocol = "http"
	ProtocolHTTP2     Protocol = "http2"
	ProtocolGRPC      Protocol = "grpc"
	ProtocolKafka     Protocol = "kafka"
	ProtocolMysql     Protocol = "mysql"
	ProtocolRawBuffer Protocol = "raw_buffer"
)

func (p Protocol) String() string { return string(p) }

func ParseProtocol[T ~string](protocol T) Protocol {
	switch proto := Protocol(strings.ToLower(string(protocol))); proto {
	case ProtocolHTTP, ProtocolHTTP2, ProtocolTCP, ProtocolTLS, ProtocolGRPC, ProtocolKafka, ProtocolMysql:
		return proto
	default:
		return ProtocolUnknown
	}
}

func IsHTTP[T ~string](protocol T) bool {
	return ParseProtocol(protocol) == ProtocolHTTP
}

func IsHTTPBased[T ~string](protocol T) bool {
	return HTTPBasedProtocols.Contains(ParseProtocol(protocol))
}

func IsHTTP2Based[T ~string](protocol T) bool {
	return HTTP2BasedProtocols.Contains(ParseProtocol(protocol))
}

// ProtocolList represents a list of Protocols.
type ProtocolList []Protocol

// SupportedProtocols is a list of supported protocols that will be communicated to a user.
var SupportedProtocols = ProtocolList{
	ProtocolGRPC,
	ProtocolHTTP,
	ProtocolHTTP2,
	ProtocolKafka,
	ProtocolTCP,
}

var HTTPBasedProtocols = ProtocolList{
	ProtocolHTTP,
	ProtocolHTTP2,
	ProtocolGRPC,
}

var HTTP2BasedProtocols = ProtocolList{
	ProtocolHTTP2,
	ProtocolGRPC,
}

func (l ProtocolList) Strings() []string {
	return util_slices.Map(l, Protocol.String)
}

func (l ProtocolList) Contains(protocol Protocol) bool {
	return slices.Contains(l, protocol)
}

// protocolStack is a mapping between a protocol and its full protocol stack, e.g.
// HTTP has a protocol stack [HTTP, TCP],
// GRPC has a protocol stack [GRPC, HTTP2, TCP],
// TCP  has a protocol stack [TCP].
var protocolStacks = map[Protocol]ProtocolList{
	ProtocolGRPC:  {ProtocolGRPC, ProtocolHTTP2, ProtocolTCP},
	ProtocolHTTP2: {ProtocolHTTP2, ProtocolTCP},
	ProtocolHTTP:  {ProtocolHTTP, ProtocolTCP},
	ProtocolKafka: {ProtocolKafka, ProtocolTCP},
	ProtocolTLS:   {ProtocolTCP},
	ProtocolTCP:   {ProtocolTCP},
}

// GetCommonProtocol returns a common protocol between given two.
//
// E.g.,
// a common protocol between HTTP and HTTP2 is HTTP2,
// a common protocol between HTTP and HTTP  is HTTP,
// a common protocol between HTTP and TCP   is TCP,
// a common protocol between GRPC and HTTP2 is HTTP2,
// a common protocol between HTTP and HTTP2 is HTTP.
func GetCommonProtocol[T ~string, U ~string](a T, b U) Protocol {
	protoA, protoB := ParseProtocol(a), ParseProtocol(b)

	switch {
	case string(a) == string(b):
		return ParseProtocol(a)
	case a == "" || b == "":
		return ProtocolUnknown
	case protoA == ProtocolUnknown || protoB == ProtocolUnknown:
		return ProtocolUnknown
	}

	var stackA, stackB ProtocolList
	var ok bool

	if stackA, ok = protocolStacks[protoA]; !ok {
		return ProtocolUnknown
	}

	if stackB, ok = protocolStacks[protoB]; !ok {
		return ProtocolUnknown
	}

	for _, pA := range stackA {
		for _, pB := range stackB {
			if pA == pB {
				return pA
			}
		}
	}

	return ProtocolUnknown
}
