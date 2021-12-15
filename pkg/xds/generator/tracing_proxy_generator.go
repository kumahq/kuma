package generator

import (
	net_url "net/url"
	"strconv"

	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/envoy/clusters"
	"github.com/kumahq/kuma/pkg/xds/envoy/names"
)

// OriginTracing is a marker to indicate by which ProxyGenerator resources were generated.
const OriginTracing = "tracing"

type TracingProxyGenerator struct {
}

var _ ResourceGenerator = TracingProxyGenerator{}

func (t TracingProxyGenerator) Generate(ctx xds_context.Context, proxy *core_xds.Proxy) (resources *core_xds.ResourceSet, err error) {
	tracingBackend := ctx.Mesh.GetTracingBackend(proxy.Policies.TrafficTrace)
	if tracingBackend == nil {
		return nil, nil
	}
	resources = core_xds.NewResourceSet()
	var endpoint *core_xds.Endpoint
	switch tracingBackend.Type {
	case mesh_proto.TracingZipkinType:
		cfg := mesh_proto.ZipkinTracingBackendConfig{}
		if err = proto.ToTyped(tracingBackend.Conf, &cfg); err != nil {
			return nil, errors.Wrap(err, "could not convert backend to zipkin")
		}
		endpoint, err = t.endpointForZipkin(&cfg)
		if err != nil {
			return nil, errors.Wrap(err, "could not generate zipkin cluster")
		}
	case mesh_proto.TracingDatadogType:
		cfg := mesh_proto.DatadogTracingBackendConfig{}
		if err = proto.ToTyped(tracingBackend.Conf, &cfg); err != nil {
			return nil, errors.Wrap(err, "could not convert backend to datadog")
		}
		endpoint, err = t.endpointForDatadog(&cfg)
		if err != nil {
			return nil, errors.Wrap(err, "could not generate datadog cluster")
		}
	}

	clusterName := names.GetTracingClusterName(tracingBackend.Name)
	res, err := clusters.NewClusterBuilder(proxy.APIVersion).
		Configure(clusters.ProvidedEndpointCluster(clusterName, proxy.Dataplane.IsIPv6(), *endpoint)).
		Configure(clusters.ClientSideTLS([]core_xds.Endpoint{*endpoint})).
		Build()
	if err != nil {
		return nil, err
	}
	resources.Add(&core_xds.Resource{Name: clusterName, Origin: OriginTracing, Resource: res})
	return resources, nil
}

func (t TracingProxyGenerator) endpointForZipkin(cfg *mesh_proto.ZipkinTracingBackendConfig) (*core_xds.Endpoint, error) {
	url, err := net_url.ParseRequestURI(cfg.Url)
	if err != nil {
		return nil, errors.Wrap(err, "invalid URL of Zipkin")
	}
	port, err := strconv.Atoi(url.Port())
	if err != nil {
		return nil, err
	}
	return &core_xds.Endpoint{
		Target: url.Hostname(),
		Port:   uint32(port),
		ExternalService: &core_xds.ExternalService{
			TLSEnabled:         url.Scheme == "https",
			AllowRenegotiation: true,
		},
	}, nil
}

func (t TracingProxyGenerator) endpointForDatadog(cfg *mesh_proto.DatadogTracingBackendConfig) (*core_xds.Endpoint, error) {
	if cfg.Port > 0xFFFF || cfg.Port < 1 {
		return nil, errors.Errorf("invalid Datadog port number %d. Must be in range 1-65535", cfg.Port)
	}
	return &core_xds.Endpoint{
		Target: cfg.Address,
		Port:   cfg.Port,
	}, nil
}
