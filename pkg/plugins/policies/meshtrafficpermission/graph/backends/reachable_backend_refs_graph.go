package backends

import (
	"golang.org/x/exp/maps"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	ms_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/matchers"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	mtp_api "github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/xds/context"
)

var log = core.Log.WithName("rms-graph")

type BackendKey struct {
	Kind string
	Name string
}

type Graph struct {
	rules map[BackendKey]core_rules.Rules
}

func NewGraph() *Graph {
	return &Graph{
		rules: map[BackendKey]core_rules.Rules{},
	}
}

func (r *Graph) CanReach(map[string]string, map[string]string) bool {
	panic("not implemented")
}

func (r *Graph) CanReachBackend(fromTags map[string]string, backendRef *mesh_proto.Dataplane_Networking_Outbound_BackendRef) bool {
	log.Info("TEST_LOG", "rules", r.rules, "fromTags", fromTags, "backendRef", backendRef)
	if backendRef.Kind == "MeshExternalService" {
		return true
	}
	rule := r.rules[BackendKey{
		Kind: backendRef.Kind,
		Name: backendRef.Name,
	}].Compute(core_rules.SubsetFromTags(fromTags))
	if rule == nil {
		return false
	}
	action := rule.Conf.(mtp_api.Conf).Action
	return action == mtp_api.Allow || action == mtp_api.AllowWithShadowDeny
}

func Builder(meshName string, resources context.Resources) context.ReachableServicesGraph {
	mtps := resources.ListOrEmpty(mtp_api.MeshTrafficPermissionType).(*mtp_api.MeshTrafficPermissionResourceList)
	ms := resources.ListOrEmpty(ms_api.MeshServiceType).(*ms_api.MeshServiceResourceList)
	return BuildGraph(ms.Items, mtps.Items)
}

func BuildGraph(meshServices []*ms_api.MeshServiceResource, mtps []*mtp_api.MeshTrafficPermissionResource) *Graph {
	graph := NewGraph()
	for _, ms := range meshServices {
		dpTags := maps.Clone(ms.Spec.Selector.DataplaneTags)
		if origin, ok := core_model.ResourceOrigin(ms.GetMeta()); ok {
			dpTags[mesh_proto.ResourceOriginLabel] = string(origin)
		}
		if ms.GetMeta().GetLabels() != nil && ms.GetMeta().GetLabels()[mesh_proto.ZoneTag] != "" {
			dpTags[mesh_proto.ZoneTag] = ms.GetMeta().GetLabels()[mesh_proto.ZoneTag]
		}
		resources := context.Resources{
			MeshLocalResources: map[core_model.ResourceType]core_model.ResourceList{
				mtp_api.MeshTrafficPermissionType: &mtp_api.MeshTrafficPermissionResourceList{
					Items: trimNotSupportedTags(mtps, dpTags),
				},
			},
		}
		// build artificial dpp for matching
		dp := mesh.NewDataplaneResource()

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

		matched, err := matchers.MatchedPolicies(mtp_api.MeshTrafficPermissionType, dp, resources)
		if err != nil {
			log.Error(err, "service could not be matched. It won't be reached by any other service", "service", ms.Meta.GetName())
			continue // it's better to ignore one service that to break the whole graph
		}

		rl, ok := matched.FromRules.Rules[core_rules.InboundListener{
			Address: "1.1.1.1",
			Port:    1234,
		}]
		if !ok {
			continue
		}

		graph.rules[BackendKey{
			Kind: "MeshService",
			Name: ms.Meta.GetName(),
		}] = rl
	}
	log.Info("TEST_LOG_BUILD", "rules", graph.rules)
	return graph
}

// trimNotSupportedTags replaces tags present in subsets of top-level target ref.
// Because we need to do policy matching on services instead of individual proxies, we have to handle subsets in a special way.
// What we do is we only support subsets with predefined tags listed in SupportedTags.
// This assumes that tags listed in SupportedTags have the same value between all instances of a given service.
// Otherwise, we trim the tags making the target ref subset wider.
//
// Alternatively, we could have computed all common tags between instances of a given service and then allow subsets with those common tags.
// However, this would require calling this function for every service.
func trimNotSupportedTags(mtps []*mtp_api.MeshTrafficPermissionResource, supportedTags map[string]string) []*mtp_api.MeshTrafficPermissionResource {
	newMtps := make([]*mtp_api.MeshTrafficPermissionResource, len(mtps))
	for i, mtp := range mtps {
		if len(mtp.Spec.TargetRef.Tags) > 0 {
			filteredTags := map[string]string{}
			for tag, val := range mtp.Spec.TargetRef.Tags {
				if _, ok := supportedTags[tag]; ok {
					filteredTags[tag] = val
				}
			}
			if len(filteredTags) != len(mtp.Spec.TargetRef.Tags) {
				mtp = &mtp_api.MeshTrafficPermissionResource{
					Meta: mtp.Meta,
					Spec: mtp.Spec.DeepCopy(),
				}
				mtp.Spec.TargetRef.Tags = filteredTags
			}
		}
		newMtps[i] = mtp
	}
	return newMtps
}
