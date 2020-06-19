package bootstrap

import (
	net_url "net/url"
	"strconv"
	"time"

	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_api_v2_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoy_api_v2_endpoint "github.com/envoyproxy/go-control-plane/envoy/api/v2/endpoint"
	envoy_bootstrap "github.com/envoyproxy/go-control-plane/envoy/config/bootstrap/v2"
	envoy_config_trace_v2 "github.com/envoyproxy/go-control-plane/envoy/config/trace/v2"
	"github.com/golang/protobuf/ptypes/duration"
	structpb "github.com/golang/protobuf/ptypes/struct"
	"github.com/pkg/errors"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/util/proto"
)

func AddTracingConfig(bootstrap *envoy_bootstrap.Bootstrap, backend *mesh_proto.TracingBackend) error {
	if backend.GetType() == mesh_proto.TracingZipkinType {
		cluster, tracingCfg, err := zipkinConfig(bootstrap, backend.Conf, backend.Name)
		if err != nil {
			return err
		}
		if bootstrap.StaticResources == nil {
			bootstrap.StaticResources = &envoy_bootstrap.Bootstrap_StaticResources{}
		}
		bootstrap.StaticResources.Clusters = append(bootstrap.StaticResources.Clusters, cluster)
		bootstrap.Tracing = tracingCfg
	}
	return nil
}

func zipkinConfig(bootstrap *envoy_bootstrap.Bootstrap, cfgStr *structpb.Struct, backendName string) (*envoy_api.Cluster, *envoy_config_trace_v2.Tracing, error) {
	cfg := mesh_proto.ZipkinTracingBackendConfig{}
	if err := proto.ToTyped(cfgStr, &cfg); err != nil {
		return nil, nil, errors.Wrap(err, "could not convert backend")
	}
	url, err := net_url.ParseRequestURI(cfg.Url)
	if err != nil {
		return nil, nil, errors.Wrap(err, "invalid URL of Zipkin")
	}

	cluster, err := zipkinCluster(backendName, url)
	if err != nil {
		return nil, nil, err
	}

	zipkinConfig := envoy_config_trace_v2.ZipkinConfig{
		CollectorCluster:         cluster.Name,
		CollectorEndpoint:        url.Path,
		TraceId_128Bit:           cfg.TraceId128Bit,
		CollectorEndpointVersion: apiVersion(&cfg, url),
	}
	zipkinConfigAny, err := proto.MarshalAnyDeterministic(&zipkinConfig)
	if err != nil {
		return nil, nil, err
	}
	tracingConfig := &envoy_config_trace_v2.Tracing{
		Http: &envoy_config_trace_v2.Tracing_Http{
			Name: "envoy.zipkin",
			ConfigType: &envoy_config_trace_v2.Tracing_Http_TypedConfig{
				TypedConfig: zipkinConfigAny,
			},
		},
	}
	return cluster, tracingConfig, nil
}

func apiVersion(zipkin *mesh_proto.ZipkinTracingBackendConfig, url *net_url.URL) envoy_config_trace_v2.ZipkinConfig_CollectorEndpointVersion {
	if zipkin.ApiVersion == "" { // try to infer it from the URL
		if url.Path == "/api/v1/spans" {
			return envoy_config_trace_v2.ZipkinConfig_HTTP_JSON_V1
		} else if url.Path == "/api/v2/spans" {
			return envoy_config_trace_v2.ZipkinConfig_HTTP_JSON
		}
	} else {
		switch zipkin.ApiVersion {
		case "httpJsonV1":
			return envoy_config_trace_v2.ZipkinConfig_HTTP_JSON_V1
		case "httpJson":
			return envoy_config_trace_v2.ZipkinConfig_HTTP_JSON
		case "httpProto":
			return envoy_config_trace_v2.ZipkinConfig_HTTP_PROTO
		}
	}
	return envoy_config_trace_v2.ZipkinConfig_HTTP_JSON
}

const zipkinClusterTimeout = 10 * time.Second

func zipkinCluster(backendName string, url *net_url.URL) (*envoy_api.Cluster, error) {
	port, err := strconv.Atoi(url.Port())
	if err != nil {
		return nil, err
	}

	cluster := &envoy_api.Cluster{
		Name:                 backendName,
		ConnectTimeout:       &duration.Duration{Seconds: int64(zipkinClusterTimeout.Seconds())},
		ClusterDiscoveryType: &envoy_api.Cluster_Type{Type: envoy_api.Cluster_STRICT_DNS},
		LbPolicy:             envoy_api.Cluster_ROUND_ROBIN,
		LoadAssignment: &envoy_api.ClusterLoadAssignment{
			ClusterName: backendName,
			Endpoints: []*envoy_api_v2_endpoint.LocalityLbEndpoints{
				{
					LbEndpoints: []*envoy_api_v2_endpoint.LbEndpoint{
						{
							HostIdentifier: &envoy_api_v2_endpoint.LbEndpoint_Endpoint{
								Endpoint: &envoy_api_v2_endpoint.Endpoint{
									Address: &envoy_api_v2_core.Address{
										Address: &envoy_api_v2_core.Address_SocketAddress{
											SocketAddress: &envoy_api_v2_core.SocketAddress{
												Address: url.Hostname(),
												PortSpecifier: &envoy_api_v2_core.SocketAddress_PortValue{
													PortValue: uint32(port),
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	return cluster, nil
}
