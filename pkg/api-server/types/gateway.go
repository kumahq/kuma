package types

import (
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/pkg/xds/envoy"
)

type PolicyMap map[core_model.ResourceType]*rest.Resource

type Destination struct {
	Tags     envoy.Tags `json:"tags"`
	Policies PolicyMap  `json:"policies"`
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
	Kind      string                        `json:"kind"`
	Gateway   ResourceKeyEntry              `json:"gateway"`
	Listeners []GatewayListenerInspectEntry `json:"listeners"`
	Policies  PolicyMap                     `json:"policies,omitempty"`
}

func NewGatewayDataplaneInspectResult() GatewayDataplaneInspectResult {
	return GatewayDataplaneInspectResult{
		Kind:      "GatewayDataplane",
		Listeners: []GatewayListenerInspectEntry{},
	}
}
