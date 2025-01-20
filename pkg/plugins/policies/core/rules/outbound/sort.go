package outbound

import (
	"slices"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/common"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/sort"
)

func Sort[T interface {
	common.PolicyAttributes
	common.Entry[ToEntry]
}](list []T) {
	slices.SortStableFunc(list, sort.Compose(
		sort.CompareByPolicyAttributes[T],
		CompareByToEntry[T],
		sort.CompareByDisplayName[T],
	))
}

func CompareByToEntry[T common.Entry[ToEntry]](a, b T) int {
	if less := a.GetEntry().GetTargetRef().Kind.Compare(b.GetEntry().GetTargetRef().Kind); less != 0 {
		return less
	}

	if a.GetEntry().GetTargetRef().Kind == common_api.MeshService {
		sectionNameToNum := func(tr common_api.TargetRef) int {
			if tr.SectionName != "" {
				return 1
			}
			return 0
		}

		if less := sectionNameToNum(a.GetEntry().GetTargetRef()) - sectionNameToNum(b.GetEntry().GetTargetRef()); less != 0 {
			return less
		}
	}

	return 0
}
