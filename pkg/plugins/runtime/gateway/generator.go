package gateway

import (
	"context"
	"fmt"
	"sort"

	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway/match"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway/merge"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	envoy_names "github.com/kumahq/kuma/pkg/xds/envoy/names"
	envoy_routes "github.com/kumahq/kuma/pkg/xds/envoy/routes"
)

const WildcardHostname = "*"

type GatewayHost struct {
	Hostname string
	Routes   []model.Resource
	// TODO(jpeach) Track TLS state for this host.
}

// Resources tracks partially-built xDS resources that can be updated
// by multiple gateway generators.
type Resources struct {
	Listener           *listeners.ListenerBuilder
	RouteConfiguration *envoy_routes.RouteConfigurationBuilder
	VirtualHost        *envoy_routes.VirtualHostBuilder

	discriminator int
}

// Discriminator returns a unique ID that can be used to disambiguate
// resources that would otherwise have conflicting names.
//
// TODO(jpeach) Whether a global discriminator makes sense probably
// depends on whether we end up processing resources in deterministic
// orders.
func (r *Resources) Discriminator() int {
	r.discriminator++
	return r.discriminator
}

type GatewayListener struct {
	Port         uint32
	Protocol     mesh_proto.Gateway_Listener_Protocol
	ResourceName string
}

type GatewayResourceInfo struct {
	Proxy     *core_xds.Proxy
	Dataplane *core_mesh.DataplaneResource
	Gateway   *core_mesh.GatewayResource
	Listener  GatewayListener
	Host      GatewayHost
	Resources Resources
}

// GatewayHostGenerator is responsible for generating xDS resources for a single GatewayHost.
type GatewayHostGenerator interface {
	GenerateHost(xds_context.Context, *GatewayResourceInfo) (*core_xds.ResourceSet, error)
	SupportsProtocol(mesh_proto.Gateway_Listener_Protocol) bool
}

// Generator generates xDS resources for an entire Gateway.
type Generator struct {
	ResourceManager core_manager.ReadOnlyResourceManager
	Generators      []GatewayHostGenerator
}

func (g Generator) Generate(ctx xds_context.Context, proxy *core_xds.Proxy) (*core_xds.ResourceSet, error) {
	mesh := ctx.Mesh.Resource.Meta.GetName()
	manager := match.ManagerForMesh(g.ResourceManager, mesh)
	gateway := match.Gateway(manager, proxy.Dataplane)

	if gateway == nil {
		log.V(1).Info("no matching gateway for dataplane",
			"name", proxy.Dataplane.Meta.GetName(),
			"mesh", proxy.Dataplane.Meta.GetMesh(),
			"service", proxy.Dataplane.Spec.GetIdentifyingService(),
		)

		return nil, nil
	}

	log.V(1).Info(fmt.Sprintf("matched gateway %q to dataplane %q",
		gateway.Meta.GetName(), proxy.Dataplane.Meta.GetName()))

	// Canonicalize the tags on each listener to be the merged resources
	// of dataplane, gateway and listener tags.
	for _, listener := range gateway.Spec.GetConf().GetListeners() {
		listener.Tags = match.MergeSelectors(
			proxy.Dataplane.Spec.GetNetworking().GetGateway().GetTags(),
			gateway.Spec.GetTags(),
			listener.GetTags(),
		)
	}

	// Multiple listener specifications can have the same port. If
	// they are compatible, then we can collapse those specifications
	// down to a single listener.
	collapsed := map[uint32][]*mesh_proto.Gateway_Listener{}
	for _, ep := range gateway.Spec.GetConf().GetListeners() {
		collapsed[ep.GetPort()] = append(collapsed[ep.GetPort()], ep)
	}

	resources := ResourceAggregator{core_xds.NewResourceSet()}

	for port, listeners := range collapsed {
		// Force all listeners on the same port to have the same protocol.
		for i := range listeners {
			if listeners[i].GetProtocol() != listeners[0].GetProtocol() {
				return nil, errors.Errorf(
					"cannot collapse listener protocols %s and %s on port %d",
					listeners[i].GetProtocol(), listeners[0].GetProtocol(), port,
				)
			}
		}

		listener, hosts, err := MakeGatewayListener(manager, gateway, listeners)
		if err != nil {
			return nil, err
		}

		hosts = RedistributeWildcardRoutes(hosts)

		// Sort by reverse hostname, so that fully qualified hostnames sort
		// before wildcard domains, and "*" is last.
		sort.Slice(hosts, func(i, j int) bool {
			return hosts[i].Hostname > hosts[j].Hostname
		})

		info := GatewayResourceInfo{
			Proxy:     proxy,
			Dataplane: proxy.Dataplane,
			Gateway:   gateway,
			Listener:  listener,
		}

		// Make a pass over the generators for each virtual host.
		for _, host := range hosts {
			info.Host = host

			// Ensure that generators don't get duplicate routes,
			// which could happen after redistributing wildcards.
			info.Host.Routes = merge.UniqueResources(info.Host.Routes)

			for _, generator := range g.Generators {
				if !generator.SupportsProtocol(listener.Protocol) {
					continue
				}

				generated, err := generator.GenerateHost(ctx, &info)
				if err != nil {
					return nil, errors.Wrapf(err, "%T failed to generate resources for dataplane %q",
						generator, proxy.Id)
				}

				resources.AddSet(generated)
			}

			info.Resources.RouteConfiguration.Configure(envoy_routes.VirtualHost(info.Resources.VirtualHost))

			// The order of generators matters, because they
			// can refer to state that is accumulated in the
			// GatewayResourceInfo. Clear the virtual host so
			// that we can't accidentally refer to a stale one.
			info.Resources.VirtualHost = nil
		}

		if err := resources.Add(BuildResourceSet(info.Resources.Listener)); err != nil {
			return nil, errors.Wrapf(err, "failed to build listener resource")
		}

		if err := resources.Add(BuildResourceSet(info.Resources.RouteConfiguration)); err != nil {
			return nil, errors.Wrapf(err, "failed to build route configuration resource")
		}
	}

	return resources.Get(), nil
}

// MakeGatewayListener converts a collapsed set of listener configurations
// in to a single configuration with a matched set of route resources. The
// given listeners must have a consistent protocol and port.
func MakeGatewayListener(
	manager *match.MeshedResourceManager,
	gateway *core_mesh.GatewayResource,
	listeners []*mesh_proto.Gateway_Listener,
) (GatewayListener, []GatewayHost, error) {
	hostsByName := map[string]GatewayHost{}
	resourcesByType := map[model.ResourceType]model.ResourceList{}

	listener := GatewayListener{
		Port:     listeners[0].GetPort(),
		Protocol: listeners[0].GetProtocol(),
		ResourceName: envoy_names.GetGatewayListenerName(
			gateway.Meta.GetName(),
			listeners[0].GetProtocol().String(),
			listeners[0].GetPort(),
		),
	}

	for _, t := range []model.ResourceType{core_mesh.TrafficRouteType, core_mesh.GatewayRouteType} {
		list, err := registry.Global().NewList(t)
		if err != nil {
			return listener, nil, err
		}

		if err := manager.List(context.Background(), list); err != nil {
			return listener, nil, err
		}

		resourcesByType[t] = list
	}

	// We don't require hostnames to be unique across listeners. As
	// long as the port and protocol matches it is OK to have multiple
	// listener entries for the same hostname, since each entry can have
	// separate tags that will select additional route resources.
	//
	// This will become a problem when multiple listeners specify
	// TLS certificates, so at that point, we might walk all this back.
	for _, l := range listeners {
		// An empty hostname is the same as "*", i.e. matches all hosts.
		hostname := l.GetHostname()
		if hostname == "" {
			hostname = WildcardHostname
		}

		host := hostsByName[hostname]
		host.Hostname = hostname

		switch listener.Protocol {
		case mesh_proto.Gateway_Listener_HTTP,
			mesh_proto.Gateway_Listener_HTTPS:
			host.Routes = append(host.Routes,
				match.Routes(resourcesByType[core_mesh.TrafficRouteType], l.GetTags())...)
			host.Routes = append(host.Routes,
				match.Routes(resourcesByType[core_mesh.GatewayRouteType], l.GetTags())...)
		default:
			// TODO(jpeach) match other route types that are appropriate to the protocol.
		}

		// TODO(jpeach) bind the listener tags to each route so
		// that generators can use the appropriate set of tags to
		// match route policies.

		hostsByName[hostname] = host
	}

	var hosts []GatewayHost
	for _, host := range hostsByName {
		hosts = append(hosts, host)
	}

	return listener, hosts, nil
}

// RedistributeWildcardRoutes takes the routes from the wildcard host
// and redistributes them to hosts with matching names, creating new
// hosts if necessary.
//
// This process is necessary because:
//
// 1. We might have a listener with hostname A and some routes, but also
//    a wildcard listener with routes for hostname A. We want all the routes
//    for hostname A in the same virtual host.
// 2. Routes with hostnames that are attached to a wildcard listener
//    should implicitly create virtual hosts so that we can generate a
//    consistent config. For example, if a wildcard listener has a route for
//    hostname A and a route for hostname B, that doesn't mean that the routes
//    are for hostnames A or B. We still want the routes to match the hostname
//    that they were specified with.
func RedistributeWildcardRoutes(
	hosts []GatewayHost,
) []GatewayHost {
	hostsByName := map[string]GatewayHost{}

	for _, h := range hosts {
		hostsByName[h.Hostname] = h
	}

	wild, ok := hostsByName[WildcardHostname]
	if !ok {
		return hosts
	}

	wildcardRoutes := wild.Routes
	wild.Routes = nil // We are rebuilding this.
	for _, r := range wildcardRoutes {
		gw, ok := r.(*core_mesh.GatewayRouteResource)
		if !ok {
			continue
		}

		names := gw.Spec.GetConf().GetHttp().GetHostnames()

		// No hostnames on this route, it stays as a wildcard route.
		if len(names) == 0 {
			wild.Routes = append(wild.Routes, r)
			continue
		}

		for _, n := range names {
			// Note that if we already have a virtualhost for this
			// name, and add the route to it, it might be a duplicate.
			host := hostsByName[n]
			host.Hostname = n
			host.Routes = append(host.Routes, r)
			hostsByName[n] = host
		}
	}

	hostsByName[WildcardHostname] = wild

	var flattened []GatewayHost
	for _, host := range hostsByName {
		flattened = append(flattened, host)
	}

	return flattened
}
