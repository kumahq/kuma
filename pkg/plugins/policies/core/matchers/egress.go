package matchers

import (
	"github.com/pkg/errors"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/plugins"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
)

func EgressMatchedPolicies(rType core_model.ResourceType, tags map[string]string, resources xds_context.Resources, opts ...plugins.MatchedPoliciesOption) (core_xds.TypedMatchingPolicies, error) {
	mpOpts := plugins.NewMatchedPoliciesConfig(opts...)

	rl := resources.ListOrEmpty(rType)
	policies := []core_model.Resource{}

	for _, item := range rl.GetItems() {
		if !mpOpts.IncludeShadow && core_model.IsShadowedResource(item) {
			continue
		}
		policies = append(policies, item)
	}

	if len(policies) == 0 {
		return core_xds.TypedMatchingPolicies{Type: rType}, nil
	}

	// pick one item to cast and figure characteristics of all policies in the list
	p := policies[0]

	if _, ok := p.GetSpec().(core_model.Policy); !ok {
		return core_xds.TypedMatchingPolicies{}, errors.Errorf("resource type %v doesn't support TargetRef matching", p.Descriptor().Name)
	}

	_, isFrom := p.GetSpec().(core_model.PolicyWithFromList)
	_, isTo := p.GetSpec().(core_model.PolicyWithToList)

	if !isFrom && !isTo {
		return core_xds.TypedMatchingPolicies{}, nil
	}

	var fr core_rules.FromRules
	var tr core_rules.ToRules
	var err error

	switch {
	case isFrom && isTo:
		// we needed a strategy to choose what rules to apply on zone egress when a policy supports both "to" and "from".
		// Picking "from" rules works for us today, because there is only MeshFaultInjection policy that has both "to"
		// and "from" and is applied on zone egress. In the future, we might want to move the strategy down to the policy plugins.
		fr, err = processFromRules(tags, policies)
	case isFrom:
		fr, err = processFromRules(tags, policies)
	case isTo:
		fr, err = processToRules(tags, policies)
		if err != nil {
			return core_xds.TypedMatchingPolicies{}, err
		}
		tr, err = processToResourceRules(policies, resources)
	}

	if err != nil {
		return core_xds.TypedMatchingPolicies{}, err
	}

	return core_xds.TypedMatchingPolicies{
		Type:      rType,
		FromRules: fr,
		ToRules:   tr,
	}, nil
}

func processFromRules(
	tags map[string]string,
	policies []core_model.Resource,
) (core_rules.FromRules, error) {
	matchedPolicies := []core_model.Resource{}

	for _, policy := range policies {
		spec := policy.GetSpec().(core_model.Policy)
		if !serviceSelectedByTargetRef(spec.GetTargetRef(), tags) {
			continue
		}
		matchedPolicies = append(matchedPolicies, policy)
	}

	SortByTargetRef(matchedPolicies)

	return core_rules.BuildFromRules(map[core_rules.InboundListener][]core_model.Resource{
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
func processToRules(tags map[string]string, policies []core_model.Resource) (core_rules.FromRules, error) {
	var matchedPolicies []core_model.Resource

	for _, policy := range policies {
		spec := policy.GetSpec().(core_model.Policy)

		to, ok := spec.(core_model.PolicyWithToList)
		if !ok {
			return core_rules.FromRules{}, nil
		}

		for _, item := range to.GetToList() {
			if serviceSelectedByTargetRef(item.GetTargetRef(), tags) {
				matchedPolicies = append(matchedPolicies, policy)
			}
		}
	}

	SortByTargetRef(matchedPolicies)

	var toList []core_rules.PolicyItemWithMeta
	for _, policy := range matchedPolicies {
		toPolicy := policy.GetSpec().(core_model.PolicyWithToList)
		for _, item := range toPolicy.GetToList() {
			if !serviceSelectedByTargetRef(item.GetTargetRef(), tags) {
				continue
			}
			// convert 'to' policyItem to 'from' policyItem
			artificial := &artificialPolicyItem{
				conf:      item.GetDefault(),
				targetRef: policy.GetSpec().(core_model.Policy).GetTargetRef(),
			}
			toList = append(toList, core_rules.BuildPolicyItemsWithMeta([]core_model.PolicyItem{artificial},
				policy.GetMeta(), toPolicy.GetTargetRef())...)
		}
	}

	rules, err := core_rules.BuildRules(toList)
	if err != nil {
		return core_rules.FromRules{}, err
	}

	return core_rules.FromRules{
		Rules: map[core_rules.InboundListener]core_rules.Rules{{}: rules},
	}, nil
}

func processToResourceRules(policies []core_model.Resource, resources xds_context.Resources) (core_rules.ToRules, error) {
	toList, err := core_rules.BuildToList(policies, resources)
	if err != nil {
		return core_rules.ToRules{}, err
	}

	resourceRules, err := core_rules.BuildResourceRules(toList, resources)
	if err != nil {
		return core_rules.ToRules{}, err
	}
	return core_rules.ToRules{
		ResourceRules: resourceRules,
	}, nil
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

func serviceSelectedByTargetRef(tr common_api.TargetRef, tags map[string]string) bool {
	switch tr.Kind {
	case common_api.Mesh:
		return true
	case common_api.MeshSubset:
		return mesh_proto.TagSelector(tr.Tags).Matches(tags)
	case common_api.MeshService:
		return tr.Name == tags[mesh_proto.ServiceTag]
	case common_api.MeshServiceSubset:
		return tr.Name == tags[mesh_proto.ServiceTag] && mesh_proto.TagSelector(tr.Tags).Matches(tags)
	}
	return false
}
