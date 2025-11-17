package k8s

import (
	chartcommon "helm.sh/helm/v4/pkg/chart/common"
	releaseutilv1 "helm.sh/helm/v4/pkg/release/v1/util"

	"github.com/kumahq/kuma/v2/pkg/util/data"
)

func SortResourcesByKind(files []data.File, kindsToSkip ...string) ([]data.File, error) {
	skippedKinds := map[string]struct{}{}
	for _, k := range kindsToSkip {
		skippedKinds[k] = struct{}{}
	}
	singleFile := data.JoinYAML(files)
	resources := releaseutilv1.SplitManifests(string(singleFile.Data))

	hooks, manifests, err := releaseutilv1.SortManifests(resources, chartcommon.VersionSet{"v1", "v1beta1", "v1alpha1"}, releaseutilv1.InstallOrder)
	if err != nil {
		return nil, err
	}

	result := make([]data.File, 0, len(manifests)+len(hooks))
	for _, manifest := range manifests {
		if _, ok := skippedKinds[manifest.Head.Kind]; !ok {
			result = append(result, data.File{Data: []byte(manifest.Content)})
		}
	}

	for _, hook := range hooks {
		result = append(result, data.File{Data: []byte(hook.Manifest)})
	}
	return result, nil
}
