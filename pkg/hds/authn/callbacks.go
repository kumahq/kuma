package authn

import (
	"context"
	"sync"
	"time"

	envoy_service_health "github.com/envoyproxy/go-control-plane/envoy/service/health/v3"
	"github.com/pkg/errors"
	"github.com/sethvargo/go-retry"
	"google.golang.org/grpc/metadata"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	hds_callbacks "github.com/kumahq/kuma/pkg/hds/callbacks"
	xds_auth "github.com/kumahq/kuma/pkg/xds/auth"
)

const authorization = "authorization"

// Inspired by pkg/xds/auth/callbacks.go

type DPNotFoundRetry struct {
	Backoff  time.Duration
	MaxTimes uint
}

func NewCallbacks(resManager core_manager.ResourceManager, authenticator xds_auth.Authenticator, dpNotFoundRetry DPNotFoundRetry) hds_callbacks.Callbacks {
	if dpNotFoundRetry.Backoff == 0 { // backoff cannot be 0
		dpNotFoundRetry.Backoff = 1 * time.Millisecond
	}
	return &authn{
		resManager:      resManager,
		authenticator:   authenticator,
		contexts:        map[core_xds.StreamID]context.Context{},
		authenticated:   map[core_xds.StreamID]string{},
		dpNotFoundRetry: dpNotFoundRetry,
	}
}

type authn struct {
	resManager      core_manager.ResourceManager
	authenticator   xds_auth.Authenticator
	dpNotFoundRetry DPNotFoundRetry

	sync.RWMutex // protects contexts and authenticated
	// contexts stores context for every stream, since Context from which we can extract auth data is only available in OnStreamOpen
	contexts map[core_xds.StreamID]context.Context
	// authenticated stores authenticated ProxyID for stream. We don't want to authenticate every because since on K8S we execute ReviewToken which is expensive
	// as long as client won't change ProxyID it's safe to authenticate only once.
	authenticated map[core_xds.StreamID]string
}

var _ hds_callbacks.Callbacks = &authn{}

func (a *authn) OnStreamOpen(ctx context.Context, streamID int64) error {
	a.Lock()
	defer a.Unlock()

	a.contexts[streamID] = ctx
	return nil
}

func (a *authn) OnHealthCheckRequest(streamID int64, req *envoy_service_health.HealthCheckRequest) error {
	if id, alreadyAuthenticated := a.authNodeId(streamID); alreadyAuthenticated {
		if req.GetNode().GetId() != "" && req.GetNode().GetId() != id {
			return errors.Errorf("stream was authenticated for ID %s. Received request is for node with ID %s. Node ID cannot be changed after stream is initialized", id, req.GetNode().GetId())
		}
		return nil
	}

	credential, err := a.credential(streamID)
	if err != nil {
		return err
	}
	err = a.authenticate(credential, req.GetNode().GetId())
	if err != nil {
		return err
	}
	a.Lock()
	a.authenticated[streamID] = req.GetNode().GetId()
	a.Unlock()
	return nil
}

func (a *authn) OnEndpointHealthResponse(streamID int64, _ *envoy_service_health.EndpointHealthResponse) error {
	if id, alreadyAuthenticated := a.authNodeId(streamID); !alreadyAuthenticated {
		return errors.Errorf("stream was not authenticated for ID %s", id)
	}
	return nil
}

func (a *authn) OnStreamClosed(streamID int64) {
	a.Lock()
	delete(a.contexts, streamID)
	delete(a.authenticated, streamID)
	a.Unlock()
}

func (a *authn) authNodeId(streamID core_xds.StreamID) (string, bool) {
	a.RLock()
	defer a.RUnlock()
	id, ok := a.authenticated[streamID]
	return id, ok
}

func (a *authn) credential(streamID core_xds.StreamID) (xds_auth.Credential, error) {
	a.RLock()
	defer a.RUnlock()

	ctx, exists := a.contexts[streamID]
	if !exists {
		return "", errors.Errorf("there is no context for stream ID %d", streamID)
	}
	credential, err := extractCredential(ctx)
	if err != nil {
		return "", errors.Wrap(err, "could not extract credential from DiscoveryRequest")
	}
	return credential, err
}

func extractCredential(ctx context.Context) (xds_auth.Credential, error) {
	metadata, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.Errorf("request has no metadata")
	}
	if values, ok := metadata[authorization]; ok {
		if len(values) != 1 {
			return "", errors.Errorf("request must have exactly 1 %q header, got %d", authorization, len(values))
		}
		return values[0], nil
	}
	return "", nil
}

func (a *authn) authenticate(credential xds_auth.Credential, nodeID string) error {
	dataplane := core_mesh.NewDataplaneResource()

	proxyId, err := core_xds.ParseProxyIdFromString(nodeID)
	if err != nil {
		return errors.Wrap(err, "HDS request must have a valid Proxy Id")
	}
	// Retry on DP not found because HDS is initiated in the parallel with XDS.
	// It is very likely that Dataplane is not yet created.
	// We could just close the stream with an error and Envoy would retry, but to have better UX (not printing confusing logs) it's better to retry
	err = retry.Do(context.Background(), retry.WithMaxRetries(uint64(a.dpNotFoundRetry.MaxTimes), retry.NewConstant(a.dpNotFoundRetry.Backoff)), func(ctx context.Context) error {
		err := a.resManager.Get(ctx, dataplane, core_store.GetBy(proxyId.ToResourceKey()))
		if core_store.IsResourceNotFound(err) {
			return retry.RetryableError(errors.New("dataplane not found. Create Dataplane in Kuma CP first or pass it as an argument to kuma-dp"))
		}
		return err
	})
	if err != nil {
		return err
	}

	if err := a.authenticator.Authenticate(context.Background(), dataplane, credential); err != nil {
		return errors.Wrap(err, "authentication failed")
	}
	return nil
}
