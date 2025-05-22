package util

import (
	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/matchers"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	mtp_api "github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/xds/context"
)

func ComputeMtpRulesForTags(
	dpTags map[string]string,
	mtpsWithTrimmedTags []*mtp_api.MeshTrafficPermissionResource,
) (core_rules.Rules, bool, error) {
	resources := context.Resources{
		MeshLocalResources: map[core_model.ResourceType]core_model.ResourceList{
			mtp_api.MeshTrafficPermissionType: &mtp_api.MeshTrafficPermissionResourceList{
				Items: filterMTPsWithKindDataplane(mtpsWithTrimmedTags),
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
		return nil, false, err
	}

	rl, ok := matched.FromRules.Rules[core_rules.InboundListener{
		Address: "1.1.1.1",
		Port:    1234,
	}]
	return rl, ok, nil
}

// TODO Currently autoreachable services functionality is based on tags. When MTP selects dataplanes by Dataplane kind
// we don't have tags to work with. We should rethink how to implement autoreachable services with Dataplane kind and new spec.rules
// This should be covered by: https://github.com/kumahq/kuma/issues/12403
func filterMTPsWithKindDataplane(mtps []*mtp_api.MeshTrafficPermissionResource) []*mtp_api.MeshTrafficPermissionResource {
	var filteredMtps []*mtp_api.MeshTrafficPermissionResource
	for _, mtp := range mtps {
		if mtp.Spec != nil && mtp.Spec.GetTargetRef().Kind == common_api.Dataplane {
			continue
		}
		filteredMtps = append(filteredMtps, mtp)
	}
	return filteredMtps
}
