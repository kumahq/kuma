package v1alpha1

import (
	"slices"

	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"

	"github.com/kumahq/kuma/v3/pkg/core"
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

var logger = core.Log.WithName("MeshPassthrough")

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
		addWarnings(proxy, policies, "policy doesn't support proxy running without transparent-proxy")
		return nil
	}
	listeners := policies_xds.GatherListeners(rs)
	warnings, err := applyToOutboundPassthrough(ctx, rs, policies.ProxyConf, listeners, proxy)
	if len(warnings) > 0 {
		// the same conflict reproduces on every proxy the policy matches and on
		// every reconciliation, so it stays at debug level like other dropped config
		logger.V(1).Info("some matches were dropped, they resolve to a filter chain another match already configures", "proxy", proxy.Id.String(), "warnings", warnings)
		addWarnings(proxy, policies, warnings...)
	}
	return err
}

// policies are kept in a map by value and the slice may be shared with the policy
// matching cache, so write an extended copy back instead of appending in place
func addWarnings(proxy *core_xds.Proxy, policies core_xds.TypedMatchingPolicies, warnings ...string) {
	if len(warnings) == 0 {
		return
	}
	policies.Warnings = append(slices.Clone(policies.Warnings), warnings...)
	proxy.Policies.Dynamic[api.MeshPassthroughType] = policies
}

func applyToOutboundPassthrough(
	_ xds_context.Context,
	rs *core_xds.ResourceSet,
	policyConf *core_rules.ProxyConf,
	listeners policies_xds.Listeners,
	proxy *core_xds.Proxy,
) ([]string, error) {
	if policyConf == nil {
		return nil, nil
	}
	conf := policyConf.Conf.(api.Conf)

	// todo: this should be handled by "base policy"
	if pointer.Deref(conf.PassthroughMode) == "" {
		conf.PassthroughMode = pointer.To[api.PassthroughMode]("Matched")
	}

	if conf.PassthroughMode != nil && pointer.Deref(conf.PassthroughMode) == "None" {
		// remove clusters because they were added in TransparentProxyGenerator
		removeDefaultPassthroughCluster(rs)
		return nil, nil
	}
	if conf.PassthroughMode != nil && pointer.Deref(conf.PassthroughMode) == "All" {
		// clusters were added in TransparentProxyGenerator, do nothing
		return nil, nil
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
			return conf.ConfWarnings(), configurer.Configure(listeners.Ipv4Passthrough, listeners.Ipv6Passthrough, rs)
		}
	}
	return nil, nil
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
