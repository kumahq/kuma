package v1alpha1

import (
	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
)

func (x *To) GetDefault() any {
	if len(x.Rules) == 0 {
		return Rule{
			Default: RuleConf{
				BackendRefs: &[]common_api.BackendRef{{
					Kind:        common_api.BackendRefKind(x.TargetRef.Kind),
					Labels:      x.TargetRef.Labels,
					SectionName: x.TargetRef.SectionName,
					Weight:      pointer.To(uint(1)),
				}},
			},
		}
	}

	return x.Rules[0]
}
