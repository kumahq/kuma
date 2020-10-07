package cla

import (
	"context"
	"fmt"
	"time"

	envoy_api_v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/patrickmn/go-cache"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/dns/lookup"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/xds/cache/once"
	"github.com/kumahq/kuma/pkg/xds/envoy/endpoints"
	"github.com/kumahq/kuma/pkg/xds/topology"
)

var (
	claCacheLog = core.Log.WithName("cla-cache")
)

type Cache struct {
	cache   *cache.Cache
	rm      manager.ReadOnlyResourceManager
	ipFunc  lookup.LookupIPFunc
	zone    string
	onceMap *once.Map
}

func NewCache(rm manager.ReadOnlyResourceManager, zone string, expirationTime time.Duration, ipFunc lookup.LookupIPFunc) *Cache {
	return &Cache{
		cache:   cache.New(expirationTime, time.Duration(int64(float64(expirationTime)*0.9))),
		rm:      rm,
		zone:    zone,
		ipFunc:  ipFunc,
		onceMap: once.NewMap(),
	}
}

func (c *Cache) GetCLA(ctx context.Context, meshName, service string) (*envoy_api_v2.ClusterLoadAssignment, error) {
	key := fmt.Sprintf("%s:%s", meshName, service)
	value, found := c.cache.Get(key)
	if found {
		return value.(*envoy_api_v2.ClusterLoadAssignment), nil
	}
	o := c.onceMap.Get(key)
	o.Do(func() (interface{}, error) {
		dataplanes, err := topology.GetDataplanes(claCacheLog, ctx, c.rm, c.ipFunc, meshName)
		if err != nil {
			return nil, err
		}
		mesh := &core_mesh.MeshResource{}
		if err := c.rm.Get(ctx, mesh, core_store.GetByKey(meshName, meshName)); err != nil {
			return nil, err
		}
		endpointMap := topology.BuildEndpointMap(dataplanes.Items, c.zone, mesh)
		cla := endpoints.CreateClusterLoadAssignment(service, endpointMap[service])
		c.cache.SetDefault(key, cla)
		c.onceMap.Delete(key)
		return cla, nil
	})
	if o.Err != nil {
		return nil, o.Err
	}
	return o.Value.(*envoy_api_v2.ClusterLoadAssignment), nil
}
