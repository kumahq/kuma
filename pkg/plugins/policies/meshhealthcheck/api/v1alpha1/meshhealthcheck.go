// +kubebuilder:object:generate=true
package v1alpha1

import (
    common_api "github.com/kumahq/kuma/api/common/v1alpha1"
    "google.golang.org/protobuf/types/known/durationpb"
    "google.golang.org/protobuf/types/known/wrapperspb"
)

// MeshHealthCheck
// +kuma:policy:skip_registration=true
type MeshHealthCheck struct {
    // TargetRef is a reference to the resource the policy takes an effect on.
    // The resource could be either a real store object or virtual resource
    // defined inplace.
    TargetRef common_api.TargetRef `json:"targetRef,omitempty"`

    // To list makes a match between the consumed services and corresponding configurations
    To []To `json:"to,omitempty"`
}

type To struct {
    // TargetRef is a reference to the resource that represents a group of
    // destinations.
    TargetRef common_api.TargetRef `json:"targetRef,omitempty"`
    // Default is a configuration specific to the group of destinations referenced in
    // 'targetRef'
    Default Conf `json:"default,omitempty"`
}

type Conf struct {
    // Interval between consecutive health checks.
    Interval *durationpb.Duration `json:"interval,omitempty"`
    // Maximum time to wait for a health check response.
    Timeout *durationpb.Duration `json:"timeout,omitempty"`
    // Number of consecutive unhealthy checks before considering a host
    // unhealthy.
    UnhealthyThreshold uint32 `json:"unhealthyThreshold,omitempty"`
    // Number of consecutive healthy checks before considering a host healthy.
    HealthyThreshold uint32 `json:"healthyThreshold,omitempty"`
    // If specified, Envoy will start health checking after for a random time in
    // ms between 0 and initialJitter. This only applies to the first health
    // check.
    InitialJitter *durationpb.Duration `json:"initialJitter,omitempty"`
    // If specified, during every interval Envoy will add IntervalJitter to the
    // wait time.
    IntervalJitter *durationpb.Duration `json:"intervalJitter,omitempty"`
    // If specified, during every interval Envoy will add IntervalJitter *
    // IntervalJitterPercent / 100 to the wait time. If IntervalJitter and
    // IntervalJitterPercent are both set, both of them will be used to
    // increase the wait time.
    IntervalJitterPercent uint32 `json:"intervalJitterPercent,omitempty"`
    // Allows to configure panic threshold for Envoy cluster. If not specified,
    // the default is 50%. To disable panic mode, set to 0%.
    HealthyPanicThreshold *wrapperspb.FloatValue `json:"healthyPanicThreshold,omitempty"`
    // If set to true, Envoy will not consider any hosts when the cluster is in
    // 'panic mode'. Instead, the cluster will fail all requests as if all hosts
    // are unhealthy. This can help avoid potentially overwhelming a failing
    // service.
    FailTrafficOnPanic *wrapperspb.BoolValue `json:"failTrafficOnPanic,omitempty"`
    // Specifies the path to the file where Envoy can log health check events.
    // If empty, no event log will be written.
    EventLogPath string `json:"eventLogPath,omitempty"`
    // If set to true, health check failure events will always be logged. If set
    // to false, only the initial health check failure event will be logged. The
    // default value is false.
    AlwaysLogHealthCheckFailures *wrapperspb.BoolValue `json:"alwaysLogHealthCheckFailures,omitempty"`
    // The "no traffic interval" is a special health check interval that is used
    // when a cluster has never had traffic routed to it. This lower interval
    // allows cluster information to be kept up to date, without sending a
    // potentially large amount of active health checking traffic for no reason.
    // Once a cluster has been used for traffic routing, Envoy will shift back
    // to using the standard health check interval that is defined. Note that
    // this interval takes precedence over any other. The default value for "no
    // traffic interval" is 60 seconds.
    NoTrafficInterval *durationpb.Duration `json:"noTrafficInterval,omitempty"`
    Tcp               *TcpHealthCheck      `json:"tcp,omitempty"`
    Http              *HttpHealthCheck     `json:"http,omitempty"`
    Grpc              *GrpcHealthCheck     `json:"grpc,omitempty"`
    // Reuse health check connection between health checks. Default is true.
    ReuseConnection *wrapperspb.BoolValue `json:"reuseConnection,omitempty"`
}

// TcpHealthCheck defines configuration for specifying bytes to send and
// expected response during the health check
type TcpHealthCheck struct {
    // If true the HealthCheck is disabled
    //  +optional
    Disabled bool `json:"disabled,omitempty"`
    // Bytes which will be sent during the health check to the target
    Send *wrapperspb.BytesValue `json:"send,omitempty"`
    // Bytes blocks expected as a response. When checking the response,
    // “fuzzy” matching is performed such that each block must be found, and
    // in the order specified, but not necessarily contiguous.
    Receive []*wrapperspb.BytesValue `json:"receive,omitempty"`
}

// HttpHealthCheck defines HTTP configuration which will instruct the service
// the health check will be made for is an HTTP service.
type HttpHealthCheck struct {
    // If true the HealthCheck is disabled
    //  +optional
    Disabled bool `json:"disabled,omitempty"`
    // The HTTP path which will be requested during the health check
    // (ie. /health)
    //  +required
    Path string `json:"path,omitempty"`
    // The list of HTTP headers which should be added to each health check
    // request
    //  +optional
    RequestHeadersToAdd []*HeaderValueOption `json:"requestHeadersToAdd,omitempty"`
    // List of HTTP response statuses which are considered healthy
    //  +optional
    ExpectedStatuses []*wrapperspb.UInt32Value `json:"expectedStatuses,omitempty"`
}

// GrpcHealthCheck defines gRPC configuration which will instruct the service
// the health check will be made for is a gRPC service.
type GrpcHealthCheck struct {
    // If true the HealthCheck is disabled
    //  +optional
    Disabled bool `json:"disabled,omitempty"`
    // Service name parameter which will be sent to gRPC service
    //  +optional
    ServiceName string `json:"serviceName,omitempty"`
    // The value of the :authority header in the gRPC health check request,
    // by default name of the cluster this health check is associated with
    //  +optional
    Authority string `json:"authority,omitempty"`
}

type HeaderValueOption struct {
    // Key/Value representation of the HTTP header
    //  +required
    Header *HeaderValue `json:"header,omitempty"`
    // If true (default) the header values should be appended to already present ones
    //  +optional
    Append *wrapperspb.BoolValue `json:"append,omitempty"`
}

type HeaderValue struct {
    // Header name
    //  +required
    Key string `json:"key,omitempty"`
    // Header value
    //  +optional
    Value string `json:"value,omitempty"`
}
