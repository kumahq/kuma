package accesslog

import (
	accesslog_config "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v2"
	accesslog_data "github.com/envoyproxy/go-control-plane/envoy/data/accesslog/v2"
)

type ResponseTrailerFormatter struct {
	HeaderFormatter
}

func (f *ResponseTrailerFormatter) FormatHttpLogEntry(entry *accesslog_data.HTTPAccessLogEntry) (string, error) {
	return f.Format(entry.GetResponse().GetResponseTrailers())
}

func (f *ResponseTrailerFormatter) FormatTcpLogEntry(entry *accesslog_data.TCPAccessLogEntry) (string, error) {
	return "", nil
}

func (f *ResponseTrailerFormatter) ConfigureHttpLog(config *accesslog_config.HttpGrpcAccessLogConfig) error {
	config.AdditionalResponseTrailersToLog = f.AppendTo(config.AdditionalResponseTrailersToLog)
	return nil
}
