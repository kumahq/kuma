package otelenv

import (
	"fmt"
	"net/url"
	"slices"
	"strings"

	core_xds "github.com/kumahq/kuma/v2/pkg/core/xds"
)

const (
	sharedEnvPrefix  = "OTEL_EXPORTER_OTLP"
	tracesEnvPrefix  = "OTEL_EXPORTER_OTLP_TRACES"
	logsEnvPrefix    = "OTEL_EXPORTER_OTLP_LOGS"
	metricsEnvPrefix = "OTEL_EXPORTER_OTLP_METRICS"
)

// Discover creates a new Config by reading environment variables
func Discover(pipeEnabled bool) Config {
	return discoverWithLookup(pipeEnabled, OSEnvReader{})
}

func discoverWithLookup(pipeEnabled bool, reader EnvReader) Config {
	cfg := Config{
		PipeEnabled: pipeEnabled,
		Shared:      readLayer(sharedEnvPrefix, reader),
		Traces:      readLayer(tracesEnvPrefix, reader),
		Logs:        readLayer(logsEnvPrefix, reader),
		Metrics:     readLayer(metricsEnvPrefix, reader),
	}
	cfg.Inventory = buildInventory(cfg)
	return cfg
}

func readLayer(prefix string, reader EnvReader) Layer {
	return Layer{
		Endpoint:          readField(reader, prefix+"_ENDPOINT"),
		Protocol:          readField(reader, prefix+"_PROTOCOL"),
		Headers:           readField(reader, prefix+"_HEADERS"),
		Timeout:           readField(reader, prefix+"_TIMEOUT"),
		Compression:       readField(reader, prefix+"_COMPRESSION"),
		Insecure:          readField(reader, prefix+"_INSECURE"),
		Certificate:       readField(reader, prefix+"_CERTIFICATE"),
		ClientCertificate: readField(reader, prefix+"_CLIENT_CERTIFICATE"),
		ClientKey:         readField(reader, prefix+"_CLIENT_KEY"),
	}
}

func readField(reader EnvReader, name string) *string {
	value, ok := reader.Lookup(name)
	if !ok {
		return nil
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func buildInventory(cfg Config) *core_xds.OtelBootstrapInventory {
	sharedInv, sharedErrs := buildLayerInventory("shared", cfg.Shared, Layer{})
	tracesInv, tracesErrs := buildLayerInventory(string(core_xds.OtelSignalTraces), cfg.Traces, cfg.Shared)
	logsInv, logsErrs := buildLayerInventory(string(core_xds.OtelSignalLogs), cfg.Logs, cfg.Shared)
	metricsInv, metricsErrs := buildLayerInventory(string(core_xds.OtelSignalMetrics), cfg.Metrics, cfg.Shared)

	return &core_xds.OtelBootstrapInventory{
		PipeEnabled:      cfg.PipeEnabled,
		Shared:           sharedInv,
		Traces:           tracesInv,
		Logs:             logsInv,
		Metrics:          metricsInv,
		ValidationErrors: slices.Concat(sharedErrs, tracesErrs, logsErrs, metricsErrs),
	}
}

func buildLayerInventory(
	name string,
	layer Layer,
	shared Layer,
) (*core_xds.OtelSignalEnvInventory, []string) {
	endpointParsedAsURL, endpointHasPath := endpointCharacteristics(layer.Endpoint)
	effectiveLayerProtocol, protoErrors := effectiveProtocol(name, layer)
	authMode, authErrors := effectiveAuthMode(name, layer)

	errors := slices.Concat(protoErrors, authErrors)

	if layer.Compression != nil {
		if _, ok := parseCompression(*layer.Compression); !ok {
			errors = append(errors, fmt.Sprintf("%s.compression", name))
		}
	}

	if layer.Timeout != nil {
		if _, ok := parseTimeout(*layer.Timeout); !ok {
			errors = append(errors, fmt.Sprintf("%s.timeout", name))
		}
	}

	if endpointParsedAsURL && endpointHasPath && effectiveProtocolForEndpoint(layer, shared) == core_xds.OtelProtocolGRPC {
		errors = append(errors, fmt.Sprintf("%s.endpoint", name))
	}

	inv := &core_xds.OtelSignalEnvInventory{
		EndpointPresent:          layer.Endpoint != nil,
		EndpointParsedAsURL:      endpointParsedAsURL,
		EndpointHasPath:          endpointHasPath,
		ProtocolPresent:          layer.Protocol != nil,
		HeadersPresent:           layer.Headers != nil,
		TimeoutPresent:           layer.Timeout != nil,
		CompressionPresent:       layer.Compression != nil,
		InsecurePresent:          layer.Insecure != nil,
		CertificatePresent:       layer.Certificate != nil,
		ClientCertificatePresent: layer.ClientCertificate != nil,
		ClientKeyPresent:         layer.ClientKey != nil,
		EffectiveProtocol:        effectiveLayerProtocol,
		EffectiveAuthMode:        authMode,
		OverrideKinds:            overrideKinds(layer, shared),
	}

	if !inv.HasAnyInput() {
		return nil, errors
	}

	return inv, errors
}

func endpointCharacteristics(field *string) (bool, bool) {
	if field == nil {
		return false, false
	}

	parsedURL, err := url.Parse(strings.TrimSpace(*field))
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return false, false
	}

	return true, parsedURL.Path != ""
}

func effectiveProtocol(name string, layer Layer) (core_xds.OtelProtocol, []string) {
	if layer.Protocol == nil {
		return "", nil
	}

	if parsed, ok := parseProtocol(*layer.Protocol); ok {
		return parsed, nil
	}

	return core_xds.OtelProtocolUnknown, []string{fmt.Sprintf("%s.protocol", name)}
}

func effectiveProtocolForEndpoint(layer Layer, shared Layer) core_xds.OtelProtocol {
	for _, field := range []*string{layer.Protocol, shared.Protocol} {
		if field == nil {
			continue
		}

		if parsed, ok := parseProtocol(*field); ok {
			return parsed
		}

		return core_xds.OtelProtocolUnknown
	}

	return core_xds.OtelProtocolGRPC
}

func effectiveAuthMode(name string, layer Layer) (core_xds.OtelAuthMode, []string) {
	var errors []string
	if (layer.ClientCertificate != nil) != (layer.ClientKey != nil) {
		errors = append(errors, fmt.Sprintf("%s.mtls", name))
	}

	switch {
	case layer.ClientCertificate != nil && layer.ClientKey != nil:
		return core_xds.OtelAuthModeMTLS, errors
	case layer.Certificate != nil:
		return core_xds.OtelAuthModeTLS, errors
	case layer.Headers != nil:
		return core_xds.OtelAuthModeHeaders, errors
	default:
		return core_xds.OtelAuthModeNone, errors
	}
}

func overrideKinds(layer Layer, shared Layer) []string {
	// Table is pre-sorted alphabetically so the result is already sorted.
	var overrides []string
	for _, pair := range []struct {
		name          string
		signal, share *string
	}{
		{"certificate", layer.Certificate, shared.Certificate},
		{"clientCertificate", layer.ClientCertificate, shared.ClientCertificate},
		{"clientKey", layer.ClientKey, shared.ClientKey},
		{"compression", layer.Compression, shared.Compression},
		{"endpoint", layer.Endpoint, shared.Endpoint},
		{"headers", layer.Headers, shared.Headers},
		{"insecure", layer.Insecure, shared.Insecure},
		{"protocol", layer.Protocol, shared.Protocol},
		{"timeout", layer.Timeout, shared.Timeout},
	} {
		if pair.signal != nil && (pair.share == nil || *pair.signal != *pair.share) {
			overrides = append(overrides, pair.name)
		}
	}
	return overrides
}
