package callbacks

import (
	"context"
	"sync"
	"time"

	util_xds "github.com/kumahq/kuma/pkg/util/xds"
)

type DeliveryTrackerCallbacks struct {
	util_xds.NoopCallbacks
	waitTimeout time.Duration

	response map[string]chan struct{}
	sync.Mutex
}

func NewDeliveryTrackerCallbacks(waitTimeout time.Duration) *DeliveryTrackerCallbacks {
	return &DeliveryTrackerCallbacks{
		response:    map[string]chan struct{}{},
		waitTimeout: waitTimeout,
	}
}

func (d *DeliveryTrackerCallbacks) WaitForResponse(ctx context.Context, version string) error {
	d.Lock()
	ch := make(chan struct{}, 1)
	d.response[version] = ch
	d.Unlock()
	ctx, cancelFn := context.WithTimeout(ctx, d.waitTimeout)
	defer func() {
		cancelFn()
		d.Lock()
		delete(d.response, version)
		d.Unlock()
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-ch:
		return nil
	}
}

func (d *DeliveryTrackerCallbacks) OnStreamResponse(i int64, request util_xds.DiscoveryRequest, response util_xds.DiscoveryResponse) {
	d.Lock()
	ch, ok := d.response[response.VersionInfo()]
	d.Unlock()
	if !ok {
		return
	}
	ch <- struct{}{}
}

var _ util_xds.Callbacks = &DeliveryTrackerCallbacks{}
