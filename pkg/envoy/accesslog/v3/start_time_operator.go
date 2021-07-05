package v3

import (
	"fmt"

	accesslog_data "github.com/envoyproxy/go-control-plane/envoy/data/accesslog/v3"
	accesslog_config "github.com/envoyproxy/go-control-plane/envoy/extensions/access_loggers/grpc/v3"
)

const (
	// defaultStartTimeFormat is a Golang's equivalent of "%Y-%m-%dT%H:%M:%E3SZ" in C++.
	defaultStartTimeFormat = "2006-01-02T15:04:05.000Z"
)

// StartTimeOperator represents a `%START_TIME%` command operator.
type StartTimeOperator string

func (f StartTimeOperator) FormatHttpLogEntry(entry *accesslog_data.HTTPAccessLogEntry) (string, error) {
	return f.format(entry.GetCommonProperties())
}

func (f StartTimeOperator) FormatTcpLogEntry(entry *accesslog_data.TCPAccessLogEntry) (string, error) {
	return f.format(entry.GetCommonProperties())
}

func (f StartTimeOperator) format(entry *accesslog_data.AccessLogCommon) (string, error) {
	if entry.GetStartTime() == nil {
		return "", nil
	}
	if err := entry.GetStartTime().CheckValid(); err != nil {
		return "", err
	}
	startTime := entry.GetStartTime().AsTime()
	// TODO(yskopets): take format string parameter into account
	return startTime.Format(defaultStartTimeFormat), nil
}

func (f StartTimeOperator) ConfigureHttpLog(config *accesslog_config.HttpGrpcAccessLogConfig) error {
	// has no effect on HttpGrpcAccessLogConfig
	return nil
}

func (f StartTimeOperator) ConfigureTcpLog(config *accesslog_config.TcpGrpcAccessLogConfig) error {
	// has no effect on TcpGrpcAccessLogConfig
	return nil
}

// String returns the canonical representation of this access log fragment.
func (f StartTimeOperator) String() string {
	if f == "" {
		return CommandOperatorDescriptor(CMD_START_TIME).String()
	}
	return fmt.Sprintf("%%START_TIME(%s)%%", string(f))
}
