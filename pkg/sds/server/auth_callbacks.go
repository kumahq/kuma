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

func newAuthCallbacks(authenticator sds_auth.Authenticator) *authCallbacks {
	return &authCallbacks{
		authenticator: authenticator,
		mutex:         sync.RWMutex{},
		contexts:      map[int64]context.Context{},
	}
}

// AuthCallback checks if the DiscoveryRequest is authorized, ie. if it has a valid Dataplane Token/Service Account Token.
type authCallbacks struct {
	authenticator sds_auth.Authenticator

	mutex    sync.RWMutex
	contexts map[int64]context.Context
}

func (a *authCallbacks) OnStreamOpen(ctx context.Context, streamID int64, s string) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	a.contexts[streamID] = ctx
	return nil
}

func (a *authCallbacks) OnStreamClosed(streamID int64) {
}

func (a *authCallbacks) OnStreamRequest(streamID int64, req *envoy_api.DiscoveryRequest) error {
	credential, err := a.credential(streamID)
	if err != nil {
		return err
	}
	return a.authenticate(credential, req)
}

func (a *authCallbacks) credential(streamID int64) (sds_auth.Credential, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

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

func (a *authCallbacks) OnStreamResponse(i int64, request *envoy_api.DiscoveryRequest, response *envoy_api.DiscoveryResponse) {
}

func (a *authCallbacks) OnFetchRequest(ctx context.Context, request *envoy_api.DiscoveryRequest) error {
	credential, err := sds_auth.ExtractCredential(ctx)
	if err != nil {
		return err
	}
	return a.authenticate(credential, request)
}

func (a *authCallbacks) OnFetchResponse(request *envoy_api.DiscoveryRequest, response *envoy_api.DiscoveryResponse) {
}

var _ envoy_server.Callbacks = &authCallbacks{}
