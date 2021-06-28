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

func NewCallbacks(resManager core_manager.ResourceManager, authenticator Authenticator, dpNotFoundRetry DPNotFoundRetry) util_xds.Callbacks {
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
	resManager      core_manager.ResourceManager
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
	credential, err := extractCredential(ctx)
	if err != nil {
		return "", errors.Wrap(err, "could not extract credential from DiscoveryRequest")
	}
	return credential, err
}

func (a *authCallbacks) authenticate(credential Credential, req util_xds.DiscoveryRequest) error {
	md := core_xds.DataplaneMetadataFromXdsMetadata(req.Metadata())
	switch md.GetProxyType() {
	case mesh_proto.IngressProxyType:
		return a.authenticateZoneIngress(credential, req)
	default:
		return a.authenticateDataplane(credential, req)
	}
}

func (a *authCallbacks) authenticateZoneIngress(credential Credential, req util_xds.DiscoveryRequest) error {
	zoneIngress := core_mesh.NewZoneIngressResource()
	md := core_xds.DataplaneMetadataFromXdsMetadata(req.Metadata())
	if md.GetZoneIngressResource() != nil {
		zoneIngress = md.GetZoneIngressResource()
	} else {
		proxyId, err := core_xds.ParseProxyIdFromString(req.NodeId())
		if err != nil {
			return errors.Wrap(err, "request must have a valid Proxy Id")
		}
		backoff, _ := retry.NewConstant(a.dpNotFoundRetry.Backoff)
		backoff = retry.WithMaxRetries(uint64(a.dpNotFoundRetry.MaxTimes), backoff)
		err = retry.Do(context.Background(), backoff, func(ctx context.Context) error {
			err := a.resManager.Get(ctx, zoneIngress, core_store.GetBy(proxyId.ToResourceKey()))
			if core_store.IsResourceNotFound(err) {
				return retry.RetryableError(errors.New("zoneIngress not found. Create ZoneIngress in Kuma CP first or pass it as an argument to kuma-dp"))
			}
			return err
		})
		if err != nil {
			return err
		}
	}

	if err := a.authenticator.AuthenticateZoneIngress(context.Background(), zoneIngress, credential); err != nil {
		return errors.Wrap(err, "authentication failed")
	}
	return nil
}

func (a *authCallbacks) authenticateDataplane(credential Credential, req util_xds.DiscoveryRequest) error {
	dataplane := core_mesh.NewDataplaneResource()
	md := core_xds.DataplaneMetadataFromXdsMetadata(req.Metadata())
	if md.GetDataplaneResource() != nil {
		dataplane = md.GetDataplaneResource()
	} else {
		proxyId, err := core_xds.ParseProxyIdFromString(req.NodeId())
		if err != nil {
			return errors.Wrap(err, "request must have a valid Proxy Id")
		}
		backoff, _ := retry.NewConstant(a.dpNotFoundRetry.Backoff)
		backoff = retry.WithMaxRetries(uint64(a.dpNotFoundRetry.MaxTimes), backoff)
		err = retry.Do(context.Background(), backoff, func(ctx context.Context) error {
			err := a.resManager.Get(ctx, dataplane, core_store.GetBy(proxyId.ToResourceKey()))
			if core_store.IsResourceNotFound(err) {
				return retry.RetryableError(errors.New("dataplane not found. Create Dataplane in Kuma CP first or pass it as an argument to kuma-dp"))
			}
			return err
		})
		if err != nil {
			return err
		}
	}

	if err := a.authenticator.Authenticate(context.Background(), dataplane, credential); err != nil {
		return errors.Wrap(err, "authentication failed")
	}
	return nil
}

func extractCredential(ctx context.Context) (Credential, error) {
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
