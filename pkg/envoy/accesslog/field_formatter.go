package accesslog

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/wrappers"

	accesslog_data "github.com/envoyproxy/go-control-plane/envoy/data/accesslog/v2"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
)

const (
	includePort = true
	excludePort = false
)

type FieldFormatter string

func (f FieldFormatter) FormatHttpLogEntry(entry *accesslog_data.HTTPAccessLogEntry) (string, error) {
	switch f {
	case "BYTES_RECEIVED":
		return f.formatUint(entry.GetRequest().GetRequestBodyBytes())
	case "BYTES_SENT":
		return f.formatUint(entry.GetResponse().GetResponseBodyBytes())
	case "PROTOCOL":
		return f.formatHttpVersion(entry.GetProtocolVersion())
	case "RESPONSE_CODE":
		return f.formatUint(uint64(entry.GetResponse().GetResponseCode().GetValue()))
	case "RESPONSE_CODE_DETAILS":
		return entry.GetResponse().GetResponseCodeDetails(), nil
	default:
		return f.formatAccessLogCommon(entry.GetCommonProperties())
	}
}

func (f FieldFormatter) FormatTcpLogEntry(entry *accesslog_data.TCPAccessLogEntry) (string, error) {
	switch f {
	case "BYTES_RECEIVED":
		return f.formatUint(entry.GetConnectionProperties().GetReceivedBytes())
	case "BYTES_SENT":
		return f.formatUint(entry.GetConnectionProperties().GetSentBytes())
	case "PROTOCOL":
		return "", nil
	case "RESPONSE_CODE":
		return "0", nil
	case "RESPONSE_CODE_DETAILS":
		return "", nil
	default:
		return f.formatAccessLogCommon(entry.GetCommonProperties())
	}
}

func (f FieldFormatter) formatAccessLogCommon(entry *accesslog_data.AccessLogCommon) (string, error) {
	switch f {
	case "UPSTREAM_TRANSPORT_FAILURE_REASON":
		return entry.GetUpstreamTransportFailureReason(), nil
	case "REQUEST_DURATION":
		return f.formatDuration(entry.GetTimeToLastRxByte())
	case "RESPONSE_DURATION":
		return f.formatDuration(entry.GetTimeToFirstUpstreamRxByte())
	case "RESPONSE_TX_DURATION":
		return f.formatDurationDelta(entry.GetTimeToLastDownstreamTxByte(), entry.GetTimeToFirstUpstreamRxByte())
	case "DURATION":
		return f.formatDuration(entry.GetTimeToLastDownstreamTxByte())
	case "RESPONSE_FLAGS":
		return f.formatResponseFlags(entry.GetResponseFlags())
	case "UPSTREAM_HOST":
		return f.formatAddress(entry.GetUpstreamRemoteAddress(), includePort)
	case "UPSTREAM_CLUSTER":
		return entry.GetUpstreamCluster(), nil
	case "UPSTREAM_LOCAL_ADDRESS":
		return f.formatAddress(entry.GetUpstreamLocalAddress(), includePort)
	case "DOWNSTREAM_LOCAL_ADDRESS":
		return f.formatAddress(entry.GetDownstreamLocalAddress(), includePort)
	case "DOWNSTREAM_LOCAL_ADDRESS_WITHOUT_PORT":
		return f.formatAddress(entry.GetDownstreamLocalAddress(), excludePort)
	case "DOWNSTREAM_REMOTE_ADDRESS":
		return f.formatAddress(entry.GetDownstreamRemoteAddress(), includePort)
	case "DOWNSTREAM_REMOTE_ADDRESS_WITHOUT_PORT":
		return f.formatAddress(entry.GetDownstreamRemoteAddress(), excludePort)
	case "DOWNSTREAM_DIRECT_REMOTE_ADDRESS":
		return f.formatAddress(entry.GetDownstreamDirectRemoteAddress(), includePort)
	case "DOWNSTREAM_DIRECT_REMOTE_ADDRESS_WITHOUT_PORT":
		return f.formatAddress(entry.GetDownstreamDirectRemoteAddress(), excludePort)
	case "REQUESTED_SERVER_NAME":
		return entry.GetTlsProperties().GetTlsSniHostname(), nil
	case "ROUTE_NAME":
		return entry.GetRouteName(), nil
	case "DOWNSTREAM_PEER_URI_SAN":
		return f.formatUriSans(entry.GetTlsProperties().GetPeerCertificateProperties().GetSubjectAltName())
	case "DOWNSTREAM_LOCAL_URI_SAN":
		return f.formatUriSans(entry.GetTlsProperties().GetLocalCertificateProperties().GetSubjectAltName())
	case "DOWNSTREAM_PEER_SUBJECT":
		return entry.GetTlsProperties().GetPeerCertificateProperties().GetSubject(), nil
	case "DOWNSTREAM_LOCAL_SUBJECT":
		return entry.GetTlsProperties().GetLocalCertificateProperties().GetSubject(), nil
	case "DOWNSTREAM_TLS_SESSION_ID":
		return entry.GetTlsProperties().GetTlsSessionId(), nil
	case "DOWNSTREAM_TLS_CIPHER":
		return f.formatTlsCipherSuite(entry.GetTlsProperties().GetTlsCipherSuite())
	case "DOWNSTREAM_TLS_VERSION":
		return f.formatTlsVersion(entry.GetTlsProperties().GetTlsVersion())
	case "DOWNSTREAM_PEER_FINGERPRINT_256",
		"DOWNSTREAM_PEER_SERIAL",
		"DOWNSTREAM_PEER_ISSUER",
		"DOWNSTREAM_PEER_CERT",
		"DOWNSTREAM_PEER_CERT_V_START",
		"DOWNSTREAM_PEER_CERT_V_END",
		"HOSTNAME":
		fallthrough // these fields have no equivalents in GrpcAccessLog
	default:
		// make it clear to the user what is happening
		return fmt.Sprintf("UNSUPPORTED_FIELD(%s)", f), nil
	}
}

func (f FieldFormatter) formatUint(value uint64) (string, error) {
	return strconv.FormatUint(value, 10), nil
}

func (f FieldFormatter) formatInt(value int64) (string, error) {
	return strconv.FormatInt(value, 10), nil
}

func (f FieldFormatter) formatDuration(dur *duration.Duration) (string, error) {
	if dur == nil {
		return "", nil
	}
	durNanos, err := ptypes.Duration(dur)
	if err != nil {
		return "", err
	}
	return f.formatInt(int64(durNanos / time.Millisecond))
}

func (f FieldFormatter) formatDurationDelta(outer *duration.Duration, inner *duration.Duration) (string, error) {
	if outer == nil || inner == nil {
		return "", nil
	}
	outerNanos, err := ptypes.Duration(outer)
	if err != nil {
		return "", err
	}
	innerNanos, err := ptypes.Duration(inner)
	if err != nil {
		return "", err
	}
	return f.formatInt(int64((outerNanos - innerNanos) / time.Millisecond))
}

func (f FieldFormatter) formatHttpVersion(value accesslog_data.HTTPAccessLogEntry_HTTPVersion) (string, error) {
	switch value {
	case accesslog_data.HTTPAccessLogEntry_PROTOCOL_UNSPECIFIED:
		return "", nil
	case accesslog_data.HTTPAccessLogEntry_HTTP10:
		return "HTTP/1.0", nil
	case accesslog_data.HTTPAccessLogEntry_HTTP11:
		return "HTTP/1.1", nil
	case accesslog_data.HTTPAccessLogEntry_HTTP2:
		return "HTTP/2", nil
	case accesslog_data.HTTPAccessLogEntry_HTTP3:
		return "HTTP/3", nil
	default:
		return value.String(), nil
	}
}

func (f FieldFormatter) formatResponseFlags(flags *accesslog_data.ResponseFlags) (string, error) {
	values := make([]string, 0)
	if flags.GetFailedLocalHealthcheck() {
		values = append(values, ResponseFlagFailedLocalHealthCheck)
	}
	if flags.GetNoHealthyUpstream() {
		values = append(values, ResponseFlagNoHealthyUpstream)
	}
	if flags.GetUpstreamRequestTimeout() {
		values = append(values, ResponseFlagUpstreamRequestTimeout)
	}
	if flags.GetLocalReset() {
		values = append(values, ResponseFlagLocalReset)
	}
	if flags.GetUpstreamRemoteReset() {
		values = append(values, ResponseFlagUpstreamRemoteReset)
	}
	if flags.GetUpstreamConnectionFailure() {
		values = append(values, ResponseFlagUpstreamConnectionFailure)
	}
	if flags.GetUpstreamConnectionTermination() {
		values = append(values, ResponseFlagUpstreamConnectionTermination)
	}
	if flags.GetUpstreamOverflow() {
		values = append(values, ResponseFlagUpstreamOverflow)
	}
	if flags.GetNoRouteFound() {
		values = append(values, ResponseFlagNoRouteFound)
	}
	if flags.GetDelayInjected() {
		values = append(values, ResponseFlagDelayInjected)
	}
	if flags.GetFaultInjected() {
		values = append(values, ResponseFlagFaultInjected)
	}
	if flags.GetRateLimited() {
		values = append(values, ResponseFlagRateLimited)
	}
	if flags.GetUnauthorizedDetails().GetReason() == accesslog_data.ResponseFlags_Unauthorized_EXTERNAL_SERVICE {
		values = append(values, ResponseFlagUnauthorizedExternalService)
	}
	if flags.GetRateLimitServiceError() {
		values = append(values, ResponseFlagRatelimitServiceError)
	}
	if flags.GetDownstreamConnectionTermination() {
		values = append(values, ResponseFlagDownstreamConnectionTermination)
	}
	if flags.GetUpstreamRetryLimitExceeded() {
		values = append(values, ResponseFlagUpstreamRetryLimitExceeded)
	}
	if flags.GetStreamIdleTimeout() {
		values = append(values, ResponseFlagStreamIdleTimeout)
	}
	if flags.GetInvalidEnvoyRequestHeaders() {
		values = append(values, ResponseFlagInvalidEnvoyRequestHeaders)
	}
	if flags.GetDownstreamProtocolError() {
		values = append(values, ResponseFlagDownstreamProtocolError)
	}
	return strings.Join(values, ","), nil
}

func (f FieldFormatter) formatAddress(address *envoy_core.Address, includePort bool) (string, error) {
	switch typ := address.GetAddress().(type) {
	case *envoy_core.Address_SocketAddress:
		if includePort {
			return net.JoinHostPort(typ.SocketAddress.GetAddress(), strconv.FormatUint(uint64(typ.SocketAddress.GetPortValue()), 10)), nil
		}
		return typ.SocketAddress.GetAddress(), nil
	case *envoy_core.Address_Pipe:
		return typ.Pipe.GetPath(), nil
	default:
		return "", nil
	}
}

func (f FieldFormatter) formatUriSans(sans []*accesslog_data.TLSProperties_CertificateProperties_SubjectAltName) (string, error) {
	values := make([]string, 0, len(sans))
	for _, san := range sans {
		switch typ := san.GetSan().(type) {
		case *accesslog_data.TLSProperties_CertificateProperties_SubjectAltName_Uri:
			values = append(values, typ.Uri)
		}
	}
	return strings.Join(values, ","), nil
}

func (f FieldFormatter) formatTlsCipherSuite(value *wrappers.UInt32Value) (string, error) {
	if value == nil || value.GetValue() == 0xFFFF {
		return "", nil
	}
	return TlsCipherSuite(value.GetValue()).String(), nil
}

func (f FieldFormatter) formatTlsVersion(value accesslog_data.TLSProperties_TLSVersion) (string, error) {
	switch value {
	case accesslog_data.TLSProperties_VERSION_UNSPECIFIED:
		return "", nil
	case accesslog_data.TLSProperties_TLSv1:
		return "TLSv1", nil
	case accesslog_data.TLSProperties_TLSv1_1:
		return "TLSv1.1", nil
	case accesslog_data.TLSProperties_TLSv1_2:
		return "TLSv1.2", nil
	case accesslog_data.TLSProperties_TLSv1_3:
		return "TLSv1.3", nil
	default:
		return value.String(), nil
	}
}
