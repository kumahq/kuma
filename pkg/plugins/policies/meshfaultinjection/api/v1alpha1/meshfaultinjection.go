// +kubebuilder:object:generate=true
package v1alpha1

import (
	k8s "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
)

// MeshFaultInjection
type MeshFaultInjection struct {
	// TargetRef is a reference to the resource the policy takes an effect on.
	// The resource could be either a real store object or virtual resource
	// defined inplace.
	TargetRef *common_api.TargetRef `json:"targetRef,omitempty"`

	// From list makes a match between clients and corresponding configurations
	From []From `json:"from,omitempty"`

	// To list makes a match between clients and corresponding configurations
	To []To `json:"to,omitempty"`
}

type From struct {
	// TargetRef is a reference to the resource that represents a group of
	// destinations.
	TargetRef common_api.TargetRef `json:"targetRef"`
	// Default is a configuration specific to the group of destinations referenced in
	// 'targetRef'
	Default Conf `json:"default,omitempty"`
}

type To struct {
	// TargetRef is a reference to the resource that represents a group of
	// destinations.
	TargetRef common_api.TargetRef `json:"targetRef"`
	// Default is a configuration specific to the group of destinations referenced in
	// 'targetRef'
	Default Conf `json:"default,omitempty"`
}

type Conf struct {
	// Http allows to define list of Http faults between dataplanes.
	Http *[]FaultInjectionConf `json:"http,omitempty"`
}

// FaultInjection defines the configuration of faults between dataplanes.
type FaultInjectionConf struct {
	// Abort defines a configuration of not delivering requests to destination
	// service and replacing the responses from destination dataplane by
	// predefined status code
	Abort *AbortConf `json:"abort,omitempty"`
	// Delay defines configuration of delaying a response from a destination
	Delay *DelayConf `json:"delay,omitempty"`
	// ResponseBandwidth defines a configuration to limit the speed of
	// responding to the requests
	ResponseBandwidth *ResponseBandwidthConf `json:"responseBandwidth,omitempty"`
}

type AbortConf struct {
	// HTTP status code which will be returned to source side
	HttpStatus int32 `json:"httpStatus"`
	// Percentage of requests on which abort will be injected, has to be
	// either int or decimal represented as string.
	Percentage intstr.IntOrString `json:"percentage"`
}

type DelayConf struct {
	// The duration during which the response will be delayed
	Value k8s.Duration `json:"value"`
	// Percentage of requests on which delay will be injected, has to be
	// either int or decimal represented as string.
	Percentage intstr.IntOrString `json:"percentage"`
}

type ResponseBandwidthConf struct {
	// Limit is represented by value measure in Gbps, Mbps, kbps, e.g.
	// 10kbps
	Limit string `json:"limit"`
	// Percentage of requests on which response bandwidth limit will be
	// either int or decimal represented as string.
	Percentage intstr.IntOrString `json:"percentage"`
}
