package accesslog

import (
	accesslog_config "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v2"
	accesslog_data "github.com/envoyproxy/go-control-plane/envoy/data/accesslog/v2"
)

type ResponseHeaderFormatter struct {
	HeaderFormatter
}

func (f *ResponseHeaderFormatter) FormatHttpLogEntry(entry *accesslog_data.HTTPAccessLogEntry) (string, error) {
	return f.Format(entry.GetResponse().GetResponseHeaders())
}

func (f *ResponseHeaderFormatter) FormatTcpLogEntry(entry *accesslog_data.TCPAccessLogEntry) (string, error) {
	return "", nil
}

func (f *ResponseHeaderFormatter) ConfigureHttpLog(config *accesslog_config.HttpGrpcAccessLogConfig) error {
	config.AdditionalResponseHeadersToLog = f.AppendTo(config.AdditionalResponseHeadersToLog)
	return nil
}
