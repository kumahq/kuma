package matchers

import (
	"errors"
	"fmt"
	"slices"
	"sort"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	meshhttproute_api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	xds_topology "github.com/kumahq/kuma/pkg/xds/topology"
)

func PolicyMatches(resource core_model.Resource, dpp *core_mesh.DataplaneResource, referencableResources xds_context.Resources) (bool, error) {
	var gateway *core_mesh.MeshGatewayResource
	if dpp.Spec.IsBuiltinGateway() {
		gateway = xds_topology.SelectGateway(referencableResources.Gateways().Items, dpp.Spec.Matches)
	}
	refPolicy, ok := resource.GetSpec().(core_model.Policy)
	if !ok {
		return false, errors.New("resource is not a targetRef policy")
	}
	selectedInbounds, selectedGatewayListener, delegatedGateway, err := dppSelectedByPolicy(resource.GetMeta(), refPolicy.GetTargetRef(), dpp, gateway, referencableResources)
	return len(selectedInbounds) != 0 || len(selectedGatewayListener) != 0 || delegatedGateway, err
}

// MatchedPolicies match policies using the standard matchers using targetRef (madr-005)
func MatchedPolicies(rType core_model.ResourceType, dpp *core_mesh.DataplaneResource, resources xds_context.Resources) (core_xds.TypedMatchingPolicies, error) {
	policies := resources.ListOrEmpty(rType)
	var warnings []string

	matchedPoliciesByInbound := map[core_rules.InboundListener][]core_model.Resource{}
	matchedPoliciesByGatewayListener := map[core_rules.InboundListenerHostname][]core_model.Resource{}
	var dpPolicies []core_model.Resource

	gateway := xds_topology.SelectGateway(resources.Gateways().Items, dpp.Spec.Matches)
	for _, policy := range policies.GetItems() {
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

		dpPolicies = append(dpPolicies, policy)

		for _, listener := range matchedGatewayListeners {
			matchedPoliciesByGatewayListener[listener] = append(matchedPoliciesByGatewayListener[listener], policy)
		}
		for _, inbound := range selectedInbounds {
			matchedPoliciesByInbound[inbound] = append(matchedPoliciesByInbound[inbound], policy)
		}
	}

	sort.Sort(ByTargetRef(dpPolicies))

	for _, ps := range matchedPoliciesByInbound {
		sort.Sort(ByTargetRef(ps))
	}

	fr, err := core_rules.BuildFromRules(matchedPoliciesByInbound)
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("couldn't create From rules: %s", err.Error()))
	}

	tr, err := core_rules.BuildToRules(dpPolicies, resources.ListOrEmpty(meshhttproute_api.MeshHTTPRouteType).GetItems())
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("couldn't create To rules: %s", err.Error()))
	}

	gr, err := core_rules.BuildGatewayRules(
		matchedPoliciesByInbound,
		matchedPoliciesByGatewayListener,
		resources.ListOrEmpty(meshhttproute_api.MeshHTTPRouteType).GetItems(),
	)
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("couldn't create Gateway rules: %s", err.Error()))
	}

	sr, err := core_rules.BuildSingleItemRules(dpPolicies)
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("couldn't create top level rules: %s", err.Error()))
	}

	return core_xds.TypedMatchingPolicies{
		Type:              rType,
		DataplanePolicies: dpPolicies,
		FromRules:         fr,
		ToRules:           tr,
		GatewayRules:      gr,
		SingleItemRules:   sr,
		Warnings:          warnings,
	}, nil
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
		return dppSelectedByPolicy(mhr.Meta, mhr.Spec.TargetRef, dpp, gateway, referencableResources)
	default:
		return nil, nil, false, fmt.Errorf("unsupported targetRef kind '%s'", ref.Kind)
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

type ByTargetRef []core_model.Resource

func (b ByTargetRef) Len() int { return len(b) }

func (b ByTargetRef) Less(i, j int) bool {
	r1, ok1 := b[i].GetSpec().(core_model.Policy)
	r2, ok2 := b[j].GetSpec().(core_model.Policy)
	if !(ok1 && ok2) {
		panic("resource doesn't support TargetRef matching")
	}

	tr1, tr2 := r1.GetTargetRef(), r2.GetTargetRef()

	if tr1.Kind != tr2.Kind {
		return tr1.Kind.Less(tr2.Kind)
	}

	o1, o2 := originToNumber(b[i]), originToNumber(b[j])
	if o1 != o2 {
		return o1 < o2
	}

	if tr1.Kind == common_api.MeshGateway {
		if len(tr1.Tags) != len(tr2.Tags) {
			return len(tr1.Tags) < len(tr2.Tags)
		}
	}

	return core_model.GetDisplayName(b[i]) > core_model.GetDisplayName(b[j])
}

func (b ByTargetRef) Swap(i, j int) { b[i], b[j] = b[j], b[i] }

// The logic of this method is to recreate the following comparison table:

// origin_1 | origin_2 | has_more_priority
// ---------|----------|-------------
// Global   | Zone     | origin_2
// Global   | Unknown  | origin_2
// Zone     | Global   | origin_1
// Zone     | Unknown  | origin_1
// Unknown  | Global   | origin_1
// Unknown  | Zone     | origin_2
//
// If we assign numbers to origins like Global=-1, Zone=1, Unknown=0, then we can compare them as numbers
// and get the same result as in the table above.
func originToNumber(r core_model.Resource) int {
	origin, _ := core_model.ResourceOrigin(r.GetMeta())
	switch origin {
	case mesh_proto.GlobalResourceOrigin:
		return -1
	case mesh_proto.ZoneResourceOrigin:
		return 1
	default:
		return 0
	}
}
