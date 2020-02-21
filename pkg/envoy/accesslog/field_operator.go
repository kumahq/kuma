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

	accesslog_config "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v2"
	accesslog_data "github.com/envoyproxy/go-control-plane/envoy/data/accesslog/v2"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
)

const (
	includePort = true
	excludePort = false
)

// FieldOperator represents a simple field command operator,
// such as `%BYTES_RECEIVED%` or `%PROTOCOL%`.
type FieldOperator string

// String returns the canonical representation of this access log fragment.
func (f FieldOperator) String() string {
	return CommandOperatorDescriptor(f).String()
}

func (f FieldOperator) ConfigureHttpLog(config *accesslog_config.HttpGrpcAccessLogConfig) error {
	// has no effect on HttpGrpcAccessLogConfig
	return nil
}

func (f FieldOperator) ConfigureTcpLog(config *accesslog_config.TcpGrpcAccessLogConfig) error {
	// has no effect on TcpGrpcAccessLogConfig
	return nil
}

func (f FieldOperator) FormatHttpLogEntry(entry *accesslog_data.HTTPAccessLogEntry) (string, error) {
	switch f {
	case CMD_BYTES_RECEIVED:
		return f.formatUint(entry.GetRequest().GetRequestBodyBytes())
	case CMD_BYTES_SENT:
		return f.formatUint(entry.GetResponse().GetResponseBodyBytes())
	case CMD_PROTOCOL:
		return f.formatHttpVersion(entry.GetProtocolVersion())
	case CMD_RESPONSE_CODE:
		return f.formatUint(uint64(entry.GetResponse().GetResponseCode().GetValue()))
	case CMD_RESPONSE_CODE_DETAILS:
		return entry.GetResponse().GetResponseCodeDetails(), nil
	case CMD_REQUEST_DURATION:
		return f.formatDuration(entry.GetCommonProperties().GetTimeToLastRxByte())
	case CMD_RESPONSE_DURATION:
		return f.formatDuration(entry.GetCommonProperties().GetTimeToFirstUpstreamRxByte())
	case CMD_RESPONSE_TX_DURATION:
		return f.formatDurationDelta(entry.GetCommonProperties().GetTimeToLastDownstreamTxByte(), entry.GetCommonProperties().GetTimeToFirstUpstreamRxByte())
	default:
		return f.formatAccessLogCommon(entry.GetCommonProperties())
	}
}

func (f FieldOperator) FormatTcpLogEntry(entry *accesslog_data.TCPAccessLogEntry) (string, error) {
	switch f {
	case CMD_BYTES_RECEIVED:
		return f.formatUint(entry.GetConnectionProperties().GetReceivedBytes())
	case CMD_BYTES_SENT:
		return f.formatUint(entry.GetConnectionProperties().GetSentBytes())
	case CMD_PROTOCOL:
		return "", nil // replicate Envoy's behaviour
	case CMD_RESPONSE_CODE:
		return "0", nil // replicate Envoy's behaviour
	case CMD_RESPONSE_CODE_DETAILS:
		return "", nil // replicate Envoy's behaviour
	case CMD_REQUEST_DURATION:
		return "", nil // replicate Envoy's behaviour
	case CMD_RESPONSE_DURATION:
		return "", nil // replicate Envoy's behaviour
	case CMD_RESPONSE_TX_DURATION:
		return "", nil // replicate Envoy's behaviour
	default:
		return f.formatAccessLogCommon(entry.GetCommonProperties())
	}
}

func (f FieldOperator) formatAccessLogCommon(entry *accesslog_data.AccessLogCommon) (string, error) {
	switch f {
	case CMD_UPSTREAM_TRANSPORT_FAILURE_REASON:
		return entry.GetUpstreamTransportFailureReason(), nil
	case CMD_DURATION:
		return f.formatDuration(entry.GetTimeToLastDownstreamTxByte())
	case CMD_RESPONSE_FLAGS:
		return f.formatResponseFlags(entry.GetResponseFlags())
	case CMD_UPSTREAM_HOST:
		return f.formatAddress(entry.GetUpstreamRemoteAddress(), includePort)
	case CMD_UPSTREAM_CLUSTER:
		return entry.GetUpstreamCluster(), nil
	case CMD_UPSTREAM_LOCAL_ADDRESS:
		return f.formatAddress(entry.GetUpstreamLocalAddress(), includePort)
	case CMD_DOWNSTREAM_LOCAL_ADDRESS:
		return f.formatAddress(entry.GetDownstreamLocalAddress(), includePort)
	case CMD_DOWNSTREAM_LOCAL_ADDRESS_WITHOUT_PORT:
		return f.formatAddress(entry.GetDownstreamLocalAddress(), excludePort)
	case CMD_DOWNSTREAM_REMOTE_ADDRESS:
		return f.formatAddress(entry.GetDownstreamRemoteAddress(), includePort)
	case CMD_DOWNSTREAM_REMOTE_ADDRESS_WITHOUT_PORT:
		return f.formatAddress(entry.GetDownstreamRemoteAddress(), excludePort)
	case CMD_DOWNSTREAM_DIRECT_REMOTE_ADDRESS:
		return f.formatAddress(entry.GetDownstreamDirectRemoteAddress(), includePort)
	case CMD_DOWNSTREAM_DIRECT_REMOTE_ADDRESS_WITHOUT_PORT:
		return f.formatAddress(entry.GetDownstreamDirectRemoteAddress(), excludePort)
	case CMD_REQUESTED_SERVER_NAME:
		return entry.GetTlsProperties().GetTlsSniHostname(), nil
	case CMD_ROUTE_NAME:
		return entry.GetRouteName(), nil
	case CMD_DOWNSTREAM_PEER_URI_SAN:
		return f.formatUriSans(entry.GetTlsProperties().GetPeerCertificateProperties().GetSubjectAltName())
	case CMD_DOWNSTREAM_LOCAL_URI_SAN:
		return f.formatUriSans(entry.GetTlsProperties().GetLocalCertificateProperties().GetSubjectAltName())
	case CMD_DOWNSTREAM_PEER_SUBJECT:
		return entry.GetTlsProperties().GetPeerCertificateProperties().GetSubject(), nil
	case CMD_DOWNSTREAM_LOCAL_SUBJECT:
		return entry.GetTlsProperties().GetLocalCertificateProperties().GetSubject(), nil
	case CMD_DOWNSTREAM_TLS_SESSION_ID:
		return entry.GetTlsProperties().GetTlsSessionId(), nil
	case CMD_DOWNSTREAM_TLS_CIPHER:
		return f.formatTlsCipherSuite(entry.GetTlsProperties().GetTlsCipherSuite())
	case CMD_DOWNSTREAM_TLS_VERSION:
		return f.formatTlsVersion(entry.GetTlsProperties().GetTlsVersion())
	case CMD_DOWNSTREAM_PEER_FINGERPRINT_256,
		CMD_DOWNSTREAM_PEER_SERIAL,
		CMD_DOWNSTREAM_PEER_ISSUER,
		CMD_DOWNSTREAM_PEER_CERT,
		CMD_DOWNSTREAM_PEER_CERT_V_START,
		CMD_DOWNSTREAM_PEER_CERT_V_END,
		CMD_HOSTNAME:
		fallthrough // these fields have no equivalent data in GrpcAccessLog
	default:
		// make it clear to the user what is happening
		return fmt.Sprintf("UNSUPPORTED_COMMAND(%s)", f), nil
	}
}

func (f FieldOperator) formatUint(value uint64) (string, error) {
	return strconv.FormatUint(value, 10), nil
}

func (f FieldOperator) formatInt(value int64) (string, error) {
	return strconv.FormatInt(value, 10), nil
}

func (f FieldOperator) formatDuration(dur *duration.Duration) (string, error) {
	if dur == nil {
		return "", nil
	}
	durNanos, err := ptypes.Duration(dur)
	if err != nil {
		return "", err
	}
	return f.formatInt(int64(durNanos / time.Millisecond))
}

func (f FieldOperator) formatDurationDelta(outer *duration.Duration, inner *duration.Duration) (string, error) {
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

func (f FieldOperator) formatHttpVersion(value accesslog_data.HTTPAccessLogEntry_HTTPVersion) (string, error) {
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

func (f FieldOperator) formatResponseFlags(flags *accesslog_data.ResponseFlags) (string, error) {
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

func (f FieldOperator) formatAddress(address *envoy_core.Address, includePort bool) (string, error) {
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

func (f FieldOperator) formatUriSans(sans []*accesslog_data.TLSProperties_CertificateProperties_SubjectAltName) (string, error) {
	values := make([]string, 0, len(sans))
	for _, san := range sans {
		switch typ := san.GetSan().(type) {
		case *accesslog_data.TLSProperties_CertificateProperties_SubjectAltName_Uri:
			values = append(values, typ.Uri)
		}
	}
	return strings.Join(values, ","), nil
}

func (f FieldOperator) formatTlsCipherSuite(value *wrappers.UInt32Value) (string, error) {
	if value == nil || value.GetValue() == 0xFFFF {
		return "", nil
	}
	return TlsCipherSuite(value.GetValue()).String(), nil
}

func (f FieldOperator) formatTlsVersion(value accesslog_data.TLSProperties_TLSVersion) (string, error) {
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
