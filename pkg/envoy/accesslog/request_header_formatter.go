package accesslog

import (
	accesslog_config "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v2"
	accesslog_data "github.com/envoyproxy/go-control-plane/envoy/data/accesslog/v2"
)

type RequestHeaderFormatter struct {
	HeaderFormatter
}

func (f *RequestHeaderFormatter) FormatHttpLogEntry(entry *accesslog_data.HTTPAccessLogEntry) (string, error) {
	// TODO(yskopets): handle pseudo headers
	return f.Format(entry.GetRequest().GetRequestHeaders())
}

func (f *RequestHeaderFormatter) FormatTcpLogEntry(entry *accesslog_data.TCPAccessLogEntry) (string, error) {
	return "", nil
}

func (f *RequestHeaderFormatter) ConfigureHttpLog(config *accesslog_config.HttpGrpcAccessLogConfig) error {
	config.AdditionalRequestHeadersToLog = f.AppendTo(config.AdditionalRequestHeadersToLog)
	return nil
}
