package v1alpha1

import (
	"github.com/pkg/errors"

	common_api "github.com/kumahq/kuma/v2/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/naming"
	core_plugins "github.com/kumahq/kuma/v2/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/v2/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/v2/pkg/core/xds"
	"github.com/kumahq/kuma/v2/pkg/plugins/policies/core/matchers"
	"github.com/kumahq/kuma/v2/pkg/plugins/policies/core/rules/subsetutils"
	api "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshproxypatch/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/util/pointer"
	xds_context "github.com/kumahq/kuma/v2/pkg/xds/context"
)

type modificator interface {
	apply(*core_xds.ResourceSet) error
}

type plugin struct{}

func (p plugin) Order() int { return api.MeshProxyPatchResourceTypeDescriptor.Order }

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
	if proxy.Dataplane.Spec.GetNetworking().HasZoneProxyListeners() {
		return applyZoneProxyDataplane(rs, proxy, policies)
	}
	rule := policies.SingleItemRules.Rules.Compute(subsetutils.MeshElement())
	conf := rule.Conf.(api.Conf)
	return ApplyMods(rs, pointer.Deref(conf.AppendModifications))
}

// applyZoneProxyDataplane walks matched policies in the order produced by
// SortByTargetRef (most-specific last, reverse-lex on display name as the
// final tie-break) and applies each policy's modifications with
// sectionName-aware scoping: when `targetRef` is `kind: Dataplane` and
// `sectionName` matches an embedded zone proxy listener, listener /
// network-filter / http-filter modifications are narrowed to that
// listener. Cluster and virtual-host modifications have no listener
// anchor and apply globally regardless of `sectionName`.
//
// Error behavior is all-or-nothing per call: the first modification
// that errors aborts the loop and the partial ResourceSet is returned
// unchanged for the remaining policies. Earlier policies' successful
// modifications are NOT rolled back — the caller owns rollback semantics.
func applyZoneProxyDataplane(rs *core_xds.ResourceSet, proxy *core_xds.Proxy, policies core_xds.TypedMatchingPolicies) error {
	networking := proxy.Dataplane.Spec.GetNetworking()
	for _, p := range policies.DataplanePolicies {
		spec, ok := p.GetSpec().(*api.MeshProxyPatch)
		if !ok {
			continue
		}
		mods := pointer.Deref(spec.Default.AppendModifications)
		if len(mods) == 0 {
			continue
		}
		listenerName := scopedZoneProxyListenerName(networking, spec.GetTargetRef())
		scoped := mods
		if listenerName != "" {
			scoped = scopeModsToListener(mods, listenerName)
		}
		if err := ApplyMods(rs, scoped); err != nil {
			return errors.Wrapf(err, "could not apply MeshProxyPatch %q", core_model.MetaToResourceKey(p.GetMeta()))
		}
	}
	return nil
}

// scopedZoneProxyListenerName returns the contextual Envoy listener name
// when `targetRef` is `kind: Dataplane` and `sectionName` matches a zone
// proxy listener (ZoneIngress or ZoneEgress) on the DPP's networking.
// Returns "" when no scoping should be applied.
func scopedZoneProxyListenerName(networking *mesh_proto.Dataplane_Networking, targetRef common_api.TargetRef) string {
	if targetRef.Kind != common_api.Dataplane {
		return ""
	}
	sectionName := pointer.Deref(targetRef.SectionName)
	if sectionName == "" {
		return ""
	}
	for _, l := range networking.GetListeners() {
		if l.GetSectionName() != sectionName {
			continue
		}
		switch l.GetType() {
		case mesh_proto.Dataplane_Networking_Listener_ZoneIngress:
			return naming.ContextualZoneIngressListenerName(sectionName)
		case mesh_proto.Dataplane_Networking_Listener_ZoneEgress:
			return naming.ContextualZoneEgressListenerName(sectionName)
		}
	}
	return ""
}

// scopeModsToListener returns a copy of `mods` with the implicit listener
// name injected on listener / network-filter / http-filter modifications.
// Cluster and virtual-host modifications are returned unchanged. A user
// match.name / match.listenerName already set is preserved; if it differs
// from `listenerName`, the conjunction never matches (same as today for
// any contradictory match).
func scopeModsToListener(mods []api.Modification, listenerName string) []api.Modification {
	out := make([]api.Modification, len(mods))
	for i, m := range mods {
		out[i] = m
		switch {
		case m.Listener != nil:
			lm := *m.Listener
			match := api.ListenerMatch{}
			if lm.Match != nil {
				match = *lm.Match
			}
			if match.Name == nil {
				match.Name = pointer.To(listenerName)
			}
			lm.Match = &match
			out[i].Listener = &lm
		case m.NetworkFilter != nil:
			nfm := *m.NetworkFilter
			match := api.NetworkFilterMatch{}
			if nfm.Match != nil {
				match = *nfm.Match
			}
			if match.ListenerName == nil {
				match.ListenerName = pointer.To(listenerName)
			}
			nfm.Match = &match
			out[i].NetworkFilter = &nfm
		case m.HTTPFilter != nil:
			hfm := *m.HTTPFilter
			match := api.HTTPFilterMatch{}
			if hfm.Match != nil {
				match = *hfm.Match
			}
			if match.ListenerName == nil {
				match.ListenerName = pointer.To(listenerName)
			}
			hfm.Match = &match
			out[i].HTTPFilter = &hfm
		}
	}
	return out
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
