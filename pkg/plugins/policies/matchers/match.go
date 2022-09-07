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

type ResourceSpecWithTargetRef interface {
	core_model.ResourceSpec
	GetTargetRef() *common_proto.TargetRef
}

// MatchedInboundPolicies returns an InboundPolicies map with a list of policies
// set for each inbound. Should be used for policies that modify inbound listeners/cluster.
func MatchedInboundPolicies(
	rType core_model.ResourceType,
	dpp *core_mesh.DataplaneResource,
	resources xds_context.Resources,
) (core_xds.TypedMatchingPolicies, error) {
	inboundPoliciesMap, err := matchedInboundPolicies(rType, dpp, resources)
	if err != nil {
		return core_xds.TypedMatchingPolicies{}, err
	}

	for inbound := range inboundPoliciesMap {
		sort.Sort(ByTargetRef(inboundPoliciesMap[inbound]))
	}

	return core_xds.TypedMatchingPolicies{
		Type:            rType,
		InboundPolicies: inboundPoliciesMap,
	}, nil
}

// MatchedDataplanePolicies returns a DataplanePolicies list with policies
// matched for the DPP. Should be used for policies that modify either outbound
// listeners/clusters or DPP overall (like ProxyTemplate).
// The result of this function is the same as MatchedInboundPolicies, but without
// grouping policies by inbounds.
func MatchedDataplanePolicies(
	rType core_model.ResourceType,
	dpp *core_mesh.DataplaneResource,
	resources xds_context.Resources,
) (core_xds.TypedMatchingPolicies, error) {
	// policies matched for DPP is a union of policies matched for every inbound
	inboundPoliciesMap, err := matchedInboundPolicies(rType, dpp, resources)
	if err != nil {
		return core_xds.TypedMatchingPolicies{}, err
	}

	matchedPoliciesByName := map[string]core_model.Resource{}
	for _, policies := range inboundPoliciesMap {
		for _, policy := range policies {
			if _, ok := matchedPoliciesByName[policy.GetMeta().GetName()]; !ok {
				matchedPoliciesByName[policy.GetMeta().GetName()] = policy
			}
		}
	}

	matchedPolicies := []core_model.Resource{}
	for _, policy := range matchedPoliciesByName {
		matchedPolicies = append(matchedPolicies, policy)
	}

	sort.Sort(ByTargetRef(matchedPolicies))

	return core_xds.TypedMatchingPolicies{
		Type:              rType,
		DataplanePolicies: matchedPolicies,
	}, nil
}

func matchedInboundPolicies(
	rType core_model.ResourceType,
	dpp *core_mesh.DataplaneResource,
	resources xds_context.Resources,
) (map[mesh_proto.InboundInterface][]core_model.Resource, error) {
	policies := resources.ListOrEmpty(rType)

	matched := map[mesh_proto.InboundInterface][]core_model.Resource{}
	for _, policy := range policies.GetItems() {
		spec, ok := policy.GetSpec().(ResourceSpecWithTargetRef)
		if !ok {
			return nil, errors.Errorf("resource type %v doesn't support TargetRef matching", rType)
		}

		targetRef := spec.GetTargetRef()
		for _, inbound := range inboundsSelectedByTargetRef(targetRef, dpp) {
			matched[inbound] = append(matched[inbound], policy)
		}
	}

	return matched, nil
}

func inboundsSelectedByTargetRef(tr *common_proto.TargetRef, dpp *core_mesh.DataplaneResource) []mesh_proto.InboundInterface {
	switch tr.GetKindEnum() {
	case common_proto.TargetRef_Mesh:
		// return all inbounds interfaces of the DPP
		result := []mesh_proto.InboundInterface{}
		for _, inbound := range dpp.Spec.GetNetworking().GetInbound() {
			result = append(result, dpp.Spec.GetNetworking().ToInboundInterface(inbound))
		}
		return result
	case common_proto.TargetRef_MeshSubset:
		return inboundsSelectedByTags(tr.GetTags(), dpp)
	case common_proto.TargetRef_MeshService:
		return inboundsSelectedByTags(map[string]string{
			mesh_proto.ServiceTag: tr.GetName(),
		}, dpp)
	case common_proto.TargetRef_MeshServiceSubset:
		trTags := tr.GetTags()
		trTags[mesh_proto.ServiceTag] = tr.GetName()
		return inboundsSelectedByTags(trTags, dpp)
	default:
		return []mesh_proto.InboundInterface{}
	}
}

func inboundsSelectedByTags(tags map[string]string, dpp *core_mesh.DataplaneResource) []mesh_proto.InboundInterface {
	result := []mesh_proto.InboundInterface{}
	for _, inbound := range dpp.Spec.GetNetworking().GetInbound() {
		if isInboundSelectedByTags(tags, inbound) {
			result = append(result, dpp.Spec.GetNetworking().ToInboundInterface(inbound))
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
	r1, ok1 := b[i].GetSpec().(ResourceSpecWithTargetRef)
	r2, ok2 := b[j].GetSpec().(ResourceSpecWithTargetRef)
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
