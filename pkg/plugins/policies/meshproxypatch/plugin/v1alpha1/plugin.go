package v1alpha1

import (
	"github.com/pkg/errors"

	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/matchers"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/subsetutils"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshproxypatch/api/v1alpha1"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
)

var Origin = "mesh-proxy-patch"

type modificator interface {
	apply(*core_xds.ResourceSet) error
}

type plugin struct{}

var _ core_plugins.PolicyPlugin = &plugin{}

func NewPlugin() core_plugins.Plugin {
	return &plugin{}
}

func (p plugin) MatchedPolicies(dataplane *core_mesh.DataplaneResource, resources xds_context.Resources, opts ...core_plugins.MatchedPoliciesOption) (core_xds.TypedMatchingPolicies, error) {
	return matchers.MatchedPolicies(api.MeshProxyPatchType, dataplane, resources, opts...)
}

func (p plugin) Apply(rs *core_xds.ResourceSet, _ xds_context.Context, proxy *core_xds.Proxy) error {
	policies, ok := proxy.Policies.Dynamic[api.MeshProxyPatchType]
	if !ok {
		return nil
	}
	if len(policies.SingleItemRules.Rules) == 0 {
		return nil
	}
	rule := policies.SingleItemRules.Rules.Compute(subsetutils.MeshElement())
	conf := rule.Conf.(api.Conf)
	if err := ApplyMods(rs, conf.AppendModifications); err != nil {
		return err
	}
	return nil
}

func ApplyMods(resources *core_xds.ResourceSet, modifications []api.Modification) error {
	for i, modification := range modifications {
		var modificator modificator
		switch {
		case modification.Cluster != nil:
			mod := clusterModificator(*modification.Cluster)
			modificator = &mod
		case modification.Listener != nil:
			mod := listenerModificator(*modification.Listener)
			modificator = &mod
		case modification.NetworkFilter != nil:
			mod := networkFilterModificator(*modification.NetworkFilter)
			modificator = &mod
		case modification.HTTPFilter != nil:
			mod := httpFilterModificator(*modification.HTTPFilter)
			modificator = &mod
		case modification.VirtualHost != nil:
			mod := virtualHostModificator(*modification.VirtualHost)
			modificator = &mod
		default:
			return errors.Errorf("invalid modification")
		}
		if err := modificator.apply(resources); err != nil {
			return errors.Wrapf(err, "could not apply %d modification", i)
		}
	}
	return nil
}
