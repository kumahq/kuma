package sort

import (
	"cmp"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/common"
)

func CompareByPolicyAttributes[T common.PolicyAttributes](a, b T) int {
	if less := a.GetTopLevel().Kind.Compare(b.GetTopLevel().Kind); less != 0 {
		return less
	}

	o1, _ := core_model.ResourceOrigin(a.GetResourceMeta())
	o2, _ := core_model.ResourceOrigin(b.GetResourceMeta())
	if less := o1.Compare(o2); less != 0 {
		return less
	}

	if less := core_model.PolicyRole(a.GetResourceMeta()).Compare(core_model.PolicyRole(b.GetResourceMeta())); less != 0 {
		return less
	}

	return 0
}

func CompareByDisplayName[T common.PolicyAttributes](a, b T) int {
	return cmp.Compare(core_model.GetDisplayName(b.GetResourceMeta()), core_model.GetDisplayName(a.GetResourceMeta()))
}

func Compose[T any](comparators ...func(a, b T) int) func(a, b T) int {
	return func(a, b T) int {
		for _, comparator := range comparators {
			if less := comparator(a, b); less != 0 {
				return less
			}
		}
		return 0
	}
}
