package accesslog

import (
	accesslog_config "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v2"
	accesslog_data "github.com/envoyproxy/go-control-plane/envoy/data/accesslog/v2"
)

// LogConfigurer adjusts configuration of
// `envoy.http_grpc_access_log` and `envoy.tcp_grpc_access_log`
// according to the format string, e.g. to capture additional HTTP headers.
type LogConfigurer interface {
	HttpLogConfigurer
	TcpLogConfigurer
}

// LogEntryFormatter formats a given HTTP or TCP log entry
// according to the format string.
type LogEntryFormatter interface {
	HttpLogEntryFormatter
	TcpLogEntryFormatter
}

// LogConfigureFormatter is a convenience type.
type LogConfigureFormatter interface {
	LogConfigurer
	LogEntryFormatter
}

// HttpLogConfigurer adjusts configuration of `envoy.http_grpc_access_log`
// according to the format string, e.g. to capture additional HTTP headers.
type HttpLogConfigurer interface {
	ConfigureHttpLog(config *accesslog_config.HttpGrpcAccessLogConfig) error
}

// TcpLogConfigurer adjusts configuration of `envoy.tcp_grpc_access_log`
// according to the format string, e.g. to capture filter state objects.
type TcpLogConfigurer interface {
	ConfigureTcpLog(config *accesslog_config.TcpGrpcAccessLogConfig) error
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
