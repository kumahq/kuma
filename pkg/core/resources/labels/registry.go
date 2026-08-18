package labels

import (
	"github.com/kumahq/kuma/v3/pkg/plugins/runtime/k8s/metadata"
)

// AllComputedLabels lists every label Compute may write: the registry keys plus
// the two labels supplied by the caller rather than derived from the spec.
//
// if changed sync with:
// https://github.com/Kong/shared-speakeasy/blob/b3ddd3ef1f31e42bfe71b96ea473493072f9742c/customtypes/kumalabels/kumalabels.go#L15
var AllComputedLabels = func() map[string]struct{} {
	all := map[string]struct{}{
		metadata.KumaServiceAccount: {},
		metadata.KumaWorkload:       {},
	}
	for key := range registry {
		all[key] = struct{}{}
	}
	return all
}()
