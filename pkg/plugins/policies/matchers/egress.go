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

	if len(policies.GetItems()) == 0 {
		return core_xds.TypedMatchingPolicies{}, nil
	}

	p := policies.GetItems()[0]

	if _, ok := p.GetSpec().(core_xds.Policy); !ok {
		return core_xds.TypedMatchingPolicies{}, errors.Errorf("resource type %v doesn't support TargetRef matching", p.Descriptor().Name)
	}

	_, isFrom := p.GetSpec().(core_xds.PolicyWithFromList)
	_, isTo := p.GetSpec().(core_xds.PolicyWithToList)

	if isFrom && isTo {
		return core_xds.TypedMatchingPolicies{}, errors.Errorf("zone egress doesn't support policies that have both 'from' and 'to'")
	}

	if !isFrom && !isTo {
		return core_xds.TypedMatchingPolicies{}, nil
	}

	var fr core_xds.FromRules
	var err error
	if isFrom {
		fr, err = processFromRules(es, policies)
	} else {
		fr, err = processToRules(es, policies)
	}
	if err != nil {
		return core_xds.TypedMatchingPolicies{}, err
	}

	return core_xds.TypedMatchingPolicies{
		Type:      rType,
		FromRules: fr,
	}, nil
}

func processFromRules(
	es *core_mesh.ExternalServiceResource,
	rl core_model.ResourceList,
) (core_xds.FromRules, error) {
	matchedPolicies := []core_model.Resource{}

	for _, policy := range rl.GetItems() {
		spec := policy.GetSpec().(core_xds.Policy)
		if !externalServiceSelectedByTargetRef(spec.GetTargetRef(), es) {
			continue
		}
		matchedPolicies = append(matchedPolicies, policy)
	}

	sort.Sort(ByTargetRef(matchedPolicies))

	return core_xds.BuildFromRules(map[core_xds.InboundListener][]core_model.Resource{
		{}: matchedPolicies, // egress always has only 1 listener, so we can use empty key
	})
}

// It's not natural for zone egress to have 'to' policies. It doesn't make sense to target
// external service in the top-level targetRef and specify 'to' array (simply because we don't
// have access to the external service outbounds). But there are situations when we're
// targeting external service in the 'to' array, and we need to make adjustments on the Egress, i.e:
//
// type: MeshLoadBalancingStrategy
// spec:
//
//	targetRef:
//	  kind: Mesh
//	to:
//	  - targetRef:
//	      kind: MeshService
//	      name: external-service-1
//	    default:
//	      localityAwareness:
//	        disabled: true
//
// In this case we need to apply the policy to the Egress. The problem is that Egress is
// a single point for multiple clients. This means we have to specify different configurations
// for different clients. In order to easily get a list of rules for 'to' policy on Egress
// we have to convert it to 'from' policy, i.e. the policy above will be converted to artificially
// created policy:
//
// spec:
//
//	targetRef:
//	  kind: MeshService
//	  name: external-service-1
//	from:
//	  - targetRef:
//	      kind: Mesh
//	    default:
//	      localityAwareness:
//	        disabled: true
//
// that's why processToRules() method produces FromRules for the Egress.
func processToRules(
	es *core_mesh.ExternalServiceResource,
	rl core_model.ResourceList,
) (core_xds.FromRules, error) {
	matchedPolicies := []core_model.Resource{}

	for _, policy := range rl.GetItems() {
		spec := policy.GetSpec().(core_xds.Policy)

		to, ok := spec.(core_xds.PolicyWithToList)
		if !ok {
			return core_xds.FromRules{}, nil
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
			// convert 'to' policyItem to 'from' policyItem
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
		return core_xds.FromRules{}, err
	}

	return core_xds.FromRules{Rules: map[core_xds.InboundListener]core_xds.Rules{
		{}: rules,
	}}, nil
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
