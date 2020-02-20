package accesslog

import (
	accesslog_config "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v2"
	accesslog_data "github.com/envoyproxy/go-control-plane/envoy/data/accesslog/v2"
)

// DynamicMetadataFormatter represents a `%FILTER_STATE(KEY):Z%` command operator.
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
	if objects := f.appendTo(config.GetCommonConfig().GetFilterStateObjectsToLog()); objects != nil {
		if config.CommonConfig == nil {
			config.CommonConfig = &accesslog_config.CommonGrpcAccessLogConfig{}
		}
		config.CommonConfig.FilterStateObjectsToLog = objects
	}
	return nil
}

func (f *FilterStateFormatter) ConfigureTcpLog(config *accesslog_config.TcpGrpcAccessLogConfig) error {
	if objects := f.appendTo(config.GetCommonConfig().GetFilterStateObjectsToLog()); objects != nil {
		if config.CommonConfig == nil {
			config.CommonConfig = &accesslog_config.CommonGrpcAccessLogConfig{}
		}
		config.CommonConfig.FilterStateObjectsToLog = objects
	}
	return nil
}

func (f *FilterStateFormatter) appendTo(values []string) []string {
	if f.Key != "" && !stringSet(values).Contains(f.Key) {
		return append(values, f.Key)
	}
	return values
}
