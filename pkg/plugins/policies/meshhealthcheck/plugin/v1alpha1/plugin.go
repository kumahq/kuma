package v1alpha1

import (
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/kri"
	core_plugins "github.com/kumahq/kuma/v3/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/outbound"
	policies_xds "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/xds"
	api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshhealthcheck/api/v1alpha1"
	plugin_xds "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshhealthcheck/plugin/xds"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
)

var _ core_plugins.PolicyPlugin = &plugin{}

type plugin struct{}

func (p plugin) Order() int { return api.MeshHealthCheckResourceTypeDescriptor.Order }

func NewPlugin() core_plugins.Plugin {
	return &plugin{}
}

func (p plugin) Apply(rs *core_xds.ResourceSet, ctx xds_context.Context, proxy *core_xds.Proxy) error {
	policies, ok := proxy.Policies.Dynamic[api.MeshHealthCheckType]
	if !ok {
		return nil
	}

	if err := applyToRealResources(rs, policies.ToRules.ResourceRules, ctx.Mesh, labelTagSet(proxy.Dataplane)); err != nil {
		return err
	}

	return nil
}

// labelTagSet wraps a Dataplane's resource labels as a MultiValueTagSet so
// they can be templated into HTTP health check headers the same way inbound
// tags used to be.
func labelTagSet(dataplane *core_mesh.DataplaneResource) mesh_proto.MultiValueTagSet {
	data := map[string][]string{}
	for key, value := range dataplane.GetMeta().GetLabels() {
		data[key] = []string{value}
	}
	return mesh_proto.MultiValueTagSetFrom(data)
}

func applyToRealResources(rs *core_xds.ResourceSet, rules outbound.ResourceRules, meshCtx xds_context.MeshContext, tagSet mesh_proto.MultiValueTagSet, filters ...func(*core_xds.Resource) bool) error {
	return policies_xds.ForEachOutboundRule(rs, rules, meshCtx.Resources, func(uri kri.Identifier, conf api.Conf, resourcesByType core_xds.ResourcesByType) error {
		for typ, resources := range resourcesByType {
			if typ == envoy_resource.ClusterType {
				if err := configureClusters(resources, conf, tagSet); err != nil {
					return err
				}
			}
		}
		return nil
	}, filters...)
}

func configureClusters(resources []*core_xds.Resource, conf api.Conf, tagSet mesh_proto.MultiValueTagSet) error {
	for _, resource := range resources {
		configurer := plugin_xds.Configurer{
			Conf:     conf,
			Protocol: resource.Protocol,
			Tags:     tagSet,
		}
		err := configurer.Configure(resource.Resource.(*envoy_cluster.Cluster))
		if err != nil {
			return err
		}
	}
	return nil
}
