package server

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"

	"github.com/kumahq/kuma/pkg/core/dns/lookup"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/xds/topology"
)

type MeshSnapshotCache struct {
	cache  *cache.Cache
	rm     manager.ReadOnlyResourceManager
	types  []core_model.ResourceType
	ipFunc lookup.LookupIPFunc

	mutexes  map[string]*sync.Mutex
	mapMutex sync.Mutex // guards "mutexes" field
}

func NewMeshSnapshotCache(rm manager.ReadOnlyResourceManager, expirationTime time.Duration, types []core_model.ResourceType, ipFunc lookup.LookupIPFunc) *MeshSnapshotCache {
	return &MeshSnapshotCache{
		rm:      rm,
		types:   types,
		ipFunc:  ipFunc,
		mutexes: map[string]*sync.Mutex{},
		cache:   cache.New(expirationTime, time.Duration(int64(float64(expirationTime)*0.9))),
	}
}

func (c *MeshSnapshotCache) GetHash(ctx context.Context, mesh string) (string, error) {
	if hash, found := c.cache.Get(mesh); !found {
		mutex := c.mutexFor(mesh)
		mutex.Lock()
		if hash, found = c.cache.Get(mesh); !found {
			snapshot, err := GetMeshSnapshot(ctx, mesh, c.rm, c.types, c.ipFunc)
			if err != nil {
				mutex.Unlock()
				return "", err
			}
			hash := snapshot.Hash()
			c.cache.SetDefault(mesh, hash)
			mutex.Unlock()
			c.cleanMutexFor(mesh) // We need to cleanup mutexes from the map, otherwise we can see the memory leak.
			return hash, nil
		} else {
			mutex.Unlock()
			c.cleanMutexFor(mesh)
			return hash.(string), nil
		}
	} else {
		return hash.(string), nil
	}
}

func (c *MeshSnapshotCache) mutexFor(key string) *sync.Mutex {
	c.mapMutex.Lock()
	defer c.mapMutex.Unlock()
	mutex, exist := c.mutexes[key]
	if !exist {
		mutex = &sync.Mutex{}
		c.mutexes[key] = mutex
	}
	return mutex
}

func (c *MeshSnapshotCache) cleanMutexFor(key string) {
	c.mapMutex.Lock()
	delete(c.mutexes, key)
	c.mapMutex.Unlock()
}

type MeshSnapshot struct {
	Mesh      *core_mesh.MeshResource
	Resources map[core_model.ResourceType]core_model.ResourceList
}

func GetMeshSnapshot(ctx context.Context, meshName string, rm manager.ReadOnlyResourceManager, types []core_model.ResourceType, ipFunc lookup.LookupIPFunc) (*MeshSnapshot, error) {
	snapshot := &MeshSnapshot{
		Resources: map[core_model.ResourceType]core_model.ResourceList{},
	}

	mesh := &core_mesh.MeshResource{}
	if err := rm.Get(ctx, mesh, core_store.GetByKey(meshName, meshName)); err != nil {
		return nil, err
	}
	snapshot.Mesh = mesh

	for _, typ := range types {
		switch typ {
		case core_mesh.DataplaneType:
			dataplanes := &core_mesh.DataplaneResourceList{}
			if err := rm.List(ctx, dataplanes); err != nil {
				return nil, err
			}
			dataplanes.Items = topology.ResolveAddresses(xdsServerLog, ipFunc, dataplanes.Items)
			meshedDpsAndIngresses := &core_mesh.DataplaneResourceList{}
			for _, d := range dataplanes.Items {
				if d.GetMeta().GetMesh() == meshName || d.Spec.IsIngress() {
					_ = meshedDpsAndIngresses.AddItem(d)
				}
			}
			snapshot.Resources[typ] = meshedDpsAndIngresses
		default:
			rlist, err := registry.Global().NewList(typ)
			if err != nil {
				return nil, err
			}
			if err := rm.List(ctx, rlist, core_store.ListByMesh(meshName)); err != nil {
				return nil, err
			}
			snapshot.Resources[typ] = rlist
		}
	}
	return snapshot, nil
}

func (m *MeshSnapshot) Hash() string {
	resources := []core_model.Resource{
		m.Mesh,
	}
	for _, rl := range m.Resources {
		resources = append(resources, rl.GetItems()...)
	}
	return hashResources(resources...)
}

func hashResources(rs ...core_model.Resource) string {
	hashes := []string{}
	for _, r := range rs {
		hashes = append(hashes, hashResource(r))
	}
	sort.Strings(hashes)
	hash := md5.Sum([]byte(strings.Join(hashes, ",")))
	return hex.EncodeToString(hash[:])
}

func hashResource(r core_model.Resource) string {
	switch v := r.(type) {
	case *core_mesh.DataplaneResource:
		return strings.Join(
			[]string{string(v.GetType()),
				v.GetMeta().GetMesh(),
				v.GetMeta().GetName(),
				v.GetMeta().GetVersion(),
				v.Spec.Networking.Address}, ":")
	default:
		return strings.Join(
			[]string{string(v.GetType()),
				v.GetMeta().GetMesh(),
				v.GetMeta().GetName(),
				v.GetMeta().GetVersion()}, ":")
	}
}
