package backends

import (
	"golang.org/x/exp/maps"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	ms_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	mtp_api "github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	graph_util "github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/graph/util"
)

var log = core.Log.WithName("rms-graph")

type BackendKey struct {
	Kind string
	Name string
}

func BuildRules(meshServices []*ms_api.MeshServiceResource, mtps []*mtp_api.MeshTrafficPermissionResource) map[BackendKey]core_rules.Rules {
	rules := map[BackendKey]core_rules.Rules{}
	for _, ms := range meshServices {
		dpTags := maps.Clone(ms.Spec.Selector.DataplaneTags)
		if origin, ok := core_model.ResourceOrigin(ms.GetMeta()); ok {
			dpTags[mesh_proto.ResourceOriginLabel] = string(origin)
		}
		if ms.GetMeta().GetLabels() != nil && ms.GetMeta().GetLabels()[mesh_proto.ZoneTag] != "" {
			dpTags[mesh_proto.ZoneTag] = ms.GetMeta().GetLabels()[mesh_proto.ZoneTag]
		}
		rl, ok, err := graph_util.ComputeMtpRulesForTags(dpTags, trimNotSupportedTags(mtps, dpTags))
		if err != nil {
			log.Error(err, "service could not be matched. It won't be reached by any other service", "service", ms.Meta.GetName())
			continue
		}
		if !ok {
			continue
		}
		rules[BackendKey{
			Kind: string(common_api.MeshService),
			Name: ms.Meta.GetName(),
		}] = rl
	}
	return rules
}

// trimNotSupportedTags removes tags that are not available in MeshService.dpTags + kuma.io/origin and kuma.io/zone
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
