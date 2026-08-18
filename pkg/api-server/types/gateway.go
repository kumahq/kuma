package types

import (
	"github.com/kumahq/kuma/v3/pkg/xds/envoy/tags"
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

// GatewayDataplaneInspectResult is the gateway-kinded variant of the response
// from the deprecated GET /meshes/{mesh}/dataplanes/{name}/policies endpoint.
// Nothing in this repo produces it anymore, but DataplaneInspectResponse's
// Marshal/UnmarshalJSON keep it as a documented, wire-compatible kind for the
// vendored GUI bundle that still calls the endpoint.
type GatewayDataplaneInspectResult struct {
	Gateway   ResourceKeyEntry              `json:"gateway"`
	Listeners []GatewayListenerInspectEntry `json:"listeners"`
	Policies  PolicyMap                     `json:"policies,omitempty"`
}

func (*GatewayDataplaneInspectResult) dataplaneInspectEntry() {
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
