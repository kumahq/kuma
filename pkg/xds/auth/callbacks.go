package auth

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sethvargo/go-retry"
	"google.golang.org/grpc/metadata"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	util_xds "github.com/kumahq/kuma/pkg/util/xds"
)

// The same logic also resides in pkg/hds/authn/callbacks.go

const authorization = "authorization"

// DPNotFoundRetry if callbacks are used in xDS other than ADS.
// It might be useful to have a retry when dataplane is not found, because on Universal ADS is creating it.
type DPNotFoundRetry struct {
	Backoff  time.Duration
	MaxTimes uint
}

func NewCallbacks(resManager core_manager.ReadOnlyResourceManager, authenticator Authenticator, dpNotFoundRetry DPNotFoundRetry) util_xds.Callbacks {
	if dpNotFoundRetry.Backoff == 0 { // backoff cannot be 0
		dpNotFoundRetry.Backoff = 1 * time.Millisecond
	}
	return &authCallbacks{
		resManager:      resManager,
		authenticator:   authenticator,
		contexts:        map[core_xds.StreamID]context.Context{},
		authenticated:   map[core_xds.StreamID]string{},
		dpNotFoundRetry: dpNotFoundRetry,
	}
}

// authCallback checks if the DiscoveryRequest is authorized, ie. if it has a valid Dataplane Token/Service Account Token.
type authCallbacks struct {
	util_xds.NoopCallbacks
	resManager      core_manager.ReadOnlyResourceManager
	authenticator   Authenticator
	dpNotFoundRetry DPNotFoundRetry

	sync.RWMutex // protects contexts and authenticated
	// contexts stores context for every stream, since Context from which we can extract auth data is only available in OnStreamOpen
	contexts map[core_xds.StreamID]context.Context
	// authenticated stores authenticated ProxyID for stream. We don't want to authenticate every because since on K8S we execute ReviewToken which is expensive
	// as long as client won't change ProxyID it's safe to authenticate only once.
	authenticated map[core_xds.StreamID]string
}

var _ util_xds.Callbacks = &authCallbacks{}

func (a *authCallbacks) OnStreamOpen(ctx context.Context, streamID core_xds.StreamID, _ string) error {
	a.Lock()
	defer a.Unlock()

	a.contexts[streamID] = ctx
	return nil
}

func (a *authCallbacks) OnStreamClosed(streamID core_xds.StreamID) {
	a.Lock()
	delete(a.contexts, streamID)
	delete(a.authenticated, streamID)
	a.Unlock()
}

func (a *authCallbacks) OnStreamRequest(streamID core_xds.StreamID, req util_xds.DiscoveryRequest) error {
	if id, alreadyAuthenticated := a.authNodeId(streamID); alreadyAuthenticated {
		if req.NodeId() != "" && req.NodeId() != id {
			return errors.Errorf("stream was authenticated for ID %s. Received request is for node with ID %s. Node ID cannot be changed after stream is initialized", id, req.NodeId())
		}
		return nil
	}

	credential, err := a.credential(streamID)
	if err != nil {
		return err
	}
	err = a.authenticate(credential, req)
	if err != nil {
		return err
	}
	a.Lock()
	a.authenticated[streamID] = req.NodeId()
	a.Unlock()
	return nil
}

func (a *authCallbacks) authNodeId(streamID core_xds.StreamID) (string, bool) {
	a.RLock()
	defer a.RUnlock()
	id, ok := a.authenticated[streamID]
	return id, ok
}

func (a *authCallbacks) credential(streamID core_xds.StreamID) (Credential, error) {
	a.RLock()
	defer a.RUnlock()

	ctx, exists := a.contexts[streamID]
	if !exists {
		return "", errors.Errorf("there is no context for stream ID %d", streamID)
	}
	credential, err := ExtractCredential(ctx)
	if err != nil {
		return "", errors.Wrap(err, "could not extract credential from DiscoveryRequest")
	}
	return credential, err
}

func (a *authCallbacks) authenticate(credential Credential, req util_xds.DiscoveryRequest) error {
	md := core_xds.DataplaneMetadataFromXdsMetadata(req.Metadata())

	// If we already have a resource from the xDS bootstrap, we can use that.
	resource := md.Resource

	// Otherwise, search for the pre-created resource.
	if resource == nil {
		proxyId, err := core_xds.ParseProxyIdFromString(req.NodeId())
		if err != nil {
			return errors.Wrap(err, "request must have a valid Proxy ID")
		}

		switch md.GetProxyType() {
		case mesh_proto.IngressProxyType:
			resource = core_mesh.NewZoneIngressResource()
		case mesh_proto.EgressProxyType:
			resource = core_mesh.NewZoneEgressResource()
		case mesh_proto.DataplaneProxyType:
			resource = core_mesh.NewDataplaneResource()
		default:
			return errors.Errorf("unsupported proxy type %q", md.GetProxyType())
		}

		backoff := retry.WithMaxRetries(uint64(a.dpNotFoundRetry.MaxTimes), retry.NewConstant(a.dpNotFoundRetry.Backoff))
		err = retry.Do(context.Background(), backoff, func(ctx context.Context) error {
			err := a.resManager.Get(ctx, resource, core_store.GetBy(proxyId.ToResourceKey()))
			if core_store.IsResourceNotFound(err) {
				return retry.RetryableError(errors.Errorf(
					"resource %q not found; create a %s in Kuma CP first or pass it as an argument to kuma-dp",
					proxyId, resource.Descriptor().Name))
			}
			return err
		})
		if err != nil {
			return err
		}
	}

	return errors.Wrap(a.authenticator.Authenticate(context.Background(), resource, credential),
		"authentication failed")
}

func ExtractCredential(ctx context.Context) (Credential, error) {
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
