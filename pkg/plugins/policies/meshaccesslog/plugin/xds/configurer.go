package xds

import (
	"fmt"
	"strconv"
	"strings"

	envoy_accesslog "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v3"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	access_loggers_file "github.com/envoyproxy/go-control-plane/envoy/extensions/access_loggers/file/v3"
	access_loggers_otel "github.com/envoyproxy/go-control-plane/envoy/extensions/access_loggers/open_telemetry/v3"
	envoy_wellknown "github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"github.com/pkg/errors"
	otlp "go.opentelemetry.io/proto/otlp/common/v1"
	"google.golang.org/protobuf/types/known/structpb"

	core_meta "github.com/kumahq/kuma/pkg/core/metadata"
	core_system_names "github.com/kumahq/kuma/pkg/core/system_names"
	"github.com/kumahq/kuma/pkg/core/validators"
	bldrs_accesslog "github.com/kumahq/kuma/pkg/envoy/builders/accesslog"
	. "github.com/kumahq/kuma/pkg/envoy/builders/common"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshaccesslog/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/util/pointer"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	listeners_v3 "github.com/kumahq/kuma/pkg/xds/envoy/listeners/v3"
)

const (
	defaultNetworkAccessLogFormat = `[%START_TIME%] %RESPONSE_FLAGS% %KUMA_MESH% %KUMA_SOURCE_ADDRESS_WITHOUT_PORT%(%KUMA_SOURCE_SERVICE%)->%UPSTREAM_HOST%(%KUMA_DESTINATION_SERVICE%) took %DURATION%ms, sent %BYTES_SENT% bytes, received: %BYTES_RECEIVED% bytes`
	defaultHttpAccessLogFormat    = `[%START_TIME%] %KUMA_MESH% "%REQ(:METHOD)% %REQ(X-ENVOY-ORIGINAL-PATH?:PATH)% %PROTOCOL%" %RESPONSE_CODE% %RESPONSE_FLAGS% %BYTES_RECEIVED% %BYTES_SENT% %DURATION% %RESP(X-ENVOY-UPSTREAM-SERVICE-TIME)% "%REQ(X-FORWARDED-FOR)%" "%REQ(USER-AGENT)%" "%REQ(X-B3-TRACEID?X-DATADOG-TRACEID)%" "%REQ(X-REQUEST-ID)%" "%REQ(:AUTHORITY)%" "%KUMA_SOURCE_SERVICE%" "%KUMA_DESTINATION_SERVICE%" "%KUMA_SOURCE_ADDRESS_WITHOUT_PORT%" "%UPSTREAM_HOST%"`
)

func DefaultFormat(protocol core_meta.Protocol) string {
	if core_meta.IsHTTPBased(protocol) {
		return defaultHttpAccessLogFormat
	} else {
		return defaultNetworkAccessLogFormat
	}
}

func BaseAccessLogBuilder(
	backend api.Backend,
	defaultFormat string,
	backendsAcc *EndpointAccumulator,
	values listeners_v3.KumaValues,
	accessLogSocketPath string,
) *Builder[envoy_accesslog.AccessLog] {
	return bldrs_accesslog.NewBuilder().
		Configure(IfNotNil(backend.Tcp, func(tcpBackend api.TCPBackend) Configurer[envoy_accesslog.AccessLog] {
			return bldrs_accesslog.Config(envoy_wellknown.FileAccessLog, bldrs_accesslog.NewFileBuilder().
				Configure(TCPBackendSFS(&tcpBackend, defaultFormat, values)).
				Configure(bldrs_accesslog.Path(accessLogSocketPath)))
		})).
		Configure(IfNotNil(backend.File, func(fileBackend api.FileBackend) Configurer[envoy_accesslog.AccessLog] {
			return bldrs_accesslog.Config(envoy_wellknown.FileAccessLog, bldrs_accesslog.NewFileBuilder().
				Configure(FileBackendSFS(&fileBackend, defaultFormat, values)).
				Configure(bldrs_accesslog.Path(fileBackend.Path)))
		})).
		Configure(IfNotNil(backend.OpenTelemetry, func(otelBackend api.OtelBackend) Configurer[envoy_accesslog.AccessLog] {
			return bldrs_accesslog.Config("envoy.access_loggers.open_telemetry", bldrs_accesslog.NewOtelBuilder().
				Configure(OtelBody(&otelBackend, defaultFormat, values)).
				Configure(OtelAttributes(&otelBackend)).
				Configure(bldrs_accesslog.CommonConfig("MeshAccessLog", string(backendsAcc.ClusterForEndpoint(
					EndpointForOtel(otelBackend.Endpoint),
				)))))
		}))
}

type EndpointAccumulator struct {
	endpoints             map[LoggingEndpoint]int
	latest                int
	UnifiedResourceNaming bool
}

type endpointClusterName string

func (acc *EndpointAccumulator) ClusterForEndpoint(endpoint LoggingEndpoint) endpointClusterName {
	ind, found := acc.endpoints[endpoint]
	if !found {
		ind = acc.latest
		if acc.endpoints == nil {
			acc.endpoints = map[LoggingEndpoint]int{}
		}
		acc.endpoints[endpoint] = ind
		acc.latest += 1
	}

	getNameOrDefault := core_system_names.GetNameOrDefault(acc.UnifiedResourceNaming)
	name := getNameOrDefault(
		core_system_names.AsSystemName("meshaccesslog_"+core_system_names.CleanName(endpoint.Address+"-"+strconv.Itoa(int(endpoint.Port)))),
		fmt.Sprintf("meshaccesslog:opentelemetry:%d", ind),
	)
	return endpointClusterName(name)
}

const defaultOpenTelemetryGRPCPort uint32 = 4317

func EndpointForOtel(endpoint string) LoggingEndpoint {
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

func TCPBackendSFS(
	backend *api.TCPBackend,
	defaultFormat string,
	values listeners_v3.KumaValues,
) Configurer[access_loggers_file.FileAccessLog] {
	return func(a *access_loggers_file.FileAccessLog) error {
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
			return errors.New(validators.MustHaveOnlyOne("format", "plain", "json"))
		}
		a.AccessLogFormat = &access_loggers_file.FileAccessLog_LogFormat{
			LogFormat: sfs,
		}
		return nil
	}
}

func FileBackendSFS(
	backend *api.FileBackend,
	defaultFormat string,
	values listeners_v3.KumaValues,
) Configurer[access_loggers_file.FileAccessLog] {
	return func(a *access_loggers_file.FileAccessLog) error {
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
			return errors.New(validators.MustHaveOnlyOne("format", "plain", "json"))
		}
		a.AccessLogFormat = &access_loggers_file.FileAccessLog_LogFormat{
			LogFormat: sfs,
		}
		return nil
	}
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

func interpolateKumaVariablesInAnyValue(val *otlp.AnyValue, values listeners_v3.KumaValues) {
	switch v := val.GetValue().(type) {
	case *otlp.AnyValue_StringValue:
		interpolated := listeners_v3.InterpolateKumaValues(v.StringValue, values)
		v.StringValue = interpolated
	case *otlp.AnyValue_ArrayValue:
		for _, kv := range v.ArrayValue.Values {
			interpolateKumaVariablesInAnyValue(kv, values)
		}
	case *otlp.AnyValue_KvlistValue:
		for _, kv := range v.KvlistValue.Values {
			interpolateKumaVariablesInAnyValue(kv.Value, values)
			key := listeners_v3.InterpolateKumaValues(kv.Key, values)
			kv.Key = key
		}
	case *otlp.AnyValue_BoolValue:
	case *otlp.AnyValue_IntValue:
	case *otlp.AnyValue_DoubleValue:
	case *otlp.AnyValue_BytesValue:
	}
}

func OtelBody(
	backend *api.OtelBackend,
	defaultFormat string,
	values listeners_v3.KumaValues,
) Configurer[access_loggers_otel.OpenTelemetryAccessLogConfig] {
	return func(a *access_loggers_otel.OpenTelemetryAccessLogConfig) error {
		defaultBody := listeners_v3.InterpolateKumaValues(defaultFormat, values)
		body := &otlp.AnyValue{
			Value: &otlp.AnyValue_StringValue{StringValue: defaultBody},
		}
		if backend.Body != nil {
			if err := util_proto.FromJSON(backend.Body.Raw, body); err == nil {
				interpolateKumaVariablesInAnyValue(body, values)
			} else {
				interpolatedRaw := listeners_v3.InterpolateKumaValues(string(backend.Body.Raw), values)
				body = &otlp.AnyValue{
					Value: &otlp.AnyValue_StringValue{StringValue: interpolatedRaw},
				}
			}
		}
		a.Body = body
		return nil
	}
}

func OtelAttributes(backend *api.OtelBackend) Configurer[access_loggers_otel.OpenTelemetryAccessLogConfig] {
	return func(a *access_loggers_otel.OpenTelemetryAccessLogConfig) error {
		attributes := &otlp.KeyValueList{}
		for _, kv := range pointer.Deref(backend.Attributes) {
			attributes.Values = append(attributes.Values, &otlp.KeyValue{
				Key: kv.Key,
				Value: &otlp.AnyValue{
					Value: &otlp.AnyValue_StringValue{StringValue: kv.Value},
				},
			})
		}
		a.Attributes = attributes
		return nil
	}
}
