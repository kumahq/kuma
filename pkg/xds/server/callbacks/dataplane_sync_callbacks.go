package callbacks

import (
	"context"
	"fmt"
	"slices"
	stdsync "sync"
	"sync/atomic"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	util_xds "github.com/kumahq/kuma/pkg/util/xds"
	"github.com/kumahq/kuma/pkg/xds/sync"
)

var dataplaneSyncLog = core.Log.WithName("xds").WithName("dataplane-sync")

// NewDataplaneSyncCallbacks creates a callbacks for XDS that will follow the lifecycle of a Dataplane.
// It will ensure that one and only one watchdog and lifecycle manager is created for a given Dataplane.
func NewDataplaneSyncCallbacks(
	dpWatchdogFactory sync.DataplaneWatchdogFactory,
	lifecycleManagerFactory DataplaneLifecycleManagerFactory,
) util_xds.Callbacks {
	return &dataplaneSyncCallbacks{
		dpWatchdogFactory:         dpWatchdogFactory,
		dpLifecycleManagerFactory: lifecycleManagerFactory,
		proxyInfos:                map[core_model.ResourceKey]*dpInfo{},
		activeStreams:             map[core_xds.StreamID]*streamInfo{},
	}
}

// DataplaneLifecycleManagerFactory is a factory for DataplaneLifecycleManager
type DataplaneLifecycleManagerFactory interface {
	New(key core_model.ResourceKey) DataplaneLifecycleManager
}

type DataplaneLifecycleManagerFunc func(key core_model.ResourceKey) DataplaneLifecycleManager

func (f DataplaneLifecycleManagerFunc) New(key core_model.ResourceKey) DataplaneLifecycleManager {
	return f(key)
}

// DataplaneLifecycleManager is responsible for managing the lifecycle of a Dataplane.
type DataplaneLifecycleManager interface {
	Register(logr.Logger, context.Context, *core_xds.DataplaneMetadata) error
	Deregister(logr.Logger, context.Context)
}

// dpInfo holds all the state for a specific dataplane
type dpInfo struct {
	stdsync.RWMutex
	dpKey core_model.ResourceKey
	// ctx is the context of the callbacks for a specific DP.
	ctx context.Context
	// cancels the ctx of any subtasks for the dataplane
	cancelFunc context.CancelFunc
	// done is closed when all subtasks are completed
	done chan struct{}
	// meta is the metadata of the dataplane
	meta             atomic.Pointer[core_xds.DataplaneMetadata]
	lifecycleManager DataplaneLifecycleManager
	// streams is a list of streamIDs that are associated with this dataplane they are ordered in the order they were opened
	streams []core_xds.StreamID
}

// dataplaneSyncCallbacks tracks XDS streams that are connected to the CP and fire up a watchdog.
// Watchdog should be run only once for given DP regardless of the number of streams.
// For ADS there is only one stream for DP.
//
// We keep
type dataplaneSyncCallbacks struct {
	stdsync.RWMutex

	dpWatchdogFactory         sync.DataplaneWatchdogFactory
	dpLifecycleManagerFactory DataplaneLifecycleManagerFactory

	activeStreams map[core_xds.StreamID]*streamInfo
	proxyInfos    map[core_model.ResourceKey]*dpInfo
}

type streamInfo struct {
	ctx       context.Context
	proxyInfo *dpInfo
}

func (t *dataplaneSyncCallbacks) OnStreamResponse(i int64, request util_xds.DiscoveryRequest, response util_xds.DiscoveryResponse) {
	// Noop
}

func (t *dataplaneSyncCallbacks) OnStreamOpen(ctx context.Context, streamID core_xds.StreamID, _ string) error {
	t.Lock()
	defer t.Unlock()
	dataplaneSyncLog.V(1).Info("stream is open", "streamID", streamID)
	if _, found := t.activeStreams[streamID]; found {
		return errors.Errorf("streamID %d is already tracked, we should never reopen a stream", streamID)
	}
	t.activeStreams[streamID] = &streamInfo{ctx: ctx, proxyInfo: nil}
	return nil
}

func (t *dataplaneSyncCallbacks) OnStreamRequest(streamID core_xds.StreamID, request util_xds.DiscoveryRequest) error {
	if request.NodeId() == "" {
		// from https://www.envoyproxy.io/docs/envoy/latest/api-docs/xds_protocol#ack-nack-and-versioning:
		// Only the first request on a stream is guaranteed to carry the node identifier.
		// The subsequent discovery requests on the same stream may carry an empty node identifier.
		// This holds true regardless of the acceptance of the discovery responses on the same stream.
		// The node identifier should always be identical if present more than once on the stream.
		// It is sufficient to only check the first message for the node identifier as a result.
		return nil
	}

	t.RLock()
	dataplaneSyncLog.V(1).Info("stream request", "streamID", streamID)
	alreadyProcessed := t.activeStreams[streamID].proxyInfo != nil
	t.RUnlock()
	if alreadyProcessed { // We fast return if we already know that this streamID is tracking a specific dataplane
		return nil
	}

	proxyId, err := core_xds.ParseProxyIdFromString(request.NodeId())
	if err != nil {
		return errors.Wrap(err, "invalid node ID")
	}
	dpKey := proxyId.ToResourceKey()
	metadata := core_xds.DataplaneMetadataFromXdsMetadata(request.Metadata())
	if metadata == nil {
		return errors.New("metadata in xDS Node cannot be nil")
	}
	if metadata.Resource != nil {
		if core_model.MetaToResourceKey(metadata.Resource.GetMeta()) != dpKey {
			return fmt.Errorf("proxyId %s does not match proxy resource %s", dpKey, metadata.Resource.GetMeta())
		}
	}
	l := dataplaneSyncLog.WithValues("dpKey", dpKey, "streamID", streamID)
	t.Lock()
	if t.activeStreams[streamID].proxyInfo != nil {
		return nil // We fast return if we already know that this streamID is tracking a specific dataplane
	}
	pInfo := t.proxyInfos[dpKey]
	if pInfo == nil {
		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan struct{})
		pInfo = &dpInfo{
			dpKey:            dpKey,
			lifecycleManager: t.dpLifecycleManagerFactory.New(dpKey),
			ctx:              ctx,
			done:             done,
			meta:             atomic.Pointer[core_xds.DataplaneMetadata]{},
			cancelFunc:       cancel,
		}
		t.proxyInfos[dpKey] = pInfo
	}
	t.activeStreams[streamID].proxyInfo = pInfo
	ctx := t.activeStreams[streamID].ctx
	pInfo.Lock()
	defer pInfo.Unlock()
	t.Unlock()
	pInfo.streams = append(pInfo.streams, streamID)
	pInfo.meta.Store(metadata)
	if len(pInfo.streams) == 1 {
		err := pInfo.lifecycleManager.Register(l, ctx, metadata)
		if err != nil {
			return err
		}

		l.V(1).Info("starting watchdog")
		go func() {
			defer close(pInfo.done)
			l.V(1).Info("watchdog started")
			if t.dpWatchdogFactory != nil {
				t.dpWatchdogFactory.New(dpKey, pInfo.meta.Load).Start(pInfo.ctx)
			}
		}()
	} else {
		l.V(1).Info("watchdog was started by a previous stream, not starting a new one", "activeStreams", pInfo.streams)
	}

	return nil
}

func (t *dataplaneSyncCallbacks) OnStreamClosed(streamID core_xds.StreamID) {
	t.Lock()
	l := dataplaneSyncLog.WithValues("streamID", streamID)
	l.V(1).Info("stream closed")
	sInfo := t.activeStreams[streamID]
	if sInfo == nil { // Should never happen as it means a stream was closed without being opened
		l.Error(errors.New("stream closed without being opened"), "stream closed without being opened")
		t.Unlock()
		return
	}
	if sInfo.proxyInfo == nil { // Never associated with a dataplane, no need to care
		delete(t.activeStreams, streamID)
		t.Unlock()
		return
	}
	pInfo := sInfo.proxyInfo
	l = dataplaneSyncLog.WithValues("dpKey", pInfo.dpKey)
	pInfo.Lock()
	t.Unlock()
	streamIdx := slices.Index(pInfo.streams, streamID)
	if streamIdx == -1 {
		l.Info("streamID not found in the list of streams", "streams", pInfo.streams)
		pInfo.Unlock()
		return
	}
	if streamIdx != len(pInfo.streams)-1 {
		l.V(1).Info("dpKey is associated with a newer stream", "streams", pInfo.streams)
		pInfo.streams = append(pInfo.streams[:streamIdx], pInfo.streams[streamIdx+1:]...) // Remove streamID from the list
		pInfo.Unlock()
		return // We are not the last stream, we don't care
	}
	if sInfo.proxyInfo.cancelFunc != nil {
		l.V(1).Info("stopping watchdog")
		sInfo.proxyInfo.cancelFunc()
	}
	sInfo.proxyInfo.lifecycleManager.Deregister(l, sInfo.ctx)
	if sInfo.proxyInfo.cancelFunc != nil {
		<-sInfo.proxyInfo.done
		l.V(1).Info("watchdog terminated")
	}
	pInfo.Unlock()
	t.Lock()
	delete(t.activeStreams, streamID)
	delete(t.proxyInfos, pInfo.dpKey)
	t.Unlock()
}
