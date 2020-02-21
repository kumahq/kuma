package accesslog

import (
	"strconv"
	"strings"

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
	return "UNSUPPORTED_COMMAND(%FILTER_STATE(KEY):Z%)", nil
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

// String returns the canonical representation of this command operator.
func (f *FilterStateFormatter) String() string {
	var builder []string
	builder = append(builder, "%FILTER_STATE(")
	builder = append(builder, f.Key)
	builder = append(builder, ")")
	if f.MaxLength != 0 {
		builder = append(builder, ":")
		builder = append(builder, strconv.FormatInt(int64(f.MaxLength), 10))
	}
	builder = append(builder, "%")
	return strings.Join(builder, "")
}
