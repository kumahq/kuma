package egress

import (
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	envoy_names "github.com/kumahq/kuma/pkg/xds/envoy/names"
)

const (
	EgressProxy = "egress-proxy"

	// OriginEgress is a marker to indicate by which ProxyGenerator resources
	// were generated.
	OriginEgress = "egress"
)

var (
	log = core.Log.WithName("xds").WithName("generator").WithName("egress")
)

type Destination struct {
	Destination envoy_common.Tags
	Weight      uint32

	IsExternalService bool
}

type Listener struct {
	Address      string
	Port         uint32
	ResourceName string
}

// Resources tracks partially-built xDS resources that can be updated
// by multiple zoneegress generators.
type Resources struct {
	Listener *envoy_listeners.ListenerBuilder
}

type ResourceInfo struct {
	Proxy *core_xds.Proxy
	Mesh  *core_mesh.MeshResource

	Listener         Listener
	Resources        Resources
	Destinations     []*Destination
	EndpointMap      core_xds.EndpointMap
	ExternalServices []*core_mesh.ExternalServiceResource
	TrafficRoutes    []*core_mesh.TrafficRouteResource
	ZoneIngresses    []*core_mesh.ZoneIngressResource
}

// ZoneEgressGenerator is responsible for generating xDS resources for
// a single ZoneEgressHost.
type ZoneEgressGenerator interface {
	Generate(xds_context.Context, *ResourceInfo) (*core_xds.ResourceSet, error)
}

// Generator generates xDS resources for an entire ZoneEgress.
type Generator struct {
	Generators []ZoneEgressGenerator
}

// ResourceBuilder is an interface commonly implemented by complex Envoy
// configuration element builders.
// TODO (bartsmykla): DRY pkg/plugins/runtime/gateway/builder.go:11
type ResourceBuilder interface {
	Build() (envoy_common.NamedResource, error)
}

// BuildResourceSet is an adaptor that triggers the resource builder,
// b, to build its resource. If the builder is successful, the result is
// wrapped in a ResourceSet.
// TODO (bartsmykla): DRY pkg/plugins/runtime/gateway/builder.go:20
func BuildResourceSet(b ResourceBuilder) (*core_xds.ResourceSet, error) {
	resource, err := b.Build()
	if err != nil {
		return nil, err
	}

	if resource.GetName() == "" {
		return nil, errors.Errorf("anonymous resource %T", resource)
	}

	set := core_xds.NewResourceSet()
	set.Add(&core_xds.Resource{
		Name:     resource.GetName(),
		Origin:   OriginEgress,
		Resource: resource,
	})

	return set, nil
}

func makeListener(zoneEgress *core_mesh.ZoneEgressResource) Listener {
	networking := zoneEgress.Spec.GetNetworking()

	address := networking.GetAddress()
	port := networking.GetPort()

	return Listener{
		Port:         port,
		Address:      address,
		ResourceName: envoy_names.GetInboundListenerName(address, port),
	}
}

func routeDestinations(
	proxy *core_xds.Proxy,
	mesh *core_mesh.MeshResource,
) []*Destination {
	var destinations []*Destination

	meshName := mesh.GetMeta().GetName()
	zoneEgressProxy := proxy.ZoneEgressProxy
	endpointMap := zoneEgressProxy.MeshEndpointMap[meshName]

	for serviceName, endpoints := range endpointMap {
		for _, endpoint := range endpoints {
			tags := map[string]string{
				mesh_proto.ServiceTag: serviceName,
				"mesh":                meshName,
			}

			destinations = append(destinations, &Destination{
				Destination:       tags,
				Weight:            endpoint.Weight,
				IsExternalService: endpoint.IsExternalService(),
			})
		}
	}

	return destinations
}

func (g Generator) Generate(
	ctx xds_context.Context,
	proxy *core_xds.Proxy,
) (*core_xds.ResourceSet, error) {
	resources := core_xds.NewResourceSet()
	zoneEgressProxy := proxy.ZoneEgressProxy
	listener := makeListener(zoneEgressProxy.ZoneEgressResource)

	for _, mesh := range zoneEgressProxy.Meshes {
		meshName := mesh.GetMeta().GetName()

		info := ResourceInfo{
			Proxy:            proxy,
			Listener:         listener,
			Mesh:             mesh,
			EndpointMap:      zoneEgressProxy.MeshEndpointMap[meshName],
			ExternalServices: zoneEgressProxy.ExternalServiceMap[meshName],
			TrafficRoutes:    zoneEgressProxy.TrafficRouteMap[meshName],
			ZoneIngresses:    zoneEgressProxy.ZoneIngresses,
			Destinations:     routeDestinations(proxy, mesh),
		}

		for _, generator := range g.Generators {
			rs, err := generator.Generate(ctx, &info)
			if err != nil {
				err := errors.Wrapf(
					err,
					"%T failed to generate resources for zone egress %q",
					generator,
					proxy.Id,
				)
				return nil, err
			}

			resources.AddSet(rs)
		}

		rs, err := BuildResourceSet(info.Resources.Listener)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to build listener resource")
		}

		resources.AddSet(rs)
	}

	return resources, nil
}
