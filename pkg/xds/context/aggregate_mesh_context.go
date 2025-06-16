package context

import (
	"context"
	"encoding/base64"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/xds/cache/sha256"
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

	hash := aggregatedHash(meshContexts)

	egressByName := map[string]*core_mesh.ZoneEgressResource{}
	if len(meshContexts) > 0 {
		for _, egress := range meshContexts[0].Resources.ZoneEgresses().Items {
			egressByName[egress.Meta.GetName()] = egress
		}
	} else {
		var egressList core_mesh.ZoneEgressResourceList
		if err := resManager.List(ctx, &egressList, core_store.ListOrdered()); err != nil {
			return AggregatedMeshContexts{}, err
		}

		for _, egress := range egressList.GetItems() {
			egressByName[egress.GetMeta().GetName()] = egress.(*core_mesh.ZoneEgressResource)
		}

		hash = base64.StdEncoding.EncodeToString(core_model.ResourceListHash(&egressList))
	}

	result := AggregatedMeshContexts{
		Hash:               hash,
		Meshes:             meshes,
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
