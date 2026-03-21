package otelenv

import (
	"fmt"
	"net/url"
	"slices"
	"strings"

	core_xds "github.com/kumahq/kuma/v2/pkg/core/xds"
)

var envPrefix = map[core_xds.OtelSignal]string{
	core_xds.OtelSignalShared:  "OTEL_EXPORTER_OTLP",
	core_xds.OtelSignalTraces:  "OTEL_EXPORTER_OTLP_TRACES",
	core_xds.OtelSignalLogs:    "OTEL_EXPORTER_OTLP_LOGS",
	core_xds.OtelSignalMetrics: "OTEL_EXPORTER_OTLP_METRICS",
}

func Discover(pipeEnabled bool) Config {
	return discoverWithLookup(pipeEnabled, OSEnvReader{})
}

func discoverWithLookup(pipeEnabled bool, reader EnvReader) Config {
	return NewConfig(pipeEnabled,
		readLayer(core_xds.OtelSignalShared, reader),
		readLayer(core_xds.OtelSignalTraces, reader),
		readLayer(core_xds.OtelSignalLogs, reader),
		readLayer(core_xds.OtelSignalMetrics, reader),
	)
}

func NewConfig(pipeEnabled bool, shared, traces, logs, metrics Layer) Config {
	sharedInv := shared.analyze(nil)
	tracesInv := traces.analyze(&shared)
	logsInv := logs.analyze(&shared)
	metricsInv := metrics.analyze(&shared)

	return Config{
		PipeEnabled: pipeEnabled,
		Shared:      shared,
		Traces:      traces,
		Logs:        logs,
		Metrics:     metrics,
		Inventory: core_xds.OtelBootstrapInventory{
			PipeEnabled: pipeEnabled,
			Shared:      sharedInv,
			Traces:      tracesInv,
			Logs:        logsInv,
			Metrics:     metricsInv,
			ValidationErrors: slices.Concat(
				sharedInv.GetValidationErrors(),
				tracesInv.GetValidationErrors(),
				logsInv.GetValidationErrors(),
				metricsInv.GetValidationErrors(),
			),
		},
	}
}

func readLayer(signal core_xds.OtelSignal, reader EnvReader) Layer {
	prefix := envPrefix[signal]
	return Layer{
		Signal:            signal,
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

func (layer Layer) analyze(shared *Layer) *core_xds.OtelSignalEnvInventory {
	inv := &core_xds.OtelSignalEnvInventory{
		Signal:                   layer.Signal,
		EndpointPresent:          layer.Endpoint != nil,
		ProtocolPresent:          layer.Protocol != nil,
		HeadersPresent:           layer.Headers != nil,
		TimeoutPresent:           layer.Timeout != nil,
		CompressionPresent:       layer.Compression != nil,
		InsecurePresent:          layer.Insecure != nil,
		CertificatePresent:       layer.Certificate != nil,
		ClientCertificatePresent: layer.ClientCertificate != nil,
		ClientKeyPresent:         layer.ClientKey != nil,
		OverrideKinds:            layer.overrideKinds(shared),
	}

	if layer.Protocol != nil {
		if parsed, ok := parseProtocol(*layer.Protocol); ok {
			inv.EffectiveProtocol = parsed
		} else {
			inv.EffectiveProtocol = core_xds.OtelProtocolUnknown
			inv.ValidationErrors = append(inv.ValidationErrors, fmt.Sprintf("%s.protocol", inv.Signal))
		}
	}

	if layer.Endpoint != nil {
		if parsed, err := url.Parse(*layer.Endpoint); err == nil && parsed.Scheme != "" && parsed.Host != "" {
			inv.EndpointParsedAsURL = true
			inv.EndpointHasPath = parsed.Path != ""
			if parsed.Path != "" && effectiveProtocolForEndpoint(layer, shared) == core_xds.OtelProtocolGRPC {
				inv.ValidationErrors = append(inv.ValidationErrors, fmt.Sprintf("%s.endpoint", inv.Signal))
			}
		}
	}

	if layer.Compression != nil {
		if _, ok := parseCompression(*layer.Compression); !ok {
			inv.ValidationErrors = append(inv.ValidationErrors, fmt.Sprintf("%s.compression", inv.Signal))
		}
	}

	if layer.Timeout != nil {
		if _, ok := parseTimeout(*layer.Timeout); !ok {
			inv.ValidationErrors = append(inv.ValidationErrors, fmt.Sprintf("%s.timeout", inv.Signal))
		}
	}

	if (layer.ClientCertificate != nil) != (layer.ClientKey != nil) {
		inv.ValidationErrors = append(inv.ValidationErrors, fmt.Sprintf("%s.mtls", inv.Signal))
	}

	switch {
	case layer.ClientCertificate != nil && layer.ClientKey != nil:
		inv.EffectiveAuthMode = core_xds.OtelAuthModeMTLS
	case layer.Certificate != nil:
		inv.EffectiveAuthMode = core_xds.OtelAuthModeTLS
	case layer.Headers != nil:
		inv.EffectiveAuthMode = core_xds.OtelAuthModeHeaders
	default:
		inv.EffectiveAuthMode = core_xds.OtelAuthModeNone
	}

	if !inv.HasAnyInput() {
		return nil
	}
	return inv
}

func (layer Layer) overrideKinds(shared *Layer) []string {
	if shared == nil {
		return nil
	}

	type fieldPair struct {
		name          string
		signal, share *string
	}

	// Pre-sorted alphabetically so the result is already sorted.
	pairs := []fieldPair{
		{"certificate", layer.Certificate, shared.Certificate},
		{"clientCertificate", layer.ClientCertificate, shared.ClientCertificate},
		{"clientKey", layer.ClientKey, shared.ClientKey},
		{"compression", layer.Compression, shared.Compression},
		{"endpoint", layer.Endpoint, shared.Endpoint},
		{"headers", layer.Headers, shared.Headers},
		{"insecure", layer.Insecure, shared.Insecure},
		{"protocol", layer.Protocol, shared.Protocol},
		{"timeout", layer.Timeout, shared.Timeout},
	}

	var kinds []string
	for pair := range slices.Values(pairs) {
		if pair.signal != nil && (pair.share == nil || *pair.signal != *pair.share) {
			kinds = append(kinds, pair.name)
		}
	}

	return kinds
}

func effectiveProtocolForEndpoint(layer Layer, shared *Layer) core_xds.OtelProtocol {
	protocols := []*string{layer.Protocol}
	if shared != nil {
		protocols = append(protocols, shared.Protocol)
	}

	for field := range slices.Values(protocols) {
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
