package server

import (
	"context"
	"sync"

	envoy_api_v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_server "github.com/envoyproxy/go-control-plane/pkg/server/v2"
	"github.com/pkg/errors"
	"google.golang.org/grpc/metadata"

	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
)

const authorityHeader = ":authority"

// ConnectionInfoTracker tracks the information about the connection itself from the data plane to the control plane
type ConnectionInfoTracker struct {
	sync.RWMutex
	connectionInfos map[core_xds.StreamID]xds_context.ConnectionInfo
}

func NewConnectionInfoTracker() *ConnectionInfoTracker {
	return &ConnectionInfoTracker{
		connectionInfos: map[core_xds.StreamID]xds_context.ConnectionInfo{},
	}
}

func (c *ConnectionInfoTracker) ConnectionInfo(streamID core_xds.StreamID) xds_context.ConnectionInfo {
	c.RLock()
	defer c.RUnlock()
	return c.connectionInfos[streamID]
}

func (c *ConnectionInfoTracker) OnStreamOpen(ctx context.Context, streamID core_xds.StreamID, _ string) error {
	metadata, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return errors.New("request has no metadata")
	}
	values := metadata.Get(authorityHeader)
	if len(values) != 1 {
		return errors.Errorf("request has no %s header", authorityHeader)
	}
	c.Lock()
	connInfo := xds_context.ConnectionInfo{
		Authority: values[0],
	}
	c.connectionInfos[streamID] = connInfo
	c.Unlock()
	return nil
}

func (c *ConnectionInfoTracker) OnStreamClosed(streamID core_xds.StreamID) {
	c.Lock()
	delete(c.connectionInfos, streamID)
	c.Unlock()
}

func (c *ConnectionInfoTracker) OnStreamRequest(core_xds.StreamID, *envoy_api_v2.DiscoveryRequest) error {
	return nil
}

func (c *ConnectionInfoTracker) OnStreamResponse(core_xds.StreamID, *envoy_api_v2.DiscoveryRequest, *envoy_api_v2.DiscoveryResponse) {
}

func (c *ConnectionInfoTracker) OnFetchRequest(context.Context, *envoy_api_v2.DiscoveryRequest) error {
	return nil
}

func (c *ConnectionInfoTracker) OnFetchResponse(*envoy_api_v2.DiscoveryRequest, *envoy_api_v2.DiscoveryResponse) {
}

var _ envoy_server.Callbacks = &ConnectionInfoTracker{}
