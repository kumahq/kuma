package v1alpha1

import (
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/kri"
	core_meta "github.com/kumahq/kuma/v3/pkg/core/metadata"
	core_plugins "github.com/kumahq/kuma/v3/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/v3/pkg/core/xds/types"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/matchers"
	core_rules "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/outbound"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/subsetutils"
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

func (p plugin) MatchedPolicies(dataplane *core_mesh.DataplaneResource, resources xds_context.Resources, opts ...core_plugins.MatchedPoliciesOption) (core_xds.TypedMatchingPolicies, error) {
	return matchers.MatchedPolicies(api.MeshHealthCheckType, dataplane, resources, opts...)
}

func (p plugin) Apply(rs *core_xds.ResourceSet, ctx xds_context.Context, proxy *core_xds.Proxy) error {
	policies, ok := proxy.Policies.Dynamic[api.MeshHealthCheckType]
	if !ok {
		return nil
	}

	clusters := policies_xds.GatherClusters(rs)

	if err := applyToOutbounds(policies.ToRules, clusters.Outbound, clusters.OutboundSplit, proxy.Outbounds, proxy.Dataplane, ctx.Mesh); err != nil {
		return err
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

func applyToOutbounds(
	rules core_rules.ToRules,
	outboundClusters map[string]*envoy_cluster.Cluster,
	outboundSplitClusters map[string][]*envoy_cluster.Cluster,
	outbounds xds_types.Outbounds,
	dataplane *core_mesh.DataplaneResource,
	meshCtx xds_context.MeshContext,
) error {
	targetedClusters := policies_xds.GatherTargetedClusters(
		outbounds,
		outboundSplitClusters,
		outboundClusters,
	)

	for cluster, serviceName := range targetedClusters {
		if err := configure(dataplane, rules.Rules, subsetutils.KumaServiceTagElement(serviceName), meshCtx.GetServiceProtocol(serviceName), cluster); err != nil {
			return err
		}
	}

	return nil
}

func configure(
	dataplane *core_mesh.DataplaneResource,
	rules core_rules.Rules,
	element subsetutils.Element,
	protocol core_meta.Protocol,
	cluster *envoy_cluster.Cluster,
) error {
	conf := core_rules.ComputeConf[api.Conf](rules, element)
	if conf == nil {
		return nil
	}

	configurer := plugin_xds.Configurer{
		Conf:     *conf,
		Protocol: protocol,
		Tags:     labelTagSet(dataplane),
	}

	if err := configurer.Configure(cluster); err != nil {
		return err
	}
	return nil
}

func applyToRealResource(meshCtx xds_context.MeshContext, rules outbound.ResourceRules, tagSet mesh_proto.MultiValueTagSet, uri kri.Identifier, resourcesByType core_xds.ResourcesByType) error {
	conf := rules.Compute(uri, meshCtx.Resources)
	if conf == nil {
		return nil
	}

	for typ, resources := range resourcesByType {
		if typ == envoy_resource.ClusterType {
			err := configureClusters(resources, conf.Conf[0].(api.Conf), tagSet)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func applyToRealResources(rs *core_xds.ResourceSet, rules outbound.ResourceRules, meshCtx xds_context.MeshContext, tagSet mesh_proto.MultiValueTagSet, filters ...func(*core_xds.Resource) bool) error {
	for uri, resType := range rs.IndexByOrigin(filters...) {
		if err := applyToRealResource(meshCtx, rules, tagSet, uri, resType); err != nil {
			return err
		}
	}
	return nil
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
