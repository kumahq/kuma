package xds

import (
	"fmt"
	"strconv"
	"strings"

	envoy_accesslog "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v3"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	access_loggers_file "github.com/envoyproxy/go-control-plane/envoy/extensions/access_loggers/file/v3"
	access_loggers_grpc "github.com/envoyproxy/go-control-plane/envoy/extensions/access_loggers/grpc/v3"
	access_loggers_otel "github.com/envoyproxy/go-control-plane/envoy/extensions/access_loggers/open_telemetry/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	envoy_tcp "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/tcp_proxy/v3"
	"github.com/pkg/errors"
	otlp "go.opentelemetry.io/proto/otlp/common/v1"
	"google.golang.org/protobuf/types/known/structpb"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/validators"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshaccesslog/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/util/pointer"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	listeners_v3 "github.com/kumahq/kuma/pkg/xds/envoy/listeners/v3"
)

const (
	defaultNetworkAccessLogFormat = `[%START_TIME%] %RESPONSE_FLAGS% %KUMA_MESH% %KUMA_SOURCE_ADDRESS_WITHOUT_PORT%(%KUMA_SOURCE_SERVICE%)->%UPSTREAM_HOST%(%KUMA_DESTINATION_SERVICE%) took %DURATION%ms, sent %BYTES_SENT% bytes, received: %BYTES_RECEIVED% bytes`
	defaultHttpAccessLogFormat    = `[%START_TIME%] %KUMA_MESH% "%REQ(:METHOD)% %REQ(X-ENVOY-ORIGINAL-PATH?:PATH)% %PROTOCOL%" %RESPONSE_CODE% %RESPONSE_FLAGS% %BYTES_RECEIVED% %BYTES_SENT% %DURATION% %RESP(X-ENVOY-UPSTREAM-SERVICE-TIME)% "%REQ(X-FORWARDED-FOR)%" "%REQ(USER-AGENT)%" "%REQ(X-B3-TRACEID?X-DATADOG-TRACEID)%" "%REQ(X-REQUEST-ID)%" "%REQ(:AUTHORITY)%" "%KUMA_SOURCE_SERVICE%" "%KUMA_DESTINATION_SERVICE%" "%KUMA_SOURCE_ADDRESS_WITHOUT_PORT%" "%UPSTREAM_HOST%"`
)

type Configurer struct {
	Mesh                string
	TrafficDirection    envoy.TrafficDirection
	SourceService       string
	DestinationService  string
	Backend             api.Backend
	Dataplane           *core_mesh.DataplaneResource
	AccessLogSocketPath string
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

func (c *Configurer) interpolateKumaVariables(formatString string) string {
	return listeners_v3.InterpolateKumaValues(formatString, c.SourceService, c.DestinationService, c.Mesh, c.TrafficDirection, c.Dataplane)
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

func (c *Configurer) envoyAccessLog(endpoints *EndpointAccumulator, defaultFormat string) (*envoy_accesslog.AccessLog, error) {
	switch {
	case c.Backend.Tcp != nil:
		return c.tcpBackend(c.Backend.Tcp, defaultFormat)
	case c.Backend.File != nil:
		return c.fileBackend(c.Backend.File, defaultFormat)
	case c.Backend.OpenTelemetry != nil:
		return c.otelAccessLog(c.Backend.OpenTelemetry, endpoints, defaultFormat)
	default:
		return nil, errors.New(validators.MustHaveOnlyOne("backend", "tcp", "file", "openTelemetry"))
	}
}

func (c *Configurer) tcpBackend(backend *api.TCPBackend, defaultFormat string) (*envoy_accesslog.AccessLog, error) {
	var sfs *envoy_core.SubstitutionFormatString

	switch {
	case backend.Format == nil:
		envoyFormat := c.interpolateKumaVariables(newLine(defaultFormat))
		sfs = c.sfsJSON(map[string]*structpb.Value{
			"address": structpb.NewStringValue(backend.Address),
			"message": structpb.NewStringValue(envoyFormat),
		}, false)
	case backend.Format.Plain != nil:
		envoyFormat := c.interpolateKumaVariables(newLine(*backend.Format.Plain))
		sfs = c.sfsJSON(map[string]*structpb.Value{
			"address": structpb.NewStringValue(backend.Address),
			"message": structpb.NewStringValue(envoyFormat),
		}, pointer.Deref(backend.Format.OmitEmptyValues))
	case backend.Format.Json != nil:
		fields := c.jsonToFields(*backend.Format.Json)
		sfs = c.sfsJSON(map[string]*structpb.Value{
			"address": structpb.NewStringValue(backend.Address),
			"message": structpb.NewStructValue(&structpb.Struct{Fields: fields}),
		}, pointer.Deref(backend.Format.OmitEmptyValues))
	default:
		return nil, errors.New(validators.MustHaveOnlyOne("format", "plain", "json"))
	}

	return fileAccessLog(sfs, c.AccessLogSocketPath)
}

func (c *Configurer) fileBackend(backend *api.FileBackend, defaultFormat string) (*envoy_accesslog.AccessLog, error) {
	var sfs *envoy_core.SubstitutionFormatString

	switch {
	case backend.Format == nil:
		sfs = c.sfsPlain(newLine(defaultFormat), false)
	case backend.Format.Plain != nil:
		sfs = c.sfsPlain(newLine(*backend.Format.Plain), pointer.Deref(backend.Format.OmitEmptyValues))
	case backend.Format.Json != nil:
		fields := c.jsonToFields(*backend.Format.Json)
		sfs = c.sfsJSON(fields, pointer.Deref(backend.Format.OmitEmptyValues))
	default:
		return nil, errors.New(validators.MustHaveOnlyOne("format", "plain", "json"))
	}

	return fileAccessLog(sfs, backend.Path)
}

func newLine(s string) string {
	return s + "\n"
}

func (c *Configurer) sfsPlain(plain string, omitEmpty bool) *envoy_core.SubstitutionFormatString {
	envoyFormat := c.interpolateKumaVariables(plain)
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

func (c *Configurer) sfsJSON(fields map[string]*structpb.Value, omitEmpty bool) *envoy_core.SubstitutionFormatString {
	return &envoy_core.SubstitutionFormatString{
		Format: &envoy_core.SubstitutionFormatString_JsonFormat{
			JsonFormat: &structpb.Struct{
				Fields: fields,
			},
		},
		OmitEmptyValues: omitEmpty,
	}
}

func (c *Configurer) jsonToFields(jsonValues []api.JsonValue) map[string]*structpb.Value {
	fields := map[string]*structpb.Value{}
	for _, kv := range jsonValues {
		if kv.Key == nil || kv.Value == nil {
			continue
		}
		interpolated := c.interpolateKumaVariables(*kv.Value)
		fields[*kv.Key] = structpb.NewStringValue(interpolated)
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

func (c *Configurer) interpolateKumaVariablesInAnyValue(val *otlp.AnyValue) error {
	switch v := val.GetValue().(type) {
	case *otlp.AnyValue_StringValue:
		interpolated := c.interpolateKumaVariables(v.StringValue)
		v.StringValue = interpolated
	case *otlp.AnyValue_ArrayValue:
		for _, kv := range v.ArrayValue.Values {
			if err := c.interpolateKumaVariablesInAnyValue(kv); err != nil {
				return err
			}
		}
	case *otlp.AnyValue_KvlistValue:
		for _, kv := range v.KvlistValue.Values {
			if err := c.interpolateKumaVariablesInAnyValue(kv.Value); err != nil {
				return err
			}
			key := c.interpolateKumaVariables(kv.Key)
			kv.Key = key
		}
	case *otlp.AnyValue_BoolValue:
	case *otlp.AnyValue_IntValue:
	case *otlp.AnyValue_DoubleValue:
	case *otlp.AnyValue_BytesValue:
	}

	return nil
}

func (c *Configurer) otelAccessLog(
	backend *api.OtelBackend,
	endpoints *EndpointAccumulator,
	defaultBodyFormat string,
) (*envoy_accesslog.AccessLog, error) {
	defaultBody := c.interpolateKumaVariables(defaultBodyFormat)
	body := &otlp.AnyValue{
		Value: &otlp.AnyValue_StringValue{StringValue: defaultBody},
	}
	if backend.Body != nil {
		if err := util_proto.FromJSON(backend.Body.Raw, body); err == nil {
			if err := c.interpolateKumaVariablesInAnyValue(body); err != nil {
				return nil, errors.Wrap(err, "couldn't interpolate OTLP any value")
			}
		} else {
			interpolatedRaw := c.interpolateKumaVariables(string(backend.Body.Raw))
			body = &otlp.AnyValue{
				Value: &otlp.AnyValue_StringValue{StringValue: interpolatedRaw},
			}
		}
	}

	attributes := otlp.KeyValueList{}
	if backend.Attributes != nil {
		for _, kv := range *backend.Attributes {
			if kv.Key == nil || kv.Value == nil {
				continue
			}
			attributes.Values = append(attributes.Values, &otlp.KeyValue{
				Key: *kv.Key,
				Value: &otlp.AnyValue{
					Value: &otlp.AnyValue_StringValue{StringValue: *kv.Value},
				},
			})
		}
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

func (c *Configurer) Configure(filterChain *envoy_listener.FilterChain, endpoints *EndpointAccumulator) error {
	httpAccessLog := func(hcm *envoy_hcm.HttpConnectionManager) error {
		accessLog, err := c.envoyAccessLog(endpoints, defaultHttpAccessLogFormat)
		if err != nil {
			return err
		}
		hcm.AccessLog = append(hcm.AccessLog, accessLog)
		return nil
	}
	tcpAccessLog := func(tcpProxy *envoy_tcp.TcpProxy) error {
		accessLog, err := c.envoyAccessLog(endpoints, defaultNetworkAccessLogFormat)
		if err != nil {
			return err
		}

		tcpProxy.AccessLog = append(tcpProxy.AccessLog, accessLog)
		return nil
	}

	if err := listeners_v3.UpdateHTTPConnectionManager(filterChain, httpAccessLog); err != nil && !errors.Is(err, &listeners_v3.UnexpectedFilterConfigTypeError{}) {
		return err
	}
	if err := listeners_v3.UpdateTCPProxy(filterChain, tcpAccessLog); err != nil && !errors.Is(err, &listeners_v3.UnexpectedFilterConfigTypeError{}) {
		return err
	}

	return nil
}
