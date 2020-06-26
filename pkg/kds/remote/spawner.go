package remote

import (
	"context"
	"sync/atomic"

	envoy_api_v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_server "github.com/envoyproxy/go-control-plane/pkg/server/v2"
	"github.com/go-logr/logr"

	"github.com/Kong/kuma/pkg/core/runtime/component"
)

type ComponentFactory func(logr.Logger, *envoy_api_v2.DiscoveryRequest) component.Component

func NewComponentSpawner(log logr.Logger, factory ComponentFactory) envoy_server.Callbacks {
	return &componentSpawner{
		log:     log,
		factory: factory,
		spawned: int32(0),
	}
}

type componentSpawner struct {
	log     logr.Logger
	factory ComponentFactory
	spawned int32
	stop    chan struct{}
}

func (c *componentSpawner) OnStreamRequest(streamID int64, req *envoy_api_v2.DiscoveryRequest) error {
	if atomic.CompareAndSwapInt32(&c.spawned, 0, 1) {
		// spawn component
		c.stop = make(chan struct{})
		comp := c.factory(c.log.WithValues("streamID", streamID), req)
		go func() {
			err := comp.Start(c.stop)
			c.log.Error(err, "component finished with an error")
			atomic.CompareAndSwapInt32(&c.spawned, 1, 0)
		}()
	}
	return nil
}

func (c *componentSpawner) OnStreamClosed(int64) {
	if atomic.CompareAndSwapInt32(&c.spawned, 1, 0) {
		close(c.stop)
	}
}

func (c *componentSpawner) OnStreamResponse(int64, *envoy_api_v2.DiscoveryRequest, *envoy_api_v2.DiscoveryResponse) {
}

func (c *componentSpawner) OnStreamOpen(context.Context, int64, string) error { return nil }

func (c *componentSpawner) OnFetchRequest(context.Context, *envoy_api_v2.DiscoveryRequest) error {
	return nil
}

func (c *componentSpawner) OnFetchResponse(*envoy_api_v2.DiscoveryRequest, *envoy_api_v2.DiscoveryResponse) {
}
