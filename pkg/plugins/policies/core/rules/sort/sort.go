package sort

import (
	"cmp"
	"strconv"
	"time"

	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/common"
	"github.com/kumahq/kuma/v3/pkg/plugins/runtime/k8s/metadata"
)

func CompareByPolicyAttributes[T common.PolicyAttributes](a, b T) int {
	if less := a.GetTopLevel().Kind.Compare(b.GetTopLevel().Kind); less != 0 {
		return less
	}

	if less := a.GetTopLevel().CompareDataplaneKind(b.GetTopLevel()); less != 0 {
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

// CompareByRouteCreationTimestamp sorts by metadata.GatewayAPIRouteCreationTimestampLabel, oldest last; missing/unparsable label sorts as oldest.
func CompareByRouteCreationTimestamp[T common.PolicyAttributes](a, b T) int {
	keyOf := func(item T) time.Time {
		label, ok := item.GetResourceMeta().GetLabels()[metadata.GatewayAPIRouteCreationTimestampLabel]
		if !ok {
			return time.Time{}
		}
		sec, err := strconv.ParseInt(label, 10, 64)
		if err != nil {
			return time.Time{}
		}
		return time.Unix(sec, 0)
	}

	return keyOf(b).Compare(keyOf(a))
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
