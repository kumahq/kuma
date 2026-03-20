package otelenv

import (
	"fmt"
	"net/url"
	"os"
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

type Config struct {
	PipeEnabled bool
	Shared      Layer
	Traces      Layer
	Logs        Layer
	Metrics     Layer
	Inventory   *core_xds.OtelBootstrapInventory
}

type Layer struct {
	Endpoint          FieldValue
	Protocol          FieldValue
	Headers           FieldValue
	Timeout           FieldValue
	Compression       FieldValue
	Insecure          FieldValue
	Certificate       FieldValue
	ClientCertificate FieldValue
	ClientKey         FieldValue
}

type FieldValue struct {
	Present bool
	Value   string
}

func Discover(pipeEnabled bool) Config {
	return discoverWithLookup(pipeEnabled, os.LookupEnv)
}

func discoverWithLookup(
	pipeEnabled bool,
	lookup func(string) (string, bool),
) Config {
	cfg := Config{
		PipeEnabled: pipeEnabled,
		Shared:      readLayer(sharedEnvPrefix, lookup),
		Traces:      readLayer(tracesEnvPrefix, lookup),
		Logs:        readLayer(logsEnvPrefix, lookup),
		Metrics:     readLayer(metricsEnvPrefix, lookup),
	}
	cfg.Inventory = buildInventory(cfg)
	return cfg
}

func readLayer(prefix string, lookup func(string) (string, bool)) Layer {
	return Layer{
		Endpoint:          readField(prefix+"_ENDPOINT", lookup),
		Protocol:          readField(prefix+"_PROTOCOL", lookup),
		Headers:           readField(prefix+"_HEADERS", lookup),
		Timeout:           readField(prefix+"_TIMEOUT", lookup),
		Compression:       readField(prefix+"_COMPRESSION", lookup),
		Insecure:          readField(prefix+"_INSECURE", lookup),
		Certificate:       readField(prefix+"_CERTIFICATE", lookup),
		ClientCertificate: readField(prefix+"_CLIENT_CERTIFICATE", lookup),
		ClientKey:         readField(prefix+"_CLIENT_KEY", lookup),
	}
}

func readField(name string, lookup func(string) (string, bool)) FieldValue {
	value, ok := lookup(name)
	if !ok {
		return FieldValue{}
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return FieldValue{}
	}
	return FieldValue{
		Present: true,
		Value:   value,
	}
}

func buildInventory(cfg Config) *core_xds.OtelBootstrapInventory {
	var validationErrors []string
	inventory := &core_xds.OtelBootstrapInventory{
		PipeEnabled: cfg.PipeEnabled,
	}

	inventory.Shared = buildLayerInventory("shared", cfg.Shared, Layer{}, &validationErrors)
	inventory.Traces = buildLayerInventory(string(core_xds.OtelSignalTraces), cfg.Traces, cfg.Shared, &validationErrors)
	inventory.Logs = buildLayerInventory(string(core_xds.OtelSignalLogs), cfg.Logs, cfg.Shared, &validationErrors)
	inventory.Metrics = buildLayerInventory(string(core_xds.OtelSignalMetrics), cfg.Metrics, cfg.Shared, &validationErrors)
	inventory.ValidationErrors = validationErrors

	if inventory.Shared != nil && !inventory.Shared.HasAnyInput() {
		inventory.Shared = nil
	}
	if inventory.Traces != nil && !inventory.Traces.HasAnyInput() {
		inventory.Traces = nil
	}
	if inventory.Logs != nil && !inventory.Logs.HasAnyInput() {
		inventory.Logs = nil
	}
	if inventory.Metrics != nil && !inventory.Metrics.HasAnyInput() {
		inventory.Metrics = nil
	}

	return inventory
}

func buildLayerInventory(
	name string,
	layer Layer,
	shared Layer,
	validationErrors *[]string,
) *core_xds.OtelSignalEnvInventory {
	endpointParsedAsURL, endpointHasPath := endpointCharacteristics(layer.Endpoint)
	effectiveLayerProtocol := effectiveProtocol(name, layer, validationErrors)
	if layer.Compression.Present {
		if _, ok := parseCompression(layer.Compression.Value); !ok {
			*validationErrors = append(*validationErrors, fmt.Sprintf("%s.compression", name))
		}
	}
	if layer.Timeout.Present {
		if _, ok := parseTimeout(layer.Timeout.Value); !ok {
			*validationErrors = append(*validationErrors, fmt.Sprintf("%s.timeout", name))
		}
	}
	if endpointParsedAsURL && endpointHasPath && effectiveProtocolForEndpoint(layer, shared) == core_xds.OtelProtocolGRPC {
		*validationErrors = append(*validationErrors, fmt.Sprintf("%s.endpoint", name))
	}
	inventory := &core_xds.OtelSignalEnvInventory{
		EndpointPresent:          layer.Endpoint.Present,
		EndpointParsedAsURL:      endpointParsedAsURL,
		EndpointHasPath:          endpointHasPath,
		ProtocolPresent:          layer.Protocol.Present,
		HeadersPresent:           layer.Headers.Present,
		TimeoutPresent:           layer.Timeout.Present,
		CompressionPresent:       layer.Compression.Present,
		InsecurePresent:          layer.Insecure.Present,
		CertificatePresent:       layer.Certificate.Present,
		ClientCertificatePresent: layer.ClientCertificate.Present,
		ClientKeyPresent:         layer.ClientKey.Present,
		EffectiveProtocol:        effectiveLayerProtocol,
		EffectiveAuthMode:        effectiveAuthMode(name, layer, validationErrors),
		OverrideKinds:            overrideKinds(layer, shared),
	}

	return inventory
}

func endpointCharacteristics(field FieldValue) (bool, bool) {
	if !field.Present {
		return false, false
	}

	parsedURL, err := url.Parse(strings.TrimSpace(field.Value))
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return false, false
	}

	return true, parsedURL.Path != ""
}

func effectiveProtocol(
	name string,
	layer Layer,
	validationErrors *[]string,
) core_xds.OtelProtocol {
	if !layer.Protocol.Present {
		return ""
	}

	if parsed, ok := parseProtocol(layer.Protocol.Value); ok {
		return parsed
	}
	*validationErrors = append(*validationErrors, fmt.Sprintf("%s.protocol", name))
	return core_xds.OtelProtocolUnknown
}

func effectiveProtocolForEndpoint(layer Layer, shared Layer) core_xds.OtelProtocol {
	if parsed, ok := parseProtocolField(layer.Protocol); ok {
		return parsed
	}
	if layer.Protocol.Present {
		return core_xds.OtelProtocolUnknown
	}
	if parsed, ok := parseProtocolField(shared.Protocol); ok {
		return parsed
	}
	if shared.Protocol.Present {
		return core_xds.OtelProtocolUnknown
	}
	return core_xds.OtelProtocolGRPC
}

func parseProtocolField(field FieldValue) (core_xds.OtelProtocol, bool) {
	if !field.Present {
		return "", false
	}
	return parseProtocol(field.Value)
}

func effectiveAuthMode(
	name string,
	layer Layer,
	validationErrors *[]string,
) core_xds.OtelAuthMode {
	if layer.ClientCertificate.Present != layer.ClientKey.Present {
		*validationErrors = append(*validationErrors, fmt.Sprintf("%s.mtls", name))
	}

	switch {
	case layer.ClientCertificate.Present && layer.ClientKey.Present:
		return core_xds.OtelAuthModeMTLS
	case layer.Certificate.Present:
		return core_xds.OtelAuthModeTLS
	case layer.Headers.Present:
		return core_xds.OtelAuthModeHeaders
	default:
		return core_xds.OtelAuthModeNone
	}
}

func overrideKinds(layer Layer, shared Layer) []string {
	var overrides []string
	appendIfDifferent(&overrides, "endpoint", layer.Endpoint, shared.Endpoint)
	appendIfDifferent(&overrides, "protocol", layer.Protocol, shared.Protocol)
	appendIfDifferent(&overrides, "headers", layer.Headers, shared.Headers)
	appendIfDifferent(&overrides, "timeout", layer.Timeout, shared.Timeout)
	appendIfDifferent(&overrides, "compression", layer.Compression, shared.Compression)
	appendIfDifferent(&overrides, "insecure", layer.Insecure, shared.Insecure)
	appendIfDifferent(&overrides, "certificate", layer.Certificate, shared.Certificate)
	appendIfDifferent(&overrides, "clientCertificate", layer.ClientCertificate, shared.ClientCertificate)
	appendIfDifferent(&overrides, "clientKey", layer.ClientKey, shared.ClientKey)
	slices.Sort(overrides)
	return overrides
}

func appendIfDifferent(overrides *[]string, field string, signal FieldValue, shared FieldValue) {
	if !signal.Present {
		return
	}

	if !shared.Present || signal.Value != shared.Value {
		*overrides = append(*overrides, field)
	}
}
