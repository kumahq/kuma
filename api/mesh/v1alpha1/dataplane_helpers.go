package v1alpha1

import (
	"fmt"
	"net"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

const (
	// Mandatory tag that has a reserved meaning in Kuma.
	ServiceTag     = "service"
	ServiceUnknown = "unknown"
	// Optional tag that has a reserved meaning in Kuma.
	// If absent, Kuma will treat application's protocol as opaque TCP.
	ProtocolTag = "protocol"
)

// ServiceTagValue represents the value of "service" tag.
//
// E.g., "web", "backend", "database" are typical values in universal case,
// "web.default.svc:80" in k8s case.
type ServiceTagValue string

func (v ServiceTagValue) HasPort() bool {
	_, _, err := net.SplitHostPort(string(v))
	return err == nil
}

func (v ServiceTagValue) HostAndPort() (string, uint32, error) {
	host, port, err := net.SplitHostPort(string(v))
	if err != nil {
		return "", 0, err
	}
	num, err := strconv.ParseUint(port, 10, 32)
	if err != nil {
		return "", 0, err
	}
	return host, uint32(num), nil
}

type InboundInterface struct {
	DataplaneIP   string
	DataplanePort uint32
	WorkloadPort  uint32
}

func (i InboundInterface) String() string {
	return fmt.Sprintf("%s:%d:%d", i.DataplaneIP, i.DataplanePort, i.WorkloadPort)
}

type OutboundInterface struct {
	DataplaneIP   string
	DataplanePort uint32
}

func (i OutboundInterface) String() string {
	return fmt.Sprintf("%s:%d", i.DataplaneIP, i.DataplanePort)
}

const inboundInterfaceBNF = `( IPv4 | '[' IPv6 ']' ) ':' DATAPLANE_PORT ':' WORKLOAD_PORT`

func ParseInboundInterface(text string) (InboundInterface, error) {
	wpInd := strings.LastIndex(text, ":")
	if wpInd < 0 {
		return InboundInterface{}, errors.Errorf("invalid format: expected %q, got %q", inboundInterfaceBNF, text)
	}
	dpInd := strings.LastIndex(text[:wpInd], ":")
	if dpInd < 0 {
		return InboundInterface{}, errors.Errorf("invalid format: expected %q, got %q", inboundInterfaceBNF, text)
	}
	host, port, err := net.SplitHostPort(text[:wpInd])
	if err != nil {
		return InboundInterface{}, errors.Errorf("invalid format: expected %q, got %q", inboundInterfaceBNF, text)
	}
	dataplaneIP, err := ParseIP(host)
	if err != nil {
		return InboundInterface{}, errors.Wrapf(err, "invalid DATAPLANE_IP in %q", text)
	}
	dataplanePort, err := ParsePort(port)
	if err != nil {
		return InboundInterface{}, errors.Wrapf(err, "invalid DATAPLANE_PORT in %q", text)
	}
	workloadPort, err := ParsePort(text[wpInd+1:])
	if err != nil {
		return InboundInterface{}, errors.Wrapf(err, "invalid WORKLOAD_PORT in %q", text)
	}
	return InboundInterface{
		DataplaneIP:   dataplaneIP,
		DataplanePort: dataplanePort,
		WorkloadPort:  workloadPort,
	}, nil
}

const outboundInterfaceBNF = `[ IPv4 | '[' IPv6 ']' ] ':' DATAPLANE_PORT`

func ParseOutboundInterface(text string) (OutboundInterface, error) {
	wpInd := strings.LastIndex(text, ":")
	if wpInd < 0 {
		return OutboundInterface{}, errors.Errorf("invalid format: expected %q, got %q", outboundInterfaceBNF, text)
	}
	port := text[wpInd+1:]
	var dataplaneIP string
	if wpInd == 0 {
		dataplaneIP = "127.0.0.1"
	} else {
		var err error
		dataplaneIP, port, err = net.SplitHostPort(text)
		if err != nil {
			return OutboundInterface{}, errors.Errorf("invalid format: expected %q, got %q", outboundInterfaceBNF, text)
		}
		dataplaneIP, err = ParseIP(dataplaneIP)
		if err != nil {
			return OutboundInterface{}, errors.Wrapf(err, "invalid DATAPLANE_IP in %q", text)
		}
	}
	dataplanePort, err := ParsePort(port)
	if err != nil {
		return OutboundInterface{}, errors.Wrapf(err, "invalid DATAPLANE_PORT in %q", text)
	}
	return OutboundInterface{
		DataplaneIP:   dataplaneIP,
		DataplanePort: dataplanePort,
	}, nil
}

func (n *Dataplane_Networking) GetOutboundInterfaces() ([]OutboundInterface, error) {
	if n == nil {
		return nil, nil
	}
	ofaces := make([]OutboundInterface, len(n.Outbound))
	for i, outbound := range n.Outbound {
		if outbound.Interface != "" { // legacy format
			oface, err := ParseOutboundInterface(outbound.Interface)
			if err != nil {
				return nil, err
			}
			ofaces[i] = oface
		} else {
			oface := OutboundInterface{
				DataplanePort: outbound.Port,
			}
			if outbound.Address != "" {
				oface.DataplaneIP = outbound.Address
			} else {
				oface.DataplaneIP = "127.0.0.1"
			}
			ofaces[i] = oface
		}
	}
	return ofaces, nil
}

func ParsePort(text string) (uint32, error) {
	port, err := strconv.ParseUint(text, 10, 32)
	if err != nil {
		return 0, errors.Wrapf(err, "%q is not a valid port number", text)
	}
	if port < 1 || 65535 < port {
		return 0, errors.Errorf("port number must be in the range [1, 65535] but got %d", port)
	}
	return uint32(port), nil
}

func ParseIP(text string) (string, error) {
	if net.ParseIP(text) == nil {
		return "", errors.Errorf("%q is not a valid IP address", text)
	}
	return text, nil
}

func (n *Dataplane_Networking) GetInboundInterface(service string) (*InboundInterface, error) {
	for i, inbound := range n.Inbound {
		if inbound.Tags[ServiceTag] != service {
			continue
		}
		iface, err := n.GetInboundInterfaceByIdx(i)
		return &iface, err
	}
	return nil, errors.Errorf("Dataplane has no Inbound Interface for service %q", service)
}

func (n *Dataplane_Networking) GetInboundInterfaces() ([]InboundInterface, error) {
	if n == nil {
		return nil, nil
	}
	ifaces := make([]InboundInterface, len(n.Inbound))
	for i, _ := range n.Inbound {
		iface, err := n.GetInboundInterfaceByIdx(i)
		if err != nil {
			return nil, err
		}
		ifaces[i] = iface
	}
	return ifaces, nil
}

func (n *Dataplane_Networking) GetInboundInterfaceByIdx(idx int) (InboundInterface, error) {
	if idx >= len(n.Inbound) {
		return InboundInterface{}, errors.Errorf("there is no inbound for index %d. Dataplane has %d inbounds", idx, len(n.Inbound))
	}
	inbound := n.Inbound[idx]
	if inbound.Interface != "" {
		return ParseInboundInterface(inbound.Interface)
	} else {
		iface := InboundInterface{
			DataplanePort: inbound.Port,
		}
		if inbound.Address != "" {
			iface.DataplaneIP = inbound.Address
		} else {
			iface.DataplaneIP = n.Address
		}
		if inbound.ServicePort != 0 {
			iface.WorkloadPort = inbound.ServicePort
		} else {
			iface.WorkloadPort = inbound.Port
		}
		return iface, nil
	}
}

// Matches is simply an alias for MatchTags to make source code more aesthetic.
func (d *Dataplane) Matches(selector TagSelector) bool {
	if d != nil {
		return d.MatchTags(selector)
	}
	return false
}

func (d *Dataplane) MatchTags(selector TagSelector) bool {
	for _, inbound := range d.GetNetworking().GetInbound() {
		if inbound.MatchTags(selector) {
			return true
		}
	}
	if d.GetNetworking().GetGateway() != nil {
		if d.Networking.Gateway.MatchTags(selector) {
			return true
		}
	}
	return false
}

func (d *Dataplane_Networking_Gateway) MatchTags(selector TagSelector) bool {
	return selector.Matches(d.Tags)
}

// GetService returns a service represented by this inbound interface.
//
// The purpose of this method is to encapsulate implementation detail
// that service is modeled as a tag rather than a separate field.
func (d *Dataplane_Networking_Inbound) GetService() string {
	if d == nil {
		return ""
	}
	return d.Tags[ServiceTag]
}

// GetProtocol returns a protocol supported by this inbound interface.
//
// The purpose of this method is to encapsulate implementation detail
// that protocol is modeled as a tag rather than a separate field.
func (d *Dataplane_Networking_Inbound) GetProtocol() string {
	if d == nil {
		return ""
	}
	return d.Tags[ProtocolTag]
}

func (d *Dataplane_Networking_Inbound) MatchTags(selector TagSelector) bool {
	return selector.Matches(d.Tags)
}

func (d *Dataplane_Networking_Outbound) MatchTags(selector TagSelector) bool {
	service := selector[ServiceTag]
	return service == MatchAllTag || service == d.Service
}

const MatchAllTag = "*"

type TagSelector map[string]string

func (s TagSelector) Matches(tags map[string]string) bool {
	if len(s) == 0 {
		return true
	}
	for tag, value := range s {
		inboundVal, exist := tags[tag]
		if !exist {
			return false
		}
		if value != inboundVal && value != MatchAllTag {
			return false
		}
	}
	return true
}

func (s TagSelector) Rank() (r TagSelectorRank) {
	for _, value := range s {
		if value == MatchAllTag {
			r.WildcardMatches++
		} else {
			r.ExactMatches++
		}
	}
	return
}

func (s TagSelector) Equal(other TagSelector) bool {
	return len(s) == 0 && len(other) == 0 || len(s) == len(other) && reflect.DeepEqual(s, other)
}

func MatchAll() TagSelector {
	return nil
}

func MatchAnyService() TagSelector {
	return MatchService(MatchAllTag)
}

func MatchService(service string) TagSelector {
	return TagSelector{ServiceTag: service}
}

func MatchTags(tags map[string]string) TagSelector {
	return TagSelector(tags)
}

// Set of tags that only allows a single value per key.
type SingleValueTagSet map[string]string

func (t SingleValueTagSet) Keys() []string {
	keys := make([]string, 0, len(t))
	for key := range t {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// Set of tags that allows multiple values per key.
type MultiValueTagSet map[string]map[string]bool

func (t MultiValueTagSet) Keys() []string {
	keys := make([]string, 0, len(t))
	for key := range t {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (t MultiValueTagSet) Values(key string) []string {
	if t == nil {
		return nil
	}
	var result []string
	for value := range t[key] {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func (d *Dataplane) Tags() MultiValueTagSet {
	tags := MultiValueTagSet{}
	for _, inbound := range d.GetNetworking().GetInbound() {
		for tag, value := range inbound.Tags {
			_, exists := tags[tag]
			if !exists {
				tags[tag] = map[string]bool{}
			}
			tags[tag][value] = true
		}
	}
	for tag, value := range d.GetNetworking().GetGateway().GetTags() {
		_, exists := tags[tag]
		if !exists {
			tags[tag] = map[string]bool{}
		}
		tags[tag][value] = true
	}
	return tags
}

func (d *Dataplane) GetIdentifyingService() string {
	services := d.Tags().Values(ServiceTag)
	if len(services) > 0 {
		return services[0]
	}
	return ServiceUnknown
}

func (t MultiValueTagSet) String() string {
	var tags []string
	for tag := range t {
		tags = append(tags, fmt.Sprintf("%s=%s", tag, strings.Join(t.Values(tag), ",")))
	}
	sort.Strings(tags)
	return strings.Join(tags, " ")
}

// TagSelectorRank helps to decide which of 2 selectors is more specific.
type TagSelectorRank struct {
	// Number of tags that match by the exact value.
	ExactMatches int
	// Number of tags that match by a wildcard ('*').
	WildcardMatches int
}

func (r TagSelectorRank) CombinedWith(other TagSelectorRank) TagSelectorRank {
	return TagSelectorRank{
		ExactMatches:    r.ExactMatches + other.ExactMatches,
		WildcardMatches: r.WildcardMatches + other.WildcardMatches,
	}
}

func (r TagSelectorRank) CompareTo(other TagSelectorRank) int {
	thisTotal := r.ExactMatches + r.WildcardMatches
	otherTotal := other.ExactMatches + other.WildcardMatches
	if thisTotal == otherTotal {
		return r.ExactMatches - other.ExactMatches
	}
	return thisTotal - otherTotal
}
