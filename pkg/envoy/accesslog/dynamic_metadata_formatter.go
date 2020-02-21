package accesslog

import (
	"strconv"
	"strings"

	accesslog_data "github.com/envoyproxy/go-control-plane/envoy/data/accesslog/v2"
)

// DynamicMetadataFormatter represents a `%DYNAMIC_METADATA(NAMESPACE:KEY*):Z%` command operator.
type DynamicMetadataFormatter struct {
	FilterNamespace string
	Path            []string
	MaxLength       int
}

func (f *DynamicMetadataFormatter) FormatHttpLogEntry(entry *accesslog_data.HTTPAccessLogEntry) (string, error) {
	return f.format(entry.GetCommonProperties())
}

func (f *DynamicMetadataFormatter) FormatTcpLogEntry(entry *accesslog_data.TCPAccessLogEntry) (string, error) {
	return f.format(entry.GetCommonProperties())
}

func (f *DynamicMetadataFormatter) format(entry *accesslog_data.AccessLogCommon) (string, error) {
	// TODO(yskopets): implement
	return "UNSUPPORTED_COMMAND(%DYNAMIC_METADATA(NAMESPACE:KEY*):Z%)", nil
}

// String returns the canonical representation of this command operator.
func (f *DynamicMetadataFormatter) String() string {
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
