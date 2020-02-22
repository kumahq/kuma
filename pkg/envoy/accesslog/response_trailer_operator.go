package accesslog

import (
	"fmt"

	accesslog_config "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v2"
	accesslog_data "github.com/envoyproxy/go-control-plane/envoy/data/accesslog/v2"
)

// ResponseTrailerOperator represents a `%TRAILER(X?Y):Z%` command operator.
type ResponseTrailerOperator struct {
	HeaderFormatter
}

func (f *ResponseTrailerOperator) FormatHttpLogEntry(entry *accesslog_data.HTTPAccessLogEntry) (string, error) {
	return f.Format(HeaderMap(entry.GetResponse().GetResponseTrailers()))
}

func (f *ResponseTrailerOperator) FormatTcpLogEntry(entry *accesslog_data.TCPAccessLogEntry) (string, error) {
	return "", nil
}

func (f *ResponseTrailerOperator) ConfigureHttpLog(config *accesslog_config.HttpGrpcAccessLogConfig) error {
	config.AdditionalResponseTrailersToLog = stringList(f.GetOperandHeaders()).AppendToSet(config.AdditionalResponseTrailersToLog)
	return nil
}

func (f *ResponseTrailerOperator) ConfigureTcpLog(config *accesslog_config.TcpGrpcAccessLogConfig) error {
	// has no effect on TcpGrpcAccessLogConfig
	return nil
}

// String returns the canonical representation of this access log fragment.
func (f *ResponseTrailerOperator) String() string {
	return fmt.Sprintf("%%TRAILER%s%%", f.HeaderFormatter.String())
}
