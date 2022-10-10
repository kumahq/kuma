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

	api "github.com/kumahq/kuma/pkg/plugins/policies/meshtrace/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	v3 "github.com/kumahq/kuma/pkg/xds/envoy/listeners/v3"
)

type Configurer struct {
	Conf *api.MeshTrace_Conf

	// Opaque string which envoy will assign to tracer collector cluster, on those
	// which support association of named "service" tags on traces. Consumed by datadog.
	Service          string
	TrafficDirection envoy_core.TrafficDirection
	Destination      string
}

var _ v3.FilterChainConfigurer = &Configurer{}

const (
	ZipkinProviderName  = "zipkin"
	DatadogProviderName = "datadog"
)

func (c *Configurer) Configure(filterChain *envoy_listener.FilterChain) error {
	if c.Conf.Backends[0] == nil {
		return nil
	}

	backend := c.Conf.Backends[0]

	return v3.UpdateHTTPConnectionManager(filterChain, func(hcm *envoy_hcm.HttpConnectionManager) error {
		hcm.Tracing = &envoy_hcm.HttpConnectionManager_Tracing{}

		if c.Conf.GetSampling() != nil {
			if c.Conf.GetSampling().GetOverall() != nil {
				hcm.Tracing.OverallSampling = &envoy_type.Percent{
					Value: float64(c.Conf.Sampling.Overall.Value),
				}
			}
			if c.Conf.GetSampling().GetClient() != nil {
				hcm.Tracing.ClientSampling = &envoy_type.Percent{
					Value: float64(c.Conf.Sampling.Client.Value),
				}
			}
			if c.Conf.GetSampling().GetRandom() != nil {
				hcm.Tracing.RandomSampling = &envoy_type.Percent{
					Value: float64(c.Conf.Sampling.Random.Value),
				}
			}
		}

		if c.Conf.Tags != nil {
			hcm.Tracing.CustomTags = mapTags(c.Conf.Tags)
		}

		if backend.GetZipkin() != nil {
			tracing, err := c.zipkinConfig(GetTracingClusterName(ZipkinProviderName))
			if err != nil {
				return err
			}
			hcm.Tracing.Provider = tracing
		}

		datadog := backend.GetDatadog()
		if datadog != nil {
			tracing, err := c.datadogConfig(GetTracingClusterName(DatadogProviderName))
			if err != nil {
				return err
			}
			hcm.Tracing.Provider = tracing
		}

		return nil
	})
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
		Name: "envoy.datadog",
		ConfigType: &envoy_trace.Tracing_Http_TypedConfig{
			TypedConfig: datadogConfigAny,
		},
	}
	return tracingConfig, nil
}

func (c *Configurer) zipkinConfig(clusterName string) (*envoy_trace.Tracing_Http, error) {
	zipkin := c.Conf.GetBackends()[0].GetZipkin()
	url, err := net_url.ParseRequestURI(zipkin.Url)
	if err != nil {
		return nil, errors.Wrap(err, "invalid URL of Zipkin")
	}

	zipkinConfig := envoy_trace.ZipkinConfig{
		CollectorCluster:         clusterName,
		CollectorEndpoint:        url.Path,
		TraceId_128Bit:           zipkin.TraceId128Bit,
		CollectorEndpointVersion: apiVersion(zipkin, url),
		SharedSpanContext:        zipkin.SharedSpanContext,
		CollectorHostname:        url.Host,
	}
	zipkinConfigAny, err := proto.MarshalAnyDeterministic(&zipkinConfig)
	if err != nil {
		return nil, err
	}
	tracingConfig := &envoy_trace.Tracing_Http{
		Name: "envoy.zipkin",
		ConfigType: &envoy_trace.Tracing_Http_TypedConfig{
			TypedConfig: zipkinConfigAny,
		},
	}
	return tracingConfig, nil
}

func (c *Configurer) createDatadogServiceName() string {
	datadog := c.Conf.GetBackends()[0].GetDatadog()

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

func apiVersion(zipkin *api.MeshTrace_ZipkinBackend, url *net_url.URL) envoy_trace.ZipkinConfig_CollectorEndpointVersion {
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

func mapTags(tags []*api.MeshTrace_Tag) []*tracingv3.CustomTag {
	var customTags []*tracingv3.CustomTag

	for _, tag := range tags {
		if tag.Header != nil {
			customTags = append(customTags, mapHeaderTag(tag.Name, tag.Header))
		} else {
			customTags = append(customTags, mapLiteralTag(tag.Name, tag.Literal))
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

func mapHeaderTag(name string, header *api.MeshTrace_HeaderTag) *tracingv3.CustomTag {
	return &tracingv3.CustomTag{
		Tag: name,
		Type: &tracingv3.CustomTag_RequestHeader{
			RequestHeader: &tracingv3.CustomTag_Header{
				Name:         header.Name,
				DefaultValue: header.Default,
			},
		},
	}
}

func GetTracingClusterName(provider string) string {
	return "meshtrace:" + provider
}
