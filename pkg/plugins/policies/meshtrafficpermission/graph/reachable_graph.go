package graph

import (
	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	ms_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	mtp_api "github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	graph_backends "github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/graph/backends"
	graph_services "github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/graph/services"
	"github.com/kumahq/kuma/pkg/xds/context"
)

type Graph struct {
	rules        map[string]core_rules.Rules
	backendRules map[core_model.TypedResourceIdentifier]core_rules.Rules
}

func NewGraph(rules map[string]core_rules.Rules, backendRules map[core_model.TypedResourceIdentifier]core_rules.Rules) *Graph {
	return &Graph{
		rules:        rules,
		backendRules: backendRules,
	}
}

func (r *Graph) CanReach(fromTags map[string]string, toTags map[string]string) bool {
	if _, crossMeshTagExist := toTags[mesh_proto.MeshTag]; crossMeshTagExist {
		// we cannot compute graph for cross mesh, so it's better to allow the traffic
		return true
	}
	rule := r.rules[toTags[mesh_proto.ServiceTag]].Compute(core_rules.SubsetFromTags(fromTags))
	if rule == nil {
		return false
	}
	action := rule.Conf.(mtp_api.Conf).Action
	return action == mtp_api.Allow || action == mtp_api.AllowWithShadowDeny
}

func (r *Graph) CanReachBackend(fromTags map[string]string, backendIdentifier core_model.TypedResourceIdentifier) bool {
	if backendIdentifier.ResourceType == core_model.ResourceType(common_api.MeshExternalService) || backendIdentifier.ResourceType == core_model.ResourceType(common_api.MeshMultiZoneService) {
		return true
	}
	noPort := core_model.TypedResourceIdentifier{
		ResourceIdentifier: backendIdentifier.ResourceIdentifier,
		ResourceType:       backendIdentifier.ResourceType,
	}
	rule := r.backendRules[noPort].Compute(core_rules.SubsetFromTags(fromTags))
	if rule == nil {
		return false
	}
	action := rule.Conf.(mtp_api.Conf).Action
	return action == mtp_api.Allow || action == mtp_api.AllowWithShadowDeny
}

func Builder(meshName string, resources context.Resources) context.ReachableServicesGraph {
	services := graph_services.BuildServices(
		meshName,
		resources.Dataplanes().Items,
		resources.ExternalServices().Items,
		resources.ZoneIngresses().Items,
	)
	ms := resources.ListOrEmpty(ms_api.MeshServiceType).(*ms_api.MeshServiceResourceList)
	mtps := resources.ListOrEmpty(mtp_api.MeshTrafficPermissionType).(*mtp_api.MeshTrafficPermissionResourceList)
	rules := graph_services.BuildRules(services, mtps.Items)
	backendRules := graph_backends.BuildRules(ms.Items, mtps.Items)
	return NewGraph(rules, backendRules)
}
