package modifications

import (
	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/api/v2/route"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v2"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

type virtualHostModificator mesh_proto.ProxyTemplate_Modifications_VirtualHost

func (c *virtualHostModificator) apply(resources *core_xds.ResourceSet) error {
	virtualHost := &envoy_route.VirtualHost{}
	if err := util_proto.FromYAML([]byte(c.Value), virtualHost); err != nil {
		return err
	}

	for _, resource := range resources.Resources(envoy_resource.RouteType) {
		if c.routeConfigurationMatches(resource) {
			routeCfg := resource.Resource.(*envoy_api.RouteConfiguration)
			switch c.Operation {
			case mesh_proto.OpAdd:
				c.add(routeCfg, virtualHost)
			case mesh_proto.OpRemove:
				c.remove(routeCfg)
			case mesh_proto.OpPatch:
				c.patch(routeCfg, virtualHost)
			default:
				return errors.Errorf("invalid operation: %s", c.Operation)
			}
		}
	}
	return nil
}

func (c *virtualHostModificator) patch(routeCfg *envoy_api.RouteConfiguration, vHostPatch *envoy_route.VirtualHost) {
	for _, vHost := range routeCfg.VirtualHosts {
		if c.virtualHostMatches(vHost) {
			proto.Merge(vHost, vHostPatch)
		}
	}
}

func (c *virtualHostModificator) remove(routeCfg *envoy_api.RouteConfiguration) {
	var vHosts []*envoy_route.VirtualHost
	for _, vHost := range routeCfg.VirtualHosts {
		if !c.virtualHostMatches(vHost) {
			vHosts = append(vHosts, vHost)
		}
	}
	routeCfg.VirtualHosts = vHosts
}

func (c *virtualHostModificator) add(routeCfg *envoy_api.RouteConfiguration, vHost *envoy_route.VirtualHost) {
	routeCfg.VirtualHosts = append(routeCfg.VirtualHosts, vHost)
}

func (c *virtualHostModificator) virtualHostMatches(vHost *envoy_route.VirtualHost) bool {
	if c.Match.GetName() != "" && c.Match.GetName() != vHost.Name {
		return false
	}
	return true
}

func (c *virtualHostModificator) routeConfigurationMatches(routeCfg *core_xds.Resource) bool {
	if c.Match.GetOrigin() != "" && c.Match.GetOrigin() != routeCfg.Origin {
		return false
	}
	if c.Match.GetRouteConfigurationName() != "" && c.Match.GetRouteConfigurationName() != routeCfg.Name {
		return false
	}
	return true
}
