package accesslog

import (
	accesslog_config "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v2"
	accesslog_data "github.com/envoyproxy/go-control-plane/envoy/data/accesslog/v2"
)

// Placeholder represents a placeholder added to an access log format string
// that must be resolved before configuring Envoy with that format string.
//
// E.g. %KUMA_SOURCE_SERVICE%, %KUMA_DESTINATION_SERVICE%,
// %KUMA_SOURCE_ADDRESS% and %KUMA_SOURCE_ADDRESS_WITHOUT_PORT%
// are examples of such placeholders.
type Placeholder string

func (f Placeholder) FormatHttpLogEntry(entry *accesslog_data.HTTPAccessLogEntry) (string, error) {
	return f.String(), nil
}

func (f Placeholder) FormatTcpLogEntry(entry *accesslog_data.TCPAccessLogEntry) (string, error) {
	return f.String(), nil
}

func (f Placeholder) ConfigureHttpLog(config *accesslog_config.HttpGrpcAccessLogConfig) error {
	// has no effect on HttpGrpcAccessLogConfig
	return nil
}

func (f Placeholder) ConfigureTcpLog(config *accesslog_config.TcpGrpcAccessLogConfig) error {
	// has no effect on TcpGrpcAccessLogConfig
	return nil
}

// String returns the canonical representation of this command operator.
func (f Placeholder) String() string {
	return CommandOperatorDescriptor(f).String()
}

// Interpolate returns an access log fragment with all placeholders resolved.
func (f Placeholder) Interpolate(variables InterpolationVariables) (AccessLogFragment, error) {
	value := variables.Get(string(f))
	return TextSpan(value), nil // turn placeholder into a text literal
}
