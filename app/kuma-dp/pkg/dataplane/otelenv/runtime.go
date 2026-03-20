package otelenv

import (
	"cmp"
	"maps"
	"net"
	"net/url"
	"os"
	"path"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"

	motb_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshopentelemetrybackend/api/v1alpha1"
	core_xds "github.com/kumahq/kuma/v2/pkg/core/xds"
)

type SignalRuntime struct {
	Enabled        bool
	BlockedReasons []string
	HTTPPath       string
	Transport      ExporterTransport
}

type BackendRuntime struct {
	Traces  SignalRuntime
	Logs    SignalRuntime
	Metrics SignalRuntime
}

type ExporterTransport struct {
	Protocol          core_xds.OtelProtocol
	Endpoint          string
	UseTLS            *bool
	Headers           map[string]string
	Compression       string
	Timeout           time.Duration
	Certificate       string
	ClientCertificate string
	ClientKey         string // #nosec G117 -- OTLP mTLS config field, not a hardcoded key
}

type exporterOverride struct {
	Endpoint          *string
	UseTLS            *bool
	HTTPPath          *string
	Headers           map[string]string
	HeadersPresent    bool
	Compression       *string
	Timeout           *time.Duration
	Certificate       *string
	ClientCertificate *string
	ClientKey         *string // #nosec G117 -- OTLP mTLS config field, not a hardcoded key
}

func (c Config) ResolveBackend(backend core_xds.OtelPipeBackend) BackendRuntime {
	return BackendRuntime{
		Traces:  c.resolveSignal(backend, core_xds.OtelSignalTraces, backend.Traces),
		Logs:    c.resolveSignal(backend, core_xds.OtelSignalLogs, backend.Logs),
		Metrics: c.resolveSignal(backend, core_xds.OtelSignalMetrics, backend.Metrics),
	}
}

func (c Config) resolveSignal(
	backend core_xds.OtelPipeBackend,
	signal core_xds.OtelSignal,
	plan *core_xds.OtelSignalRuntimePlan,
) SignalRuntime {
	if plan == nil || !plan.Enabled {
		return SignalRuntime{}
	}

	runtime := SignalRuntime{
		Enabled:        true,
		BlockedReasons: slices.Clone(plan.BlockedReasons),
	}

	if isHardBlocked(plan) {
		return runtime
	}

	explicit := explicitTransport(backend, signal)
	sharedAllowed := sharedEnvAllowed(plan)
	signalAllowed := signalEnvAllowed(plan)
	preferEnv := backend.EnvPolicy != nil && backend.EnvPolicy.Precedence != motb_api.EnvPrecedenceExplicitFirst

	protocol := explicit.Transport.Protocol
	if sharedAllowed {
		protocol = pickProtocol(protocol, c.Shared.Protocol, preferEnv)
	}
	sigLayer := layerForSignal(c, signal)
	if signalAllowed {
		protocol = pickProtocol(protocol, sigLayer.Protocol, preferEnv)
	}

	runtime.Transport = ExporterTransport{
		Protocol: protocol,
	}
	runtime.Transport.Endpoint = explicit.Transport.Endpoint
	runtime.Transport.UseTLS = copyBoolPtr(explicit.Transport.UseTLS)
	runtime.Transport.Headers = maps.Clone(explicit.Transport.Headers)
	runtime.Transport.Compression = explicit.Transport.Compression
	runtime.Transport.Timeout = explicit.Transport.Timeout
	runtime.Transport.Certificate = explicit.Transport.Certificate
	runtime.Transport.ClientCertificate = explicit.Transport.ClientCertificate
	runtime.Transport.ClientKey = explicit.Transport.ClientKey
	runtime.HTTPPath = explicit.HTTPPath
	if runtime.Transport.Protocol == core_xds.OtelProtocolHTTPProtobuf && runtime.HTTPPath == "" {
		runtime.HTTPPath = path.Join("/", backend.Path, "v1", string(signal))
	}

	if sharedAllowed {
		applyOverride(&runtime, endpointOverrideForLayer(c.Shared, signal, protocol, false), preferEnv)
		applyOverride(&runtime, transportOverrideForLayer(c.Shared), preferEnv)
	}
	if signalAllowed {
		applyOverride(&runtime, endpointOverrideForLayer(sigLayer, signal, protocol, true), preferEnv)
		applyOverride(&runtime, transportOverrideForLayer(sigLayer), preferEnv)
	}

	return runtime
}

func explicitTransport(backend core_xds.OtelPipeBackend, signal core_xds.OtelSignal) SignalRuntime {
	protocol := core_xds.OtelProtocolGRPC
	if backend.UseHTTP {
		protocol = core_xds.OtelProtocolHTTPProtobuf
	}

	runtime := SignalRuntime{
		HTTPPath: path.Join("/", backend.Path, "v1", string(signal)),
		Transport: ExporterTransport{
			Protocol: protocol,
			Endpoint: resolveEndpointAddress(backend.Endpoint),
			UseTLS:   boolPtr(backend.UseHTTPS),
		},
	}
	if protocol == core_xds.OtelProtocolGRPC {
		runtime.HTTPPath = ""
	}

	return runtime
}

// resolveEndpointAddress fills in the host portion of the endpoint when the CP
// sent an empty host (e.g. ":4317"). Uses HOST_IP env var, falling back to
// 127.0.0.1.
func resolveEndpointAddress(endpoint string) string {
	host, port, err := net.SplitHostPort(endpoint)
	if err != nil || host != "" {
		return endpoint
	}
	hostIP := cmp.Or(os.Getenv("HOST_IP"), "127.0.0.1")
	return net.JoinHostPort(hostIP, port)
}

func pickProtocol(current core_xds.OtelProtocol, field FieldValue, preferEnv bool) core_xds.OtelProtocol {
	if !field.Present {
		return current
	}

	parsed, ok := parseProtocol(field.Value)
	if !ok {
		return current
	}
	if preferEnv || current == "" {
		return parsed
	}
	return current
}

func applyOverride(runtime *SignalRuntime, override exporterOverride, preferEnv bool) {
	if override.Endpoint != nil && (preferEnv || runtime.Transport.Endpoint == "") {
		runtime.Transport.Endpoint = *override.Endpoint
	}
	if override.UseTLS != nil && (preferEnv || runtime.Transport.UseTLS == nil) {
		v := *override.UseTLS
		runtime.Transport.UseTLS = &v
	}
	if override.HTTPPath != nil && (preferEnv || runtime.HTTPPath == "") {
		runtime.HTTPPath = *override.HTTPPath
	}
	if override.HeadersPresent && (preferEnv || len(runtime.Transport.Headers) == 0) {
		runtime.Transport.Headers = maps.Clone(override.Headers)
	}
	if override.Compression != nil && (preferEnv || runtime.Transport.Compression == "") {
		runtime.Transport.Compression = *override.Compression
	}
	if override.Timeout != nil && (preferEnv || runtime.Transport.Timeout == 0) {
		runtime.Transport.Timeout = *override.Timeout
	}
	if override.Certificate != nil && (preferEnv || runtime.Transport.Certificate == "") {
		runtime.Transport.Certificate = *override.Certificate
	}
	if override.ClientCertificate != nil && (preferEnv || runtime.Transport.ClientCertificate == "") {
		runtime.Transport.ClientCertificate = *override.ClientCertificate
	}
	if override.ClientKey != nil && (preferEnv || runtime.Transport.ClientKey == "") {
		runtime.Transport.ClientKey = *override.ClientKey
	}
}

func endpointOverrideForLayer(
	layer Layer,
	signal core_xds.OtelSignal,
	protocol core_xds.OtelProtocol,
	signalSpecific bool,
) exporterOverride {
	if !layer.Endpoint.Present {
		return exporterOverride{}
	}

	value := strings.TrimSpace(layer.Endpoint.Value)
	parsedURL, err := url.Parse(value)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return exporterOverride{
			Endpoint: &value,
		}
	}

	switch protocol {
	case core_xds.OtelProtocolHTTPProtobuf:
		override := exporterOverride{
			Endpoint: &parsedURL.Host,
		}
		if strings.EqualFold(parsedURL.Scheme, "http") || strings.EqualFold(parsedURL.Scheme, "unix") {
			override.UseTLS = boolPtr(false)
		} else {
			override.UseTLS = boolPtr(true)
		}

		pathValue := parsedURL.Path
		if signalSpecific {
			if pathValue == "" {
				pathValue = "/"
			}
			override.HTTPPath = &pathValue
			return override
		}

		if pathValue != "" {
			pathValue = path.Join(pathValue, "v1", string(signal))
			override.HTTPPath = &pathValue
		}
		return override
	default:
		target := path.Join(parsedURL.Host, parsedURL.Path)
		override := exporterOverride{
			Endpoint: &target,
		}
		if strings.EqualFold(parsedURL.Scheme, "http") || strings.EqualFold(parsedURL.Scheme, "unix") {
			override.UseTLS = boolPtr(false)
		} else {
			override.UseTLS = boolPtr(true)
		}
		return override
	}
}

func transportOverrideForLayer(layer Layer) exporterOverride {
	override := exporterOverride{}

	if layer.Insecure.Present {
		override.UseTLS = boolPtr(!strings.EqualFold(strings.TrimSpace(layer.Insecure.Value), "true"))
	}
	if layer.Headers.Present {
		override.Headers = parseHeaders(layer.Headers.Value)
		override.HeadersPresent = true
	}
	if layer.Compression.Present {
		if compression, ok := parseCompression(layer.Compression.Value); ok {
			override.Compression = &compression
		}
	}
	if layer.Timeout.Present {
		if timeout, ok := parseTimeout(layer.Timeout.Value); ok {
			override.Timeout = &timeout
		}
	}
	if layer.Certificate.Present {
		override.Certificate = &layer.Certificate.Value
	}
	if layer.ClientCertificate.Present && layer.ClientKey.Present {
		override.ClientCertificate = &layer.ClientCertificate.Value
		override.ClientKey = &layer.ClientKey.Value
	}

	return override
}

func parseProtocol(value string) (core_xds.OtelProtocol, bool) {
	switch strings.TrimSpace(value) {
	case string(core_xds.OtelProtocolGRPC):
		return core_xds.OtelProtocolGRPC, true
	case string(core_xds.OtelProtocolHTTPProtobuf):
		return core_xds.OtelProtocolHTTPProtobuf, true
	default:
		return "", false
	}
}

func parseCompression(value string) (string, bool) {
	v := strings.ToLower(strings.TrimSpace(value))
	switch v {
	case "gzip":
		return "gzip", true
	case "none", "":
		return "", true
	default:
		return "", false
	}
}

func parseTimeout(value string) (time.Duration, bool) {
	timeoutMS, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || timeoutMS < 0 {
		return 0, false
	}
	return time.Duration(timeoutMS) * time.Millisecond, true
}

func parseHeaders(value string) map[string]string {
	headers := map[string]string{}
	for header := range strings.SplitSeq(value, ",") {
		name, headerValue, ok := strings.Cut(header, "=")
		if !ok {
			continue
		}

		name = strings.TrimSpace(name)
		if !isValidHeaderKey(name) {
			continue
		}

		unescapedValue, err := url.PathUnescape(headerValue)
		if err != nil {
			continue
		}
		headers[name] = strings.TrimSpace(unescapedValue)
	}
	return headers
}

func isHardBlocked(plan *core_xds.OtelSignalRuntimePlan) bool {
	if plan == nil {
		return false
	}
	if len(plan.MissingFields) > 0 {
		return true
	}
	return slices.Contains(plan.BlockedReasons, core_xds.OtelBlockedReasonRequiredEnvMissing)
}

func sharedEnvAllowed(plan *core_xds.OtelSignalRuntimePlan) bool {
	if plan == nil {
		return false
	}
	for _, reason := range plan.BlockedReasons {
		switch reason {
		case core_xds.OtelBlockedReasonEnvDisabledByPolicy,
			core_xds.OtelBlockedReasonMultipleBackends:
			return false
		}
	}
	return true
}

func signalEnvAllowed(plan *core_xds.OtelSignalRuntimePlan) bool {
	if !sharedEnvAllowed(plan) || plan == nil {
		return false
	}
	return !slices.Contains(plan.BlockedReasons, core_xds.OtelBlockedReasonSignalOverridesBlocked)
}

func layerForSignal(cfg Config, signal core_xds.OtelSignal) Layer {
	switch signal {
	case core_xds.OtelSignalTraces:
		return cfg.Traces
	case core_xds.OtelSignalLogs:
		return cfg.Logs
	case core_xds.OtelSignalMetrics:
		return cfg.Metrics
	default:
		return Layer{}
	}
}

func boolPtr(value bool) *bool {
	return &value
}

func copyBoolPtr(p *bool) *bool {
	if p == nil {
		return nil
	}
	v := *p
	return &v
}

func isValidHeaderKey(key string) bool {
	if key == "" {
		return false
	}
	for _, c := range key {
		if !isTokenChar(c) {
			return false
		}
	}
	return true
}

func isTokenChar(c rune) bool {
	return c <= unicode.MaxASCII && (unicode.IsLetter(c) ||
		unicode.IsDigit(c) ||
		c == '!' || c == '#' || c == '$' || c == '%' || c == '&' || c == '\'' || c == '*' ||
		c == '+' || c == '-' || c == '.' || c == '^' || c == '_' || c == '`' || c == '|' || c == '~')
}
