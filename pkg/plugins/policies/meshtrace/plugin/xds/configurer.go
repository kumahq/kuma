package xds

import (
	net_url "net/url"
	"strings"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_trace "github.com/envoyproxy/go-control-plane/envoy/config/trace/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	tracingv3 "github.com/envoyproxy/go-control-plane/envoy/type/tracing/v3"
	envoy_type "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"k8s.io/apimachinery/pkg/util/intstr"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/kri"
	core_system_names "github.com/kumahq/kuma/pkg/core/system_names"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshtrace/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/util/pointer"
	"github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	v3 "github.com/kumahq/kuma/pkg/xds/envoy/listeners/v3"
)

type Configurer struct {
	Conf api.Conf

	// Opaque string which envoy will assign to tracer collector cluster, on those
	// which support association of named "service" tags on traces. Consumed by datadog.
	Service               string
	TrafficDirection      envoy_core.TrafficDirection
	Destination           string
	IsGateway             bool
	UnifiedResourceNaming bool
	KriWithoutSection     *kri.Identifier
}

var _ v3.FilterChainConfigurer = &Configurer{}

const (
	ZipkinProviderName        = "zipkin"
	DatadogProviderName       = "datadog"
	OpenTelemetryProviderName = "opentelemetry"
)

func (c *Configurer) Configure(filterChain *envoy_listener.FilterChain) error {
	var backend api.Backend
	if backends := pointer.Deref(c.Conf.Backends); len(backends) == 0 {
		return nil
	} else {
		backend = backends[0]
	}
	getNameOrDefault := core_system_names.GetNameOrDefault((c.UnifiedResourceNaming) && c.KriWithoutSection != nil)

	return v3.UpdateHTTPConnectionManager(filterChain, func(hcm *envoy_hcm.HttpConnectionManager) error {
		hcm.Tracing = &envoy_hcm.HttpConnectionManager_Tracing{
			SpawnUpstreamSpan: wrapperspb.Bool(c.IsGateway),
		}

		if c.Conf.Sampling != nil {
			if overall := c.Conf.Sampling.Overall; overall != nil {
				percent, err := c.envoyPercent(*overall)
				if err != nil {
					return err
				}
				hcm.Tracing.OverallSampling = percent
			}
			if client := c.Conf.Sampling.Client; client != nil {
				percent, err := c.envoyPercent(*client)
				if err != nil {
					return err
				}
				hcm.Tracing.ClientSampling = percent
			}
			if random := c.Conf.Sampling.Random; random != nil {
				percent, err := c.envoyPercent(*random)
				if err != nil {
					return err
				}
				hcm.Tracing.RandomSampling = percent
			}
		}

		if c.Conf.Tags != nil {
			hcm.Tracing.CustomTags = mapTags(pointer.Deref(c.Conf.Tags))
		}

		if backend.Zipkin != nil {
			name := getNameOrDefault(
				core_system_names.AsSystemName(kri.WithSectionName(pointer.Deref(c.KriWithoutSection), core_system_names.CleanName(backend.Zipkin.Url))),
				GetTracingClusterName(ZipkinProviderName),
			)
			tracing, err := c.zipkinConfig(name)
			if err != nil {
				return err
			}
			hcm.Tracing.Provider = tracing
		}

		if backend.Datadog != nil {
			name := getNameOrDefault(
				core_system_names.AsSystemName(kri.WithSectionName(pointer.Deref(c.KriWithoutSection), core_system_names.CleanName(backend.Datadog.Url))),
				GetTracingClusterName(DatadogProviderName),
			)
			tracing, err := c.datadogConfig(name)
			if err != nil {
				return err
			}
			hcm.Tracing.Provider = tracing
		}

		if backend.OpenTelemetry != nil {
			name := getNameOrDefault(
				core_system_names.AsSystemName(kri.WithSectionName(pointer.Deref(c.KriWithoutSection), core_system_names.CleanName(backend.OpenTelemetry.Endpoint))),
				GetTracingClusterName(OpenTelemetryProviderName),
			)
			tracing, err := c.opentelemetryConfig(name)
			if err != nil {
				return err
			}
			hcm.Tracing.Provider = tracing
		}

		return nil
	})
}

func (c *Configurer) envoyPercent(intOrStr intstr.IntOrString) (*envoy_type.Percent, error) {
	decimal, err := common_api.NewDecimalFromIntOrString(intOrStr)
	if err != nil {
		return nil, err
	}
	value, _ := decimal.Float64()
	return &envoy_type.Percent{
		Value: value,
	}, nil
}

func (c *Configurer) datadogConfig(clusterName string) (*envoy_trace.Tracing_Http, error) {
	datadogConfig := envoy_trace.DatadogConfig{
		CollectorCluster: clusterName,
		ServiceName:      c.createDatadogServiceName(),
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

func (c *Configurer) opentelemetryConfig(clusterName string) (*envoy_trace.Tracing_Http, error) {
	otelConfig := envoy_trace.OpenTelemetryConfig{
		ServiceName: c.Service,
		GrpcService: &envoy_core.GrpcService{
			TargetSpecifier: &envoy_core.GrpcService_EnvoyGrpc_{
				EnvoyGrpc: &envoy_core.GrpcService_EnvoyGrpc{
					ClusterName: clusterName,
				},
			},
		},
	}
	otelConfigAny, err := proto.MarshalAnyDeterministic(&otelConfig)
	if err != nil {
		return nil, err
	}
	tracingConfig := &envoy_trace.Tracing_Http{
		Name: "envoy.tracers.opentelemetry",
		ConfigType: &envoy_trace.Tracing_Http_TypedConfig{
			TypedConfig: otelConfigAny,
		},
	}
	return tracingConfig, nil
}

func (c *Configurer) zipkinConfig(clusterName string) (*envoy_trace.Tracing_Http, error) {
	zipkin := pointer.Deref(c.Conf.Backends)[0].Zipkin
	url, err := net_url.ParseRequestURI(zipkin.Url)
	if err != nil {
		return nil, errors.Wrap(err, "invalid URL of Zipkin")
	}

	ssc := wrapperspb.Bool(zipkin.SharedSpanContext)
	zipkinConfig := envoy_trace.ZipkinConfig{
		CollectorCluster:         clusterName,
		CollectorEndpoint:        url.Path,
		TraceId_128Bit:           zipkin.TraceId128Bit,
		CollectorEndpointVersion: apiVersion(zipkin),
		SharedSpanContext:        ssc,
		CollectorHostname:        url.Host,
		SplitSpansForRequest:     c.IsGateway,
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

func (c *Configurer) createDatadogServiceName() string {
	datadog := pointer.Deref(c.Conf.Backends)[0].Datadog

	if datadog.SplitService {
		var datadogServiceName []string
		switch c.TrafficDirection {
		case envoy_core.TrafficDirection_INBOUND:
			datadogServiceName = []string{c.Service, string(envoy.TrafficDirectionInbound)}
		case envoy_core.TrafficDirection_OUTBOUND:
			datadogServiceName = []string{c.Service, string(envoy.TrafficDirectionOutbound), c.Destination}
		default:
			return c.Service
		}
		return strings.Join(datadogServiceName, "_")
	} else {
		return c.Service
	}
}

func apiVersion(zipkin *api.ZipkinBackend) envoy_trace.ZipkinConfig_CollectorEndpointVersion {
	switch zipkin.ApiVersion {
	case "httpJson":
		return envoy_trace.ZipkinConfig_HTTP_JSON
	case "httpProto":
		return envoy_trace.ZipkinConfig_HTTP_PROTO
	default:
		return envoy_trace.ZipkinConfig_HTTP_JSON
	}
}

func mapTags(tags []api.Tag) []*tracingv3.CustomTag {
	var customTags []*tracingv3.CustomTag

	for _, tag := range tags {
		if tag.Header != nil {
			customTags = append(customTags, mapHeaderTag(tag.Name, tag.Header))
		} else {
			customTags = append(customTags, mapLiteralTag(tag.Name, pointer.Deref(tag.Literal)))
		}
	}

	return customTags
}

func mapLiteralTag(name, literal string) *tracingv3.CustomTag {
	return &tracingv3.CustomTag{
		Tag: name,
		Type: &tracingv3.CustomTag_Literal_{
			Literal: &tracingv3.CustomTag_Literal{
				Value: literal,
			},
		},
	}
}

func mapHeaderTag(name string, header *api.HeaderTag) *tracingv3.CustomTag {
	return &tracingv3.CustomTag{
		Tag: name,
		Type: &tracingv3.CustomTag_RequestHeader{
			RequestHeader: &tracingv3.CustomTag_Header{
				Name:         header.Name,
				DefaultValue: pointer.Deref(header.Default),
			},
		},
	}
}

func GetTracingClusterName(provider string) string {
	return "meshtrace:" + provider
}
