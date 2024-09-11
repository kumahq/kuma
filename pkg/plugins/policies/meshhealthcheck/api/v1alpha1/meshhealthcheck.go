// +kubebuilder:object:generate=true
package v1alpha1

import (
	k8s "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
)

// MeshHealthCheck defines health checking policies between different data plane
// proxies.
type MeshHealthCheck struct {
	// TargetRef is a reference to the resource the policy takes an effect on.
	// The resource could be either a real store object or virtual resource
	// defined inplace.
	TargetRef *common_api.TargetRef `json:"targetRef,omitempty"`

	// To list makes a match between the consumed services and corresponding configurations
	To []To `json:"to,omitempty"`
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
	// Interval between consecutive health checks.
	// +kubebuilder:default="1m"
	Interval *k8s.Duration `json:"interval,omitempty"`
	// Maximum time to wait for a health check response.
	// +kubebuilder:default="15s"
	Timeout *k8s.Duration `json:"timeout,omitempty"`
	// Number of consecutive unhealthy checks before considering a host
	// unhealthy.
	// +kubebuilder:default=5
	UnhealthyThreshold *int32 `json:"unhealthyThreshold,omitempty"`
	// Number of consecutive healthy checks before considering a host healthy.
	// +kubebuilder:default=1
	HealthyThreshold *int32 `json:"healthyThreshold,omitempty"`
	// If specified, Envoy will start health checking after a random time in
	// ms between 0 and initialJitter. This only applies to the first health
	// check.
	InitialJitter *k8s.Duration `json:"initialJitter,omitempty"`
	// If specified, during every interval Envoy will add IntervalJitter to the
	// wait time.
	IntervalJitter *k8s.Duration `json:"intervalJitter,omitempty"`
	// If specified, during every interval Envoy will add IntervalJitter *
	// IntervalJitterPercent / 100 to the wait time. If IntervalJitter and
	// IntervalJitterPercent are both set, both of them will be used to
	// increase the wait time.
	IntervalJitterPercent *int32 `json:"intervalJitterPercent,omitempty"`
	// Allows to configure panic threshold for Envoy cluster. If not specified,
	// the default is 50%. To disable panic mode, set to 0%.
	// Either int or decimal represented as string.
	HealthyPanicThreshold *intstr.IntOrString `json:"healthyPanicThreshold,omitempty"`
	// If set to true, Envoy will not consider any hosts when the cluster is in
	// 'panic mode'. Instead, the cluster will fail all requests as if all hosts
	// are unhealthy. This can help avoid potentially overwhelming a failing
	// service.
	FailTrafficOnPanic *bool `json:"failTrafficOnPanic,omitempty"`
	// Specifies the path to the file where Envoy can log health check events.
	// If empty, no event log will be written.
	EventLogPath *string `json:"eventLogPath,omitempty"`
	// If set to true, health check failure events will always be logged. If set
	// to false, only the initial health check failure event will be logged. The
	// default value is false.
	AlwaysLogHealthCheckFailures *bool `json:"alwaysLogHealthCheckFailures,omitempty"`
	// The "no traffic interval" is a special health check interval that is used
	// when a cluster has never had traffic routed to it. This lower interval
	// allows cluster information to be kept up to date, without sending a
	// potentially large amount of active health checking traffic for no reason.
	// Once a cluster has been used for traffic routing, Envoy will shift back
	// to using the standard health check interval that is defined. Note that
	// this interval takes precedence over any other. The default value for "no
	// traffic interval" is 60 seconds.
	NoTrafficInterval *k8s.Duration    `json:"noTrafficInterval,omitempty"`
	Tcp               *TcpHealthCheck  `json:"tcp,omitempty"`
	Http              *HttpHealthCheck `json:"http,omitempty"`
	Grpc              *GrpcHealthCheck `json:"grpc,omitempty"`
	// Reuse health check connection between health checks. Default is true.
	ReuseConnection *bool `json:"reuseConnection,omitempty"`
}

// TcpHealthCheck defines configuration for specifying bytes to send and
// expected response during the health check
type TcpHealthCheck struct {
	// If true the TcpHealthCheck is disabled
	Disabled *bool `json:"disabled,omitempty"`
	// Base64 encoded content of the message which will be sent during the health check to the target
	Send *string `json:"send,omitempty"`
	// List of Base64 encoded blocks of strings expected as a response. When checking the response,
	// "fuzzy" matching is performed such that each block must be found, and
	// in the order specified, but not necessarily contiguous.
	// If not provided or empty, checks will be performed as "connect only" and be marked as successful when TCP connection is successfully established.
	Receive *[]string `json:"receive,omitempty"`
}

// HttpHealthCheck defines HTTP configuration which will instruct the service
// the health check will be made for is an HTTP service.
type HttpHealthCheck struct {
	// If true the HttpHealthCheck is disabled
	Disabled *bool `json:"disabled,omitempty"`
	// The HTTP path which will be requested during the health check
	// (ie. /health)
	// +kubebuilder:default="/"
	Path *string `json:"path,omitempty"`
	// The list of HTTP headers which should be added to each health check
	// request
	RequestHeadersToAdd *HeaderModifier `json:"requestHeadersToAdd,omitempty"`
	// List of HTTP response statuses which are considered healthy
	ExpectedStatuses *[]int32 `json:"expectedStatuses,omitempty"`
}

// GrpcHealthCheck defines gRPC configuration which will instruct the service
// the health check will be made for is a gRPC service.
type GrpcHealthCheck struct {
	// If true the GrpcHealthCheck is disabled
	Disabled *bool `json:"disabled,omitempty"`
	// Service name parameter which will be sent to gRPC service
	ServiceName *string `json:"serviceName,omitempty"`
	// The value of the :authority header in the gRPC health check request,
	// by default name of the cluster this health check is associated with
	Authority *string `json:"authority,omitempty"`
}

type HeaderKeyValue struct {
	Name  common_api.HeaderName  `json:"name"`
	Value common_api.HeaderValue `json:"value"`
}

// Configuration to set or add multiple values for a header must use RFC 7230
// header value formatting, separating each value with a comma.
type HeaderModifier struct {
	// +listType=map
	// +listMapKey=name
	// +kubebuilder:validation:MaxItems=16
	Set []HeaderKeyValue `json:"set,omitempty"`
	// +listType=map
	// +listMapKey=name
	// +kubebuilder:validation:MaxItems=16
	Add []HeaderKeyValue `json:"add,omitempty"`
}
