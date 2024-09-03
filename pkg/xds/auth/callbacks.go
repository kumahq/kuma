package auth

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sethvargo/go-retry"
	"google.golang.org/grpc/metadata"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/user"
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

func NewCallbacks(resManager core_manager.ReadOnlyResourceManager, authenticator Authenticator, dpNotFoundRetry DPNotFoundRetry) util_xds.MultiXDSCallbacks {
	if dpNotFoundRetry.Backoff == 0 { // backoff cannot be 0
		dpNotFoundRetry.Backoff = 1 * time.Millisecond
	}
	return &authCallbacks{
		resManager:      resManager,
		authenticator:   authenticator,
		streams:         map[core_xds.StreamID]stream{},
		deltaStreams:    map[core_xds.StreamID]stream{},
		dpNotFoundRetry: dpNotFoundRetry,
	}
}

// authCallback checks if the DiscoveryRequest is authorized, ie. if it has a valid Dataplane Token/Service Account Token.
type authCallbacks struct {
	util_xds.NoopCallbacks
	resManager      core_manager.ReadOnlyResourceManager
	authenticator   Authenticator
	dpNotFoundRetry DPNotFoundRetry

	sync.RWMutex // protects streams
	streams      map[core_xds.StreamID]stream
	deltaStreams map[core_xds.StreamID]stream
}

type stream struct {
	// context of the stream that contains a credential
	ctx context.Context
	// Dataplane / ZoneIngress / ZoneEgress associated with this XDS stream.
	resource model.Resource
	// nodeID of the stream. Has to be the same for the whole life of a stream.
	nodeID string
}

var _ util_xds.MultiXDSCallbacks = &authCallbacks{}

func (a *authCallbacks) OnStreamOpen(ctx context.Context, streamID core_xds.StreamID, _ string) error {
	a.Lock()
	defer a.Unlock()

	a.streams[streamID] = stream{
		ctx:      ctx,
		resource: nil,
	}
	return nil
}

func (a *authCallbacks) OnStreamClosed(streamID core_xds.StreamID) {
	a.Lock()
	delete(a.streams, streamID)
	a.Unlock()
}

func (a *authCallbacks) OnStreamRequest(streamID core_xds.StreamID, req util_xds.DiscoveryRequest) error {
	return a.onStreamRequest(streamID, req, false)
}

func (a *authCallbacks) OnDeltaStreamOpen(ctx context.Context, streamID core_xds.StreamID, _ string) error {
	a.Lock()
	defer a.Unlock()

	a.deltaStreams[streamID] = stream{
		ctx:      ctx,
		resource: nil,
	}

	core.Log.V(1).Info("OnDeltaStreamOpen", "streamID", streamID)
	return nil
}

func (a *authCallbacks) OnDeltaStreamClosed(streamID int64) {
	a.Lock()
	delete(a.deltaStreams, streamID)
	a.Unlock()
}

func (a *authCallbacks) OnStreamDeltaRequest(streamID core_xds.StreamID, req util_xds.DeltaDiscoveryRequest) error {
	return a.onStreamRequest(streamID, req, true)
}

func (a *authCallbacks) onStreamRequest(streamID core_xds.StreamID, req util_xds.Request, isDelta bool) error {
	s, err := a.stream(streamID, req, isDelta)
	if err != nil {
		return err
	}
	core.Log.V(1).Info("OnStreamDeltaRequest auth", "req", req)

	credential, err := ExtractCredential(s.ctx)
	if err != nil {
		return errors.Wrap(err, "could not extract credential from DiscoveryRequest")
	}
	if err := a.authenticator.Authenticate(user.Ctx(s.ctx, user.ControlPlane), s.resource, credential); err != nil {
		return errors.Wrap(err, "authentication failed")
	}
	a.Lock()
	if isDelta {
		a.deltaStreams[streamID] = s
	} else {
		a.streams[streamID] = s
	}
	a.Unlock()
	return nil
}

func (a *authCallbacks) stream(streamID core_xds.StreamID, req util_xds.Request, isDelta bool) (stream, error) {
	a.RLock()
	var s stream
	var ok bool
	if isDelta {
		s, ok = a.deltaStreams[streamID]
	} else {
		s, ok = a.streams[streamID]
	}
	a.RUnlock()
	if !ok {
		return stream{}, errors.New("stream is not present")
	}

	if s.nodeID == "" {
		s.nodeID = req.NodeId()
	}

	if s.nodeID != req.NodeId() {
		return stream{}, errors.Errorf("stream was authenticated for ID %s. Received request is for node with ID %s. Node ID cannot be changed after stream is initialized", s.nodeID, req.NodeId())
	}

	if s.resource == nil {
		md := core_xds.DataplaneMetadataFromXdsMetadata(req.Metadata())
		res, err := a.resource(user.Ctx(s.ctx, user.ControlPlane), md, req.NodeId())
		if err != nil {
			return stream{}, err
		}
		s.resource = res
	}
	return s, nil
}

func (a *authCallbacks) resource(ctx context.Context, md *core_xds.DataplaneMetadata, nodeID string) (model.Resource, error) {
	if md.Resource != nil {
		return md.Resource, nil
	}

	// Otherwise, search for the pre-created resource.
	proxyId, err := core_xds.ParseProxyIdFromString(nodeID)
	if err != nil {
		return nil, errors.Wrap(err, "request must have a valid Proxy ID")
	}

	var resource model.Resource
	switch md.GetProxyType() {
	case mesh_proto.IngressProxyType:
		resource = core_mesh.NewZoneIngressResource()
	case mesh_proto.EgressProxyType:
		resource = core_mesh.NewZoneEgressResource()
	case mesh_proto.DataplaneProxyType:
		resource = core_mesh.NewDataplaneResource()
	default:
		return nil, errors.Errorf("unsupported proxy type %q", md.GetProxyType())
	}

	backoff := retry.WithMaxRetries(uint64(a.dpNotFoundRetry.MaxTimes), retry.NewConstant(a.dpNotFoundRetry.Backoff))
	err = retry.Do(ctx, backoff, func(ctx context.Context) error {
		err := a.resManager.Get(ctx, resource, core_store.GetBy(proxyId.ToResourceKey()))
		if core_store.IsResourceNotFound(err) {
			return retry.RetryableError(errors.Errorf(
				"resource %q not found; create a %s in Kuma CP first or pass it as an argument to kuma-dp",
				proxyId, resource.Descriptor().Name))
		}
		return err
	})
	return resource, err
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
