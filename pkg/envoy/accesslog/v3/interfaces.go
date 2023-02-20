package v3

import (
	accesslog_data "github.com/envoyproxy/go-control-plane/envoy/data/accesslog/v3"
	accesslog_config "github.com/envoyproxy/go-control-plane/envoy/extensions/access_loggers/grpc/v3"
)

// AccessLogFragment represents a fragment of an Envoy access log format string,
// such as a command operator or a span of plain text.
//
// See https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log/usage#command-operators
type AccessLogFragment interface {
	HttpLogEntryFormatter
	TcpLogEntryFormatter
	HttpLogConfigurer
	TcpLogConfigurer
	// String returns the canonical representation of this fragment.
	String() string
}

// HttpLogEntryFormatter formats a given HTTP log entry
// according to the format string.
type HttpLogEntryFormatter interface {
	FormatHttpLogEntry(entry *accesslog_data.HTTPAccessLogEntry) (string, error)
}

// TcpLogEntryFormatter formats a given TCP log entry
// according to the format string.
type TcpLogEntryFormatter interface {
	FormatTcpLogEntry(entry *accesslog_data.TCPAccessLogEntry) (string, error)
}

// HttpLogConfigurer adjusts configuration of `envoy.access_loggers.http_grpc`
// according to the format string, e.g. to capture additional HTTP headers.
type HttpLogConfigurer interface {
	ConfigureHttpLog(config *accesslog_config.HttpGrpcAccessLogConfig) error
}

// TcpLogConfigurer adjusts configuration of `envoy.tcp_grpc_access_log`
// according to the format string, e.g. to capture filter state objects.
type TcpLogConfigurer interface {
	ConfigureTcpLog(config *accesslog_config.TcpGrpcAccessLogConfig) error
}

// AccessLogFragmentInterpolator interpolates placeholders
// added to an access log format string.
// E.g. %KUMA_SOURCE_SERVICE%, %KUMA_DESTINATION_SERVICE%,
// %KUMA_SOURCE_ADDRESS% and %KUMA_SOURCE_ADDRESS_WITHOUT_PORT%
// are examples of such placeholders.
type AccessLogFragmentInterpolator interface {
	// Interpolate returns an access log fragment with all placeholders resolved.
	Interpolate(variables InterpolationVariables) (AccessLogFragment, error)
}

// InterpolationVariables represents a context of Interpolate() operation
// as a map of variables.
type InterpolationVariables map[string]string

func (m InterpolationVariables) Get(variable string) string {
	return m[variable]
}
