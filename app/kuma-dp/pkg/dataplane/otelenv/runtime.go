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

	"go.opentelemetry.io/otel/baggage"

	motb_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshopentelemetrybackend/api/v1alpha1"
	core_xds "github.com/kumahq/kuma/v2/pkg/core/xds"
)

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

	runtime.Transport = explicit.Transport
	runtime.Transport.Protocol = protocol
	runtime.Transport.UseTLS = new(*explicit.Transport.UseTLS)
	runtime.Transport.Headers = maps.Clone(explicit.Transport.Headers)

	runtime.HTTPPath = explicit.HTTPPath
	if runtime.Transport.Protocol == core_xds.OtelProtocolHTTPProtobuf && runtime.HTTPPath == "" {
		runtime.HTTPPath = path.Join("/", backend.Path, "v1", string(signal))
	}

	if sharedAllowed {
		runtime.mergeLayer(c.Shared, signal, false, preferEnv)
	}

	if signalAllowed {
		runtime.mergeLayer(sigLayer, signal, true, preferEnv)
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
			UseTLS:   new(backend.UseHTTPS),
		},
	}

	if protocol == core_xds.OtelProtocolGRPC {
		runtime.HTTPPath = ""
	}

	return runtime
}

type resolvedEndpoint struct {
	Host     string
	HTTPPath string
	UseTLS   *bool // non-nil when URL scheme determined TLS
}

// resolveEndpoint parses the layer's endpoint for the given protocol and
// signal context. Returns ok=false when absent or incompatible (gRPC + path).
func (layer Layer) resolveEndpoint(
	protocol core_xds.OtelProtocol,
	signal core_xds.OtelSignal,
	signalSpecific bool,
) (resolvedEndpoint, bool) {
	if layer.Endpoint == nil {
		return resolvedEndpoint{}, false
	}
	value := strings.TrimSpace(*layer.Endpoint)
	if value == "" {
		return resolvedEndpoint{}, false
	}

	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return resolvedEndpoint{Host: value}, true
	}

	if parsed.Path != "" && protocol != core_xds.OtelProtocolHTTPProtobuf {
		return resolvedEndpoint{}, false
	}

	ep := resolvedEndpoint{Host: parsed.Host}
	if strings.EqualFold(parsed.Scheme, "http") || strings.EqualFold(parsed.Scheme, "unix") {
		ep.UseTLS = new(false)
	} else {
		ep.UseTLS = new(true)
	}

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

// mergeLayer applies an env layer's parsed values onto the runtime transport.
func (r *SignalRuntime) mergeLayer(layer Layer, signal core_xds.OtelSignal, signalSpecific, preferEnv bool) {
	tlsFromScheme := false
	if ep, ok := layer.resolveEndpoint(r.Transport.Protocol, signal, signalSpecific); ok {
		if preferEnv || r.Transport.Endpoint == "" {
			r.Transport.Endpoint = ep.Host
		}
		if ep.HTTPPath != "" && (preferEnv || r.HTTPPath == "") {
			r.HTTPPath = ep.HTTPPath
		}
		if ep.UseTLS != nil && (preferEnv || r.Transport.UseTLS == nil) {
			r.Transport.UseTLS = new(*ep.UseTLS)
			tlsFromScheme = true
		}
	}

	if !tlsFromScheme && layer.Insecure != nil && (preferEnv || r.Transport.UseTLS == nil) {
		r.Transport.UseTLS = new(!strings.EqualFold(strings.TrimSpace(*layer.Insecure), "true"))
	}

	if layer.Headers != nil && (preferEnv || len(r.Transport.Headers) == 0) {
		if headers := parseHeaders(*layer.Headers); len(headers) > 0 {
			r.Transport.Headers = headers
		}
	}

	if layer.Compression != nil && (preferEnv || r.Transport.Compression == "") {
		if c, ok := parseCompression(*layer.Compression); ok {
			r.Transport.Compression = c
		}
	}

	if layer.Timeout != nil && (preferEnv || r.Transport.Timeout == 0) {
		if t, ok := parseTimeout(*layer.Timeout); ok {
			r.Transport.Timeout = t
		}
	}

	mergeString(&r.Transport.Certificate, layer.Certificate, preferEnv)
	if layer.ClientCertificate != nil && layer.ClientKey != nil && (preferEnv || r.Transport.ClientCertificate == "") {
		r.Transport.ClientCertificate = *layer.ClientCertificate
		r.Transport.ClientKey = *layer.ClientKey
	}
}

func mergeString(target *string, source *string, preferEnv bool) {
	if source != nil && (preferEnv || *target == "") {
		*target = *source
	}
}

// resolveEndpointAddress fills in the host portion when the CP sent an empty
// host (e.g. ":4317"). Uses HOST_IP env var, falling back to 127.0.0.1.
func resolveEndpointAddress(endpoint string) string {
	host, port, err := net.SplitHostPort(endpoint)
	if err != nil || host != "" {
		return endpoint
	}
	hostIP := cmp.Or(os.Getenv("HOST_IP"), "127.0.0.1")
	return net.JoinHostPort(hostIP, port)
}

func pickProtocol(current core_xds.OtelProtocol, field *string, preferEnv bool) core_xds.OtelProtocol {
	if field == nil {
		return current
	}

	parsed, ok := parseProtocol(*field)
	if !ok {
		return current
	}

	if preferEnv || current == "" {
		return parsed
	}

	return current
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
	parsed, _ := baggage.Parse(value)
	for _, member := range parsed.Members() {
		headers[member.Key()] = member.Value()
	}
	return headers
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
