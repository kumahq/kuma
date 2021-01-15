package v3

import (
	accesslog_data "github.com/envoyproxy/go-control-plane/envoy/data/accesslog/v3"
	accesslog_config "github.com/envoyproxy/go-control-plane/envoy/extensions/access_loggers/grpc/v3"
)

// TextSpan represents a span of plain text.
type TextSpan string

func (f TextSpan) FormatHttpLogEntry(entry *accesslog_data.HTTPAccessLogEntry) (string, error) {
	return f.String(), nil
}

func (f TextSpan) FormatTcpLogEntry(entry *accesslog_data.TCPAccessLogEntry) (string, error) {
	return f.String(), nil
}

func (f TextSpan) ConfigureHttpLog(config *accesslog_config.HttpGrpcAccessLogConfig) error {
	// has no effect on HttpGrpcAccessLogConfig
	return nil
}

func (f TextSpan) ConfigureTcpLog(config *accesslog_config.TcpGrpcAccessLogConfig) error {
	// has no effect on TcpGrpcAccessLogConfig
	return nil
}

// String returns the canonical representation of this access log fragment.
func (f TextSpan) String() string {
	return string(f)
}
