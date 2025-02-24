package v3

import (
	net_url "net/url"
	"strings"

	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_trace "github.com/envoyproxy/go-control-plane/envoy/config/trace/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	envoy_type "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/util/pointer"
	"github.com/kumahq/kuma/pkg/util/proto"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/envoy/names"
)

type TracingConfigurer struct {
	Backend *mesh_proto.TracingBackend

	// Opaque string which envoy will assign to tracer collector cluster, on those
	// which support association of named "service" tags on traces. Consumed by datadog.
	Service          string
	TrafficDirection envoy_common.TrafficDirection
	Destination      string

	SpawnUpstreamSpan bool
}

var _ FilterChainConfigurer = &TracingConfigurer{}

func (c *TracingConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	if c.Backend == nil {
		return nil
	}

	return UpdateHTTPConnectionManager(filterChain, func(hcm *envoy_hcm.HttpConnectionManager) error {
		hcm.Tracing = &envoy_hcm.HttpConnectionManager_Tracing{
			SpawnUpstreamSpan: wrapperspb.Bool(c.SpawnUpstreamSpan),
		}
		if c.Backend.Sampling != nil {
			hcm.Tracing.OverallSampling = &envoy_type.Percent{
				Value: c.Backend.Sampling.Value,
			}
		}
		switch c.Backend.Type {
		case mesh_proto.TracingZipkinType:
			tracing, err := zipkinConfig(c.Backend.Conf, c.Backend.Name)
			if err != nil {
				return err
			}
			hcm.Tracing.Provider = tracing
		case mesh_proto.TracingDatadogType:
			tracing, err := datadogConfig(c.Backend.Conf, c.Backend.Name, c.Service, c.TrafficDirection, c.Destination)
			if err != nil {
				return err
			}
			hcm.Tracing.Provider = tracing
		}
		return nil
	})
}

func datadogConfig(cfgStr *structpb.Struct, backendName string, serviceName string, direction envoy_common.TrafficDirection, destination string) (*envoy_trace.Tracing_Http, error) {
	cfg := mesh_proto.DatadogTracingBackendConfig{}
	if err := proto.ToTyped(cfgStr, &cfg); err != nil {
		return nil, errors.Wrap(err, "could not convert backend")
	}

	datadogConfig := envoy_trace.DatadogConfig{
		CollectorCluster: names.GetTracingClusterName(backendName),
		ServiceName:      createDatadogServiceName(&cfg, serviceName, direction, destination),
	}
	datadogConfigAny, err := proto.MarshalAnyDeterministic(&datadogConfig)
	if err != nil {
		return nil, err
	}
	tracingConfig := &envoy_trace.Tracing_Http{
		Name: "envoy.tracers.datadog",
		ConfigType: &envoy_trace.Tracing_Http_TypedConfig{
			TypedConfig: datadogConfigAny,
		},
	}
	return tracingConfig, nil
}

func zipkinConfig(cfgStr *structpb.Struct, backendName string) (*envoy_trace.Tracing_Http, error) {
	cfg := mesh_proto.ZipkinTracingBackendConfig{}
	if err := proto.ToTyped(cfgStr, &cfg); err != nil {
		return nil, errors.Wrap(err, "could not convert backend")
	}
	url, err := net_url.ParseRequestURI(cfg.Url)
	if err != nil {
		return nil, errors.Wrap(err, "invalid URL of Zipkin")
	}

	zipkinConfig := envoy_trace.ZipkinConfig{
		CollectorCluster:         names.GetTracingClusterName(backendName),
		CollectorEndpoint:        url.Path,
		TraceId_128Bit:           pointer.Deref(cfg.TraceId128Bit).Value,
		CollectorEndpointVersion: apiVersion(&cfg, url),
		SharedSpanContext:        cfg.SharedSpanContext,
		CollectorHostname:        url.Host,
	}
	zipkinConfigAny, err := proto.MarshalAnyDeterministic(&zipkinConfig)
	if err != nil {
		return nil, err
	}
	tracingConfig := &envoy_trace.Tracing_Http{
		Name: "envoy.tracers.zipkin",
		ConfigType: &envoy_trace.Tracing_Http_TypedConfig{
			TypedConfig: zipkinConfigAny,
		},
	}
	return tracingConfig, nil
}

func apiVersion(zipkin *mesh_proto.ZipkinTracingBackendConfig, url *net_url.URL) envoy_trace.ZipkinConfig_CollectorEndpointVersion {
	if zipkin.ApiVersion == "" { // try to infer it from the URL
		if url.Path == "/api/v2/spans" {
			return envoy_trace.ZipkinConfig_HTTP_JSON
		}
	} else {
		switch zipkin.ApiVersion {
		case "httpJson":
			return envoy_trace.ZipkinConfig_HTTP_JSON
		case "httpProto":
			return envoy_trace.ZipkinConfig_HTTP_PROTO
		}
	}
	return envoy_trace.ZipkinConfig_HTTP_JSON
}

func createDatadogServiceName(datadog *mesh_proto.DatadogTracingBackendConfig, serviceName string, direction envoy_common.TrafficDirection, destination string) string {
	if pointer.Deref(datadog.SplitService).Value {
		var datadogServiceName []string
		switch direction {
		case envoy_common.TrafficDirectionInbound:
			datadogServiceName = []string{serviceName, string(direction)}
		case envoy_common.TrafficDirectionOutbound:
			datadogServiceName = []string{serviceName, string(direction), destination}
		default:
			return serviceName
		}
		return strings.Join(datadogServiceName, "_")
	} else {
		return serviceName
	}
}
