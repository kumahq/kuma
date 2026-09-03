package v1alpha1

import (
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"

	"github.com/kumahq/kuma/v3/pkg/core/naming"
	core_plugins "github.com/kumahq/kuma/v3/pkg/core/plugins"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/v3/pkg/core/xds/types"
	core_rules "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules"
	policies_xds "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/xds"
	api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshpassthrough/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/meshpassthrough/plugin/xds"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
)

var _ core_plugins.PolicyPlugin = &plugin{}

type plugin struct{}

func (p plugin) Order() int { return api.MeshPassthroughResourceTypeDescriptor.Order }

func NewPlugin() core_plugins.Plugin {
	return &plugin{}
}

func (p plugin) Apply(rs *core_xds.ResourceSet, ctx xds_context.Context, proxy *core_xds.Proxy) error {
	if proxy.Dataplane == nil {
		return nil
	}
	policies, ok := proxy.Policies.Dynamic[api.MeshPassthroughType]
	if !ok {
		return nil
	}
	if !proxy.GetTransparentProxy().Enabled() || proxy.Metadata.HasFeature(xds_types.FeatureBindOutbounds) {
		policies.Warnings = append(policies.Warnings, "policy doesn't support proxy running without transparent-proxy")
		return nil
	}
	listeners := policies_xds.GatherListeners(rs)
	if err := applyToOutboundPassthrough(ctx, rs, policies.ProxyConf, listeners, proxy); err != nil {
		return err
	}
	return nil
}

func applyToOutboundPassthrough(
	_ xds_context.Context,
	rs *core_xds.ResourceSet,
	policyConf *core_rules.ProxyConf,
	listeners policies_xds.Listeners,
	proxy *core_xds.Proxy,
) error {
	if policyConf == nil {
		return nil
	}
	conf := policyConf.Conf.(api.Conf)

	// todo: this should be handled by "base policy"
	if pointer.Deref(conf.PassthroughMode) == "" {
		conf.PassthroughMode = pointer.To[api.PassthroughMode]("Matched")
	}

	if conf.PassthroughMode != nil && pointer.Deref(conf.PassthroughMode) == "None" {
		// remove clusters because they were added in TransparentProxyGenerator
		removeDefaultPassthroughCluster(rs)
		return nil
	}
	if conf.PassthroughMode != nil && pointer.Deref(conf.PassthroughMode) == "All" {
		// clusters were added in TransparentProxyGenerator, do nothing
		return nil
	}

	if conf.PassthroughMode != nil && pointer.Deref(conf.PassthroughMode) == "Matched" || conf.PassthroughMode == nil {
		removeDefaultPassthroughCluster(rs)
		if len(pointer.Deref(conf.AppendMatch)) > 0 {
			configurer := xds.Configurer{
				APIVersion:        proxy.APIVersion,
				InternalAddresses: proxy.InternalAddresses,
				Conf:              conf,
				IPv6Enabled:       proxy.Metadata.IPv6Enabled,
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
	rs.Remove(
		envoy_resource.ClusterType,
		naming.ContextualTransparentProxyName("outbound", 4),
	)
	rs.Remove(
		envoy_resource.ClusterType,
		naming.ContextualTransparentProxyName("outbound", 6),
	)
}
