package callbacks

import (
	"time"

	"github.com/golang/protobuf/proto"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
)

var sinkLog = core.Log.WithName("xds").WithName("sink")

type DataplaneInsightSink interface {
	Start(stop <-chan struct{})
}

type DataplaneInsightStore interface {
	// Upsert creates or updates the subscription, storing it with
	// the key dataplaneID. dataplaneType gives the resource type of
	// the dataplane proxy that has subscribed.
	Upsert(dataplaneType core_model.ResourceType, dataplaneID core_model.ResourceKey, subscription *mesh_proto.DiscoverySubscription) error
}

func NewDataplaneInsightSink(
	dataplaneType core_model.ResourceType,
	accessor SubscriptionStatusAccessor,
	newTicker func() *time.Ticker,
	flushBackoff time.Duration,
	store DataplaneInsightStore) DataplaneInsightSink {
	return &dataplaneInsightSink{
		newTicker:     newTicker,
		dataplaneType: dataplaneType,
		accessor:      accessor,
		flushBackoff:  flushBackoff,
		store:         store,
	}
}

var _ DataplaneInsightSink = &dataplaneInsightSink{}

type dataplaneInsightSink struct {
	newTicker     func() *time.Ticker
	dataplaneType core_model.ResourceType
	accessor      SubscriptionStatusAccessor
	store         DataplaneInsightStore
	flushBackoff  time.Duration
}

func (s *dataplaneInsightSink) Start(stop <-chan struct{}) {
	ticker := s.newTicker()
	defer ticker.Stop()

	var lastStoredState *mesh_proto.DiscoverySubscription

	flush := func(closing bool) {
		dataplaneID, currentState := s.accessor.GetStatus()
		if proto.Equal(currentState, lastStoredState) {
			return
		}

		copy := proto.Clone(currentState).(*mesh_proto.DiscoverySubscription)
		if err := s.store.Upsert(s.dataplaneType, dataplaneID, copy); err != nil {
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
					"dataplaneid", dataplaneID,
					"err", err)
			case store.IsResourceConflict(err):
				sinkLog.V(1).Info("failed to flush DataplaneInsight because it was updated in other place. Will retry in the next tick",
					"dataplaneid", dataplaneID)
			case store.IsResourcePreconditionFailed(err):
				sinkLog.V(1).Info("failed to flush DataplaneInsight for unsupported resource",
					"dataplaneid", dataplaneID,
					"err", err,
				)
			default:
				sinkLog.Error(err, "failed to flush DataplaneInsight", "dataplaneid", dataplaneID)
			}
		} else {
			sinkLog.V(1).Info("DataplaneInsight saved", "dataplaneid", dataplaneID, "subscription", currentState)
			lastStoredState = currentState
		}
	}

	for {
		select {
		case <-ticker.C:
			flush(false)
			// On Kubernetes, because of the cache subsequent Get, Update requests can fail, because the cache is not strongly consistent.
			// We handle the Resource Conflict logging on V1, but we can try to avoid the situation with backoff
			time.Sleep(s.flushBackoff)
		case <-stop:
			flush(true)
			return
		}
	}
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
	dataplaneType core_model.ResourceType,
	dataplaneID core_model.ResourceKey,
	subscription *mesh_proto.DiscoverySubscription,
) error {
	switch dataplaneType {
	case core_mesh.ZoneIngressType:
		return manager.Upsert(s.resManager, dataplaneID, core_mesh.NewZoneIngressInsightResource(), func(resource core_model.Resource) {
			insight := resource.(*core_mesh.ZoneIngressInsightResource)
			insight.Spec.UpdateSubscription(subscription)
		})
	case core_mesh.DataplaneType:
		return manager.Upsert(s.resManager, dataplaneID, core_mesh.NewDataplaneInsightResource(), func(resource core_model.Resource) {
			insight := resource.(*core_mesh.DataplaneInsightResource)
			insight.Spec.UpdateSubscription(subscription)
		})
	default:
		// Return a designated precondition error since we don't expect other dataplane types.
		return store.ErrorResourcePreconditionFailed(dataplaneType, dataplaneID.Name, dataplaneID.Mesh)
	}
}
