package callbacks

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/xds/secrets"
)

var sinkLog = core.Log.WithName("xds").WithName("sink")

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
		secretsInfo *secrets.Info,
	) error
}

func NewDataplaneInsightSink(
	xdsMetadata *structpb.Struct,
	accessor SubscriptionStatusAccessor,
	secrets secrets.Secrets,
	newTicker func() *time.Ticker,
	generationTicker func() *time.Ticker,
	flushBackoff time.Duration,
	store DataplaneInsightStore,
) DataplaneInsightSink {
	metadata := core_xds.DataplaneMetadataFromXdsMetadata(xdsMetadata)

	var dpType core_model.ResourceType
	// If the dataplane was started with a resource YAML, then it will be serialized in
	// the node metadata and we would know the underlying type directly. Since that
	// is optional, we can't depend on it here, so we map from the proxy type, which is
	// guaranteed.
	switch metadata.GetProxyType() {
	case mesh_proto.IngressProxyType:
		dpType = core_mesh.ZoneIngressType
	case mesh_proto.DataplaneProxyType:
		dpType = core_mesh.DataplaneType
	case mesh_proto.EgressProxyType:
		dpType = core_mesh.ZoneEgressType
	}

	return &dataplaneInsightSink{
		flushTicker:      newTicker,
		generationTicker: generationTicker,
		dataplaneType:    dpType,
		accessor:         accessor,
		secrets:          secrets,
		flushBackoff:     flushBackoff,
		store:            store,
		xdsMetadata:      xdsMetadata,
	}
}

var _ DataplaneInsightSink = &dataplaneInsightSink{}

type dataplaneInsightSink struct {
	flushTicker      func() *time.Ticker
	generationTicker func() *time.Ticker
	dataplaneType    core_model.ResourceType
	accessor         SubscriptionStatusAccessor
	secrets          secrets.Secrets
	store            DataplaneInsightStore
	flushBackoff     time.Duration
	xdsMetadata      *structpb.Struct
}

func (s *dataplaneInsightSink) Start(stop <-chan struct{}) {
	flushTicker := s.flushTicker()
	defer flushTicker.Stop()

	generationTicker := s.generationTicker()
	defer generationTicker.Stop()

	var lastStoredState *mesh_proto.DiscoverySubscription
	var lastStoredSecretsInfo *secrets.Info
	var generation uint32

	proxyType, err := core_mesh.ProxyTypeFromResourceType(s.dataplaneType)
	if err != nil {
		sinkLog.Error(err, "failed to create dataplaneInsightSink")
		return
	}

	flush := func(closing bool) {
		dataplaneID, currentState := s.accessor.GetStatus()
		secretsInfo := s.secrets.Info(proxyType, dataplaneID)

		select {
		case <-generationTicker.C:
			generation++
		default:
		}
		currentState.Generation = generation

		if proto.Equal(currentState, lastStoredState) && secretsInfo == lastStoredSecretsInfo {
			return
		}

		ctx := context.TODO()

		if err := s.store.Upsert(ctx, s.xdsMetadata, s.dataplaneType, dataplaneID, currentState, secretsInfo); err != nil {
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
			case store.IsAlreadyExists(err) || store.IsConflict(err):
				sinkLog.V(1).Info("failed to flush DataplaneInsight because it was updated in other place. Will retry in the next tick",
					"dataplaneid", dataplaneID)
			default:
				sinkLog.Error(err, "failed to flush DataplaneInsight", "dataplaneid", dataplaneID)
			}
		} else {
			sinkLog.V(1).Info("DataplaneInsight saved", "dataplaneid", dataplaneID, "subscription", currentState)
			lastStoredState = currentState
			lastStoredSecretsInfo = secretsInfo
		}
	}

	// flush the first insight as quickly as possible so
	// 1) user sees that DP is online in kumactl/GUI (even without any XDS updates)
	// 2) we can have lower deregistrationDelay, see pkg/xds/server/callbacks/dataplane_lifecycle.go#deregisterProxy
	flush(false)

	for {
		select {
		case <-flushTicker.C:
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
	ctx context.Context,
	xdsMetadata *structpb.Struct,
	dataplaneType core_model.ResourceType,
	dataplaneID core_model.ResourceKey,
	subscription *mesh_proto.DiscoverySubscription,
	secretsInfo *secrets.Info,
) error {
	switch dataplaneType {
	case core_mesh.ZoneIngressType:
		return manager.Upsert(ctx, s.resManager, dataplaneID, core_mesh.NewZoneIngressInsightResource(), func(resource core_model.Resource) error {
			insight := resource.(*core_mesh.ZoneIngressInsightResource)
			return insight.Spec.UpdateSubscription(subscription)
		})
	case core_mesh.ZoneEgressType:
		return manager.Upsert(ctx, s.resManager, dataplaneID, core_mesh.NewZoneEgressInsightResource(), func(resource core_model.Resource) error {
			insight := resource.(*core_mesh.ZoneEgressInsightResource)
			return insight.Spec.UpdateSubscription(subscription)
		})
	case core_mesh.DataplaneType:
		return manager.Upsert(ctx, s.resManager, dataplaneID, core_mesh.NewDataplaneInsightResource(), func(resource core_model.Resource) error {
			insight := resource.(*core_mesh.DataplaneInsightResource)
			if err := insight.Spec.UpdateSubscription(subscription); err != nil {
				return err
			}

			insight.Spec.Metadata = xdsMetadata

			if secretsInfo == nil { // it means mTLS was disabled, we need to clear stats
				insight.Spec.MTLS = nil
			} else if insight.Spec.MTLS == nil ||
				insight.Spec.MTLS.CertificateExpirationTime.AsTime() != secretsInfo.Expiration ||
				insight.Spec.MTLS.IssuedBackend != secretsInfo.IssuedBackend ||
				!reflect.DeepEqual(insight.Spec.MTLS.SupportedBackends, secretsInfo.SupportedBackends) {
				if err := insight.Spec.UpdateCert(secretsInfo.Generation, secretsInfo.Expiration, secretsInfo.IssuedBackend, secretsInfo.SupportedBackends); err != nil {
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
