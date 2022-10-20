package k8s

import (
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/releaseutil"

	"github.com/kumahq/kuma/app/kumactl/pkg/install/data"
)

func SortResourcesByKind(files []data.File, kindsToSkip ...string) ([]data.File, error) {
	skippedKinds := map[string]struct{}{}
	for _, k := range kindsToSkip {
		skippedKinds[k] = struct{}{}
	}
	singleFile := data.JoinYAML(files)
	resources := releaseutil.SplitManifests(string(singleFile.Data))

	hooks, manifests, err := releaseutil.SortManifests(resources, chartutil.VersionSet{"v1", "v1beta1", "v1alpha1"}, releaseutil.InstallOrder)
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
