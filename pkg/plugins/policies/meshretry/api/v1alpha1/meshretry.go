// +kubebuilder:object:generate=true
package v1alpha1

import (
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
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

var HttpMethodPrefix HTTPRetryOn = "HttpMethod"

var (
	// All5xx means Envoy will attempt a retry if the upstream server responds with
	// any 5xx response code, or does not respond at all (disconnect/reset/read timeout).
	// (Includes ConnectFailure and RefusedStream)
	All5xx HTTPRetryOn = "5xx"

	// GatewayError is similar to the 5XX policy but will only retry requests
	// that result in a 502, 503, or 504.
	GatewayError HTTPRetryOn = "GatewayError"

	// Reset means Envoy will attempt a retry if the upstream server does not respond at all
	// (disconnect/reset/read timeout.)
	Reset HTTPRetryOn = "Reset"

	// Retriable4xx means Envoy will attempt a retry if the upstream server responds with
	// a retriable 4xx response code. Currently, the only response code in this category is 409.
	Retriable4xx HTTPRetryOn = "Retriable4xx"

	// ConnectFailure means Envoy will attempt a retry if a request is failed because of
	// a connection failure to the upstream server (connect timeout, etc.). (Included in 5XX)
	ConnectFailure HTTPRetryOn = "ConnectFailure"

	// EnvoyRatelimited means Envoy will retry if the header x-envoy-ratelimited is present.
	EnvoyRatelimited HTTPRetryOn = "EnvoyRatelimited"

	// RefusedStream means Envoy will attempt a retry if the upstream server resets the stream
	// with a RefusedStream error code. This reset type indicates that a request is safe to retry.
	// (Included in 5XX)
	RefusedStream HTTPRetryOn = "RefusedStream"

	// Http3PostConnectFailure means Envoy will attempt a retry if a request is sent over
	// HTTP/3 to the upstream server and failed after getting connected.
	Http3PostConnectFailure HTTPRetryOn = "Http3PostConnectFailure"

	// HttpMethodConnect means Envoy will attempt a retry if the used HTTP method is CONNECT.
	HttpMethodConnect = HttpMethodPrefix + "Connect"

	// HttpMethodDelete means Envoy will attempt a retry if the used HTTP method is DELETE.
	HttpMethodDelete = HttpMethodPrefix + "Delete"

	// HttpMethodGet means Envoy will attempt a retry if the used HTTP method is GET.
	HttpMethodGet = HttpMethodPrefix + "Get"

	// HttpMethodHead means Envoy will attempt a retry if the used HTTP method is HEAD.
	HttpMethodHead = HttpMethodPrefix + "Head"

	// HttpMethodOptions means Envoy will attempt a retry if the used HTTP method is OPTIONS.
	HttpMethodOptions = HttpMethodPrefix + "Options"

	// HttpMethodPatch means Envoy will attempt a retry if the used HTTP method is PATCH.
	HttpMethodPatch = HttpMethodPrefix + "Patch"

	// HttpMethodPost means Envoy will attempt a retry if the used HTTP method is POST.
	HttpMethodPost = HttpMethodPrefix + "Post"

	// HttpMethodPut means Envoy will attempt a retry if the used HTTP method is PUT.
	HttpMethodPut = HttpMethodPrefix + "Put"

	// HttpMethodTrace means Envoy will attempt a retry if the used HTTP method is TRACE.
	HttpMethodTrace = HttpMethodPrefix + "Trace"
)

var HttpRetryOnEnumToEnvoyValue = map[HTTPRetryOn]string{
	All5xx:                  "5xx",
	GatewayError:            "gateway-error",
	Reset:                   "reset",
	Retriable4xx:            "retriable-4xx",
	ConnectFailure:          "connect-failure",
	EnvoyRatelimited:        "envoy-ratelimited",
	RefusedStream:           "refused-stream",
	Http3PostConnectFailure: "http3-post-connect-failure",
	HttpMethodConnect:       "CONNECT",
	HttpMethodDelete:        "DELETE",
	HttpMethodGet:           "GET",
	HttpMethodHead:          "HEAD",
	HttpMethodOptions:       "OPTIONS",
	HttpMethodPatch:         "PATCH",
	HttpMethodPost:          "POST",
	HttpMethodPut:           "PUT",
	HttpMethodTrace:         "TRACE",
}

type HTTP struct {
	// NumRetries is the number of attempts that will be made on failed (and retriable) requests
	NumRetries *uint32 `json:"numRetries,omitempty"`
	// PerTryTimeout is the amount of time after which retry attempt should timeout.
	// Setting this timeout to 0 will disable it. Default is 15s.
	PerTryTimeout *k8s.Duration `json:"perTryTimeout,omitempty"`
	// BackOff is a configuration of durations which will be used in exponential
	// backoff strategy between retries
	BackOff *BackOff `json:"backOff,omitempty"`
	// RateLimitedBackOff is a configuration of backoff which will be used
	// when the upstream returns one of the headers configured.
	RateLimitedBackOff *RateLimitedBackOff `json:"rateLimitedBackOff,omitempty"`
	// RetryOn is a list of conditions which will cause a retry. Available values are:
	// [5XX, GatewayError, Reset, Retriable4xx, ConnectFailure, EnvoyRatelimited,
	// RefusedStream, Http3PostConnectFailure, HttpMethodConnect, HttpMethodDelete,
	// HttpMethodGet, HttpMethodHead, HttpMethodOptions, HttpMethodPatch,
	// HttpMethodPost, HttpMethodPut, HttpMethodTrace].
	// Also, any HTTP status code (500, 503, etc).
	RetryOn *[]HTTPRetryOn `json:"retryOn,omitempty"`
	// RetriableResponseHeaders is an HTTP response headers that trigger a retry
	// if present in the response. A retry will be triggered if any of the header
	// matches match the upstream response headers.
	RetriableResponseHeaders *[]common_api.HeaderMatch `json:"retriableResponseHeaders,omitempty"`
	// RetriableRequestHeaders is an HTTP headers which must be present in the request
	// for retries to be attempted.
	RetriableRequestHeaders *[]common_api.HeaderMatch `json:"retriableRequestHeaders,omitempty"`
}
type GRPCRetryOn string

var (
	// Canceled means Envoy will attempt a retry if the gRPC status code in
	// the response headers is “cancelled” (1)
	Canceled GRPCRetryOn = "Canceled"

	// DeadlineExceeded Envoy will attempt a retry if the gRPC status code in
	// the response headers is “deadline-exceeded” (4)
	DeadlineExceeded GRPCRetryOn = "DeadlineExceeded"

	// Internal Envoy will attempt to retry if the gRPC status code in the
	// response headers is “internal” (13)
	Internal GRPCRetryOn = "Internal"

	// ResourceExhausted means Envoy will attempt a retry if the gRPC status code
	// in the response headers is “resource-exhausted” (8)
	ResourceExhausted GRPCRetryOn = "ResourceExhausted"

	// Unavailable means Envoy will attempt a retry if the gRPC status code in
	// the response headers is “unavailable” (14)
	Unavailable GRPCRetryOn = "Unavailable"
)

var GrpcRetryOnEnumToEnvoyValue = map[GRPCRetryOn]string{
	Canceled:          "canceled",
	DeadlineExceeded:  "deadline-exceeded",
	Internal:          "internal",
	ResourceExhausted: "resource-exhausted",
	Unavailable:       "unavailable",
}

type GRPC struct {
	// NumRetries is the number of attempts that will be made on failed (and retriable) requests.
	NumRetries *uint32 `json:"numRetries,omitempty"`
	// PerTryTimeout is the amount of time after which retry attempt should timeout.
	// Setting this timeout to 0 will disable it. Default is 15s.
	PerTryTimeout *k8s.Duration `json:"perTryTimeout,omitempty"`
	// BackOff is a configuration of durations which will be used in exponential
	// backoff strategy between retries.
	BackOff *BackOff `json:"backOff,omitempty"`
	// RateLimitedBackOff is a configuration of backoff which will be used
	// when the upstream returns one of the headers configured.
	RateLimitedBackOff *RateLimitedBackOff `json:"rateLimitedBackOff,omitempty"`
	// RetryOn is a list of conditions which will cause a retry. Available values are:
	// [Canceled, DeadlineExceeded, Internal, ResourceExhausted, Unavailable].
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

type RateLimitedBackOff struct {
	// ResetHeaders specifies the list of headers (like Retry-After or X-RateLimit-Reset) to match against the response.
	// Headers are tried in order, and matched case-insensitive. The first header to be parsed successfully is used.
	// If no headers match the default exponential BackOff is used instead.
	ResetHeaders *[]ResetHeader `json:"resetHeaders,omitempty"`
	// MaxInterval is a maximal amount of time which will be taken between retries.
	// Default is 300 seconds.
	MaxInterval *k8s.Duration `json:"maxInterval,omitempty"`
}

// +kubebuilder:validation:Enum=Seconds;UnixTimestamp
type RateLimitFormat string

var (
	// Seconds is an integer that represents the number of seconds to wait before retrying.
	Seconds RateLimitFormat = "Seconds"

	// UnixTimestamp is an integer that represents the point in time at which to retry, as a Unix timestamp in seconds.
	UnixTimestamp RateLimitFormat = "UnixTimestamp"
)

var RateLimitFormatEnumToEnvoyValue = map[RateLimitFormat]envoy_route.RetryPolicy_ResetHeaderFormat{
	Seconds:       envoy_route.RetryPolicy_SECONDS,
	UnixTimestamp: envoy_route.RetryPolicy_UNIX_TIMESTAMP,
}

type ResetHeader struct {
	// The Name of the reset header.
	Name common_api.HeaderName `json:"name"`
	// The format of the reset header, either Seconds or UnixTimestamp.
	Format RateLimitFormat `json:"format"`
}
