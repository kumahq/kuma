package server

import (
	"context"
	"github.com/Kong/kuma/pkg/core/xds"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/envoyproxy/go-control-plane/pkg/server"
	"sync"
)

type dataplaneMetadataTracker struct {
	mutex            sync.RWMutex
	metadataForProxy map[xds.ProxyId]*xds.DataplaneMetadata
}

func newDataplaneMetadataTracker() *dataplaneMetadataTracker {
	return &dataplaneMetadataTracker{
		mutex:            sync.RWMutex{},
		metadataForProxy: map[xds.ProxyId]*xds.DataplaneMetadata{},
	}
}

func (d *dataplaneMetadataTracker) Metadata(proxyId xds.ProxyId) *xds.DataplaneMetadata {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	return d.metadataForProxy[proxyId]
}

var _ server.Callbacks = &dataplaneMetadataTracker{}

func (d *dataplaneMetadataTracker) OnStreamOpen(context.Context, int64, string) error {
	return nil
}

func (d *dataplaneMetadataTracker) OnStreamClosed(int64) {
}

func (d *dataplaneMetadataTracker) OnStreamRequest(stream int64, req *v2.DiscoveryRequest) error {
	proxyId, err := xds.ParseProxyId(req.Node)
	if err != nil {
		return err
	}

	d.mutex.Lock()
	d.metadataForProxy[*proxyId] = xds.DataplaneMetadataFromNode(req.Node)
	d.mutex.Unlock()
	return nil
}

func (d *dataplaneMetadataTracker) OnStreamResponse(int64, *v2.DiscoveryRequest, *v2.DiscoveryResponse) {
}

func (d *dataplaneMetadataTracker) OnFetchRequest(context.Context, *v2.DiscoveryRequest) error {
	return nil
}

func (d *dataplaneMetadataTracker) OnFetchResponse(*v2.DiscoveryRequest, *v2.DiscoveryResponse) {
}
