package types

import (
	"github.com/kumahq/kuma/pkg/xds/envoy/tags"
)

type Destination struct {
	Tags     tags.Tags `json:"tags"`
	Policies PolicyMap `json:"policies"`
}

type RouteInspectEntry struct {
	Route        string        `json:"route"`
	Destinations []Destination `json:"destinations"`
}

type HostInspectEntry struct {
	HostName string              `json:"hostName"`
	Routes   []RouteInspectEntry `json:"routes"`
}

type GatewayListenerInspectEntry struct {
	Port     uint32             `json:"port"`
	Protocol string             `json:"protocol"`
	Hosts    []HostInspectEntry `json:"hosts"`
}

type GatewayDataplaneInspectResult struct {
	Gateway   ResourceKeyEntry              `json:"gateway"`
	Listeners []GatewayListenerInspectEntry `json:"listeners"`
	Policies  PolicyMap                     `json:"policies,omitempty"`
}

func (*GatewayDataplaneInspectResult) dataplaneInspectEntry() {
}

func NewGatewayDataplaneInspectResult() GatewayDataplaneInspectResult {
	return GatewayDataplaneInspectResult{
		Listeners: []GatewayListenerInspectEntry{},
	}
}

type PolicyInspectGatewayRouteEntry struct {
	Route        string      `json:"route"`
	Destinations []tags.Tags `json:"destinations"`
}

type PolicyInspectGatewayHostEntry struct {
	HostName string                           `json:"hostName"`
	Routes   []PolicyInspectGatewayRouteEntry `json:"routes"`
}

type PolicyInspectGatewayListenerEntry struct {
	Port     uint32                          `json:"port"`
	Protocol string                          `json:"protocol"`
	Hosts    []PolicyInspectGatewayHostEntry `json:"hosts"`
}

type PolicyInspectGatewayEntry struct {
	DataplaneKey ResourceKeyEntry                    `json:"dataplane"`
	Gateway      ResourceKeyEntry                    `json:"gateway,omitempty"`
	Listeners    []PolicyInspectGatewayListenerEntry `json:"listeners,omitempty"`
}

func (*PolicyInspectGatewayEntry) policyInspectEntry() {
}

func NewPolicyInspectGatewayEntry(key, gateway ResourceKeyEntry) PolicyInspectGatewayEntry {
	return PolicyInspectGatewayEntry{
		DataplaneKey: key,
		Gateway:      gateway,
	}
}
