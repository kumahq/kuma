// +kubebuilder:object:generate=true
package v1alpha1

import (
	k8s "k8s.io/apimachinery/pkg/apis/meta/v1"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
)

// MeshFaultInjection
type MeshFaultInjection struct {
	// TargetRef is a reference to the resource the policy takes an effect on.
	// The resource could be either a real store object or virtual resource
	// defined inplace.
	TargetRef common_api.TargetRef `json:"targetRef"`

	From []From `json:"from,omitempty"`
}

type From struct {
	// TargetRef is a reference to the resource that represents a group of
	// destinations.
	TargetRef common_api.TargetRef `json:"targetRef"`
	// Default is a configuration specific to the group of destinations referenced in
	// 'targetRef'
	Default Conf `json:"default,omitempty"`
}

type Conf struct {
	Http []FaultInjectionConf `json:"http,omitempty"`
}

// FaultInjection defines the configuration of faults between dataplanes.
type FaultInjectionConf struct {
	Abort             *AbortConf             `json:"abort,omitempty"`
	Delay             *DelayConf             `json:"delay,omitempty"`
	ResponseBandwidth *ResponseBandwidthConf `json:"responseBandwidth,omitempty"`
}

// Abort defines a configuration of not delivering requests to destination
// service and replacing the responses from destination dataplane by
// predefined status code
type AbortConf struct {
	// HTTP status code which will be returned to source side
	HttpStatus *int32 `json:"httpStatus,omitempty"`
	// Percentage of requests on which abort will be injected, has to be in
	// [0.0 - 100.0] range
	Percentage *int32 `json:"percentage,omitempty"`
}

// Delay defines configuration of delaying a response from a destination
type DelayConf struct {
	// The duration during which the response will be delayed
	Value *k8s.Duration `json:"value,omitempty"`
	// Percentage of requests on which delay will be injected, has to be in
	// [0.0 - 100.0] range
	Percentage *int32 `json:"percentage,omitempty"`
}

// ResponseBandwidth defines a configuration to limit the speed of
// responding to the requests
type ResponseBandwidthConf struct {
	// Limit is represented by value measure in gbps, mbps, kbps or bps, e.g.
	// 10kbps
	Limit *string `json:"limit,omitempty"`
	// Percentage of requests on which response bandwidth limit will be
	// injected, has to be in [0.0 - 100.0] range
	Percentage *int32 `json:"percentage,omitempty"`
}
