package accesslog

import (
	"fmt"

	accesslog_config "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v2"
	accesslog_data "github.com/envoyproxy/go-control-plane/envoy/data/accesslog/v2"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
)

const (
	HeaderMethod             = ":method"
	HeaderScheme             = ":scheme"
	HeaderAuthority          = ":authority"
	HeaderPath               = ":path"
	HeaderUserAgent          = "user-agent"
	HeaderReferer            = "referer"
	HeaderXForwardedFor      = "x-forwarded-for"
	HeaderXRequestID         = "x-request-id"
	HeaderXEnvoyOriginalPath = "x-envoy-original-path"
)

var (
	isAlwaysCapturedRequestHeader = func(header string) bool {
		switch header {
		case HeaderMethod,
			HeaderScheme,
			HeaderAuthority,
			HeaderPath,
			HeaderUserAgent,
			HeaderReferer,
			HeaderXForwardedFor,
			HeaderXRequestID,
			HeaderXEnvoyOriginalPath:
			return true
		default:
			return false
		}
	}
	isNotCapturedByDefaultRequestHeader = func(header string) bool {
		return !isAlwaysCapturedRequestHeader(header)
	}
)

// RequestHeaderOperator represents a `%REQ(X?Y):Z%` command operator.
type RequestHeaderOperator struct {
	HeaderFormatter
}

func (f *RequestHeaderOperator) FormatHttpLogEntry(entry *accesslog_data.HTTPAccessLogEntry) (string, error) {
	return f.Format(&RequestHeaders{entry.GetRequest()})
}

func (f *RequestHeaderOperator) FormatTcpLogEntry(entry *accesslog_data.TCPAccessLogEntry) (string, error) {
	return "", nil
}

func (f *RequestHeaderOperator) ConfigureHttpLog(config *accesslog_config.HttpGrpcAccessLogConfig) error {
	config.AdditionalRequestHeadersToLog = stringList(f.GetOperandHeaders()).
		Filter(isNotCapturedByDefaultRequestHeader).
		AppendToSet(config.AdditionalRequestHeadersToLog)
	return nil
}

func (f *RequestHeaderOperator) ConfigureTcpLog(config *accesslog_config.TcpGrpcAccessLogConfig) error {
	// has no effect on TcpGrpcAccessLogConfig
	return nil
}

// String returns the canonical representation of this access log fragment.
func (f *RequestHeaderOperator) String() string {
	return fmt.Sprintf("%%REQ%s%%", f.HeaderFormatter.String())
}

// RequestHeaders represents a set of HTTP request headers
// that includes both regular headers, such as `referer` and `user-agent`,
// and pseudo headers, such as `:method`, `:authority` and `:path`.
type RequestHeaders struct {
	*accesslog_data.HTTPRequestProperties
}

func (h *RequestHeaders) Get(name string) (string, bool) {
	switch name {
	case HeaderMethod:
		return h.formatHttpMethod(h.GetRequestMethod())
	case HeaderScheme:
		return h.optionalString(h.GetScheme())
	case HeaderAuthority:
		return h.optionalString(h.GetAuthority())
	case HeaderPath:
		return h.optionalString(h.GetPath())
	case HeaderUserAgent:
		return h.optionalString(h.GetUserAgent())
	case HeaderReferer:
		return h.optionalString(h.GetReferer())
	case HeaderXForwardedFor:
		return h.optionalString(h.GetForwardedFor())
	case HeaderXRequestID:
		return h.optionalString(h.GetRequestId())
	case HeaderXEnvoyOriginalPath:
		return h.optionalString(h.GetOriginalPath())
	default:
		value, exists := h.GetRequestHeaders()[name]
		return value, exists
	}
}

func (h *RequestHeaders) formatHttpMethod(method envoy_core.RequestMethod) (string, bool) {
	switch method {
	case envoy_core.RequestMethod_METHOD_UNSPECIFIED:
		return "", false
	default:
		return method.String(), true
	}
}

func (h *RequestHeaders) optionalString(value string) (string, bool) {
	return value, value != ""
}
