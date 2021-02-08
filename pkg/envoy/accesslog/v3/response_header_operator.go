package v3

import (
	"fmt"

	accesslog_data "github.com/envoyproxy/go-control-plane/envoy/data/accesslog/v3"
	accesslog_config "github.com/envoyproxy/go-control-plane/envoy/extensions/access_loggers/grpc/v3"
)

// ResponseHeaderOperator represents a `%RESP(X?Y):Z%` command operator.
type ResponseHeaderOperator struct {
	HeaderFormatter
}

func (f *ResponseHeaderOperator) FormatHttpLogEntry(entry *accesslog_data.HTTPAccessLogEntry) (string, error) {
	return f.Format(HeaderMap(entry.GetResponse().GetResponseHeaders()))
}

func (f *ResponseHeaderOperator) FormatTcpLogEntry(entry *accesslog_data.TCPAccessLogEntry) (string, error) {
	return "", nil
}

func (f *ResponseHeaderOperator) ConfigureHttpLog(config *accesslog_config.HttpGrpcAccessLogConfig) error {
	config.AdditionalResponseHeadersToLog = stringList(f.GetOperandHeaders()).AppendToSet(config.AdditionalResponseHeadersToLog)
	return nil
}

func (f *ResponseHeaderOperator) ConfigureTcpLog(config *accesslog_config.TcpGrpcAccessLogConfig) error {
	// has no effect on TcpGrpcAccessLogConfig
	return nil
}

// String returns the canonical representation of this access log fragment.
func (f *ResponseHeaderOperator) String() string {
	return fmt.Sprintf("%%RESP%s%%", f.HeaderFormatter.String())
}
