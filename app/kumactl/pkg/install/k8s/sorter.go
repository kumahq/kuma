package k8s

import (
	"log"

	"github.com/kumahq/kuma/app/kumactl/pkg/install/data"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/releaseutil"
)

func SortResourcesByKind(files []data.File) []data.File {
	singleFile := data.JoinYAML(files)
	resources := releaseutil.SplitManifests(string(singleFile.Data))

	hooks, manifests, err := releaseutil.SortManifests(resources, chartutil.VersionSet{"v1", "v1beta1", "v1alpha1"}, releaseutil.InstallOrder)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	result := make([]data.File, len(manifests)+len(hooks))
	for i, manifest := range manifests {
		result[i].Data = []byte(manifest.Content)
	}

	baseIdx := len(manifests) - 1
	for i, hook := range hooks {
		result[baseIdx+i].Data = []byte(hook.Manifest)
	}
	return result
}
