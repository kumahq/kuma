package callbacks

import (
	"context"
	"sync"

	"github.com/pkg/errors"
	"google.golang.org/grpc/metadata"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
)

const authorityHeader = ":authority"

// ConnectionInfoTracker tracks the information about the connection itself from the data plane to the control plane
type ConnectionInfoTracker struct {
	sync.RWMutex
	connectionInfos map[core_model.ResourceKey]*xds_context.ConnectionInfo
}

var _ DataplaneCallbacks = &ConnectionInfoTracker{}

func NewConnectionInfoTracker() *ConnectionInfoTracker {
	return &ConnectionInfoTracker{
		connectionInfos: map[core_model.ResourceKey]*xds_context.ConnectionInfo{},
	}
}

func (c *ConnectionInfoTracker) ConnectionInfo(dpKey core_model.ResourceKey) *xds_context.ConnectionInfo {
	c.RLock()
	defer c.RUnlock()
	return c.connectionInfos[dpKey]
}

func (c *ConnectionInfoTracker) OnProxyReconnected(_ core_xds.StreamID, dpKey core_model.ResourceKey, ctx context.Context, _ core_xds.DataplaneMetadata) error {
	return c.processConnectionInfo(dpKey, ctx)
}

func (c *ConnectionInfoTracker) OnProxyConnected(_ core_xds.StreamID, dpKey core_model.ResourceKey, ctx context.Context, _ core_xds.DataplaneMetadata) error {
	return c.processConnectionInfo(dpKey, ctx)
}

func (c *ConnectionInfoTracker) processConnectionInfo(dpKey core_model.ResourceKey, ctx context.Context) error {
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
	c.connectionInfos[dpKey] = &connInfo
	c.Unlock()
	return nil
}

func (c *ConnectionInfoTracker) OnProxyDisconnected(_ core_xds.StreamID, dpKey core_model.ResourceKey) {
	c.Lock()
	delete(c.connectionInfos, dpKey)
	c.Unlock()
}
