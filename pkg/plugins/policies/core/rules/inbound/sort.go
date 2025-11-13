package inbound

import (
	"slices"

	"github.com/kumahq/kuma/v2/pkg/plugins/policies/core/rules/common"
	"github.com/kumahq/kuma/v2/pkg/plugins/policies/core/rules/sort"
)

func Sort[T common.PolicyAttributes](list []T) {
	slices.SortStableFunc(list, sort.Compose(
		sort.CompareByPolicyAttributes[T],
		sort.CompareByDisplayName[T],
	))
}
