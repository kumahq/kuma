package accesslog

import (
	"fmt"

	accesslog_config "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v2"
	accesslog_data "github.com/envoyproxy/go-control-plane/envoy/data/accesslog/v2"
)

// ResponseTrailerFormatter represents a `%TRAILER(X?Y):Z%` command operator.
type ResponseTrailerFormatter struct {
	HeaderFormatter
}

func (f *ResponseTrailerFormatter) FormatHttpLogEntry(entry *accesslog_data.HTTPAccessLogEntry) (string, error) {
	return f.Format(HeaderMap(entry.GetResponse().GetResponseTrailers()))
}

func (f *ResponseTrailerFormatter) FormatTcpLogEntry(entry *accesslog_data.TCPAccessLogEntry) (string, error) {
	return "", nil
}

func (f *ResponseTrailerFormatter) ConfigureHttpLog(config *accesslog_config.HttpGrpcAccessLogConfig) error {
	config.AdditionalResponseTrailersToLog = f.AppendTo(config.AdditionalResponseTrailersToLog)
	return nil
}

// String returns the canonical representation of this command operator.
func (f *ResponseTrailerFormatter) String() string {
	return fmt.Sprintf("%%TRAILER%s%%", f.HeaderFormatter.String())
}
