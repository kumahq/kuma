package v1alpha1

import (
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"

	"github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/pkg/core/xds/types"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/matchers"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	policies_xds "github.com/kumahq/kuma/pkg/plugins/policies/core/xds"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshpassthrough/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshpassthrough/plugin/xds"
	"github.com/kumahq/kuma/pkg/util/pointer"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

var _ core_plugins.PolicyPlugin = &plugin{}

type plugin struct{}

func NewPlugin() core_plugins.Plugin {
	return &plugin{}
}

func (p plugin) MatchedPolicies(dataplane *core_mesh.DataplaneResource, resources xds_context.Resources, opts ...core_plugins.MatchedPoliciesOption) (core_xds.TypedMatchingPolicies, error) {
	return matchers.MatchedPolicies(api.MeshPassthroughType, dataplane, resources, opts...)
}

func (p plugin) Apply(rs *core_xds.ResourceSet, ctx xds_context.Context, proxy *core_xds.Proxy) error {
	if proxy.Dataplane == nil {
		return nil
	}
	policies, ok := proxy.Policies.Dynamic[api.MeshPassthroughType]
	if !ok {
		return nil
	}
	if proxy.Dataplane.Spec.Networking.Gateway != nil && proxy.Dataplane.Spec.Networking.Gateway.Type == v1alpha1.Dataplane_Networking_Gateway_BUILTIN {
		policies.Warnings = append(policies.Warnings, "policy doesn't support builtin gateway")
		return nil
	}
	if proxy.Dataplane != nil && (proxy.Dataplane.Spec.Networking.TransparentProxying == nil || proxy.Metadata.HasFeature(xds_types.FeatureBindOutbounds)) {
		policies.Warnings = append(policies.Warnings, "policy doesn't support proxy running without transparent-proxy")
		return nil
	}
	listeners := policies_xds.GatherListeners(rs)
	if err := applyToOutboundPassthrough(ctx, rs, policies.SingleItemRules, listeners, proxy); err != nil {
		return err
	}
	return nil
}

func applyToOutboundPassthrough(
	ctx xds_context.Context,
	rs *core_xds.ResourceSet,
	rules core_rules.SingleItemRules,
	listeners policies_xds.Listeners,
	proxy *core_xds.Proxy,
) error {
	if len(rules.Rules) == 0 {
		return nil
	}
	rawConf := rules.Rules[0].Conf
	conf := rawConf.(api.Conf)

	// todo: this should be handled by "base policy"
	if pointer.Deref(conf.PassthroughMode) == "" {
		conf.PassthroughMode = pointer.To[api.PassthroughMode]("Matched")
	}

	if disableDefaultPassthrough(conf, ctx.Mesh.Resource.Spec.IsPassthrough()) {
		// remove clusters because they were added in TransparentProxyGenerator
		removeDefaultPassthroughCluster(rs)
		return nil
	}
	if enableDefaultPassthrough(conf, ctx.Mesh.Resource.Spec.IsPassthrough()) {
		// add clusters because they were not added in TransparentProxyGenerator
		return addDefaultPassthroughClusters(rs, proxy.APIVersion)
	}
	if ctx.Mesh.Resource.Spec.IsPassthrough() && conf.PassthroughMode != nil && pointer.Deref(conf.PassthroughMode) == api.PassthroughMode("All") {
		// clusters were added in TransparentProxyGenerator, do nothing
		return nil
	}

	if conf.PassthroughMode != nil && pointer.Deref(conf.PassthroughMode) == api.PassthroughMode("Matched") || conf.PassthroughMode == nil {
		removeDefaultPassthroughCluster(rs)
		if len(pointer.Deref(conf.AppendMatch)) > 0 {
			configurer := xds.Configurer{
				APIVersion:        proxy.APIVersion,
				InternalAddresses: proxy.InternalAddresses,
				Conf:              conf,
			}
			err := configurer.Configure(listeners.Ipv4Passthrough, listeners.Ipv6Passthrough, rs)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func removeDefaultPassthroughCluster(rs *core_xds.ResourceSet) {
	rs.Remove(envoy_resource.ClusterType, generator.OutboundNameIPv4)
	rs.Remove(envoy_resource.ClusterType, generator.OutboundNameIPv6)
}

func addDefaultPassthroughClusters(rs *core_xds.ResourceSet, apiVersion core_xds.APIVersion) error {
	outboundPassThroughCluster, err := xds.CreateCluster(apiVersion, generator.OutboundNameIPv4, core_mesh.ProtocolTCP)
	if err != nil {
		return err
	}
	rs.Add(&core_xds.Resource{
		Name:     outboundPassThroughCluster.GetName(),
		Origin:   generator.OriginTransparent,
		Resource: outboundPassThroughCluster,
	})
	outboundPassThroughCluster, err = xds.CreateCluster(apiVersion, generator.OutboundNameIPv6, core_mesh.ProtocolTCP)
	if err != nil {
		return err
	}
	rs.Add(&core_xds.Resource{
		Name:     outboundPassThroughCluster.GetName(),
		Origin:   generator.OriginTransparent,
		Resource: outboundPassThroughCluster,
	})
	return nil
}

func disableDefaultPassthrough(conf api.Conf, meshPassthroughEnabled bool) bool {
	return meshPassthroughEnabled && conf.PassthroughMode != nil && pointer.Deref[api.PassthroughMode](conf.PassthroughMode) == api.PassthroughMode("None")
}

func enableDefaultPassthrough(conf api.Conf, meshPassthroughEnabled bool) bool {
	return !meshPassthroughEnabled && conf.PassthroughMode != nil && pointer.Deref[api.PassthroughMode](conf.PassthroughMode) == api.PassthroughMode("All")
}
