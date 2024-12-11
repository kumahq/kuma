package server

import (
	"context"
	"errors"
	"sync"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_sd "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	envoy_xds "github.com/envoyproxy/go-control-plane/pkg/server/v3"

	"github.com/kumahq/kuma/pkg/kds/v2/util"
	"github.com/kumahq/kuma/pkg/multitenant"
	util_xds_v3 "github.com/kumahq/kuma/pkg/util/xds/v3"
)

type tenancyCallbacks struct {
	tenants multitenant.Tenants

	sync.RWMutex
	streamToCtx map[int64]context.Context
	util_xds_v3.NoopCallbacks
}

func NewTenancyCallbacks(tenants multitenant.Tenants) envoy_xds.Callbacks {
	return &tenancyCallbacks{
		tenants:     tenants,
		streamToCtx: map[int64]context.Context{},
	}
}

func (c *tenancyCallbacks) OnDeltaStreamOpen(ctx context.Context, streamID int64, _ string) error {
	c.Lock()
	c.streamToCtx[streamID] = ctx
	c.Unlock()
	return nil
}

func (c *tenancyCallbacks) OnStreamDeltaRequest(streamID int64, request *envoy_sd.DeltaDiscoveryRequest) error {
	c.RLock()
	defer c.RUnlock()
	ctx, ok := c.streamToCtx[streamID]
	if !ok {
		// it should not happen, but just in case it's better to fail
		return errors.New("context is missing")
	}
	tenantID, err := c.tenants.GetID(ctx)
	if err != nil {
		return err
	}
	util.FillTenantMetadata(tenantID, request.Node)
	return nil
}

func (c *tenancyCallbacks) OnDeltaStreamClosed(streamID int64, node *envoy_core.Node) {
	c.Lock()
	delete(c.streamToCtx, streamID)
	c.Unlock()
}
