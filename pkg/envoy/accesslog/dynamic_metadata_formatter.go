package accesslog

import (
	accesslog_data "github.com/envoyproxy/go-control-plane/envoy/data/accesslog/v2"
)

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
	return "UNSUPPORTED(DYNAMIC_METADATA)", nil
}
