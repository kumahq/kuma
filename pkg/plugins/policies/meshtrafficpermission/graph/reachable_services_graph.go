package graph

import (
	"golang.org/x/exp/maps"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/matchers"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	mtp_api "github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/xds/context"
)

var log = core.Log.WithName("rs-graph")

var SupportedTags = map[string]struct{}{
	mesh_proto.KubeNamespaceTag: {},
	mesh_proto.KubeServiceTag:   {},
	mesh_proto.KubePortTag:      {},
}

type Graph struct {
	rules map[string]core_rules.Rules
}

func NewGraph() *Graph {
	return &Graph{
		rules: map[string]core_rules.Rules{},
	}
}

func (r *Graph) CanReach(fromTags map[string]string, toTags map[string]string) bool {
	if _, crossMeshTagExist := toTags[mesh_proto.MeshTag]; crossMeshTagExist {
		// we cannot compute graph for cross mesh, so it's better to allow the traffic
		return true
	}
	rule := r.rules[toTags[mesh_proto.ServiceTag]].Compute(core_rules.Element(fromTags))
	if rule == nil {
		return false
	}
	action := rule.Conf.(mtp_api.Conf).Action
	return action == mtp_api.Allow || action == mtp_api.AllowWithShadowDeny
}

func Builder(meshName string, resources context.Resources) context.ReachableServicesGraph {
	services := BuildServices(
		meshName,
		resources.Dataplanes().Items,
		resources.ExternalServices().Items,
		resources.ZoneIngresses().Items,
	)
	mtps := resources.ListOrEmpty(mtp_api.MeshTrafficPermissionType).(*mtp_api.MeshTrafficPermissionResourceList)
	return BuildGraph(services, mtps.Items)
}

// BuildServices we could just take result of xds_topology.VIPOutbounds, however it does not have a context of additional tags
func BuildServices(
	meshName string,
	dataplanes []*mesh.DataplaneResource,
	externalServices []*mesh.ExternalServiceResource,
	zoneIngresses []*mesh.ZoneIngressResource,
) map[string]mesh_proto.SingleValueTagSet {
	services := map[string]mesh_proto.SingleValueTagSet{}
	addSvc := func(tags map[string]string) {
		svc := tags[mesh_proto.ServiceTag]
		if _, ok := services[svc]; ok {
			return
		}
		services[svc] = map[string]string{}
		for tag := range SupportedTags {
			if value := tags[tag]; value != "" {
				services[svc][tag] = value
			}
		}
	}

	for _, dp := range dataplanes {
		for _, tagSet := range dp.Spec.SingleValueTagSets() {
			addSvc(tagSet)
		}
	}
	for _, zi := range zoneIngresses {
		for _, availableSvc := range zi.Spec.GetAvailableServices() {
			if meshName != availableSvc.Mesh {
				continue
			}
			addSvc(availableSvc.Tags)
		}
	}
	for _, es := range externalServices {
		addSvc(es.Spec.Tags)
	}
	return services
}

func BuildGraph(services map[string]mesh_proto.SingleValueTagSet, mtps []*mtp_api.MeshTrafficPermissionResource) *Graph {
	resources := context.Resources{
		MeshLocalResources: map[core_model.ResourceType]core_model.ResourceList{
			mtp_api.MeshTrafficPermissionType: &mtp_api.MeshTrafficPermissionResourceList{
				Items: trimNotSupportedTags(mtps),
			},
		},
	}

	graph := NewGraph()

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

		matched, err := matchers.MatchedPolicies(mtp_api.MeshTrafficPermissionType, dp, resources)
		if err != nil {
			log.Error(err, "service could not be matched. It won't be reached by any other service", "service", service)
			continue // it's better to ignore one service that to break the whole graph
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
func trimNotSupportedTags(mtps []*mtp_api.MeshTrafficPermissionResource) []*mtp_api.MeshTrafficPermissionResource {
	newMtps := make([]*mtp_api.MeshTrafficPermissionResource, len(mtps))
	for i, mtp := range mtps {
		if len(mtp.Spec.TargetRef.Tags) > 0 {
			filteredTags := map[string]string{}
			for tag, val := range mtp.Spec.TargetRef.Tags {
				if _, ok := SupportedTags[tag]; ok {
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
