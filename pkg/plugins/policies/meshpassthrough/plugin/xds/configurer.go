package xds

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"

	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshpassthrough/api/v1alpha1"
)

const OriginMeshPassthrough = "meshpassthrough"

type Configurer struct {
	APIVersion core_xds.APIVersion
	Conf       api.Conf
}

func (c Configurer) Configure(ipv4 *envoy_listener.Listener, _ *envoy_listener.Listener, rs *core_xds.ResourceSet) error {
	matcherConfigurer := FilterChainMatcherConfigurer{
		Conf: c.Conf,
	}
	clustersAccumulator := map[string]bool{}
	filterChainsToGenerate := matcherConfigurer.Configure(ipv4)
	for name, config := range filterChainsToGenerate {
		configurer := FilterChainConfigurer{
			Name:       name,
			Protocol:   mesh.ParseProtocol(config.Protocol),
			Routes:     config.Routes,
			APIVersion: c.APIVersion,
		}
		for _, route := range config.Routes {
			clustersAccumulator[route.ClusterName] = true
		}
		err := configurer.Configure(ipv4)
		if err != nil {
			return err
		}
	}
	for name := range clustersAccumulator {
		config, err := CreateCluster(c.APIVersion, name)
		if err != nil {
			return err
		}
		rs.Add(&core_xds.Resource{
			Name:     config.GetName(),
			Origin:   OriginMeshPassthrough,
			Resource: config,
		})
	}
	return nil
}
