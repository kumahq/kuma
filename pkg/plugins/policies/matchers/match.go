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

// MatchedPolicies match policies using the standard matchers using targetRef (madr-005)
func MatchedPolicies(rType core_model.ResourceType, dpp *core_mesh.DataplaneResource, resources xds_context.Resources) (core_xds.TypedMatchingPolicies, error) {
	policies := resources.ListOrEmpty(rType)

	matchedPolicies := []core_model.Resource{}
	for _, policy := range policies.GetItems() {
		spec, ok := policy.GetSpec().(ResourceSpecWithTargetRef)
		if !ok {
			return core_xds.TypedMatchingPolicies{}, errors.Errorf("resource type %v doesn't support TargetRef matching", rType)
		}

		targetRef := spec.GetTargetRef()
		if isDataplaneSelectedByTargetRef(targetRef, dpp) {
			matchedPolicies = append(matchedPolicies, policy)
		}
	}

	sort.Sort(ByTargetRef(matchedPolicies))

	return core_xds.TypedMatchingPolicies{
		Type:              rType,
		DataplanePolicies: matchedPolicies,
	}, nil
}

func isDataplaneSelectedByTargetRef(tr *common_proto.TargetRef, dpp *core_mesh.DataplaneResource) bool {
	switch tr.GetKindEnum() {
	case common_proto.TargetRef_Mesh:
		return true
	case common_proto.TargetRef_MeshSubset:
		return isDataplaneSelectedByTags(tr.GetTags(), dpp)
	case common_proto.TargetRef_MeshService:
		return isDataplaneSelectedByTags(map[string]string{
			mesh_proto.ServiceTag: tr.GetName(),
		}, dpp)
	case common_proto.TargetRef_MeshServiceSubset:
		tags := map[string]string{
			mesh_proto.ServiceTag: tr.GetName(),
		}
		for k, v := range tr.GetTags() {
			tags[k] = v
		}
		return isDataplaneSelectedByTags(tags, dpp)
	default:
		return false
	}
}

func isDataplaneSelectedByTags(tags map[string]string, dpp *core_mesh.DataplaneResource) bool {
	for _, inbound := range dpp.Spec.GetNetworking().GetInbound() {
		if isInboundSelectedBy(tags, inbound) {
			return true
		}
	}
	return false
}

func isInboundSelectedBy(tags map[string]string, inbound *mesh_proto.Dataplane_Networking_Inbound) bool {
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
