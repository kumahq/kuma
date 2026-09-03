package types

import (
	"github.com/kumahq/kuma/v3/pkg/xds/envoy/tags"
)

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
