package util

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/matchers"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	mtp_api "github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/xds/context"
)

func ComputeMtpRulesForTags(
	dpTags map[string]string,
	mtpsWithTrimmedTags []*mtp_api.MeshTrafficPermissionResource,
) (rules.Rules, bool, error) {
	resources := context.Resources{
		MeshLocalResources: map[core_model.ResourceType]core_model.ResourceList{
			mtp_api.MeshTrafficPermissionType: &mtp_api.MeshTrafficPermissionResourceList{
				Items: mtpsWithTrimmedTags,
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
