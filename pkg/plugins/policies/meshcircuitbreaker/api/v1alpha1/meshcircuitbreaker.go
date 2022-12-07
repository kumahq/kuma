// +kubebuilder:object:generate=true
package v1alpha1

import (
	k8s "k8s.io/apimachinery/pkg/apis/meta/v1"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
)

// MeshCircuitBreaker
// +kuma:policy:skip_registration=true
type MeshCircuitBreaker struct {
	// TargetRef is a reference to the resource the policy takes an effect on.
	// The resource could be either a real store object or virtual resource
	// defined in place.
	TargetRef common_api.TargetRef `json:"targetRef,omitempty"`

	// To list makes a match between the consumed services and corresponding
	// configurations
	To []To `json:"to,omitempty"`

	// From list makes a match between clients and corresponding configurations
	From []From `json:"from,omitempty"`
}

type To struct {
	// TargetRef is a reference to the resource that represents a group of
	// destinations.
	TargetRef common_api.TargetRef `json:"targetRef,omitempty"`
	// Default is a configuration specific to the group of destinations
	// referenced in 'targetRef'
	Default Conf `json:"default,omitempty"`
}

type From struct {
	// TargetRef is a reference to the resource that represents a group of
	// destinations.
	TargetRef common_api.TargetRef `json:"targetRef,omitempty"`
	// Default is a configuration specific to the group of destinations
	// referenced in 'targetRef'
	Default Conf `json:"default,omitempty"`
}

type Conf struct {
	ConnectionLimits *ConnectionLimits `json:"connectionLimits,omitempty"`
	OutlierDetection *OutlierDetection `json:"outlierDetection,omitempty"`
}

type ConnectionLimits struct {
	// The maximum number of connections allowed to be made to the upstream
	// cluster.
	MaxConnections *uint32 `json:"maxConnections,omitempty"`
	// The maximum number of connection pools per cluster that are concurrently
	// supported at once. Set this for clusters which create a large number of
	// connection pools.
	MaxConnectionPools *uint32 `json:"maxConnectionPools,omitempty"`
	// The maximum number of pending requests that are allowed to the upstream
	// cluster. This limit is applied as a connection limit for non-HTTP
	// traffic.
	MaxPendingRequests *uint32 `json:"maxPendingRequests,omitempty"`
	// The maximum number of parallel retries that will be allowed to
	// the upstream cluster.
	MaxRetries *uint32 `json:"maxRetries,omitempty"`
	// The maximum number of parallel requests that are allowed to be made
	// to the upstream cluster. This limit does not apply to non-HTTP traffic.
	MaxRequests *uint32 `json:"maxRequests,omitempty"`
}

type OutlierDetection struct {
	// When set to true, outlierDetection configuration won't take any effect
	Disabled *bool `json:"disabled,omitempty"`
	// The time interval between ejection analysis sweeps. This can result in
	// both new ejections and hosts being returned to service.
	Interval *k8s.Duration `json:"interval,omitempty"`
	// The base time that a host is ejected for. The real time is equal to
	// the base time multiplied by the number of times the host has been
	// ejected.
	BaseEjectionTime *k8s.Duration `json:"baseEjectionTime,omitempty"`
	// The maximum % of an upstream cluster that can be ejected due to outlier
	// detection. Defaults to 10% but will eject at least one host regardless of
	// the value.
	MaxEjectionPercent *uint32 `json:"maxEjectionPercent,omitempty"`
	// Determines whether to distinguish local origin failures from external
	// errors. If set to true the following configuration parameters are taken
	// into account: detectors.localOriginFailures.consecutive
	SplitExternalAndLocalErrors *bool      `json:"splitExternalAndLocalErrors,omitempty"`
	Detectors                   *Detectors `json:"detectors,omitempty"`
}

type Detectors struct {
	TotalFailures       *DetectorTotalFailures             `json:"totalFailures,omitempty"`
	GatewayFailures     *DetectorGatewayFailures           `json:"gatewayFailures,omitempty"`
	LocalOriginFailures *DetectorLocalOriginFailures       `json:"localOriginFailures,omitempty"`
	SuccessRate         *DetectorSuccessRateFailures       `json:"successRate,omitempty"`
	FailurePercentage   *DetectorFailurePercentageFailures `json:"failurePercentage,omitempty"`
}

type DetectorTotalFailures struct {
	// The number of consecutive server-side error responses (for HTTP traffic,
	// 5xx responses; for TCP traffic, connection failures; for Redis, failure
	// to respond PONG; etc.) before a consecutive total failure ejection
	// occurs.
	Consecutive *uint32 `json:"consecutive,omitempty"`
}

type DetectorGatewayFailures struct {
	// The number of consecutive gateway failures (502, 503, 504 status codes)
	// before a consecutive gateway failure ejection occurs.
	Consecutive *uint32 `json:"consecutive,omitempty"`
}

type DetectorLocalOriginFailures struct {
	// The number of consecutive locally originated failures before ejection
	// occurs. Parameter takes effect only when splitExternalAndLocalErrors
	// is set to true.
	Consecutive *uint32 `json:"consecutive,omitempty"`
}

type DetectorSuccessRateFailures struct {
	// The number of hosts in a cluster that must have enough request volume to
	// detect success rate outliers. If the number of hosts is less than this
	// setting, outlier detection via success rate statistics is not performed
	// for any host in the cluster.
	MinimumHosts *uint32 `json:"minimumHosts,omitempty"`
	// The minimum number of total requests that must be collected in one
	// interval (as defined by the interval duration configured in
	// outlierDetection section) to include this host in success rate based
	// outlier detection. If the volume is lower than this setting, outlier
	// detection via success rate statistics is not performed for that host.
	RequestVolume *uint32 `json:"requestVolume,omitempty"`
	// This factor is used to determine the ejection threshold for success rate
	// outlier ejection. The ejection threshold is the difference between
	// the mean success rate, and the product of this factor and the standard
	// deviation of the mean success rate: mean - (standard_deviation *
	// success_rate_standard_deviation_factor). This factor is divided by a
	// thousand to get a double. That is, if the desired factor is 1.9,
	// the runtime value should be 1900.
	StandardDeviationFactor *uint32 `json:"standardDeviationFactor,omitempty"`
}

type DetectorFailurePercentageFailures struct {
	// The minimum number of hosts in a cluster in order to perform failure
	// percentage-based ejection. If the total number of hosts in the cluster is
	// less than this value, failure percentage-based ejection will not be
	// performed.
	MinimumHosts *uint32 `json:"minimumHosts,omitempty"`
	// The minimum number of total requests that must be collected in one
	// interval (as defined by the interval duration above) to perform failure
	// percentage-based ejection for this host. If the volume is lower than this
	// setting, failure percentage-based ejection will not be performed for this
	// host.
	RequestVolume *uint32 `json:"requestVolume,omitempty"`
	// The failure percentage to use when determining failure percentage-based
	// outlier detection. If the failure percentage of a given host is greater
	// than or equal to this value, it will be ejected.
	Threshold *uint32 `json:"threshold,omitempty"`
}
