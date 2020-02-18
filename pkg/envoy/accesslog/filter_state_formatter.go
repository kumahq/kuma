package accesslog

import (
	accesslog_config "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v2"
	accesslog_data "github.com/envoyproxy/go-control-plane/envoy/data/accesslog/v2"
)

type FilterStateFormatter struct {
	Key       string
	MaxLength int
}

func (f *FilterStateFormatter) FormatHttpLogEntry(entry *accesslog_data.HTTPAccessLogEntry) (string, error) {
	return f.format(entry.GetCommonProperties())
}

func (f *FilterStateFormatter) FormatTcpLogEntry(entry *accesslog_data.TCPAccessLogEntry) (string, error) {
	return f.format(entry.GetCommonProperties())
}

func (f *FilterStateFormatter) format(entry *accesslog_data.AccessLogCommon) (string, error) {
	// TODO(yskopets): implement
	return "UNSUPPORTED(FILTER_STATE)", nil
}

func (f *FilterStateFormatter) ConfigureHttpLog(config *accesslog_config.HttpGrpcAccessLogConfig) error {
	if config.CommonConfig == nil {
		config.CommonConfig = &accesslog_config.CommonGrpcAccessLogConfig{}
	}
	config.CommonConfig.FilterStateObjectsToLog = append(config.CommonConfig.FilterStateObjectsToLog, f.Key)
	return f.configure(config.CommonConfig)
}

func (f *FilterStateFormatter) ConfigureTcpLog(config *accesslog_config.TcpGrpcAccessLogConfig) error {
	if config.CommonConfig == nil {
		config.CommonConfig = &accesslog_config.CommonGrpcAccessLogConfig{}
	}
	return f.configure(config.CommonConfig)
}

func (f *FilterStateFormatter) configure(config *accesslog_config.CommonGrpcAccessLogConfig) error {
	config.FilterStateObjectsToLog = append(config.FilterStateObjectsToLog, f.Key)
	return nil
}
