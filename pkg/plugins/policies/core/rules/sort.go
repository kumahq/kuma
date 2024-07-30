package rules

import (
	"cmp"
	"slices"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
)

func SortByTargetRefV2(list []PolicyItemWithMeta) {
	slices.SortStableFunc(list, func(a, b PolicyItemWithMeta) int {
		if less := a.TopLevel.Kind.Compare(b.TopLevel.Kind); less != 0 {
			return less
		}

		o1, _ := core_model.ResourceOrigin(a.ResourceMeta)
		o2, _ := core_model.ResourceOrigin(b.ResourceMeta)
		if less := o1.Compare(o2); less != 0 {
			return less
		}

		if less := core_model.PolicyRole(a.ResourceMeta).Compare(core_model.PolicyRole(b.ResourceMeta)); less != 0 {
			return less
		}

		if less := a.PolicyItem.GetTargetRef().Kind.Compare(b.PolicyItem.GetTargetRef().Kind); less != 0 {
			return less
		}

		if a.PolicyItem.GetTargetRef().Kind == common_api.MeshService {
			sectionNameToNum := func(tr common_api.TargetRef) int {
				if tr.SectionName != "" {
					return 1
				}
				return 0
			}

			if less := sectionNameToNum(a.PolicyItem.GetTargetRef()) - sectionNameToNum(b.PolicyItem.GetTargetRef()); less != 0 {
				return less
			}
		}

		return cmp.Compare(core_model.GetDisplayName(b.ResourceMeta), core_model.GetDisplayName(a.ResourceMeta))
	})
}
