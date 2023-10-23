package context

import (
	"golang.org/x/exp/maps"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/matchers"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	v1alpha1 "github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
)

type ReachableServicesGraph struct {
	rules map[string]core_rules.Rules
}

func NewReachableServicesGraph() *ReachableServicesGraph {
	return &ReachableServicesGraph{
		rules: map[string]core_rules.Rules{},
	}
}

func (r *ReachableServicesGraph) CanReach(fromTags map[string]string, toSvc string) bool {
	rule := r.rules[toSvc].Compute(core_rules.SubsetFromTags(fromTags))
	if rule == nil {
		return false
	}
	action := rule.Conf.(v1alpha1.Conf).Action
	return action == v1alpha1.Allow || action == v1alpha1.AllowWithShadowDeny
}

func (r *ReachableServicesGraph) CanReachFromAny(fromTagSets []mesh_proto.SingleValueTagSet, to string) bool {
	for _, from := range fromTagSets {
		if r.CanReach(from, to) {
			return true
		}
	}
	return false
}

func BuildReachableServicesGraph(services map[string]mesh_proto.SingleValueTagSet, mtps []*v1alpha1.MeshTrafficPermissionResource) (*ReachableServicesGraph, error) {
	resources := Resources{
		MeshLocalResources: map[core_model.ResourceType]core_model.ResourceList{
			v1alpha1.MeshTrafficPermissionType: &v1alpha1.MeshTrafficPermissionResourceList{
				Items: replaceSubsets(mtps),
			},
		},
	}

	graph := NewReachableServicesGraph()

	for service, tags := range services {
		// build artificial dpp for matching
		dp := mesh.NewDataplaneResource()
		dpTags := maps.Clone(tags)
		dpTags[mesh_proto.ServiceTag] = service
		dp.Spec = &mesh_proto.Dataplane{
			Networking: &mesh_proto.Dataplane_Networking{
				Address: "1.1.1.1",
				Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
					{
						Tags: dpTags,
						Port: 1234,
					},
				},
			},
		}

		matched, err := matchers.MatchedPolicies(v1alpha1.MeshTrafficPermissionType, dp, resources)
		if err != nil {
			return nil, err
		}

		rl, ok := matched.FromRules.Rules[core_rules.InboundListener{
			Address: "1.1.1.1",
			Port:    1234,
		}]
		if !ok {
			continue
		}

		graph.rules[service] = rl
	}

	return graph, nil
}

func replaceSubsets(mtps []*v1alpha1.MeshTrafficPermissionResource) []*v1alpha1.MeshTrafficPermissionResource {
	newMtps := make([]*v1alpha1.MeshTrafficPermissionResource, len(mtps))
	for i, mtp := range mtps {
		if mtp.Spec.TargetRef.Kind == common_api.MeshSubset || mtp.Spec.TargetRef.Kind == common_api.MeshServiceSubset {
			mtp = &v1alpha1.MeshTrafficPermissionResource{
				Meta: mtp.Meta,
				Spec: mtp.Spec.DeepCopy(),
			}
		}

		if mtp.Spec.TargetRef.Kind == common_api.MeshSubset {
			mtp.Spec.TargetRef.Kind = common_api.Mesh
			mtp.Spec.TargetRef.Tags = nil
		}
		if mtp.Spec.TargetRef.Kind == common_api.MeshServiceSubset {
			mtp.Spec.TargetRef.Kind = common_api.MeshService
			mtp.Spec.TargetRef.Tags = nil
		}
		newMtps[i] = mtp
	}
	return newMtps
}
