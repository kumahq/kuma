package k8s

import (
	helm_manifest "k8s.io/helm/pkg/manifest"
	helm_releaseutil "k8s.io/helm/pkg/releaseutil"

	"github.com/Kong/kuma/app/kumactl/pkg/install/data"
)

func SortResourcesByKind(files []data.File) []data.File {
	singleFile := data.JoinYAML(files)
	resources := helm_releaseutil.SplitManifests(string(singleFile.Data))
	manifests := helm_manifest.SplitManifests(resources)
	SortByKind(manifests)

	result := make([]data.File, len(manifests))
	for i, manifest := range manifests {
		result[i].Data = []byte(manifest.Content)
	}
	return result
}
