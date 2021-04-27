package mesh

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/kumahq/kuma/pkg/metrics"

	"github.com/kumahq/kuma/pkg/xds/cache/once"

	"github.com/patrickmn/go-cache"

	"github.com/kumahq/kuma/pkg/core/dns/lookup"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
)

// Cache is needed to share and cache Hashes among goroutines which
// reconcile Dataplane's state. Calculating hash is a heavy operation
// that requires fetching all the resources belonging to the Mesh.
type Cache struct {
	cache   *cache.Cache
	rm      manager.ReadOnlyResourceManager
	types   []core_model.ResourceType
	ipFunc  lookup.LookupIPFunc
	onceMap *once.Map
	metrics *prometheus.GaugeVec
}

func NewCache(
	rm manager.ReadOnlyResourceManager,
	expirationTime time.Duration,
	types []core_model.ResourceType,
	ipFunc lookup.LookupIPFunc,
	metrics metrics.Metrics,
) (*Cache, error) {
	metric := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "mesh_cache",
		Help: "Summary of Mesh Cache",
	}, []string{"operation", "result"})
	if err := metrics.Register(metric); err != nil {
		return nil, err
	}
	return &Cache{
		rm:      rm,
		types:   types,
		ipFunc:  ipFunc,
		onceMap: once.NewMap(),
		cache:   cache.New(expirationTime, time.Duration(int64(float64(expirationTime)*0.9))),
		metrics: metric,
	}, nil
}

func (c *Cache) GetHash(ctx context.Context, mesh string) (string, error) {
	hash, found := c.cache.Get(mesh)
	if found {
		c.metrics.WithLabelValues("get", "hit").Inc()
		return hash.(string), nil
	}
	o := c.onceMap.Get(mesh)
	c.metrics.WithLabelValues("get", "hit-wait").Inc()
	o.Do(func() (interface{}, error) {
		c.metrics.WithLabelValues("get", "hit-wait").Dec()
		c.metrics.WithLabelValues("get", "miss").Inc()
		snapshot, err := GetMeshSnapshot(ctx, mesh, c.rm, c.types, c.ipFunc)
		if err != nil {
			// Don't cache errors
			c.onceMap.Delete(mesh)
			return nil, err
		}
		hash = snapshot.hash()
		c.cache.SetDefault(mesh, hash)
		c.onceMap.Delete(mesh)
		return hash, nil
	})
	if o.Err != nil {
		return "", o.Err
	}
	return o.Value.(string), nil
}
