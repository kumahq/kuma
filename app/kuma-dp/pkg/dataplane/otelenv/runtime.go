package otelenv

import (
	"cmp"
	"net"
	"net/url"
	"os"
	"path"
	"slices"
	"strconv"
	"strings"
	"time"

	motb_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshopentelemetrybackend/api/v1alpha1"
	core_xds "github.com/kumahq/kuma/v2/pkg/core/xds"
)

// runtimeOption applies a configuration value to a SignalRuntime.
// Options are collected in precedence order and applied sequentially -
// last writer wins, matching the OTel SDK's option pattern.
type runtimeOption func(*SignalRuntime)

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

	if plan.IsHardBlocked() {
		return runtime
	}

	preferEnv := backend.EnvPolicy != nil &&
		backend.EnvPolicy.Precedence != motb_api.EnvPrecedenceExplicitFirst

	type source struct{ protocol, fields []runtimeOption }

	explicit := source{
		protocol: []runtimeOption{withProtocol(explicitProtocol(backend))},
		fields:   explicitFieldOptions(backend, signal),
	}

	var env []source
	if plan.SharedEnvAllowed() {
		p, f := c.Shared.runtimeOptions(signal, false)
		env = append(env, source{p, f})
	}

	if plan.SignalEnvAllowed() {
		p, f := layerForSignal(c, signal).runtimeOptions(signal, true)
		env = append(env, source{p, f})
	}

	// Lowest precedence first - last applied wins.
	var sources []source
	if preferEnv {
		sources = slices.Concat([]source{explicit}, env)
	} else {
		sources = append(env, explicit)
	}

	// Phase 1: protocol (endpoint path handling depends on it).
	for src := range slices.Values(sources) {
		for opt := range slices.Values(src.protocol) {
			opt(&runtime)
		}
	}

	// Phase 2: all other fields.
	for src := range slices.Values(sources) {
		for opt := range slices.Values(src.fields) {
			opt(&runtime)
		}
	}

	if runtime.Transport.Protocol == core_xds.OtelProtocolHTTPProtobuf && runtime.HTTPPath == "" {
		runtime.HTTPPath = path.Join("/", backend.Path, "v1", string(signal))
	}

	return runtime
}

// --- option constructors ---

func withProtocol(p core_xds.OtelProtocol) runtimeOption {
	return func(r *SignalRuntime) { r.Transport.Protocol = p }
}

func withEndpoint(endpoint string) runtimeOption {
	return func(r *SignalRuntime) { r.Transport.Endpoint = endpoint }
}

func withUseTLS(useTLS bool) runtimeOption {
	return func(r *SignalRuntime) { r.Transport.UseTLS = new(useTLS) }
}

func withHTTPPath(p string) runtimeOption {
	return func(r *SignalRuntime) { r.HTTPPath = p }
}

func withHeaders(headers map[string]string) runtimeOption {
	return func(r *SignalRuntime) { r.Transport.Headers = headers }
}

func withCompression(compression string) runtimeOption {
	return func(r *SignalRuntime) { r.Transport.Compression = compression }
}

func withTimeout(timeout time.Duration) runtimeOption {
	return func(r *SignalRuntime) { r.Transport.Timeout = timeout }
}

func withCertificate(cert string) runtimeOption {
	return func(r *SignalRuntime) { r.Transport.Certificate = cert }
}

func withClientCerts(cert, key string) runtimeOption {
	return func(r *SignalRuntime) {
		r.Transport.ClientCertificate = cert
		r.Transport.ClientKey = key
	}
}

// --- option sources ---

func explicitProtocol(backend core_xds.OtelPipeBackend) core_xds.OtelProtocol {
	if backend.UseHTTP {
		return core_xds.OtelProtocolHTTPProtobuf
	}
	return core_xds.OtelProtocolGRPC
}

func explicitFieldOptions(backend core_xds.OtelPipeBackend, signal core_xds.OtelSignal) []runtimeOption {
	opts := []runtimeOption{
		withEndpoint(resolveEndpointAddress(backend.Endpoint)),
		withUseTLS(backend.UseHTTPS),
	}

	if backend.UseHTTP {
		opts = append(opts, withHTTPPath(path.Join("/", backend.Path, "v1", string(signal))))
	}

	return opts
}

// runtimeOptions produces options by parsing the layer's raw env var values.
// Protocol options are returned separately because endpoint resolution
// depends on the fully-resolved protocol.
func (layer Layer) runtimeOptions(signal core_xds.OtelSignal, signalSpecific bool) ([]runtimeOption, []runtimeOption) {
	var protocolOpts, fieldOpts []runtimeOption

	if layer.Protocol != nil {
		if p, ok := parseProtocol(*layer.Protocol); ok {
			protocolOpts = append(protocolOpts, withProtocol(p))
		}
	}

	fieldOpts = append(fieldOpts, layer.endpointAndTLSOption(signal, signalSpecific))

	if layer.Headers != nil {
		if headers := parseHeaders(*layer.Headers); len(headers) > 0 {
			fieldOpts = append(fieldOpts, withHeaders(headers))
		}
	}

	if layer.Compression != nil {
		if c, ok := parseCompression(*layer.Compression); ok {
			fieldOpts = append(fieldOpts, withCompression(c))
		}
	}

	if layer.Timeout != nil {
		if t, ok := parseTimeout(*layer.Timeout); ok {
			fieldOpts = append(fieldOpts, withTimeout(t))
		}
	}

	if layer.Certificate != nil {
		fieldOpts = append(fieldOpts, withCertificate(*layer.Certificate))
	}

	if layer.ClientCertificate != nil && layer.ClientKey != nil {
		fieldOpts = append(fieldOpts, withClientCerts(*layer.ClientCertificate, *layer.ClientKey))
	}

	return protocolOpts, fieldOpts
}

// endpointAndTLSOption returns a single option that handles both endpoint
// resolution and TLS. These interact because URL scheme-derived TLS
// suppresses the Insecure env var within the same layer.
func (layer Layer) endpointAndTLSOption(signal core_xds.OtelSignal, signalSpecific bool) runtimeOption {
	return func(r *SignalRuntime) {
		tlsFromScheme := false
		if ep, ok := layer.resolveEndpoint(r.Transport.Protocol, signal, signalSpecific); ok {
			r.Transport.Endpoint = ep.Host
			if ep.HTTPPath != "" {
				r.HTTPPath = ep.HTTPPath
			}
			if ep.UseTLS != nil {
				r.Transport.UseTLS = new(*ep.UseTLS)
				tlsFromScheme = true
			}
		}

		if !tlsFromScheme && layer.Insecure != nil {
			if insecure, err := strconv.ParseBool(*layer.Insecure); err == nil {
				r.Transport.UseTLS = new(!insecure)
			}
		}
	}
}

// --- endpoint resolution ---

type resolvedEndpoint struct {
	Host     string
	HTTPPath string
	UseTLS   *bool // non-nil when URL scheme determined TLS
}

func (layer Layer) resolveEndpoint(
	protocol core_xds.OtelProtocol,
	signal core_xds.OtelSignal,
	signalSpecific bool,
) (resolvedEndpoint, bool) {
	if layer.Endpoint == nil {
		return resolvedEndpoint{}, false
	}

	parsed, err := url.Parse(*layer.Endpoint)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return resolvedEndpoint{Host: *layer.Endpoint}, true
	}

	if parsed.Path != "" && protocol != core_xds.OtelProtocolHTTPProtobuf {
		return resolvedEndpoint{}, false
	}

	ep := resolvedEndpoint{Host: parsed.Host}
	ep.UseTLS = new(!isInsecureScheme(parsed.Scheme))

	if protocol == core_xds.OtelProtocolHTTPProtobuf {
		p := parsed.Path
		switch {
		case signalSpecific && p == "":
			ep.HTTPPath = "/"
		case signalSpecific:
			ep.HTTPPath = p
		case p != "":
			ep.HTTPPath = path.Join(p, "v1", string(signal))
		}
	}

	return ep, true
}

func isInsecureScheme(scheme string) bool {
	switch strings.ToLower(scheme) {
	case "http", "unix":
		return true
	default:
		return false
	}
}

// --- parse helpers ---

func resolveEndpointAddress(endpoint string) string {
	host, port, err := net.SplitHostPort(endpoint)
	if err != nil || host != "" {
		return endpoint
	}
	return net.JoinHostPort(cmp.Or(os.Getenv("HOST_IP"), "127.0.0.1"), port)
}

func parseProtocol(value string) (core_xds.OtelProtocol, bool) {
	switch value {
	case string(core_xds.OtelProtocolGRPC):
		return core_xds.OtelProtocolGRPC, true
	case string(core_xds.OtelProtocolHTTPProtobuf):
		return core_xds.OtelProtocolHTTPProtobuf, true
	default:
		return "", false
	}
}

func parseCompression(value string) (string, bool) {
	switch strings.ToLower(value) {
	case "gzip":
		return "gzip", true
	case "none", "":
		return "", true
	default:
		return "", false
	}
}

func parseTimeout(value string) (time.Duration, bool) {
	timeoutMS, err := strconv.Atoi(value)
	if err != nil || timeoutMS < 0 {
		return 0, false
	}

	return time.Duration(timeoutMS) * time.Millisecond, true
}

// parseHeaders parses OTEL_EXPORTER_OTLP_HEADERS format: comma-separated
// key=value pairs with percent-encoded values (OTel env var spec, not W3C
// baggage which rejects unencoded spaces in values).
func parseHeaders(value string) map[string]string {
	var headers map[string]string
	for pair := range strings.SplitSeq(value, ",") {
		k, v, ok := strings.Cut(pair, "=")
		if !ok {
			continue
		}
		k = strings.TrimSpace(k)
		if k == "" {
			continue
		}
		if decoded, err := url.PathUnescape(v); err == nil {
			v = decoded
		}
		if headers == nil {
			headers = map[string]string{}
		}
		headers[k] = v
	}
	return headers
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
