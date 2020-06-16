package server

import (
	"context"
	"sync"

	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_server "github.com/envoyproxy/go-control-plane/pkg/server/v2"
	"github.com/pkg/errors"

	core_xds "github.com/Kong/kuma/pkg/core/xds"
	sds_auth "github.com/Kong/kuma/pkg/sds/auth"
)

func newAuthCallbacks(authenticator sds_auth.Authenticator) envoy_server.Callbacks {
	return &authCallbacks{
		authenticator: authenticator,
		contexts:      map[core_xds.StreamID]context.Context{},
	}
}

// authCallback checks if the DiscoveryRequest is authorized, ie. if it has a valid Dataplane Token/Service Account Token.
type authCallbacks struct {
	sync.RWMutex
	authenticator sds_auth.Authenticator
	contexts      map[core_xds.StreamID]context.Context
}

var _ envoy_server.Callbacks = &authCallbacks{}

func (a *authCallbacks) OnStreamOpen(ctx context.Context, streamID core_xds.StreamID, s string) error {
	a.Lock()
	defer a.Unlock()

	a.contexts[streamID] = ctx
	return nil
}

func (a *authCallbacks) OnStreamClosed(streamID core_xds.StreamID) {
}

func (a *authCallbacks) OnStreamRequest(streamID core_xds.StreamID, req *envoy_api.DiscoveryRequest) error {
	credential, err := a.credential(streamID)
	if err != nil {
		return err
	}
	return a.authenticate(credential, req)
}

func (a *authCallbacks) credential(streamID core_xds.StreamID) (sds_auth.Credential, error) {
	a.RLock()
	defer a.RUnlock()

	ctx, exists := a.contexts[streamID]
	if !exists {
		return "", errors.Errorf("there is no context for stream ID %d", streamID)
	}
	credential, err := sds_auth.ExtractCredential(ctx)
	if err != nil {
		return "", err
	}
	return credential, err
}

func (a *authCallbacks) authenticate(credential sds_auth.Credential, req *envoy_api.DiscoveryRequest) error {
	proxyId, err := core_xds.ParseProxyId(req.Node)
	if err != nil {
		return errors.Wrap(err, "SDS request must have a valid Proxy Id")
	}

	_, err = a.authenticator.Authenticate(context.Background(), *proxyId, credential)
	if err != nil {
		return err
	}
	return nil
}

func (a *authCallbacks) OnStreamResponse(core_xds.StreamID, *envoy_api.DiscoveryRequest, *envoy_api.DiscoveryResponse) {
}

func (a *authCallbacks) OnFetchRequest(ctx context.Context, request *envoy_api.DiscoveryRequest) error {
	credential, err := sds_auth.ExtractCredential(ctx)
	if err != nil {
		return err
	}
	return a.authenticate(credential, request)
}

func (a *authCallbacks) OnFetchResponse(*envoy_api.DiscoveryRequest, *envoy_api.DiscoveryResponse) {
}

var _ envoy_server.Callbacks = &authCallbacks{}
