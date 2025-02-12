package v1alpha1

import (
	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	"github.com/kumahq/kuma/pkg/util/pointer"
)

func (x *To) GetDefault() interface{} {
	if len(pointer.Deref(x.Rules)) == 0 {
		return Rule{
			Default: RuleConf{
				BackendRefs: []common_api.BackendRef{{
					TargetRef: x.TargetRef,
					Weight:    pointer.To(uint(1)),
				}},
			},
		}
	}

	return pointer.Deref(x.Rules)[0]
}
