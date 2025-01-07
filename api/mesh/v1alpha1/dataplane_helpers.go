package v1alpha1

import (
	"encoding"
	"fmt"
	"net"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

const (
	KubeNamespaceTag = "k8s.kuma.io/namespace"
	KubeServiceTag   = "k8s.kuma.io/service-name"
	KubePortTag      = "k8s.kuma.io/service-port"
)

const (
	KubernetesEnvironment = "kubernetes"
	UniversalEnvironment  = "universal"
)

const (
	// Mandatory tag that has a reserved meaning in Kuma.
	ServiceTag     = "kuma.io/service"
	ServiceUnknown = "unknown"

	// Locality related tags
	ZoneTag = "kuma.io/zone"

	// EnvTag defines whether a zone is universal or Kubernetes
	EnvTag = "kuma.io/env"

	MeshTag = "kuma.io/mesh"

	// Optional tag that has a reserved meaning in Kuma.
	// If absent, Kuma will treat application's protocol as opaque TCP.
	ProtocolTag = "kuma.io/protocol"
	// InstanceTag is set only for Dataplanes that implements headless services
	InstanceTag = "kuma.io/instance"

	// External service tag
	ExternalServiceTag = "kuma.io/external-service-name"

	// Listener tag is used to select Gateway listeners
	ListenerTag = "gateways.kuma.io/listener-name"

	// Port tag is used to select Gateway listeners
	PortTag = "gateways.kuma.io/listener-port"

	// Used for Service-less dataplanes
	TCPPortReserved = 49151 // IANA Reserved

	// DisplayName is a standard label that can be used to easier recognize policy name.
	// On Kubernetes, Kuma resource name contains namespace. Display name is original name without namespace.
	// The name contains hash when the resource is synced from global to zone. In this case, display name is original name from originated CP.
	DisplayName = "kuma.io/display-name"

	// ResourceOriginLabel is a standard label that has information about the origin of the resource.
	// It can be either "global" or "zone".
	ResourceOriginLabel = "kuma.io/origin"

	// EffectLabel is a standard label that configures what effect the policy has on the DPPs. The only supported value
	// at this moment is "shadow". When effect is "shadow" the policy doesn't change the DPPs configs, but could be
	// observed using the Inspect API.
	EffectLabel = "kuma.io/effect"

	// PolicyRoleLabel is a standard label that reflects the role of the policy. The value is automatically set by the
	// Kuma CP based on the policy spec. Supported values are "producer", "consumer", "system" and "workload-owner".
	PolicyRoleLabel = "kuma.io/policy-role"

	// ProxyTypeLabel is a standard label that reflects the type of proxy. Supported values are "sidecar", "gateway",
	// "zoneingress", "zoneegress"
	ProxyTypeLabel = "kuma.io/proxy-type"

	// ManagedByLabel is used when a MeshService is auto-generated
	ManagedByLabel = "kuma.io/managed-by"

	// DeletionGracePeriodStartedLabel is used when generating MeshServices on
	// universal, it's here to avoid import cycles
	DeletionGracePeriodStartedLabel string = "kuma.io/deletion-grace-period-started-at"
)

type ResourceOrigin string

const (
	UnknownResourceOrigin ResourceOrigin = ""
	GlobalResourceOrigin  ResourceOrigin = "global"
	ZoneResourceOrigin    ResourceOrigin = "zone"
)

var originOrder = map[ResourceOrigin]int{
	GlobalResourceOrigin:  1,
	UnknownResourceOrigin: 2,
	ZoneResourceOrigin:    3,
}

// Compare recreates the following comparison table:
//
// origin_1 | origin_2 | has_more_priority
// ---------|----------|-------------
// Global   | Zone     | origin_2
// Global   | Unknown  | origin_2
// Zone     | Global   | origin_1
// Zone     | Unknown  | origin_1
// Unknown  | Global   | origin_1
// Unknown  | Zone     | origin_2
//
// If we assign numbers to origins like Global=1, Zone=3, Unknown=2, then we can compare them as numbers
// and get the same result as in the table above.
func (o ResourceOrigin) Compare(other ResourceOrigin) int {
	return originOrder[o] - originOrder[other]
}

func (o ResourceOrigin) IsValid() error {
	switch o {
	case GlobalResourceOrigin, ZoneResourceOrigin:
		return nil
	default:
		return errors.Errorf("unknown resource origin %q", o)
	}
}

type PolicyRole string

const (
	SystemPolicyRole        PolicyRole = "system"
	ProducerPolicyRole      PolicyRole = "producer"
	ConsumerPolicyRole      PolicyRole = "consumer"
	WorkloadOwnerPolicyRole PolicyRole = "workload-owner"
)

var roleOrder = map[PolicyRole]int{
	SystemPolicyRole:        1,
	ProducerPolicyRole:      2,
	ConsumerPolicyRole:      3,
	WorkloadOwnerPolicyRole: 4,
}

func (r PolicyRole) Compare(o PolicyRole) int {
	return roleOrder[r] - roleOrder[o]
}

type ProxyType string

const (
	DataplaneProxyType ProxyType = "dataplane"
	IngressProxyType   ProxyType = "ingress"
	EgressProxyType    ProxyType = "egress"
)

func (t ProxyType) IsValid() error {
	switch t {
	case DataplaneProxyType, IngressProxyType, EgressProxyType:
		return nil
	}
	return errors.Errorf("%s is not a valid proxy type", t)
}

type InboundInterface struct {
	DataplaneAdvertisedIP string
	DataplaneIP           string
	DataplanePort         uint32
	WorkloadIP            string
	WorkloadPort          uint32
}

// We need to implement TextMarshaler because InboundInterface is used
// as a key for maps that are JSON encoded for logging.
var _ encoding.TextMarshaler = InboundInterface{}

func (i InboundInterface) MarshalText() ([]byte, error) {
	return []byte(i.String()), nil
}

func (i InboundInterface) String() string {
	return fmt.Sprintf("%s:%d:%d", i.DataplaneIP, i.DataplanePort, i.WorkloadPort)
}

func (i *InboundInterface) IsServiceLess() bool {
	return i.DataplanePort == TCPPortReserved
}

type OutboundInterface struct {
	DataplaneIP   string
	DataplanePort uint32
}

// We need to implement TextMarshaler because OutboundInterface is used
// as a key for maps that are JSON encoded for logging.
var _ encoding.TextMarshaler = OutboundInterface{}

func (i OutboundInterface) MarshalText() ([]byte, error) {
	return []byte(i.String()), nil
}

func (i OutboundInterface) String() string {
	return net.JoinHostPort(i.DataplaneIP,
		strconv.FormatUint(uint64(i.DataplanePort), 10))
}

func NonBackendRefFilter(outbound *Dataplane_Networking_Outbound) bool {
	return outbound.BackendRef == nil
}

func (n *Dataplane_Networking) GetOutbounds(filters ...func(*Dataplane_Networking_Outbound) bool) []*Dataplane_Networking_Outbound {
	var result []*Dataplane_Networking_Outbound
	for _, outbound := range n.GetOutbound() {
		add := true
		for _, filter := range filters {
			if !filter(outbound) {
				add = false
			}
		}
		if add {
			result = append(result, outbound)
		}
	}
	return result
}

func (n *Dataplane_Networking) GetOutboundInterfaces() []OutboundInterface {
	if n == nil {
		return nil
	}
	ofaces := make([]OutboundInterface, len(n.Outbound))
	for i, outbound := range n.Outbound {
		ofaces[i] = n.ToOutboundInterface(outbound)
	}
	return ofaces
}

func (n *Dataplane_Networking) ToOutboundInterface(outbound *Dataplane_Networking_Outbound) OutboundInterface {
	oface := OutboundInterface{
		DataplanePort: outbound.Port,
	}
	if outbound.Address != "" {
		oface.DataplaneIP = outbound.Address
	} else {
		oface.DataplaneIP = "127.0.0.1"
	}
	return oface
}

func (n *Dataplane_Networking) GetInboundInterface(service string) (*InboundInterface, error) {
	for _, inbound := range n.Inbound {
		if inbound.Tags[ServiceTag] != service {
			continue
		}
		iface := n.ToInboundInterface(inbound)
		return &iface, nil
	}
	return nil, errors.Errorf("Dataplane has no Inbound Interface for service %q", service)
}

func (n *Dataplane_Networking) GetInboundInterfaces() []InboundInterface {
	if n == nil {
		return nil
	}
	ifaces := make([]InboundInterface, len(n.Inbound))
	for i, inbound := range n.Inbound {
		ifaces[i] = n.ToInboundInterface(inbound)
	}
	return ifaces
}

func (n *Dataplane_Networking) GetInboundForPort(port uint32) *Dataplane_Networking_Inbound {
	for _, inbound := range n.Inbound {
		if port == inbound.Port {
			return inbound
		}
	}
	return nil
}

func (n *Dataplane_Networking) ToInboundInterface(inbound *Dataplane_Networking_Inbound) InboundInterface {
	iface := InboundInterface{
		DataplanePort: inbound.Port,
	}
	if inbound.Address != "" {
		iface.DataplaneIP = inbound.Address
	} else {
		iface.DataplaneIP = n.Address
	}
	if n.AdvertisedAddress != "" {
		iface.DataplaneAdvertisedIP = n.AdvertisedAddress
	} else {
		iface.DataplaneAdvertisedIP = iface.DataplaneIP
	}
	if inbound.ServiceAddress != "" {
		iface.WorkloadIP = inbound.ServiceAddress
	} else {
		iface.WorkloadIP = iface.DataplaneIP
	}
	if inbound.ServicePort != 0 {
		iface.WorkloadPort = inbound.ServicePort
	} else {
		iface.WorkloadPort = inbound.Port
	}
	return iface
}

func (n *Dataplane_Networking) GetHealthyInbounds() []*Dataplane_Networking_Inbound {
	var inbounds []*Dataplane_Networking_Inbound
	for _, inbound := range n.GetInbound() {
		if inbound.GetState() != Dataplane_Networking_Inbound_Ready {
			continue
		}
		if inbound.Health != nil && !inbound.Health.Ready {
			continue
		}
		inbounds = append(inbounds, inbound)
	}
	return inbounds
}

// Matches is simply an alias for MatchTags to make source code more aesthetic.
func (d *Dataplane) Matches(selector TagSelector) bool {
	if d == nil {
		return false
	}
	for _, inbound := range d.GetNetworking().GetInbound() {
		if selector.Matches(inbound.Tags) {
			return true
		}
	}
	return selector.Matches(d.GetNetworking().GetGateway().GetTags())
}

func (d *Dataplane) MatchTagsFuzzy(selector TagSelector) bool {
	if d == nil {
		return false
	}
	for _, inbound := range d.GetNetworking().GetInbound() {
		if selector.MatchesFuzzy(inbound.Tags) {
			return true
		}
	}
	return selector.MatchesFuzzy(d.GetNetworking().GetGateway().GetTags())
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

// GetService returns a service name represented by this outbound interface.
//
// The purpose of this method is to encapsulate implementation detail
// that service is modeled as a tag rather than a separate field.
func (d *Dataplane_Networking_Outbound) GetService() string {
	if d == nil || d.GetTags() == nil {
		return ""
	}
	return d.GetTags()[ServiceTag]
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

func (s TagSelector) MatchesFuzzy(tags map[string]string) bool {
	if len(s) == 0 {
		return true
	}
	for tag, value := range s {
		inboundVal, exist := tags[tag]
		if !exist {
			return false
		}
		if !strings.Contains(inboundVal, value) && value != MatchAllTag {
			return false
		}
	}
	return true
}

func (s TagSelector) Rank() TagSelectorRank {
	var r TagSelectorRank

	for _, value := range s {
		if value == MatchAllTag {
			r.WildcardMatches++
		} else {
			r.ExactMatches++
		}
	}
	return r
}

func (s TagSelector) Equal(other TagSelector) bool {
	return len(s) == 0 && len(other) == 0 || len(s) == len(other) && reflect.DeepEqual(s, other)
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

func Merge[TagSet ~map[string]string](other ...TagSet) TagSet {
	// Small optimization, to not iterate over the whole map if only one
	// argument is provided
	if len(other) == 1 {
		return other[0]
	}

	merged := TagSet{}

	for _, t := range other {
		for k, v := range t {
			merged[k] = v
		}
	}

	return merged
}

// MergeAs is just syntactic sugar which converts merged result to assumed type
func MergeAs[R ~map[string]string, T ~map[string]string](other ...T) R {
	return R(Merge(other...))
}

func (t SingleValueTagSet) Exclude(key string) SingleValueTagSet {
	rv := SingleValueTagSet{}
	for k, v := range t {
		if k == key {
			continue
		}
		rv[k] = v
	}
	return rv
}

func (t SingleValueTagSet) String() string {
	var tags []string
	for tag, value := range t {
		tags = append(tags, fmt.Sprintf("%s=%s", tag, value))
	}
	sort.Strings(tags)
	return strings.Join(tags, " ")
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

func (t MultiValueTagSet) UniqueValues(key string) []string {
	if t == nil {
		return nil
	}
	alreadyFound := map[string]bool{}
	var result []string
	for value := range t[key] {
		if !alreadyFound[value] {
			result = append(result, value)
			alreadyFound[value] = true
		}
	}
	sort.Strings(result)
	return result
}

func MultiValueTagSetFrom(data map[string][]string) MultiValueTagSet {
	set := MultiValueTagSet{}
	for tagName, values := range data {
		for _, value := range values {
			m, ok := set[tagName]
			if !ok {
				m = map[string]bool{}
			}
			m[value] = true
			set[tagName] = m
		}
	}
	return set
}

func (d *Dataplane) TagSet() MultiValueTagSet {
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

func (d *Dataplane) SingleValueTagSets() []SingleValueTagSet {
	var sets []SingleValueTagSet
	for _, inbound := range d.GetNetworking().GetInbound() {
		sets = append(sets, SingleValueTagSet(inbound.Tags))
	}
	if gateway := d.GetNetworking().GetGateway(); gateway != nil {
		sets = append(sets, gateway.GetTags())
	}
	return sets
}

func (d *Dataplane) GetIdentifyingService() string {
	services := d.TagSet().Values(ServiceTag)
	if len(services) > 0 {
		return services[0]
	}
	return ServiceUnknown
}

func (d *Dataplane) IsDelegatedGateway() bool {
	return d.GetNetworking().GetGateway() != nil &&
		d.GetNetworking().GetGateway().GetType() == Dataplane_Networking_Gateway_DELEGATED
}

func (d *Dataplane) IsBuiltinGateway() bool {
	return d.GetNetworking().GetGateway() != nil &&
		d.GetNetworking().GetGateway().GetType() == Dataplane_Networking_Gateway_BUILTIN
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
