package services

import (
	"golang.org/x/exp/maps"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	mtp_api "github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	graph_util "github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/graph/util"
)

var log = core.Log.WithName("rs-graph")

var SupportedTags = map[string]struct{}{
	mesh_proto.KubeNamespaceTag: {},
	mesh_proto.KubeServiceTag:   {},
	mesh_proto.KubePortTag:      {},
	mesh_proto.ServiceTag:       {},
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

func BuildRules(services map[string]mesh_proto.SingleValueTagSet, mtps []*mtp_api.MeshTrafficPermissionResource) map[string]core_rules.Rules {
	trimmedMtps := trimNotSupportedTags(mtps)
	rules := map[string]core_rules.Rules{}
	for service, tags := range services {
		// build artificial dpp for matching
		dpTags := maps.Clone(tags)
		dpTags[mesh_proto.ServiceTag] = service
		rl, ok, err := graph_util.ComputeMtpRulesForTags(dpTags, trimmedMtps)
		if err != nil {
			log.Error(err, "service could not be matched. It won't be reached by any other service", "service", service)
			continue // it's better to ignore one service that to break the whole graph
		}
		if !ok {
			continue
		}
		rules[service] = rl
	}

	return rules
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
		if mtp.Spec != nil && mtp.Spec.TargetRef != nil && len(mtp.Spec.TargetRef.Tags) > 0 {
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
