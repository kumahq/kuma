package context

import (
	"context"
	"strings"

	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v3/pkg/core/resources/manager"
	core_store "github.com/kumahq/kuma/v3/pkg/core/resources/store"
	"github.com/kumahq/kuma/v3/pkg/xds/cache/sha256"
)

type meshContextFetcher = func(ctx context.Context, meshName string) (MeshContext, error)

func AggregateMeshContexts(
	ctx context.Context,
	resManager manager.ReadOnlyResourceManager,
	fetcher meshContextFetcher,
) (AggregatedMeshContexts, error) {
	var meshList core_mesh.MeshResourceList
	if err := resManager.List(ctx, &meshList, core_store.ListOrdered()); err != nil {
		return AggregatedMeshContexts{}, err
	}

	var meshContexts []MeshContext
	meshContextsByName := map[string]MeshContext{}
	var meshes []*core_mesh.MeshResource
	for _, mesh := range meshList.Items {
		meshCtx, err := fetcher(ctx, mesh.GetMeta().GetName())
		if err != nil {
			if core_store.IsNotFound(err) {
				// When the mesh no longer exists it's likely because it was removed since, let's just skip it.
				continue
			}
			return AggregatedMeshContexts{}, err
		}
		meshContexts = append(meshContexts, meshCtx)
		meshContextsByName[mesh.Meta.GetName()] = meshCtx
		meshes = append(meshes, mesh)
	}

	result := AggregatedMeshContexts{
		Hash:               aggregatedHash(meshContexts),
		Meshes:             meshes,
		MeshContextsByName: meshContextsByName,
	}
	return result, nil
}

func aggregatedHash(meshContexts []MeshContext) string {
	var hash strings.Builder
	for _, meshCtx := range meshContexts {
		hash.WriteString(meshCtx.Hash)
	}
	return sha256.Hash(hash.String())
}
