package callbacks

import (
	"context"
	"sync"

	"github.com/pkg/errors"
	"google.golang.org/grpc/metadata"

	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	util_xds "github.com/kumahq/kuma/pkg/util/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
)

const authorityHeader = ":authority"

// ConnectionInfoTracker tracks the information about the connection itself from the data plane to the control plane
type ConnectionInfoTracker struct {
	util_xds.NoopCallbacks
	sync.RWMutex
	connectionInfos map[core_xds.StreamID]xds_context.ConnectionInfo
}

var _ util_xds.Callbacks = &ConnectionInfoTracker{}

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
