package matchers

import (
	"cmp"
	"errors"
	"fmt"
	"slices"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	meshhttproute_api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/util/pointer"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	xds_topology "github.com/kumahq/kuma/pkg/xds/topology"
)

func PolicyMatches(resource core_model.Resource, dpp *core_mesh.DataplaneResource, referencableResources xds_context.Resources) (bool, error) {
	var gateway *core_mesh.MeshGatewayResource
	if dpp.Spec.IsBuiltinGateway() {
		zoneGateways := filterGatewaysByZone(referencableResources.Gateways().Items, dpp)
		gateway = xds_topology.SelectGateway(zoneGateways, dpp.Spec.Matches)
	}
	refPolicy, ok := resource.GetSpec().(core_model.Policy)
	if !ok {
		return false, errors.New("resource is not a targetRef policy")
	}
	selectedInbounds, selectedGatewayListener, delegatedGateway, err := dppSelectedByPolicy(resource.GetMeta(), refPolicy.GetTargetRef(), dpp, gateway, referencableResources)
	return len(selectedInbounds) != 0 || len(selectedGatewayListener) != 0 || delegatedGateway, err
}

// MatchedPolicies match policies using the standard matchers using targetRef (madr-005)
func MatchedPolicies(
	rType core_model.ResourceType,
	dpp *core_mesh.DataplaneResource,
	resources xds_context.Resources,
	opts ...plugins.MatchedPoliciesOption,
) (core_xds.TypedMatchingPolicies, error) {
	mpOpts := plugins.NewMatchedPoliciesConfig(opts...)

	policies := resources.ListOrEmpty(rType)
	var warnings []string

	matchedPoliciesByInbound := map[core_rules.InboundListener]core_model.ResourceList{}
	matchedPoliciesByGatewayListener := map[core_rules.InboundListenerHostname]core_model.ResourceList{}
	dpPolicies, err := registry.Global().NewList(rType)
	if err != nil {
		return core_xds.TypedMatchingPolicies{}, err
	}

	zoneGateways := filterGatewaysByZone(resources.Gateways().Items, dpp)
	gateway := xds_topology.SelectGateway(zoneGateways, dpp.Spec.Matches)
	for _, policy := range policies.GetItems() {
		if !mpOpts.IncludeShadow && core_model.IsShadowedResource(policy) {
			continue
		}

		refPolicy := policy.GetSpec().(core_model.Policy)
		selectedInbounds, matchedGatewayListeners, delegatedGatewaySelected, err := dppSelectedByPolicy(policy.GetMeta(), refPolicy.GetTargetRef(), dpp, gateway, resources)
		if err != nil {
			warnings = append(warnings,
				fmt.Sprintf("unable to resolve TargetRef on policy: mesh:%s name:%s error:%q",
					policy.GetMeta().GetMesh(), policy.GetMeta().GetName(), err.Error(),
				),
			)
		}
		if len(selectedInbounds) == 0 && len(matchedGatewayListeners) == 0 && !delegatedGatewaySelected {
			// DPP is not matched by the policy
			continue
		}

		if err := dpPolicies.AddItem(policy); err != nil {
			return core_xds.TypedMatchingPolicies{}, err
		}

		for _, listener := range matchedGatewayListeners {
			if _, ok := matchedPoliciesByGatewayListener[listener]; !ok {
				matchedPoliciesByGatewayListener[listener], err = registry.Global().NewList(rType)
			}
			if err := matchedPoliciesByGatewayListener[listener].AddItem(policy); err != nil {
				return core_xds.TypedMatchingPolicies{}, err
			}
		}
		for _, inbound := range selectedInbounds {
			if _, ok := matchedPoliciesByInbound[inbound]; !ok {
				matchedPoliciesByInbound[inbound], err = registry.Global().NewList(rType)
			}
			if err := matchedPoliciesByInbound[inbound].AddItem(policy); err != nil {
				return core_xds.TypedMatchingPolicies{}, err
			}
		}
	}

	dpPolicies = SortByTargetRef(dpPolicies)

	for inbound, ps := range matchedPoliciesByInbound {
		matchedPoliciesByInbound[inbound] = SortByTargetRef(ps)
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
		matchedPoliciesByInbound,
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

	return core_xds.TypedMatchingPolicies{
		Type:              rType,
		DataplanePolicies: dpPolicies.GetItems(),
		FromRules:         fr,
		ToRules:           tr,
		GatewayRules:      gr,
		SingleItemRules:   sr,
		Warnings:          warnings,
	}, nil
}

func filterGatewaysByZone(gateways []*core_mesh.MeshGatewayResource, dpp *core_mesh.DataplaneResource) []*core_mesh.MeshGatewayResource {
	if gateways == nil {
		return gateways
	}
	var filtered []*core_mesh.MeshGatewayResource
	dppZone, dppZoneOk := dpp.GetMeta().GetLabels()[mesh_proto.ZoneTag]
	for _, gateway := range gateways {
		gwOrigin, ok := gateway.GetMeta().GetLabels()[mesh_proto.ResourceOriginLabel]
		if !ok || gwOrigin == string(mesh_proto.GlobalResourceOrigin) {
			filtered = append(filtered, gateway)
			continue
		}
		if !dppZoneOk || core_model.IsLocalZoneResource(gateway.GetMeta().GetLabels(), dppZone) {
			filtered = append(filtered, gateway)
		}
	}
	return filtered
}

// dppSelectedByPolicy returns a list of inbounds of DPP that are selected by the top-level targetRef
// and whether a delegated gateway is selected
func dppSelectedByPolicy(
	meta core_model.ResourceMeta,
	ref common_api.TargetRef,
	dpp *core_mesh.DataplaneResource,
	gateway *core_mesh.MeshGatewayResource,
	referencableResources xds_context.Resources,
) ([]core_rules.InboundListener, []core_rules.InboundListenerHostname, bool, error) {
	if !dppSelectedByZone(meta, dpp, gateway) {
		return []core_rules.InboundListener{}, nil, false, nil
	}
	if !dppSelectedByNamespace(meta, dpp) {
		return []core_rules.InboundListener{}, nil, false, nil
	}
	switch ref.Kind {
	case common_api.Mesh:
		if isSupportedProxyType(ref.ProxyTypes, resolveDataplaneProxyType(dpp)) {
			inbounds, gwListeners, gateway := inboundsSelectedByTags(nil, dpp, gateway)
			return inbounds, gwListeners, gateway, nil
		}
		return []core_rules.InboundListener{}, nil, false, nil
	case common_api.MeshSubset:
		if isSupportedProxyType(ref.ProxyTypes, resolveDataplaneProxyType(dpp)) {
			inbounds, gwListeners, gateway := inboundsSelectedByTags(ref.Tags, dpp, gateway)
			return inbounds, gwListeners, gateway, nil
		}
		return []core_rules.InboundListener{}, nil, false, nil
	case common_api.MeshService:
		inbounds, gwListeners, gateway := inboundsSelectedByTags(map[string]string{
			mesh_proto.ServiceTag: ref.Name,
		}, dpp, gateway)
		return inbounds, gwListeners, gateway, nil
	case common_api.MeshServiceSubset:
		tags := map[string]string{
			mesh_proto.ServiceTag: ref.Name,
		}
		for k, v := range ref.Tags {
			tags[k] = v
		}
		inbounds, gwListeners, gateway := inboundsSelectedByTags(tags, dpp, gateway)
		return inbounds, gwListeners, gateway, nil
	case common_api.MeshGateway:
		if gateway == nil || !dpp.Spec.IsBuiltinGateway() || !core_model.IsReferenced(meta, ref.Name, gateway.GetMeta()) {
			return nil, nil, false, nil
		}
		inbounds, gwListeners, _ := inboundsSelectedByTags(ref.Tags, dpp, gateway)
		return inbounds, gwListeners, false, nil
	case common_api.MeshHTTPRoute:
		mhr := resolveMeshHTTPRouteRef(meta, ref.Name, referencableResources.ListOrEmpty(meshhttproute_api.MeshHTTPRouteType))
		if mhr == nil {
			return nil, nil, false, fmt.Errorf("couldn't resolve MeshHTTPRoute targetRef with name '%s'", ref.Name)
		}
		return dppSelectedByPolicy(mhr.Meta, pointer.DerefOr(mhr.Spec.TargetRef, common_api.TargetRef{Kind: common_api.Mesh}), dpp, gateway, referencableResources)
	default:
		return nil, nil, false, fmt.Errorf("unsupported targetRef kind '%s'", ref.Kind)
	}
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

func dppSelectedByZone(policyMeta core_model.ResourceMeta, dpp *core_mesh.DataplaneResource, gateway *core_mesh.MeshGatewayResource) bool {
	switch core_model.PolicyRole(policyMeta) {
	case mesh_proto.ProducerPolicyRole:
		return true
	default:
		if dpp.GetMeta() == nil && gateway == nil {
			return true
		}
		meta := dpp.GetMeta()
		if gateway != nil {
			meta = gateway.GetMeta()
		}
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

// inboundsSelectedByTags returns which inbounds are selected and whether a
// delegated gateway is selected
func inboundsSelectedByTags(tagsSelector mesh_proto.TagSelector, dpp *core_mesh.DataplaneResource, gateway *core_mesh.MeshGatewayResource) ([]core_rules.InboundListener, []core_rules.InboundListenerHostname, bool) {
	inbounds := []core_rules.InboundListener{}
	for _, inbound := range dpp.Spec.GetNetworking().GetInbound() {
		if inbound.State == mesh_proto.Dataplane_Networking_Inbound_Ignored {
			continue
		}
		if tagsSelector.Matches(inbound.Tags) {
			intf := dpp.Spec.GetNetworking().ToInboundInterface(inbound)
			inbounds = append(inbounds, core_rules.InboundListener{
				Address: intf.DataplaneIP,
				Port:    intf.DataplanePort,
			})
		}
	}
	gwListeners := []core_rules.InboundListenerHostname{}
	inboundSet := map[core_rules.InboundListener]struct{}{}
	if gateway != nil {
		for _, listener := range gateway.Spec.GetConf().GetListeners() {
			listenerTags := mesh_proto.Merge(
				dpp.Spec.GetNetworking().GetGateway().GetTags(),
				gateway.Spec.GetTags(),
				listener.GetTags(),
			)
			if tagsSelector.Matches(listenerTags) {
				inboundListener := core_rules.InboundListener{
					Address: dpp.Spec.GetNetworking().GetAddress(),
					Port:    listener.Port,
				}
				if _, ok := inboundSet[inboundListener]; !ok {
					inbounds = append(inbounds, inboundListener)
					inboundSet[inboundListener] = struct{}{}
				}
				gwListeners = append(gwListeners, core_rules.NewInboundListenerHostname(
					dpp.Spec.GetNetworking().GetAddress(),
					listener.Port,
					listener.GetNonEmptyHostname(),
				))
			}
		}
	}
	delegatedGatewaySelected := dpp.Spec.IsDelegatedGateway() && tagsSelector.Matches(dpp.Spec.GetNetworking().GetGateway().Tags)
	return inbounds, gwListeners, delegatedGatewaySelected
}

func SortByTargetRef(rl core_model.ResourceList) core_model.ResourceList {
	rs := rl.GetItems()
	slices.SortFunc(rs, func(r1, r2 core_model.Resource) int {
		p1, ok1 := r1.GetSpec().(core_model.Policy)
		p2, ok2 := r2.GetSpec().(core_model.Policy)
		if !(ok1 && ok2) {
			panic("resource doesn't support TargetRef matching")
		}

		tr1, tr2 := p1.GetTargetRef(), p2.GetTargetRef()
		if less := tr1.Kind.Compare(tr2.Kind); less != 0 {
			return less
		}

		o1, _ := core_model.ResourceOrigin(r1.GetMeta())
		o2, _ := core_model.ResourceOrigin(r2.GetMeta())
		if less := o1.Compare(o2); less != 0 {
			return less
		}

		if tr1.Kind == common_api.MeshGateway {
			if less := len(tr1.Tags) - len(tr2.Tags); less != 0 {
				return less
			}
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
