package accesslog

import (
	"github.com/golang/protobuf/ptypes"

	accesslog_data "github.com/envoyproxy/go-control-plane/envoy/data/accesslog/v2"
)

const (
	// defaultStartTimeFormat is a Golang's quivalent of "%Y-%m-%dT%H:%M:%E3SZ" in C++.
	defaultStartTimeFormat = "2006-01-02T15:04:05.000-0700"
)

type StartTimeFormatter string

func (f StartTimeFormatter) FormatHttpLogEntry(entry *accesslog_data.HTTPAccessLogEntry) (string, error) {
	return f.format(entry.GetCommonProperties())
}

func (f StartTimeFormatter) FormatTcpLogEntry(entry *accesslog_data.TCPAccessLogEntry) (string, error) {
	return f.format(entry.GetCommonProperties())
}

func (f StartTimeFormatter) format(entry *accesslog_data.AccessLogCommon) (string, error) {
	if entry.GetStartTime() == nil {
		return "", nil
	}
	startTime, err := ptypes.Timestamp(entry.GetStartTime())
	if err != nil {
		return "", err
	}
	// TODO(yskopets): take format string parameter into account
	return startTime.Format(defaultStartTimeFormat), nil
}
