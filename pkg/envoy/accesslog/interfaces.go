package accesslog

import (
	accesslog_config "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v2"
	accesslog_data "github.com/envoyproxy/go-control-plane/envoy/data/accesslog/v2"
)

type LogConfigurer interface {
	HttpLogConfigurer
	TcpLogConfigurer
}

type LogEntryFormatter interface {
	HttpLogEntryFormatter
	TcpLogEntryFormatter
}

type LogConfigureFormatter interface {
	LogConfigurer
	LogEntryFormatter
}

type HttpLogConfigurer interface {
	ConfigureHttpLog(config *accesslog_config.HttpGrpcAccessLogConfig) error
}

type TcpLogConfigurer interface {
	ConfigureTcpLog(config *accesslog_config.TcpGrpcAccessLogConfig) error
}

type HttpLogEntryFormatter interface {
	FormatHttpLogEntry(entry *accesslog_data.HTTPAccessLogEntry) (string, error)
}

type TcpLogEntryFormatter interface {
	FormatTcpLogEntry(entry *accesslog_data.TCPAccessLogEntry) (string, error)
}
