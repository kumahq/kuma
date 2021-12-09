package once

import (
	"context"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/kumahq/kuma/pkg/metrics"
)

type Cache struct {
	cache   *cache.Cache
	onceMap *omap
	metrics *prometheus.GaugeVec
}

// New creates a cache where items are evicted after being present for `expirationTime`.
// `name` is used to name the gauge used for metrics reporting.
func New(expirationTime time.Duration, name string, metrics metrics.Metrics) (*Cache, error) {
	metric := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: name,
		Help: "Summary of " + name,
	}, []string{"operation", "result"})
	if err := metrics.Register(metric); err != nil {
		return nil, err
	}
	return &Cache{
		cache:   cache.New(expirationTime, time.Duration(int64(float64(expirationTime)*0.9))),
		onceMap: newMap(),
		metrics: metric,
	}, nil
}

type Retriever interface {
	// Call method called when a cache miss happens which will return the actual value that needs to be cached
	Call(ctx context.Context, key string) (interface{}, error)
}
type RetrieverFunc func(context.Context, string) (interface{}, error)

func (f RetrieverFunc) Call(ctx context.Context, key string) (interface{}, error) {
	return f(ctx, key)
}

// GetOrRetrieve will return the cached value and if it isn't present will call `Retriever`.
// It is guaranteed there will only on one concurrent call to `Retriever` for each key, other accesses to the key will be blocked until `Retriever.Call` returns.
// If `Retriever.Call` fails the error will not be cached and subsequent calls will call the `Retriever` again.
func (c *Cache) GetOrRetrieve(ctx context.Context, key string, retriever Retriever) (interface{}, error) {
	v, found := c.cache.Get(key)
	if found {
		c.metrics.WithLabelValues("get", "hit").Inc()
		return v, nil
	}
	o, stored := c.onceMap.Get(key)
	if !stored {
		c.metrics.WithLabelValues("get", "hit-wait").Inc()
	}
	o.Do(func() (interface{}, error) {
		defer c.onceMap.Delete(key)
		val, found := c.cache.Get(key)
		if found {
			c.metrics.WithLabelValues("get", "hit").Inc()
			return val, nil
		}
		c.metrics.WithLabelValues("get", "miss").Inc()
		res, err := retriever.Call(ctx, key)
		if err != nil {
			c.metrics.WithLabelValues("get", "error").Inc()
			return nil, err
		}
		c.cache.SetDefault(key, res)
		return res, nil
	})
	if o.Err != nil {
		return "", o.Err
	}
	v = o.Value
	return v, nil
}
