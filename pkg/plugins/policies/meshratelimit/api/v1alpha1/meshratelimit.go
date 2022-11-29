// +kubebuilder:object:generate=true
package v1alpha1

import (
	k8s "k8s.io/apimachinery/pkg/apis/meta/v1"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
)

// MeshRateLimit
type MeshRateLimit struct {
	// TargetRef is a reference to the resource the policy takes an effect on.
	// The resource could be either a real store object or virtual resource
	// defined inplace.
	TargetRef common_api.TargetRef `json:"targetRef,omitempty"`
	// From list makes a match between clients and corresponding configurations
	From []From `json:"from,omitempty"`
}

type From struct {
	// TargetRef is a reference to the resource that represents a group of
	// clients.
	TargetRef common_api.TargetRef `json:"targetRef,omitempty"`
	// Default is a configuration specific to the group of clients referenced in
	// 'targetRef'
	Default Conf `json:"default,omitempty"`
}

type Conf struct {
	Local *Local `json:"local,omitempty"`
}

// LocalConf defines local http or/and tcp rate limit configuration
type Local struct {
	HTTP *LocalHTTP `json:"http,omitempty"`
	TCP  *LocalTCP  `json:"tcp,omitempty"`
}

// LocalHTTP defines confguration of local HTTP rate limiting
// https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/local_rate_limit_filter
type LocalHTTP struct {
	// Define if rate limiting should be disabled.
	// Default: false
	// +optional
	Disabled bool `json:"disabled,omitempty"`

	// The number of HTTP requests this RateLimiter allows
	// +required
	Requests uint32 `json:"requests,omitempty"`

	// The the interval for which `requests` will be accounted.
	// +required
	Interval k8s.Duration `json:"interval,omitempty"`

	// Describes the actions to take on a rate limit event
	// +optional
	// +nullable
	OnRateLimit *OnRateLimit `json:"onRateLimit,omitempty"`
}

type OnRateLimit struct {
	// The HTTP status code to be set on a rate limit event
	// +optional
	// +nullable
	Status *uint32 `json:"status,omitempty"`

	// The Headers to be added to the HTTP response on a rate limit event
	// +optional
	// +nullable
	Headers []HeaderValue `json:"headers,omitempty"`
}

type HeaderValue struct {
	// Header name
	// +optional
	Key string `json:"key,omitempty"`

	// Header value
	// +optional
	Value string `json:"value,omitempty"`

	// Should the header be appended
	// +optional
	// +nullable
	Append *bool `json:"append,omitempty"`
}

// LocalTCP defines confguration of local TCP rate limiting
// https://www.envoyproxy.io/docs/envoy/latest/configuration/listeners/network_filters/local_rate_limit_filter
type LocalTCP struct {
	// Define if rate limiting should be disabled.
	// Default: false
	// +optional
	Disabled bool `json:"disabled,omitempty"`

	// The number of connections that RateLimiter allows
	// +required
	Connections uint32 `json:"connections,omitempty"`

	// The interval of adding tokens into bucket. Must be >= 50ms
	// +required
	Interval k8s.Duration `json:"interval,omitempty"`
}
