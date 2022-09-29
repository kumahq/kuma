package matchers

import (
	"sort"

	"github.com/pkg/errors"

	common_proto "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
)

// MatchedPolicies match policies using the standard matchers using targetRef (madr-005)
func MatchedPolicies(rType core_model.ResourceType, dpp *core_mesh.DataplaneResource, resources xds_context.Resources) (core_xds.TypedMatchingPolicies, error) {
	policies := resources.ListOrEmpty(rType)

	matchedPoliciesByInbound := map[core_xds.InboundListener][]core_model.Resource{}
	dpPolicies := []core_model.Resource{}

	for _, policy := range policies.GetItems() {
		spec, ok := policy.GetSpec().(core_xds.Policy)
		if !ok {
			return core_xds.TypedMatchingPolicies{}, errors.Errorf("resource type %v doesn't support TargetRef matching", rType)
		}

		targetRef := spec.GetTargetRef()
		selectedInbounds := inboundsSelectedByTargetRef(targetRef, dpp)
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

	fr, err := fromRules(matchedPoliciesByInbound)
	if err != nil {
		return core_xds.TypedMatchingPolicies{}, err
	}

	tr, err := toRules(dpPolicies)
	if err != nil {
		return core_xds.TypedMatchingPolicies{}, err
	}

	return core_xds.TypedMatchingPolicies{
		Type:              rType,
		DataplanePolicies: dpPolicies,
		FromRules: core_xds.FromRules{
			Rules: fr,
		},
		ToRules: core_xds.ToRules{
			Rules: tr,
		},
	}, nil
}

func fromRules(
	matchedPoliciesByInbound map[core_xds.InboundListener][]core_model.Resource,
) (map[core_xds.InboundListener]core_xds.Rules, error) {
	rulesByInbound := map[core_xds.InboundListener]core_xds.Rules{}
	for inbound, policies := range matchedPoliciesByInbound {
		fromList := []core_xds.PolicyItem{}
		for _, p := range policies {
			policyWithFrom, ok := p.GetSpec().(core_xds.PolicyWithFromList)
			if !ok {
				return nil, errors.Errorf("%s doesn't support 'from' list", p.Descriptor().Name)
			}
			fromList = append(fromList, policyWithFrom.GetFromList()...)
		}
		rules, err := core_xds.BuildRules(fromList)
		if err != nil {
			return nil, err
		}
		rulesByInbound[inbound] = rules
	}
	return rulesByInbound, nil
}

func toRules(matchedPolicies []core_model.Resource) (core_xds.Rules, error) {
	toList := []core_xds.PolicyItem{}
	for _, mp := range matchedPolicies {
		policyWithTo, ok := mp.GetSpec().(core_xds.PolicyWithToList)
		if !ok {
			return nil, errors.Errorf("%s doesn't support 'to' list", mp.Descriptor().Name)
		}
		toList = append(toList, policyWithTo.GetToList()...)
	}
	return core_xds.BuildRules(toList)
}

// inboundsSelectedByTargetRef returns a list of inbounds of DPP that are selected by the targetRef
func inboundsSelectedByTargetRef(tr *common_proto.TargetRef, dpp *core_mesh.DataplaneResource) []core_xds.InboundListener {
	switch tr.GetKindEnum() {
	case common_proto.TargetRef_Mesh:
		// return all inbounds interfaces of the DPP
		result := []core_xds.InboundListener{}
		for _, inbound := range dpp.Spec.GetNetworking().GetInbound() {
			intf := dpp.Spec.GetNetworking().ToInboundInterface(inbound)
			result = append(result, core_xds.InboundListener{
				Address: intf.DataplaneIP,
				Port:    intf.DataplanePort,
			})
		}
		return result
	case common_proto.TargetRef_MeshSubset:
		return inboundsSelectedByTags(tr.GetTags(), dpp)
	case common_proto.TargetRef_MeshService:
		return inboundsSelectedByTags(map[string]string{
			mesh_proto.ServiceTag: tr.GetName(),
		}, dpp)
	case common_proto.TargetRef_MeshServiceSubset:
		tags := map[string]string{
			mesh_proto.ServiceTag: tr.GetName(),
		}
		for k, v := range tr.GetTags() {
			tags[k] = v
		}
		return inboundsSelectedByTags(tags, dpp)
	default:
		return []core_xds.InboundListener{}
	}
}

func inboundsSelectedByTags(tags map[string]string, dpp *core_mesh.DataplaneResource) []core_xds.InboundListener {
	result := []core_xds.InboundListener{}
	for _, inbound := range dpp.Spec.GetNetworking().GetInbound() {
		if isInboundSelectedByTags(tags, inbound) {
			intf := dpp.Spec.GetNetworking().ToInboundInterface(inbound)
			result = append(result, core_xds.InboundListener{
				Address: intf.DataplaneIP,
				Port:    intf.DataplanePort,
			})
		}
	}
	return result
}

func isInboundSelectedByTags(tags map[string]string, inbound *mesh_proto.Dataplane_Networking_Inbound) bool {
	for k, v := range tags {
		if inboundValue, ok := inbound.Tags[k]; !ok || inboundValue != v {
			return false
		}
	}
	return true
}

type ByTargetRef []core_model.Resource

func (b ByTargetRef) Len() int { return len(b) }

func (b ByTargetRef) Less(i, j int) bool {
	r1, ok1 := b[i].GetSpec().(core_xds.Policy)
	r2, ok2 := b[j].GetSpec().(core_xds.Policy)
	if !(ok1 && ok2) {
		panic("resource doesn't support TargetRef matching")
	}

	tr1, tr2 := r1.GetTargetRef(), r2.GetTargetRef()

	kind1, kind2 := tr1.GetKindEnum(), tr2.GetKindEnum()
	if kind1 != kind2 {
		return kind1 < kind2
	}

	return b[i].GetMeta().GetName() < b[j].GetMeta().GetName()
}

func (b ByTargetRef) Swap(i, j int) { b[i], b[j] = b[j], b[i] }
