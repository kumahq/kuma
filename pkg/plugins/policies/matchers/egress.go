package matchers

import (
	"sort"

	"github.com/pkg/errors"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
)

func EgressMatchedPolicies(rType core_model.ResourceType, es *core_mesh.ExternalServiceResource, resources xds_context.Resources) (core_xds.TypedMatchingPolicies, error) {
	policies := resources.ListOrEmpty(rType)

	fr, err := processFromRules(es, policies)
	if err != nil {
		return core_xds.TypedMatchingPolicies{}, err
	}

	tr, err := processToRules(es, policies)
	if err != nil {
		return core_xds.TypedMatchingPolicies{}, err
	}

	return core_xds.TypedMatchingPolicies{
		FromRules: fr,
		ToRules:   tr,
	}, nil
}

func processFromRules(
	es *core_mesh.ExternalServiceResource,
	rl core_model.ResourceList,
) (core_xds.FromRules, error) {
	matchedPolicies := []core_model.Resource{}

	for _, policy := range rl.GetItems() {
		spec, ok := policy.GetSpec().(core_xds.Policy)
		if !ok {
			return core_xds.FromRules{}, errors.Errorf("resource type %v doesn't support TargetRef matching", rl.GetItemType())
		}

		if !externalServiceSelectedByTargetRef(spec.GetTargetRef(), es) {
			continue
		}

		matchedPolicies = append(matchedPolicies, policy)
	}

	sort.Sort(ByTargetRef(matchedPolicies))

	return core_xds.BuildFromRules(map[core_xds.InboundListener][]core_model.Resource{
		core_xds.InboundListener{}: matchedPolicies, // egress always has only 1 listener, so we can use empty key
	})
}

func processToRules(
	es *core_mesh.ExternalServiceResource,
	rl core_model.ResourceList,
) (core_xds.ToRules, error) {
	matchedPolicies := []core_model.Resource{}

	for _, policy := range rl.GetItems() {
		spec, ok := policy.GetSpec().(core_xds.Policy)
		if !ok {
			return core_xds.ToRules{}, errors.Errorf("resource type %v doesn't support TargetRef matching", rl.GetItemType())
		}

		to, ok := spec.(core_xds.PolicyWithToList)
		if !ok {
			return core_xds.ToRules{}, nil
		}

		for _, item := range to.GetToList() {
			if externalServiceSelectedByTargetRef(item.GetTargetRef(), es) {
				matchedPolicies = append(matchedPolicies, policy)
			}
		}
	}

	sort.Sort(ByTargetRef(matchedPolicies))

	toList := []core_xds.PolicyItemWithMeta{}
	for _, policy := range matchedPolicies {
		for _, item := range policy.GetSpec().(core_xds.PolicyWithToList).GetToList() {
			if !externalServiceSelectedByTargetRef(item.GetTargetRef(), es) {
				continue
			}
			artificial := &artificialPolicyItem{
				conf:      item.GetDefault(),
				targetRef: policy.GetSpec().(core_xds.Policy).GetTargetRef(),
			}
			toList = append(toList, core_xds.BuildPolicyItemsWithMeta([]core_xds.PolicyItem{artificial},
				policy.GetMeta())...)
		}
	}

	rules, err := core_xds.BuildRules(toList)
	if err != nil {
		return core_xds.ToRules{}, err
	}

	return core_xds.ToRules{Rules: rules}, nil
}

type artificialPolicyItem struct {
	conf      interface{}
	targetRef common_api.TargetRef
}

func (a *artificialPolicyItem) GetTargetRef() common_api.TargetRef {
	return a.targetRef
}

func (a *artificialPolicyItem) GetDefault() interface{} {
	return a.conf
}

func externalServiceSelectedByTargetRef(tr common_api.TargetRef, es *core_mesh.ExternalServiceResource) bool {
	switch tr.Kind {
	case common_api.Mesh:
		return true
	case common_api.MeshSubset:
		return mesh_proto.TagSelector(tr.Tags).Matches(es.Spec.GetTags())
	case common_api.MeshService:
		return tr.Name == es.Spec.GetTags()[mesh_proto.ServiceTag]
	case common_api.MeshServiceSubset:
		return tr.Name == es.Spec.GetTags()[mesh_proto.ServiceTag] &&
			mesh_proto.TagSelector(tr.Tags).Matches(es.Spec.GetTags())
	}
	return false
}
