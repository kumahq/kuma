package generator

import (
	"fmt"
	net_url "net/url"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/envoy/clusters"
	"github.com/kumahq/kuma/pkg/xds/envoy/names"
)

// OriginTracing is a marker to indicate by which ProxyGenerator resources were generated.
const OriginTracing = "tracing"

type TracingProxyGenerator struct {
}

var _ ResourceGenerator = TracingProxyGenerator{}

func (t TracingProxyGenerator) Generate(_ xds_context.Context, proxy *core_xds.Proxy) (*core_xds.ResourceSet, error) {
	if proxy.Policies.TracingBackend == nil {
		return nil, nil
	}
	resources := core_xds.NewResourceSet()
	switch proxy.Policies.TracingBackend.Type {
	case mesh_proto.TracingZipkinType:
		res, err := t.zipkinCluster(proxy.Policies.TracingBackend, proxy.APIVersion)
		if err != nil {
			return nil, errors.Wrap(err, "could not generate zipkin cluster")
		}
		resources.Add(res)
	case mesh_proto.TracingDatadogType:
		res, err := t.datadogCluster(proxy.Policies.TracingBackend, proxy.APIVersion)
		if err != nil {
			return nil, errors.Wrap(err, "could not generate datadog cluster")
		}
		resources.Add(res)
	}

	return resources, nil
}

func (t TracingProxyGenerator) zipkinCluster(backend *mesh_proto.TracingBackend, apiVersion envoy.APIVersion) (*core_xds.Resource, error) {
	cfg := mesh_proto.ZipkinTracingBackendConfig{}
	if err := proto.ToTyped(backend.Conf, &cfg); err != nil {
		return nil, errors.Wrap(err, "could not convert backend")
	}
	url, err := net_url.ParseRequestURI(cfg.Url)
	if err != nil {
		return nil, errors.Wrap(err, "invalid URL of Zipkin")
	}
	port, err := strconv.Atoi(url.Port())
	if err != nil {
		return nil, err
	}

	clusterName := names.GetTracingClusterName(backend.Name)
	cluster, err := clusters.NewClusterBuilder(apiVersion).
		Configure(clusters.DNSCluster(clusterName, url.Hostname(), uint32(port))).
		Build()
	if err != nil {
		return nil, err
	}

	return &core_xds.Resource{
		Name:     clusterName,
		Origin:   OriginTracing,
		Resource: cluster,
	}, nil
}

func (t TracingProxyGenerator) datadogCluster(backend *mesh_proto.TracingBackend, apiVersion envoy.APIVersion) (*core_xds.Resource, error) {
	cfg := mesh_proto.DatadogTracingBackendConfig{}
	if err := proto.ToTyped(backend.Conf, &cfg); err != nil {
		return nil, errors.Wrap(err, "could not convert backend")
	}

	var cluster envoy.NamedResource
	var err error

	clusterName := names.GetTracingClusterName(backend.Name)

	if strings.HasPrefix(cfg.Address, "unix:") {
		prefix := ""
		if cfg.HostVolumeRelative {
			prefix = "/datadog-mount"
		}
		socket_path := fmt.Sprintf("%s/%s", prefix, strings.TrimPrefix(cfg.Address, "unix:"))
		cluster, err = clusters.NewClusterBuilder(apiVersion).
			Configure(clusters.StaticClusterUnixSocket(clusterName, socket_path)).
			Build()
		if err != nil {
			return nil, errors.Wrap(err, "cannot create cluster")
		}
	} else {
		elms := strings.Split(cfg.Address, ":")
		if len(elms) < 2 {
			return nil, errors.Errorf("invalid Datadog agent address '%s'", cfg.Address)
		}
		port_str := elms[len(elms)-1]
		port, err := strconv.ParseUint(port_str, 10, 16)
		if err != nil {
			return nil, errors.Wrap(err, "invalid Datadog port")
		}

		address := strings.TrimSuffix(cfg.Address, fmt.Sprintf(":%s", port_str))

		cluster, err = clusters.NewClusterBuilder(apiVersion).
			Configure(clusters.DNSCluster(clusterName, address, uint32(port))).
			Build()
		if err != nil {
			return nil, errors.Wrap(err, "cannot create cluster")
		}
	}

	return &core_xds.Resource{
		Name:     clusterName,
		Origin:   OriginTracing,
		Resource: cluster,
	}, nil
}
