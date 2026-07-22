package matchers

import (
	"cmp"
	"errors"
	"fmt"
	"slices"

	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/kri"
	"github.com/kumahq/kuma/v3/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/registry"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	core_rules "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/resolve"
	meshhttproute_api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
	util_slices "github.com/kumahq/kuma/v3/pkg/util/slices"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
)

func PolicyMatches(resource core_model.Resource, dpp *core_mesh.DataplaneResource, referencableResources xds_context.Resources) (bool, error) {
	refPolicy, ok := resource.GetSpec().(core_model.Policy)
	if !ok {
		return false, errors.New("resource is not a targetRef policy")
	}
	selectedInbounds, delegatedGateway, err := DppSelectedByPolicy(resource.GetMeta(), refPolicy.GetTargetRef(), dpp, referencableResources)
	selectedGatewayInbounds := builtinGatewayListenersSelectedByPolicy(resource.GetMeta(), refPolicy.GetTargetRef(), dpp)
	return len(selectedInbounds) != 0 || len(selectedGatewayInbounds) != 0 || delegatedGateway, err
}

// MatchedPolicies match policies using the standard matchers using targetRef (madr-005)
func MatchedPolicies(
	rType core_model.ResourceType,
	dpp *core_mesh.DataplaneResource,
	resources xds_context.Resources,
	opts ...plugins.MatchedPoliciesOption,
) (core_xds.TypedMatchingPolicies, error) {
	mpOpts := plugins.NewMatchedPoliciesConfig(opts...)

	var cacheKey string
	if mpOpts.Cache != nil {
		cacheKey = BuildCacheKey(string(rType), mpOpts, dpp)
		if cached, ok := mpOpts.Cache.GetIfPresent(cacheKey); ok {
			return cached, nil
		}
	}

	policies := resources.ListOrEmpty(rType)
	var warnings []string

	matchedPoliciesByInbound := map[core_rules.InboundListener]core_model.ResourceList{}
	matchedPoliciesByGatewayInbound := map[core_rules.InboundListener]core_model.ResourceList{}
	matchedPoliciesByGatewayListener := map[core_rules.InboundListenerHostname]core_model.ResourceList{}
	dpPolicies, err := registry.Global().NewList(rType)
	if err != nil {
		return core_xds.TypedMatchingPolicies{}, err
	}

	for _, policy := range policies.GetItems() {
		if !mpOpts.IncludeShadow && core_model.IsShadowedResource(policy) {
			continue
		}

		refPolicy := policy.GetSpec().(core_model.Policy)
		selectedInbounds, delegatedGatewaySelected, err := DppSelectedByPolicy(policy.GetMeta(), refPolicy.GetTargetRef(), dpp, resources)
		selectedGatewayInbounds := builtinGatewayListenersSelectedByPolicy(policy.GetMeta(), refPolicy.GetTargetRef(), dpp)
		if err != nil {
			warnings = append(warnings,
				fmt.Sprintf("unable to resolve TargetRef on policy: mesh:%s name:%s error:%q",
					policy.GetMeta().GetMesh(), policy.GetMeta().GetName(), err.Error(),
				),
			)
		}
		if len(selectedInbounds) == 0 && len(selectedGatewayInbounds) == 0 && !delegatedGatewaySelected {
			// DPP is not matched by the policy
			continue
		}

		if err := dpPolicies.AddItem(policy); err != nil {
			return core_xds.TypedMatchingPolicies{}, err
		}

		for _, inbound := range selectedInbounds {
			if _, ok := matchedPoliciesByInbound[inbound]; !ok {
				matchedPoliciesByInbound[inbound], err = registry.Global().NewList(rType)
				if err != nil {
					return core_xds.TypedMatchingPolicies{}, err
				}
			}
			if err := matchedPoliciesByInbound[inbound].AddItem(policy); err != nil {
				return core_xds.TypedMatchingPolicies{}, err
			}
			matchedPoliciesByGatewayInbound[inbound] = matchedPoliciesByInbound[inbound]
		}
		for _, inbound := range selectedGatewayInbounds {
			if _, ok := matchedPoliciesByGatewayInbound[inbound]; !ok {
				matchedPoliciesByGatewayInbound[inbound], err = registry.Global().NewList(rType)
				if err != nil {
					return core_xds.TypedMatchingPolicies{}, err
				}
			}
			if err := matchedPoliciesByGatewayInbound[inbound].AddItem(policy); err != nil {
				return core_xds.TypedMatchingPolicies{}, err
			}
		}
	}

	dpPolicies = SortByTargetRef(dpPolicies)

	for inbound, ps := range matchedPoliciesByInbound {
		matchedPoliciesByInbound[inbound] = SortByTargetRef(ps)
	}
	for inbound, ps := range matchedPoliciesByGatewayInbound {
		matchedPoliciesByGatewayInbound[inbound] = SortByTargetRef(ps)
	}

	fr, err := core_rules.BuildFromRules(matchedPoliciesByInbound)
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("couldn't create From rules: %s", err.Error()))
	}

	tr, err := core_rules.BuildToRules(dpPolicies, resources)
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("couldn't create To rules: %s", err.Error()))
	}

	gr, err := core_rules.BuildGatewayRules(
		matchedPoliciesByGatewayInbound,
		matchedPoliciesByGatewayListener,
		resources,
	)
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("couldn't create Gateway rules: %s", err.Error()))
	}

	sr, err := core_rules.BuildSingleItemRules(dpPolicies.GetItems())
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("couldn't create top level rules: %s", err.Error()))
	}

	result := core_xds.TypedMatchingPolicies{
		Type:              rType,
		DataplanePolicies: dpPolicies.GetItems(),
		FromRules:         fr,
		ToRules:           tr,
		GatewayRules:      gr,
		SingleItemRules:   sr,
		Warnings:          warnings,
	}
	if mpOpts.Cache != nil {
		mpOpts.Cache.Put(cacheKey, result)
	}
	return result, nil
}

// DppSelectedByPolicy returns a list of inbounds of DPP that are selected by the top-level targetRef
// and whether a delegated gateway is selected
func DppSelectedByPolicy(
	meta core_model.ResourceMeta,
	ref common_api.TargetRef,
	dpp *core_mesh.DataplaneResource,
	referencableResources xds_context.Resources,
) ([]core_rules.InboundListener, bool, error) {
	if !dppSelectedByZone(meta, dpp) {
		return []core_rules.InboundListener{}, false, nil
	}
	if !dppSelectedByNamespace(meta, dpp) {
		return []core_rules.InboundListener{}, false, nil
	}
	switch ref.Kind {
	case common_api.Mesh:
		if isSupportedProxyType(pointer.Deref(ref.ProxyTypes), resolveDataplaneProxyType(dpp)) {
			inbounds := allInboundListeners(dpp)
			inbounds = append(inbounds, embeddedListenersAsInboundListeners(dpp)...)
			return inbounds, dpp.Spec.IsDelegatedGateway(), nil
		}
		return []core_rules.InboundListener{}, false, nil
	case common_api.Dataplane:
		if allDataplanesSelected(ref) || isSelectedByResourceIdentifier(dpp, ref, meta) || isSelectedByLabels(dpp, ref) {
			inboundInterfaces := dpp.Spec.GetNetworking().InboundsSelectedBySectionName(pointer.Deref(ref.SectionName))
			inbounds := util_slices.Map(inboundInterfaces, func(i mesh_proto.InboundInterface) core_rules.InboundListener {
				return core_rules.InboundListener{Address: i.DataplaneIP, Port: i.DataplanePort}
			})
			sectionName := pointer.Deref(ref.SectionName)
			for _, l := range dpp.Spec.GetNetworking().GetListeners() {
				if sectionName != "" && l.GetSectionName() != sectionName {
					continue
				}
				addr := l.GetAddress()
				if addr == "" {
					addr = dpp.Spec.GetNetworking().GetAddress()
				}
				inbounds = append(inbounds, core_rules.InboundListener{Address: addr, Port: l.GetPort()})
			}
			return inbounds, dpp.Spec.IsDelegatedGateway(), nil
		}
		return []core_rules.InboundListener{}, false, nil
	case common_api.MeshHTTPRoute:
		mhr := resolveMeshHTTPRouteRef(meta, pointer.Deref(ref.Name), referencableResources.ListOrEmpty(meshhttproute_api.MeshHTTPRouteType))
		if mhr == nil {
			return nil, false, fmt.Errorf("couldn't resolve MeshHTTPRoute targetRef with name '%s'", pointer.Deref(ref.Name))
		}
		return DppSelectedByPolicy(mhr.Meta, pointer.DerefOr(mhr.Spec.TargetRef, common_api.TargetRef{Kind: common_api.Mesh}), dpp, referencableResources)
	default:
		return nil, false, fmt.Errorf("unsupported targetRef kind '%s'", ref.Kind)
	}
}

func allDataplanesSelected(ref common_api.TargetRef) bool {
	return pointer.Deref(ref.Name) == "" && pointer.Deref(ref.Namespace) == "" && pointer.Deref(ref.Labels) == nil
}

// TODO this is common functionality with selecting MeshService by labels, we should refactor this and extract to some common function
func isSelectedByLabels(dpp *core_mesh.DataplaneResource, ref common_api.TargetRef) bool {
	if pointer.Deref(ref.Labels) == nil {
		return false
	}

	for label, value := range pointer.Deref(ref.Labels) {
		if dpp.GetMeta().GetLabels()[label] != value {
			return false
		}
	}
	return true
}

func isSelectedByResourceIdentifier(dpp *core_mesh.DataplaneResource, ref common_api.TargetRef, meta core_model.ResourceMeta) bool {
	if pointer.Deref(ref.Name) == "" {
		return false
	}
	return kri.From(dpp) == resolve.TargetRefToKRI(kri.FromResourceMeta(meta, core_model.ResourceType(ref.Kind)), ref)
}

func dppSelectedByNamespace(meta core_model.ResourceMeta, dpp *core_mesh.DataplaneResource) bool {
	switch core_model.PolicyRole(meta) {
	case mesh_proto.ConsumerPolicyRole, mesh_proto.WorkloadOwnerPolicyRole:
		ns, ok := meta.GetLabels()[mesh_proto.KubeNamespaceTag]
		return ok && ns == dpp.GetMeta().GetLabels()[mesh_proto.KubeNamespaceTag]
	default:
		return true
	}
}

func dppSelectedByZone(policyMeta core_model.ResourceMeta, dpp *core_mesh.DataplaneResource) bool {
	switch core_model.PolicyRole(policyMeta) {
	case mesh_proto.ProducerPolicyRole:
		return true
	default:
		if dpp.GetMeta() == nil {
			return true
		}
		meta := dpp.GetMeta()
		// we should return true once dpp has no origin.
		// Resource that cannot be created on zone(global one) doesn't have it
		origin, ok := meta.GetLabels()[mesh_proto.ResourceOriginLabel]
		if !ok || origin == string(mesh_proto.GlobalResourceOrigin) {
			return true
		}
		policyOrigin, ok := policyMeta.GetLabels()[mesh_proto.ResourceOriginLabel]
		if ok && policyOrigin == string(mesh_proto.ZoneResourceOrigin) {
			zone, ok := policyMeta.GetLabels()[mesh_proto.ZoneTag]
			if !ok {
				return true
			}
			return core_model.IsLocalZoneResource(meta.GetLabels(), zone)
		}
		return true
	}
}

func resolveMeshHTTPRouteRef(refMeta core_model.ResourceMeta, refName string, mhrs core_model.ResourceList) *meshhttproute_api.MeshHTTPRouteResource {
	for _, item := range mhrs.GetItems() {
		if core_model.IsReferenced(refMeta, refName, item.GetMeta()) {
			return item.(*meshhttproute_api.MeshHTTPRouteResource)
		}
	}
	return nil
}

func resolveDataplaneProxyType(dpp *core_mesh.DataplaneResource) common_api.TargetRefProxyType {
	if dpp.Spec.GetNetworking().GetGateway() != nil {
		return common_api.Gateway
	}
	return common_api.Sidecar
}

func isSupportedProxyType(supportedTypes []common_api.TargetRefProxyType, dppType common_api.TargetRefProxyType) bool {
	return len(supportedTypes) == 0 || slices.Contains(supportedTypes, dppType)
}

func builtinGatewayListenersSelectedByPolicy(
	meta core_model.ResourceMeta,
	ref common_api.TargetRef,
	dpp *core_mesh.DataplaneResource,
) []core_rules.InboundListener {
	if ref.Kind != common_api.Mesh {
		return nil
	}
	if !dppSelectedByZone(meta, dpp) || !dppSelectedByNamespace(meta, dpp) {
		return nil
	}
	if !isSupportedProxyType(pointer.Deref(ref.ProxyTypes), resolveDataplaneProxyType(dpp)) {
		return nil
	}
	if !dpp.Spec.IsBuiltinGateway() {
		return nil
	}
	return []core_rules.InboundListener{{
		Address: dpp.Spec.GetNetworking().GetAddress(),
	}}
}

// allInboundListeners returns every inbound of the dataplane as an InboundListener
func allInboundListeners(dpp *core_mesh.DataplaneResource) []core_rules.InboundListener {
	inbounds := []core_rules.InboundListener{}
	for _, inbound := range dpp.Spec.GetNetworking().GetInbound() {
		intf := dpp.Spec.GetNetworking().ToInboundInterface(inbound)
		inbounds = append(inbounds, core_rules.InboundListener{
			Address: intf.DataplaneIP,
			Port:    intf.DataplanePort,
		})
	}
	return inbounds
}

func SortByTargetRef(rl core_model.ResourceList) core_model.ResourceList {
	rs := rl.GetItems()
	slices.SortFunc(rs, func(r1, r2 core_model.Resource) int {
		p1, ok1 := r1.GetSpec().(core_model.Policy)
		p2, ok2 := r2.GetSpec().(core_model.Policy)
		if !ok1 || !ok2 {
			panic("resource doesn't support TargetRef matching")
		}

		tr1, tr2 := p1.GetTargetRef(), p2.GetTargetRef()
		if less := tr1.Kind.Compare(tr2.Kind); less != 0 {
			return less
		}

		if less := tr1.CompareDataplaneKind(tr2); less != 0 {
			return less
		}

		o1, _ := core_model.ResourceOrigin(r1.GetMeta())
		o2, _ := core_model.ResourceOrigin(r2.GetMeta())
		if less := o1.Compare(o2); less != 0 {
			return less
		}

		if less := core_model.PolicyRole(r1.GetMeta()).Compare(core_model.PolicyRole(r2.GetMeta())); less != 0 {
			return less
		}

		return cmp.Compare(core_model.GetDisplayName(r2.GetMeta()), core_model.GetDisplayName(r1.GetMeta()))
	})
	rv := registry.Global().MustNewList(rl.GetItemType())
	for _, r := range rs {
		_ = rv.AddItem(r)
	}
	return rv
}

func embeddedListenersAsInboundListeners(dpp *core_mesh.DataplaneResource) []core_rules.InboundListener {
	var result []core_rules.InboundListener
	for _, l := range dpp.Spec.GetNetworking().GetListeners() {
		addr := l.GetAddress()
		if addr == "" {
			addr = dpp.Spec.GetNetworking().GetAddress()
		}
		result = append(result, core_rules.InboundListener{Address: addr, Port: l.GetPort()})
	}
	return result
}
