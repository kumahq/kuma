package accesslog

import (
	accesslog_config "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v2"
	accesslog_data "github.com/envoyproxy/go-control-plane/envoy/data/accesslog/v2"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
)

type RequestHeaderFormatter struct {
	HeaderFormatter
}

func (f *RequestHeaderFormatter) FormatHttpLogEntry(entry *accesslog_data.HTTPAccessLogEntry) (string, error) {
	return f.Format(&RequestHeaders{entry.GetRequest()})
}

func (f *RequestHeaderFormatter) FormatTcpLogEntry(entry *accesslog_data.TCPAccessLogEntry) (string, error) {
	return "", nil
}

func (f *RequestHeaderFormatter) ConfigureHttpLog(config *accesslog_config.HttpGrpcAccessLogConfig) error {
	config.AdditionalRequestHeadersToLog = f.AppendTo(config.AdditionalRequestHeadersToLog)
	return nil
}

type RequestHeaders struct {
	*accesslog_data.HTTPRequestProperties
}

func (h *RequestHeaders) Get(name string) (string, bool) {
	switch name {
	case ":METHOD":
		return h.formatHttpMethod(h.GetRequestMethod())
	case ":SCHEME":
		return h.optionalString(h.GetScheme())
	case ":AUTHORITY":
		return h.optionalString(h.GetAuthority())
	case ":PATH":
		return h.optionalString(h.GetPath())
	case "USER-AGENT":
		return h.optionalString(h.GetUserAgent())
	case "REFERER":
		return h.optionalString(h.GetReferer())
	case "X-FORWARDED-FOR":
		return h.optionalString(h.GetForwardedFor())
	case "X-REQUEST-ID":
		return h.optionalString(h.GetRequestId())
	case "X-ENVOY-ORIGINAL-PATH":
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
