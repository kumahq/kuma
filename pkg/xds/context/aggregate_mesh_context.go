package context

import (
	"context"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/xds/cache/sha256"
)

type meshContextFetcher = func(ctx context.Context, meshName string) (MeshContext, error)

func AggregateMeshContexts(
	ctx context.Context,
	resManager manager.ReadOnlyResourceManager,
	fetcher meshContextFetcher,
) (MeshContexts, error) {
	var meshList core_mesh.MeshResourceList
	if err := resManager.List(ctx, &meshList, core_store.ListOrdered()); err != nil {
		return MeshContexts{}, err
	}

	var meshContexts []MeshContext
	meshContextsByName := map[string]MeshContext{}
	for _, mesh := range meshList.Items {
		meshCtx, err := fetcher(ctx, mesh.GetMeta().GetName())
		if err != nil {
			return MeshContexts{}, err
		}
		meshContexts = append(meshContexts, meshCtx)
		meshContextsByName[mesh.Meta.GetName()] = meshCtx
	}

	egressByName := map[string]*core_mesh.ZoneEgressResource{}
	if len(meshContexts) > 0 {
		for _, egress := range meshContexts[0].Resources.ZoneEgresses().Items {
			egressByName[egress.Meta.GetName()] = egress
		}
	}

	result := MeshContexts{
		Hash:               aggregatedHash(meshContexts),
		Meshes:             meshList.Items,
		MeshContextsByName: meshContextsByName,
		ZoneEgressByName:   egressByName,
	}
	return result, nil
}

func aggregatedHash(meshContexts []MeshContext) string {
	var hash string
	for _, meshCtx := range meshContexts {
		hash += meshCtx.Hash
	}
	return sha256.Hash(hash)
}
