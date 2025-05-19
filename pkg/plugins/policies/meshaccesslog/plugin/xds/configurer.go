package xds

import (
	"fmt"
	"strconv"
	"strings"

	envoy_accesslog "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v3"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	access_loggers_file "github.com/envoyproxy/go-control-plane/envoy/extensions/access_loggers/file/v3"
	access_loggers_grpc "github.com/envoyproxy/go-control-plane/envoy/extensions/access_loggers/grpc/v3"
	access_loggers_otel "github.com/envoyproxy/go-control-plane/envoy/extensions/access_loggers/open_telemetry/v3"
	"github.com/pkg/errors"
	otlp "go.opentelemetry.io/proto/otlp/common/v1"
	"google.golang.org/protobuf/types/known/structpb"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/validators"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshaccesslog/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/util/pointer"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	listeners_v3 "github.com/kumahq/kuma/pkg/xds/envoy/listeners/v3"
)

const (
	defaultNetworkAccessLogFormat = `[%START_TIME%] %RESPONSE_FLAGS% %KUMA_MESH% %KUMA_SOURCE_ADDRESS_WITHOUT_PORT%(%KUMA_SOURCE_SERVICE%)->%UPSTREAM_HOST%(%KUMA_DESTINATION_SERVICE%) took %DURATION%ms, sent %BYTES_SENT% bytes, received: %BYTES_RECEIVED% bytes`
	defaultHttpAccessLogFormat    = `[%START_TIME%] %KUMA_MESH% "%REQ(:METHOD)% %REQ(X-ENVOY-ORIGINAL-PATH?:PATH)% %PROTOCOL%" %RESPONSE_CODE% %RESPONSE_FLAGS% %BYTES_RECEIVED% %BYTES_SENT% %DURATION% %RESP(X-ENVOY-UPSTREAM-SERVICE-TIME)% "%REQ(X-FORWARDED-FOR)%" "%REQ(USER-AGENT)%" "%REQ(X-B3-TRACEID?X-DATADOG-TRACEID)%" "%REQ(X-REQUEST-ID)%" "%REQ(:AUTHORITY)%" "%KUMA_SOURCE_SERVICE%" "%KUMA_DESTINATION_SERVICE%" "%KUMA_SOURCE_ADDRESS_WITHOUT_PORT%" "%UPSTREAM_HOST%"`
)

func defaultFormat(protocol core_mesh.Protocol) string {
	switch protocol {
	case core_mesh.ProtocolHTTP, core_mesh.ProtocolHTTP2, core_mesh.ProtocolGRPC:
		return defaultHttpAccessLogFormat
	default:
		return defaultNetworkAccessLogFormat
	}
}

type EndpointAccumulator struct {
	endpoints map[LoggingEndpoint]int
	latest    int
}

type endpointClusterName string

func (acc *EndpointAccumulator) clusterForEndpoint(endpoint LoggingEndpoint) endpointClusterName {
	ind, found := acc.endpoints[endpoint]
	if !found {
		ind = acc.latest
		if acc.endpoints == nil {
			acc.endpoints = map[LoggingEndpoint]int{}
		}
		acc.endpoints[endpoint] = ind
		acc.latest += 1
	}

	return endpointClusterName(fmt.Sprintf("meshaccesslog:opentelemetry:%d", ind))
}

const defaultOpenTelemetryGRPCPort uint32 = 4317

func endpointForOtel(endpoint string) LoggingEndpoint {
	target := strings.Split(endpoint, ":")
	port := defaultOpenTelemetryGRPCPort
	if len(target) > 1 {
		val, _ := strconv.ParseInt(target[1], 10, 32)
		port = uint32(val)
	}

	return LoggingEndpoint{
		Address: target[0],
		Port:    port,
	}
}

func EnvoyAccessLog(
	backend api.Backend,
	endpoints *EndpointAccumulator,
	protocol core_mesh.Protocol,
	values listeners_v3.KumaValues,
	accessLogSocketPath string,
) (*envoy_accesslog.AccessLog, error) {
	switch {
	case backend.Tcp != nil:
		return tcpBackend(backend.Tcp, defaultFormat(protocol), values, accessLogSocketPath)
	case backend.File != nil:
		return fileBackend(backend.File, defaultFormat(protocol), values)
	case backend.OpenTelemetry != nil:
		return otelAccessLog(backend.OpenTelemetry, endpoints, defaultFormat(protocol), values)
	default:
		return nil, errors.New(validators.MustHaveOnlyOne("backend", "tcp", "file", "openTelemetry"))
	}
}

func tcpBackend(
	backend *api.TCPBackend,
	defaultFormat string,
	values listeners_v3.KumaValues,
	accessLogSocketPath string,
) (*envoy_accesslog.AccessLog, error) {
	var sfs *envoy_core.SubstitutionFormatString

	switch {
	case backend.Format == nil:
		envoyFormat := listeners_v3.InterpolateKumaValues(newLine(defaultFormat), values)
		sfs = sfsJSON(map[string]*structpb.Value{
			"address": structpb.NewStringValue(backend.Address),
			"message": structpb.NewStringValue(envoyFormat),
		}, false)
	case backend.Format.Plain != nil:
		envoyFormat := listeners_v3.InterpolateKumaValues(newLine(*backend.Format.Plain), values)
		sfs = sfsJSON(map[string]*structpb.Value{
			"address": structpb.NewStringValue(backend.Address),
			"message": structpb.NewStringValue(envoyFormat),
		}, backend.Format.OmitEmptyValues)
	case backend.Format.Json != nil:
		fields := jsonToFields(*backend.Format.Json, values)
		sfs = sfsJSON(map[string]*structpb.Value{
			"address": structpb.NewStringValue(backend.Address),
			"message": structpb.NewStructValue(&structpb.Struct{Fields: fields}),
		}, backend.Format.OmitEmptyValues)
	default:
		return nil, errors.New(validators.MustHaveOnlyOne("format", "plain", "json"))
	}

	return fileAccessLog(sfs, accessLogSocketPath)
}

func fileBackend(backend *api.FileBackend, defaultFormat string, values listeners_v3.KumaValues) (*envoy_accesslog.AccessLog, error) {
	var sfs *envoy_core.SubstitutionFormatString

	switch {
	case backend.Format == nil:
		sfs = sfsPlain(newLine(defaultFormat), false, values)
	case backend.Format.Plain != nil:
		sfs = sfsPlain(newLine(*backend.Format.Plain), backend.Format.OmitEmptyValues, values)
	case backend.Format.Json != nil:
		fields := jsonToFields(*backend.Format.Json, values)
		sfs = sfsJSON(fields, backend.Format.OmitEmptyValues)
	default:
		return nil, errors.New(validators.MustHaveOnlyOne("format", "plain", "json"))
	}

	return fileAccessLog(sfs, backend.Path)
}

func newLine(s string) string {
	return s + "\n"
}

func sfsPlain(plain string, omitEmpty bool, values listeners_v3.KumaValues) *envoy_core.SubstitutionFormatString {
	envoyFormat := listeners_v3.InterpolateKumaValues(plain, values)
	return &envoy_core.SubstitutionFormatString{
		Format: &envoy_core.SubstitutionFormatString_TextFormatSource{
			TextFormatSource: &envoy_core.DataSource{
				Specifier: &envoy_core.DataSource_InlineString{
					InlineString: envoyFormat,
				},
			},
		},
		OmitEmptyValues: omitEmpty,
	}
}

func sfsJSON(fields map[string]*structpb.Value, omitEmpty bool) *envoy_core.SubstitutionFormatString {
	return &envoy_core.SubstitutionFormatString{
		Format: &envoy_core.SubstitutionFormatString_JsonFormat{
			JsonFormat: &structpb.Struct{
				Fields: fields,
			},
		},
		OmitEmptyValues: omitEmpty,
	}
}

func jsonToFields(jsonValues []api.JsonValue, values listeners_v3.KumaValues) map[string]*structpb.Value {
	fields := map[string]*structpb.Value{}
	for _, kv := range jsonValues {
		interpolated := listeners_v3.InterpolateKumaValues(kv.Value, values)
		fields[kv.Key] = structpb.NewStringValue(interpolated)
	}
	return fields
}

func fileAccessLog(logFormat *envoy_core.SubstitutionFormatString, path string) (*envoy_accesslog.AccessLog, error) {
	fileAccessLog := &access_loggers_file.FileAccessLog{
		AccessLogFormat: &access_loggers_file.FileAccessLog_LogFormat{
			LogFormat: logFormat,
		},
		Path: path,
	}

	marshaled, err := util_proto.MarshalAnyDeterministic(fileAccessLog)
	if err != nil {
		return nil, errors.Wrapf(err, "could not marshall %T", fileAccessLog)
	}
	return &envoy_accesslog.AccessLog{
		Name: "envoy.access_loggers.file",
		ConfigType: &envoy_accesslog.AccessLog_TypedConfig{
			TypedConfig: marshaled,
		},
	}, nil
}

func interpolateKumaVariablesInAnyValue(val *otlp.AnyValue, values listeners_v3.KumaValues) error {
	switch v := val.GetValue().(type) {
	case *otlp.AnyValue_StringValue:
		interpolated := listeners_v3.InterpolateKumaValues(v.StringValue, values)
		v.StringValue = interpolated
	case *otlp.AnyValue_ArrayValue:
		for _, kv := range v.ArrayValue.Values {
			if err := interpolateKumaVariablesInAnyValue(kv, values); err != nil {
				return err
			}
		}
	case *otlp.AnyValue_KvlistValue:
		for _, kv := range v.KvlistValue.Values {
			if err := interpolateKumaVariablesInAnyValue(kv.Value, values); err != nil {
				return err
			}
			key := listeners_v3.InterpolateKumaValues(kv.Key, values)
			kv.Key = key
		}
	case *otlp.AnyValue_BoolValue:
	case *otlp.AnyValue_IntValue:
	case *otlp.AnyValue_DoubleValue:
	case *otlp.AnyValue_BytesValue:
	}

	return nil
}

func otelAccessLog(
	backend *api.OtelBackend,
	endpoints *EndpointAccumulator,
	defaultBodyFormat string,
	values listeners_v3.KumaValues,
) (*envoy_accesslog.AccessLog, error) {
	defaultBody := listeners_v3.InterpolateKumaValues(defaultBodyFormat, values)
	body := &otlp.AnyValue{
		Value: &otlp.AnyValue_StringValue{StringValue: defaultBody},
	}
	if backend.Body != nil {
		if err := util_proto.FromJSON(backend.Body.Raw, body); err == nil {
			if err := interpolateKumaVariablesInAnyValue(body, values); err != nil {
				return nil, errors.Wrap(err, "couldn't interpolate OTLP any value")
			}
		} else {
			interpolatedRaw := listeners_v3.InterpolateKumaValues(string(backend.Body.Raw), values)
			body = &otlp.AnyValue{
				Value: &otlp.AnyValue_StringValue{StringValue: interpolatedRaw},
			}
		}
	}

	attributes := otlp.KeyValueList{}
	for _, kv := range pointer.Deref(backend.Attributes) {
		attributes.Values = append(attributes.Values, &otlp.KeyValue{
			Key: kv.Key,
			Value: &otlp.AnyValue{
				Value: &otlp.AnyValue_StringValue{StringValue: kv.Value},
			},
		})
	}

	log := &access_loggers_otel.OpenTelemetryAccessLogConfig{
		CommonConfig: &access_loggers_grpc.CommonGrpcAccessLogConfig{
			LogName:             "MeshAccessLog",
			TransportApiVersion: envoy_core.ApiVersion_V3,
			GrpcService: &envoy_core.GrpcService{
				TargetSpecifier: &envoy_core.GrpcService_EnvoyGrpc_{
					EnvoyGrpc: &envoy_core.GrpcService_EnvoyGrpc{
						ClusterName: string(endpoints.clusterForEndpoint(
							endpointForOtel(backend.Endpoint),
						)),
					},
				},
			},
		},
		Body:       body,
		Attributes: &attributes,
	}

	marshaled, err := util_proto.MarshalAnyDeterministic(log)
	if err != nil {
		return nil, errors.Wrapf(err, "could not marshall %T", log)
	}
	return &envoy_accesslog.AccessLog{
		Name: "envoy.access_loggers.open_telemetry",
		ConfigType: &envoy_accesslog.AccessLog_TypedConfig{
			TypedConfig: marshaled,
		},
	}, nil
}
