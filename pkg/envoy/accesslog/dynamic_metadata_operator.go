package accesslog

import (
	"strconv"
	"strings"

	accesslog_config "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v2"
	accesslog_data "github.com/envoyproxy/go-control-plane/envoy/data/accesslog/v2"
)

// DynamicMetadataOperator represents a `%DYNAMIC_METADATA(NAMESPACE:KEY*):Z%` command operator.
type DynamicMetadataOperator struct {
	FilterNamespace string
	Path            []string
	MaxLength       int
}

func (f *DynamicMetadataOperator) FormatHttpLogEntry(entry *accesslog_data.HTTPAccessLogEntry) (string, error) {
	return f.format(entry.GetCommonProperties())
}

func (f *DynamicMetadataOperator) FormatTcpLogEntry(entry *accesslog_data.TCPAccessLogEntry) (string, error) {
	return f.format(entry.GetCommonProperties())
}

func (f *DynamicMetadataOperator) format(entry *accesslog_data.AccessLogCommon) (string, error) {
	// TODO(yskopets): implement
	return "UNSUPPORTED_COMMAND(%DYNAMIC_METADATA(NAMESPACE:KEY*):Z%)", nil
}

func (f *DynamicMetadataOperator) ConfigureHttpLog(config *accesslog_config.HttpGrpcAccessLogConfig) error {
	// has no effect on HttpGrpcAccessLogConfig
	return nil
}

func (f *DynamicMetadataOperator) ConfigureTcpLog(config *accesslog_config.TcpGrpcAccessLogConfig) error {
	// has no effect on TcpGrpcAccessLogConfig
	return nil
}

// String returns the canonical representation of this access log fragment.
func (f *DynamicMetadataOperator) String() string {
	var builder []string
	builder = append(builder, "%DYNAMIC_METADATA(")
	builder = append(builder, f.FilterNamespace)
	if len(f.Path) > 0 {
		for _, segment := range f.Path {
			builder = append(builder, ":")
			builder = append(builder, segment)
		}
	}
	builder = append(builder, ")")
	if f.MaxLength != 0 {
		builder = append(builder, ":")
		builder = append(builder, strconv.FormatInt(int64(f.MaxLength), 10))
	}
	builder = append(builder, "%")
	return strings.Join(builder, "")
}
