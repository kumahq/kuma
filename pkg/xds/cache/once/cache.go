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

func (c *Cache) Get(ctx context.Context, key string, fn func(context.Context, string) (interface{}, error)) (interface{}, error) {
	v, found := c.cache.Get(key)
	if found {
		c.metrics.WithLabelValues("get", "hit").Inc()
		return v, nil
	}
	o := c.onceMap.Get(key)
	c.metrics.WithLabelValues("get", "hit-wait").Inc()
	o.Do(func() (interface{}, error) {
		defer c.onceMap.Delete(key)
		c.metrics.WithLabelValues("get", "hit-wait").Dec()
		c.metrics.WithLabelValues("get", "miss").Inc()
		res, err := fn(ctx, key)
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
	return o.Value, nil
}
