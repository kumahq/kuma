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
	NoopDataplaneCallbacks
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

func (c *ConnectionInfoTracker) OnStreamConnected(_ core_xds.StreamID, dpKey core_model.ResourceKey, ctx context.Context, _ core_xds.DataplaneMetadata) error {
	// We use OnStreamConnected, not OnFirstStreamConnected because if there are many xDS streams, we want to follow ConnectionInfo from the newest stream.
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

func (c *ConnectionInfoTracker) OnLastStreamDisconnected(_ core_xds.StreamID, dpKey core_model.ResourceKey) {
	c.Lock()
	delete(c.connectionInfos, dpKey)
	c.Unlock()
}
