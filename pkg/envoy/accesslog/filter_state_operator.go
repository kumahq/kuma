package accesslog

import (
	"strconv"
	"strings"

	accesslog_config "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v2"
	accesslog_data "github.com/envoyproxy/go-control-plane/envoy/data/accesslog/v2"
)

// FilterStateOperator represents a `%FILTER_STATE(KEY):Z%` command operator.
type FilterStateOperator struct {
	Key       string
	MaxLength int
}

func (f *FilterStateOperator) FormatHttpLogEntry(entry *accesslog_data.HTTPAccessLogEntry) (string, error) {
	return f.format(entry.GetCommonProperties())
}

func (f *FilterStateOperator) FormatTcpLogEntry(entry *accesslog_data.TCPAccessLogEntry) (string, error) {
	return f.format(entry.GetCommonProperties())
}

func (f *FilterStateOperator) format(entry *accesslog_data.AccessLogCommon) (string, error) {
	// TODO(yskopets): implement
	return "UNSUPPORTED_COMMAND(%FILTER_STATE(KEY):Z%)", nil
}

func (f *FilterStateOperator) ConfigureHttpLog(config *accesslog_config.HttpGrpcAccessLogConfig) error {
	if objects := f.appendToSet(config.GetCommonConfig().GetFilterStateObjectsToLog()); objects != nil {
		if config.CommonConfig == nil {
			config.CommonConfig = &accesslog_config.CommonGrpcAccessLogConfig{}
		}
		config.CommonConfig.FilterStateObjectsToLog = objects
	}
	return nil
}

func (f *FilterStateOperator) ConfigureTcpLog(config *accesslog_config.TcpGrpcAccessLogConfig) error {
	if objects := f.appendToSet(config.GetCommonConfig().GetFilterStateObjectsToLog()); objects != nil {
		if config.CommonConfig == nil {
			config.CommonConfig = &accesslog_config.CommonGrpcAccessLogConfig{}
		}
		config.CommonConfig.FilterStateObjectsToLog = objects
	}
	return nil
}

func (f *FilterStateOperator) appendToSet(dest []string) []string {
	if f.Key == "" {
		return dest
	}
	return stringList{f.Key}.AppendToSet(dest)
}

// String returns the canonical representation of this access log fragment.
func (f *FilterStateOperator) String() string {
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
