// +kubebuilder:object:generate=true
package v1alpha1

import (
	k8s "k8s.io/apimachinery/pkg/apis/meta/v1"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
)

// MeshRetry
type MeshRetry struct {
	// TargetRef is a reference to the resource the policy takes an effect on.
	// The resource could be either a real store object or virtual resource
	// defined inplace.
	TargetRef common_api.TargetRef `json:"targetRef"`
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
	// TCP defines a configuration of retries for TCP traffic
	TCP *TCP `json:"tcp,omitempty"`
	// HTTP defines a configuration of retries for HTTP traffic
	HTTP *HTTP `json:"http,omitempty"`
	// GRPC defines a configuration of retries for GRPC traffic
	GRPC *GRPC `json:"grpc,omitempty"`
}

type TCP struct {
	// MaxConnectAttempt is a maximal amount of TCP connection attempts
	// which will be made before giving up
	MaxConnectAttempt *uint32 `json:"maxConnectAttempt,omitempty"`
}

type HTTPRetryOn string

var (
	// ALL_5XX means Envoy will attempt a retry if the upstream server responds with
	// any 5xx response code, or does not respond at all (disconnect/reset/read timeout).
	// (Includes CONNECT_FAILURE and REFUSED_STREAM)
	ALL_5XX HTTPRetryOn = "5XX"

	// GATEWAY_ERROR is similar to the 5XX policy but will only retry requests
	// that result in a 502, 503, or 504.
	GATEWAY_ERROR HTTPRetryOn = "GATEWAY_ERROR"

	// RESET means Envoy will attempt a retry if the upstream server does not respond at all
	// (disconnect/reset/read timeout.)
	RESET HTTPRetryOn = "RESET"

	// RETRIABLE_4XX means Envoy will attempt a retry if the upstream server responds with
	// a retriable 4xx response code. Currently, the only response code in this category is 409.
	RETRIABLE_4XX HTTPRetryOn = "RETRIABLE_4XX"

	// CONNECT_FAILURE means Envoy will attempt a retry if a request is failed because of
	// a connection failure to the upstream server (connect timeout, etc.). (Included in 5XX)
	CONNECT_FAILURE HTTPRetryOn = "CONNECT_FAILURE"

	// ENVOY_RATELIMITED means Envoy will retry if the header x-envoy-ratelimited is present.
	ENVOY_RATELIMITED HTTPRetryOn = "ENVOY_RATELIMITED"

	// REFUSED_STREAM means Envoy will attempt a retry if the upstream server resets the stream
	// with a REFUSED_STREAM error code. This reset type indicates that a request is safe to retry.
	// (Included in 5XX)
	REFUSED_STREAM HTTPRetryOn = "REFUSED_STREAM"

	// HTTP3_POST_CONNECT_FAILURE means Envoy will attempt a retry if a request is sent over
	// HTTP/3 to the upstream server and failed after getting connected.
	HTTP3_POST_CONNECT_FAILURE HTTPRetryOn = "HTTP3_POST_CONNECT_FAILURE"

	// HTTP_METHOD_CONNECT means Envoy will attempt a retry if the used HTTP method is CONNECT.
	HTTP_METHOD_CONNECT HTTPRetryOn = "HTTP_METHOD_CONNECT"

	// HTTP_METHOD_DELETE means Envoy will attempt a retry if the used HTTP method is DELETE.
	HTTP_METHOD_DELETE HTTPRetryOn = "HTTP_METHOD_DELETE"

	// HTTP_METHOD_GET means Envoy will attempt a retry if the used HTTP method is GET.
	HTTP_METHOD_GET HTTPRetryOn = "HTTP_METHOD_GET"

	// HTTP_METHOD_HEAD means Envoy will attempt a retry if the used HTTP method is HEAD.
	HTTP_METHOD_HEAD HTTPRetryOn = "HTTP_METHOD_HEAD"

	// HTTP_METHOD_OPTIONS means Envoy will attempt a retry if the used HTTP method is OPTIONS.
	HTTP_METHOD_OPTIONS HTTPRetryOn = "HTTP_METHOD_OPTIONS"

	// HTTP_METHOD_PATCH means Envoy will attempt a retry if the used HTTP method is PATCH.
	HTTP_METHOD_PATCH HTTPRetryOn = "HTTP_METHOD_PATCH"

	// HTTP_METHOD_POST means Envoy will attempt a retry if the used HTTP method is POST.
	HTTP_METHOD_POST HTTPRetryOn = "HTTP_METHOD_POST"

	// HTTP_METHOD_PUT means Envoy will attempt a retry if the used HTTP method is PUT.
	HTTP_METHOD_PUT HTTPRetryOn = "HTTP_METHOD_PUT"

	// HTTP_METHOD_TRACE means Envoy will attempt a retry if the used HTTP method is TRACE.
	HTTP_METHOD_TRACE HTTPRetryOn = "HTTP_METHOD_TRACE"
)

type HTTP struct {
	// NumRetries is the number of attempts that will be made on failed (and retriable) requests
	NumRetries *uint32 `json:"numRetries,omitempty"`
	// PerTryTimeout is the amount of time after which retry attempt should timeout.
	// Setting this timeout to 0 will disable it. Default is 15s.
	PerTryTimeout *k8s.Duration `json:"perTryTimeout,omitempty"`
	// BackOff is a configuration of durations which will be used in exponential
	// backoff strategy between retries
	BackOff *BackOff `json:"backOff,omitempty"`
	// RetryOn is a list of conditions which will cause a retry. Available values are:
	// [5XX, GATEWAY_ERROR, RESET, RETRIABLE_4XX, CONNECT_FAILURE, ENVOY_RATELIMITED,
	// REFUSED_STREAM, HTTP3_POST_CONNECT_FAILURE, HTTP_METHOD_CONNECT, HTTP_METHOD_DELETE,
	// HTTP_METHOD_GET, HTTP_METHOD_HEAD, HTTP_METHOD_OPTIONS, HTTP_METHOD_PATCH,
	// HTTP_METHOD_POST, HTTP_METHOD_PUT, HTTP_METHOD_TRACE].
	// Also, any HTTP status code (500, 503, etc).
	RetryOn *[]HTTPRetryOn `json:"retryOn,omitempty"`
	// RetriableResponseHeaders is an HTTP response headers that trigger a retry
	// if present in the response. A retry will be triggered if any of the header
	// matches match the upstream response headers.
	RetriableResponseHeaders *[]common_api.HeaderMatcher `json:"retriableResponseHeaders,omitempty"`
	// RetriableRequestHeaders is an HTTP headers which must be present in the request
	// for retries to be attempted.
	RetriableRequestHeaders *[]common_api.HeaderMatcher `json:"retriableRequestHeaders,omitempty"`
}

type GRPCRetryOn string

var (
	// CANCELED means Envoy will attempt a retry if the gRPC status code in
	// the response headers is “cancelled” (1)
	CANCELED GRPCRetryOn = "CANCELED"

	// DEADLINE_EXCEEDED Envoy will attempt a retry if the gRPC status code in
	// the response headers is “deadline-exceeded” (4)
	DEADLINE_EXCEEDED GRPCRetryOn = "DEADLINE_EXCEEDED"

	// INTERNAL Envoy will attempt to retry if the gRPC status code in the
	// response headers is “internal” (13)
	INTERNAL GRPCRetryOn = "INTERNAL"

	// RESOURCE_EXHAUSTED means Envoy will attempt a retry if the gRPC status code
	// in the response headers is “resource-exhausted” (8)
	RESOURCE_EXHAUSTED GRPCRetryOn = "RESOURCE_EXHAUSTED"

	// UNAVAILABLE means Envoy will attempt a retry if the gRPC status code in
	// the response headers is “unavailable” (14)
	UNAVAILABLE GRPCRetryOn = "UNAVAILABLE"
)

type GRPC struct {
	// NumRetries is the number of attempts that will be made on failed (and retriable) requests
	NumRetries *uint32 `json:"numRetries,omitempty"`
	// PerTryTimeout is the amount of time after which retry attempt should timeout.
	// Setting this timeout to 0 will disable it. Default is 15s.
	PerTryTimeout *k8s.Duration `json:"perTryTimeout,omitempty"`
	// BackOff is a configuration of durations which will be used in exponential
	// backoff strategy between retries
	BackOff *BackOff `json:"backOff,omitempty"`
	// RetryOn is a list of conditions which will cause a retry. Available values are:
	// [CANCELED, DEADLINE_EXCEEDED, INTERNAL, RESOURCE_EXHAUSTED, UNAVAILABLE]
	RetryOn *[]GRPCRetryOn `json:"retryOn,omitempty"`
}

type BackOff struct {
	// BaseInterval is an amount of time which should be taken between retries.
	// Must be greater than zero. Values less than 1 ms are rounded up to 1 ms.
	// Default is 25ms.
	BaseInterval *k8s.Duration `json:"baseInterval,omitempty"`
	// MaxInterval is a maximal amount of time which will be taken between retries.
	// Default is 10 times the "BaseInterval".
	MaxInterval *k8s.Duration `json:"maxInterval,omitempty"`
}
