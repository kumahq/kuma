package accesslog

import (
	accesslog_data "github.com/envoyproxy/go-control-plane/envoy/data/accesslog/v2"
)

type TextLiteralFormatter string

func (f TextLiteralFormatter) FormatHttpLogEntry(entry *accesslog_data.HTTPAccessLogEntry) (string, error) {
	return f.format()
}

func (f TextLiteralFormatter) FormatTcpLogEntry(entry *accesslog_data.TCPAccessLogEntry) (string, error) {
	return f.format()
}

func (f TextLiteralFormatter) format() (string, error) {
	return string(f), nil
}
