package matchers

import (
	"sort"

	"github.com/pkg/errors"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	xds_topology "github.com/kumahq/kuma/pkg/xds/topology"
)

// MatchedPolicies match policies using the standard matchers using targetRef (madr-005)
func MatchedPolicies(rType core_model.ResourceType, dpp *core_mesh.DataplaneResource, resources xds_context.Resources) (core_xds.TypedMatchingPolicies, error) {
	policies := resources.ListOrEmpty(rType)

	var gateway *core_mesh.MeshGatewayResource
	if dpp.Spec.IsBuiltinGateway() {
		gateways := resources.Gateways()
		gateway = xds_topology.SelectGateway(gateways.Items, dpp.Spec.Matches)
	}

	matchedPoliciesByInbound := map[core_rules.InboundListener][]core_model.Resource{}
	dpPolicies := []core_model.Resource{}

	for _, policy := range policies.GetItems() {
		spec, ok := policy.GetSpec().(core_model.Policy)
		if !ok {
			return core_xds.TypedMatchingPolicies{}, errors.Errorf("resource type %v doesn't support TargetRef matching", rType)
		}

		selectedInbounds := inboundsSelectedByTargetRef(spec.GetTargetRef(), dpp, gateway)
		if len(selectedInbounds) == 0 {
			// DPP is not matched by the policy
			continue
		}

		dpPolicies = append(dpPolicies, policy)

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
		return core_xds.TypedMatchingPolicies{}, err
	}

	tr, err := core_rules.BuildToRules(dpPolicies)
	if err != nil {
		return core_xds.TypedMatchingPolicies{}, err
	}

	sr, err := core_rules.BuildSingleItemRules(dpPolicies)
	if err != nil {
		return core_xds.TypedMatchingPolicies{}, err
	}

	return core_xds.TypedMatchingPolicies{
		Type:              rType,
		DataplanePolicies: dpPolicies,
		FromRules:         fr,
		ToRules:           tr,
		SingleItemRules:   sr,
	}, nil
}

// inboundsSelectedByTargetRef returns a list of inbounds of DPP that are selected by the targetRef
func inboundsSelectedByTargetRef(tr common_api.TargetRef, dpp *core_mesh.DataplaneResource, gateway *core_mesh.MeshGatewayResource) []core_rules.InboundListener {
	switch tr.Kind {
	case common_api.Mesh:
		return inboundsSelectedByTags(nil, dpp, gateway)
	case common_api.MeshSubset:
		return inboundsSelectedByTags(tr.Tags, dpp, gateway)
	case common_api.MeshService:
		return inboundsSelectedByTags(map[string]string{
			mesh_proto.ServiceTag: tr.Name,
		}, dpp, gateway)
	case common_api.MeshServiceSubset:
		tags := map[string]string{
			mesh_proto.ServiceTag: tr.Name,
		}
		for k, v := range tr.Tags {
			tags[k] = v
		}
		return inboundsSelectedByTags(tags, dpp, gateway)
	default:
		return []core_rules.InboundListener{}
	}
}

func inboundsSelectedByTags(tags map[string]string, dpp *core_mesh.DataplaneResource, gateway *core_mesh.MeshGatewayResource) []core_rules.InboundListener {
	result := []core_rules.InboundListener{}
	for _, inbound := range dpp.Spec.GetNetworking().GetInbound() {
		if mesh_proto.TagSelector(tags).Matches(inbound.Tags) {
			intf := dpp.Spec.GetNetworking().ToInboundInterface(inbound)
			result = append(result, core_rules.InboundListener{
				Address: intf.DataplaneIP,
				Port:    intf.DataplanePort,
			})
		}
	}
	if gateway != nil {
		for _, listener := range gateway.Spec.GetConf().GetListeners() {
			listenerTags := mesh_proto.Merge(
				dpp.Spec.GetNetworking().GetGateway().GetTags(),
				gateway.Spec.GetTags(),
				listener.GetTags(),
			)
			if mesh_proto.TagSelector(tags).Matches(listenerTags) {
				result = append(result, core_rules.InboundListener{
					Address: dpp.Spec.GetNetworking().GetAddress(),
					Port:    listener.Port,
				})
			}
		}
	}
	return result
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

	return b[i].GetMeta().GetName() < b[j].GetMeta().GetName()
}

func (b ByTargetRef) Swap(i, j int) { b[i], b[j] = b[j], b[i] }
