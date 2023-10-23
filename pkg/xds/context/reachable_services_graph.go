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
	FromAll     map[string]struct{}
	Connections map[string]map[string]struct{}
	rules       map[string]core_rules.Rules
}

func NewReachableServicesGraph() *ReachableServicesGraph {
	return &ReachableServicesGraph{
		FromAll:     map[string]struct{}{},
		Connections: map[string]map[string]struct{}{},
	}
}

func (r *ReachableServicesGraph) CanReach(fromTags map[string]string, toSvc string) bool {
	_, fromAllOk := r.FromAll[toSvc]
	if fromAllOk {
		return true
	}
	connectionsTo, ok := r.Connections[fromTags[mesh_proto.ServiceTag]]
	if ok {
		_, canReach := connectionsTo[toSvc]
		return canReach
	}
	return false
	//rule := r.rules[toSvc].Compute(core_rules.SubsetFromTags(fromTags))
	//return ruleAllowsTraffic(rule)
}

func (r *ReachableServicesGraph) MarkReachableFromAll(svc string) {
	r.FromAll[svc] = struct{}{}
}

func (r *ReachableServicesGraph) MarkReachable(from, to string) {
	_, ok := r.Connections[from]
	if !ok {
		r.Connections[from] = map[string]struct{}{}
	}
	r.Connections[from][to] = struct{}{}
}

func (r *ReachableServicesGraph) CanReachFromAny(fromSvcs []string, to string) bool {
	for _, from := range fromSvcs {
		fromTags := map[string]string{
			mesh_proto.ServiceTag: from,
		}
		if r.CanReach(fromTags, to) {
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

		if meshRule := rl.Compute(core_rules.MeshSubset()); ruleAllowsTraffic(meshRule) {
			graph.MarkReachableFromAll(service)
		}

		var reachableFrom []string
		for fromSvc, _ := range services {
			if meshRule := rl.Compute(core_rules.MeshService(fromSvc)); ruleAllowsTraffic(meshRule) {
				reachableFrom = append(reachableFrom, fromSvc)
			}
		}

		if len(reachableFrom) == len(services) {
			graph.MarkReachableFromAll(service)
		} else {
			for _, fromSvc := range reachableFrom {
				graph.MarkReachable(fromSvc, service)
			}
		}
	}

	return graph, nil
}

func ruleAllowsTraffic(rule *core_rules.Rule) bool {
	if rule == nil {
		return false
	}
	return actionAllowsTraffic(rule.Conf.(v1alpha1.Conf).Action)
}

func actionAllowsTraffic(action v1alpha1.Action) bool {
	return action == v1alpha1.Allow || action == v1alpha1.AllowWithShadowDeny
}

func replaceSubsets(mtps []*v1alpha1.MeshTrafficPermissionResource) []*v1alpha1.MeshTrafficPermissionResource {
	newMtps := make([]*v1alpha1.MeshTrafficPermissionResource, len(mtps))
	for i, mtp := range mtps {
		var newFroms []v1alpha1.From
		anyFromModified := false
		for _, from := range mtp.Spec.From {
			if from.TargetRef.Kind == common_api.MeshSubset {
				anyFromModified = true
				if actionAllowsTraffic(from.Default.Action) {
					from.TargetRef.Kind = common_api.Mesh
					from.TargetRef.Tags = nil
				} else {
					continue // ignore deny for subsets
				}
			}
			if from.TargetRef.Kind == common_api.MeshServiceSubset {
				anyFromModified = true
				if actionAllowsTraffic(from.Default.Action) {
					from.TargetRef.Kind = common_api.MeshService
					from.TargetRef.Tags = nil
				} else {
					continue // ignore deny for subsets
				}
			}
			newFroms = append(newFroms, from)
		}
		if anyFromModified || mtp.Spec.TargetRef.Kind == common_api.MeshSubset || mtp.Spec.TargetRef.Kind == common_api.MeshServiceSubset {
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
		if anyFromModified {
			mtp.Spec.From = newFroms
		}
		newMtps[i] = mtp
	}
	return newMtps
}
