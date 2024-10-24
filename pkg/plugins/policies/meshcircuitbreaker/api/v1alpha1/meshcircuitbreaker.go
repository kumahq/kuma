// +kubebuilder:object:generate=true
package v1alpha1

import (
	k8s "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
)

// MeshCircuitBreaker
type MeshCircuitBreaker struct {
	// TargetRef is a reference to the resource the policy takes an effect on.
	// The resource could be either a real store object or virtual resource
	// defined in place.
	TargetRef *common_api.TargetRef `json:"targetRef,omitempty"`

	// To list makes a match between the consumed services and corresponding
	// configurations
	To []To `json:"to,omitempty"`

	// From list makes a match between clients and corresponding configurations
	From []From `json:"from,omitempty"`
}

type To struct {
	// TargetRef is a reference to the resource that represents a group of
	// destinations.
	TargetRef common_api.TargetRef `json:"targetRef"`
	// Default is a configuration specific to the group of destinations
	// referenced in 'targetRef'
	Default Conf `json:"default,omitempty"`
}

type From struct {
	// TargetRef is a reference to the resource that represents a group of
	// destinations.
	TargetRef common_api.TargetRef `json:"targetRef"`
	// Default is a configuration specific to the group of destinations
	// referenced in 'targetRef'
	Default Conf `json:"default,omitempty"`
}

type Conf struct {
	// ConnectionLimits contains configuration of each circuit breaking limit,
	// which when exceeded makes the circuit breaker to become open (no traffic
	// is allowed like no current is allowed in the circuits when physical
	// circuit breaker ir open)
	ConnectionLimits *ConnectionLimits `json:"connectionLimits,omitempty"`
	// OutlierDetection contains the configuration of the process of dynamically
	// determining whether some number of hosts in an upstream cluster are
	// performing unlike the others and removing them from the healthy load
	// balancing set. Performance might be along different axes such as
	// consecutive failures, temporal success rate, temporal latency, etc.
	// Outlier detection is a form of passive health checking.
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
	SplitExternalAndLocalErrors *bool `json:"splitExternalAndLocalErrors,omitempty"`
	// Contains configuration for supported outlier detectors
	Detectors *Detectors `json:"detectors,omitempty"`
}

type Detectors struct {
	// In the default mode (outlierDetection.splitExternalAndLocalErrors is
	// false) this detection type takes into account all generated errors:
	// locally originated and externally originated (transaction) errors.
	// In split mode (outlierDetection.splitExternalLocalOriginErrors is true)
	// this detection type takes into account only externally originated
	// (transaction) errors, ignoring locally originated errors.
	// If an upstream host is an HTTP-server, only 5xx types of error are taken
	// into account (see Consecutive Gateway Failure for exceptions).
	// Properly formatted responses, even when they carry an operational error
	// (like index not found, access denied) are not taken into account.
	TotalFailures *DetectorTotalFailures `json:"totalFailures,omitempty"`
	// In the default mode (outlierDetection.splitExternalLocalOriginErrors is
	// false) this detection type takes into account a subset of 5xx errors,
	// called "gateway errors" (502, 503 or 504 status code) and local origin
	// failures, such as timeout, TCP reset etc.
	// In split mode (outlierDetection.splitExternalLocalOriginErrors is true)
	// this detection type takes into account a subset of 5xx errors, called
	// "gateway errors" (502, 503 or 504 status code) and is supported only by
	// the http router.
	GatewayFailures *DetectorGatewayFailures `json:"gatewayFailures,omitempty"`
	// This detection type is enabled only when
	// outlierDetection.splitExternalLocalOriginErrors is true and takes into
	// account only locally originated errors (timeout, reset, etc).
	// If Envoy repeatedly cannot connect to an upstream host or communication
	// with the upstream host is repeatedly interrupted, it will be ejected.
	// Various locally originated problems are detected: timeout, TCP reset,
	// ICMP errors, etc. This detection type is supported by http router and
	// tcp proxy.
	LocalOriginFailures *DetectorLocalOriginFailures `json:"localOriginFailures,omitempty"`
	// Success Rate based outlier detection aggregates success rate data from
	// every host in a cluster. Then at given intervals ejects hosts based on
	// statistical outlier detection. Success Rate outlier detection will not be
	// calculated for a host if its request volume over the aggregation interval
	// is less than the outlierDetection.detectors.successRate.requestVolume
	// value.
	// Moreover, detection will not be performed for a cluster if the number of
	// hosts with the minimum required request volume in an interval is less
	// than the outlierDetection.detectors.successRate.minimumHosts value.
	// In the default configuration mode
	// (outlierDetection.splitExternalLocalOriginErrors is false) this detection
	// type takes into account all types of errors: locally and externally
	// originated.
	// In split mode (outlierDetection.splitExternalLocalOriginErrors is true),
	// locally originated errors and externally originated (transaction) errors
	// are counted and treated separately.
	SuccessRate *DetectorSuccessRateFailures `json:"successRate,omitempty"`
	// Failure Percentage based outlier detection functions similarly to success
	// rate detection, in that it relies on success rate data from each host in
	// a cluster. However, rather than compare those values to the mean success
	// rate of the cluster as a whole, they are compared to a flat
	// user-configured threshold. This threshold is configured via the
	// outlierDetection.failurePercentageThreshold field.
	// The other configuration fields for failure percentage based detection are
	// similar to the fields for success rate detection. As with success rate
	// detection, detection will not be performed for a host if its request
	// volume over the aggregation interval is less than the
	// outlierDetection.detectors.failurePercentage.requestVolume value.
	// Detection also will not be performed for a cluster if the number of hosts
	// with the minimum required request volume in an interval is less than the
	// outlierDetection.detectors.failurePercentage.minimumHosts value.
	FailurePercentage *DetectorFailurePercentageFailures `json:"failurePercentage,omitempty"`
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
	// success_rate_standard_deviation_factor).
	// Either int or decimal represented as string.
	StandardDeviationFactor *intstr.IntOrString `json:"standardDeviationFactor,omitempty"`
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
