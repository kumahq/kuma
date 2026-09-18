package callbacks

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core"
	"github.com/kumahq/kuma/v3/pkg/core/kri"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	meshtrust_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshtrust/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/store"
	"github.com/kumahq/kuma/v3/pkg/events"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
	otelstatus "github.com/kumahq/kuma/v3/pkg/xds/otel/status"
)

var sinkLog = core.Log.WithName("xds").WithName("sink")

// MTLSInfo describes the identity certificate currently served to a proxy.
// It is the source of DataplaneInsight.mTLS.
type MTLSInfo struct {
	Expiration time.Time
	Generation time.Time

	IssuedBackend     string
	SupportedBackends []string

	// ManagedExternally is set when the certificate is delivered by an external SDS server,
	// so the control plane doesn't know its lifetime.
	ManagedExternally bool
}

type DataplaneInsightSink interface {
	Start(stop <-chan struct{}) // TODO error
}

type DataplaneInsightStore interface {
	// Upsert creates or updates the subscription, storing it with
	// the key dataplaneID. dataplaneType gives the resource type of
	// the dataplane proxy that has subscribed.
	Upsert(
		ctx context.Context,
		xdsMetadata *structpb.Struct,
		dataplaneType core_model.ResourceType,
		dataplaneID core_model.ResourceKey,
		subscription *mesh_proto.DiscoverySubscription,
		mtlsInfo *MTLSInfo,
		otel *mesh_proto.DataplaneInsight_OpenTelemetry,
	) error
}

func NewDataplaneInsightSink(
	xdsMetadata *structpb.Struct,
	accessor SubscriptionStatusAccessor,
	otelStatusCache *otelstatus.Cache,
	newTicker func() *time.Ticker,
	generationTicker func() *time.Ticker,
	flushBackoff time.Duration,
	store DataplaneInsightStore,
	eventFactory events.ListenerFactory,
	roResManager manager.ReadOnlyResourceManager,
) DataplaneInsightSink {
	// Dataplane is the only proxy type that can connect to the control plane,
	// so the insight always belongs to a Dataplane.
	dpType := core_mesh.DataplaneType

	return &dataplaneInsightSink{
		flushTicker:      newTicker,
		generationTicker: generationTicker,
		dataplaneType:    dpType,
		accessor:         accessor,
		otelStatusCache:  otelStatusCache,
		flushBackoff:     flushBackoff,
		store:            store,
		xdsMetadata:      xdsMetadata,
		eventFactory:     eventFactory,
		roResManager:     roResManager,
	}
}

var _ DataplaneInsightSink = &dataplaneInsightSink{}

// send events
type dataplaneInsightSink struct {
	flushTicker      func() *time.Ticker
	generationTicker func() *time.Ticker
	dataplaneType    core_model.ResourceType
	accessor         SubscriptionStatusAccessor
	otelStatusCache  *otelstatus.Cache
	store            DataplaneInsightStore
	flushBackoff     time.Duration
	xdsMetadata      *structpb.Struct
	eventFactory     events.ListenerFactory
	roResManager     manager.ReadOnlyResourceManager
}

func (s *dataplaneInsightSink) Start(stop <-chan struct{}) {
	flushTicker := s.flushTicker()
	defer flushTicker.Stop()

	generationTicker := s.generationTicker()
	defer generationTicker.Stop()

	var lastStoredState *mesh_proto.DiscoverySubscription
	var lastStoredMTLSInfo *MTLSInfo
	var lastStoredOtel *mesh_proto.DataplaneInsight_OpenTelemetry
	var lastEvent *events.WorkloadIdentityChangedEvent
	var generation uint32

	dataplaneID, _ := s.accessor.GetStatus()
	listener := s.eventFactory.Subscribe(func(event events.Event) bool {
		if e, ok := event.(events.WorkloadIdentityChangedEvent); ok && e.ResourceKey == dataplaneID {
			return true
		}
		return false
	})
	defer listener.Close()

	flush := func(closing bool, event *events.WorkloadIdentityChangedEvent) {
		ctx := context.TODO()
		dataplaneID, currentState := s.accessor.GetStatus()
		var mtlsInfo *MTLSInfo
		switch {
		case event != nil && event.Operation == events.Create:
			mtlsInfo = &MTLSInfo{
				IssuedBackend: event.Origin.String(),
			}
			if event.ExpirationTime != nil && event.GenerationTime != nil {
				mtlsInfo.Expiration = pointer.Deref(event.ExpirationTime)
				mtlsInfo.Generation = pointer.Deref(event.GenerationTime)
				mtlsInfo.SupportedBackends = s.listMeshTrustBackends(ctx, dataplaneID.Mesh)
			} else {
				mtlsInfo.ManagedExternally = true
			}
			lastEvent = event
		default:
			// Either the proxy has no workload identity or it was just removed,
			// in both cases the insight must not advertise a certificate.
			mtlsInfo = nil
			// we don't need to store event once delete
			lastEvent = nil
		}
		select {
		case <-generationTicker.C:
			generation++
		default:
		}
		currentState.Generation = generation

		otel := s.otelStatusCache.Get(dataplaneID)

		otelChanged := !proto.Equal(otel, lastStoredOtel)
		if otelChanged {
			sinkLog.V(1).Info("OTel status changed", "dataplaneID", dataplaneID)
		}

		if proto.Equal(currentState, lastStoredState) && mtlsInfo == lastStoredMTLSInfo && !otelChanged {
			// We compare mtlsInfo and lastStoredMTLSInfo as pointers. It makes sense to short-circuit if flush() runs
			// on tick without any workload identity, so both are nil.
			return
		}

		if err := s.store.Upsert(ctx, s.xdsMetadata, s.dataplaneType, dataplaneID, currentState, mtlsInfo, otel); err != nil {
			switch {
			case closing:
				// When XDS stream is closed, Dataplane Status Tracker executes OnStreamClose which closes stop channel
				// The problem is that close() does not wait for this sink to do it's final work
				// In the meantime Dataplane Lifecycle executes OnStreamClose which can remove Dataplane entity (and Insights due to ownership). Therefore both scenarios can happen:
				// 1) upsert fail because it successfully retrieved DataplaneInsight but cannot Update because by this time, Insight is gone (ResourceConflict error)
				// 2) upsert fail because it tries to create a new insight, but there is no Dataplane so ownership returns an error
				// We could build a synchronous mechanism that waits for Sink to be stopped before moving on to next Callbacks, but this is potentially dangerous
				// that we could block waiting for storage instead of executing next callbacks.
				sinkLog.V(1).Info("failed to flush Dataplane status on stream close. It can happen when Dataplane is deleted at the same time",
					"dataplaneID", dataplaneID,
					"err", err)
			case store.IsAlreadyExists(err) || store.IsConflict(err):
				sinkLog.V(1).Info("failed to flush DataplaneInsight because it was updated in other place. Will retry in the next tick",
					"dataplaneID", dataplaneID)
			case store.IsNamespaceTerminating(err):
				sinkLog.V(1).Info("failed to flush DataplaneInsight because its namespace is terminating",
					"dataplaneID", dataplaneID)
			default:
				sinkLog.Error(err, "failed to flush DataplaneInsight", "dataplaneID", dataplaneID)
			}
		} else {
			sinkLog.V(1).Info("DataplaneInsight saved", "dataplaneID", dataplaneID, "subscription", currentState)
			lastStoredState = currentState
			lastStoredMTLSInfo = mtlsInfo
			lastStoredOtel = otel
		}
	}

	// flush the first insight as quickly as possible so
	// 1) user sees that DP is online in kumactl/GUI (even without any XDS updates)
	// 2) we can have lower deregistrationDelay, see pkg/xds/server/callbacks/dataplane_lifecycle.go#deregisterProxy
	flush(false, nil)

	for {
		select {
		case <-flushTicker.C:
			flush(false, lastEvent)
			// On Kubernetes, because of the cache subsequent Get, Update requests can fail, because the cache is not strongly consistent.
			// We handle the Resource Conflict logging on V1, but we can try to avoid the situation with backoff
			time.Sleep(s.flushBackoff)
		case e := <-listener.Recv():
			workloadIdentity := e.(events.WorkloadIdentityChangedEvent)
			flush(false, &workloadIdentity)
		case <-stop:
			flush(true, lastEvent)
			return
		}
	}
}

// We use a ReadOnlyResourceManager, which is backed by a cache,
// so performance is not a concern as data is always retrieved from the cache.
func (s *dataplaneInsightSink) listMeshTrustBackends(ctx context.Context, mesh string) []string {
	meshTrusts := &meshtrust_api.MeshTrustResourceList{}
	if err := s.roResManager.List(ctx, meshTrusts, store.ListByMesh(mesh)); err != nil {
		sinkLog.Error(err, "cannot list MeshTrusts")
		return nil
	}

	var backends []string
	for _, trust := range meshTrusts.Items {
		backends = append(backends, kri.From(trust).String())
	}
	sort.SliceStable(backends, func(i, j int) bool {
		return backends[i] < backends[j]
	})
	return backends
}

func NewDataplaneInsightStore(resManager manager.ResourceManager) DataplaneInsightStore {
	return &dataplaneInsightStore{
		resManager: resManager,
	}
}

var _ DataplaneInsightStore = &dataplaneInsightStore{}

type dataplaneInsightStore struct {
	resManager manager.ResourceManager
}

func (s *dataplaneInsightStore) Upsert(
	ctx context.Context,
	xdsMetadata *structpb.Struct,
	dataplaneType core_model.ResourceType,
	dataplaneID core_model.ResourceKey,
	subscription *mesh_proto.DiscoverySubscription,
	mtlsInfo *MTLSInfo,
	otel *mesh_proto.DataplaneInsight_OpenTelemetry,
) error {
	switch dataplaneType {
	case core_mesh.DataplaneType:
		return manager.Upsert(ctx, s.resManager, dataplaneID, core_mesh.NewDataplaneInsightResource(), func(resource core_model.Resource) error {
			insight := resource.(*core_mesh.DataplaneInsightResource)
			if err := insight.Spec.UpdateSubscription(subscription); err != nil {
				return err
			}

			insight.Spec.Metadata = xdsMetadata
			insight.Spec.OpenTelemetry = otel
			if mtlsInfo == nil { // it means the proxy has no identity, we need to clear stats
				insight.Spec.MTLS = nil
			} else if insight.Spec.MTLS == nil ||
				(!mtlsInfo.ManagedExternally && !insight.Spec.MTLS.CertificateExpirationTime.AsTime().Equal(mtlsInfo.Expiration)) ||
				insight.Spec.MTLS.IssuedBackend != mtlsInfo.IssuedBackend ||
				!reflect.DeepEqual(insight.Spec.MTLS.SupportedBackends, mtlsInfo.SupportedBackends) {
				if err := insight.Spec.UpdateCert(mtlsInfo.Generation, mtlsInfo.Expiration, mtlsInfo.IssuedBackend, mtlsInfo.SupportedBackends, mtlsInfo.ManagedExternally); err != nil {
					return err
				}
			}
			return nil
		})
	default:
		// Return a designated precondition error since we don't expect other dataplane types.
		return store.ErrorInvalid(fmt.Sprintf("resource 'type=%q name=%q mesh=%q' is not expected to be stored in the insight resources", dataplaneType, dataplaneID.Name, dataplaneID.Mesh))
	}
}
