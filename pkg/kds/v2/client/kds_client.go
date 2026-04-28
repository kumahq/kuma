package client

import (
	std_errors "errors"
	"os"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/core/resources/validator"
)

type UpstreamResponse struct {
	ControlPlaneId      string
	Nonce               string
	Type                core_model.ResourceType
	AddedResources      core_model.ResourceList
	InvalidResourcesKey []core_model.ResourceKey
	RemovedResourcesKey []core_model.ResourceKey
	IsInitialRequest    bool
}

const debugKDSPayloadDumpEnv = "KUMA_DEBUG_KDS_DUMP"

func (u *UpstreamResponse) Validate() error {
	if u.AddedResources == nil {
		return nil
	}
	var err error
	for _, res := range u.AddedResources.GetItems() {
		if validationErr := validator.Validate(res); validationErr != nil {
			err = std_errors.Join(err, validationErr)
			u.InvalidResourcesKey = append(u.InvalidResourcesKey, core_model.MetaToResourceKey(res.GetMeta()))
		}
	}
	return err
}

type Callbacks struct {
	OnResourcesReceived func(upstream UpstreamResponse) (error, error)
	OnNACK              func(resourceType core_model.ResourceType) // optional; called each time a NACK is sent
}

// All methods other than Receive() are non-blocking. It does not wait until the peer CP receives the message.
type DeltaKDSStream interface {
	DeltaDiscoveryRequest(resourceType core_model.ResourceType) error
	Receive() (UpstreamResponse, error)
	ACK(resourceType core_model.ResourceType) error
	NACK(resourceType core_model.ResourceType, err error) error
	CloseSend() error
}

type KDSSyncClient interface {
	Receive() error
}

type kdsSyncClient struct {
	log              logr.Logger
	resourceTypes    []core_model.ResourceType
	callbacks        *Callbacks
	kdsStream        DeltaKDSStream
	responseBackoff  time.Duration
	debugPayloadDump bool
}

func NewKDSSyncClient(
	log logr.Logger,
	rt []core_model.ResourceType,
	kdsStream DeltaKDSStream,
	cb *Callbacks,
	responseBackoff time.Duration,
) KDSSyncClient {
	return &kdsSyncClient{
		log:              log,
		resourceTypes:    rt,
		kdsStream:        kdsStream,
		callbacks:        cb,
		responseBackoff:  responseBackoff,
		debugPayloadDump: os.Getenv(debugKDSPayloadDumpEnv) == "true",
	}
}

func (s *kdsSyncClient) Receive() error {
	for _, typ := range s.resourceTypes {
		s.log.V(1).Info("sending DeltaDiscoveryRequest", "type", typ)
		if err := s.kdsStream.DeltaDiscoveryRequest(typ); err != nil {
			return errors.Wrap(err, "discovering failed")
		}
	}

	for {
		received, err := s.kdsStream.Receive()
		if err != nil {
			return errors.Wrap(err, "failed to receive a discovery response")
		}
		s.logReceivedResponse(received)
		validationErrors := received.Validate()

		if s.callbacks == nil {
			if validationErrors != nil {
				s.log.Info("received resource is invalid, sending NACK", "err", validationErrors, "nonce", received.Nonce)
				if err := s.kdsStream.NACK(received.Type, validationErrors); err != nil {
					return errors.Wrap(err, "failed to NACK a discovery response")
				}
				continue
			}
			s.log.Info("no callback set, sending ACK", "type", string(received.Type), "nonce", received.Nonce)
			if err := s.kdsStream.ACK(received.Type); err != nil {
				return errors.Wrap(err, "failed to ACK a discovery response")
			}
			continue
		}
		err, nackError := s.callbacks.OnResourcesReceived(received)
		if err != nil {
			return errors.Wrapf(err, "failed to store %s resources", received.Type)
		}
		if nackError != nil || validationErrors != nil {
			combinedErrors := std_errors.Join(nackError, validationErrors)
			s.log.Info("received resource is invalid, sending NACK", "err", combinedErrors, "nonce", received.Nonce)
			if s.callbacks.OnNACK != nil {
				s.callbacks.OnNACK(received.Type)
			}
			if err := s.kdsStream.NACK(received.Type, combinedErrors); err != nil {
				return errors.Wrap(err, "failed to NACK a discovery response")
			}
			continue
		}
		if !received.IsInitialRequest {
			// Execute backoff only on subsequent request.
			// When client first connects, the server sends empty DeltaDiscoveryResponse for every resource type.
			time.Sleep(s.responseBackoff)
		}
		s.log.V(1).Info("sending ACK", "type", received.Type, "nonce", received.Nonce)
		if err := s.kdsStream.ACK(received.Type); err != nil {
			return errors.Wrap(err, "failed to ACK a discovery response")
		}
	}
}

func (s *kdsSyncClient) logReceivedResponse(received UpstreamResponse) {
	if s.debugPayloadDump {
		s.log.V(1).Info("DeltaDiscoveryResponse received", "response", received)
		return
	}

	s.log.V(1).Info(
		"DeltaDiscoveryResponse received",
		"type", received.Type,
		"nonce", received.Nonce,
		"addedResourcesCount", resourceCount(received.AddedResources),
		"removedResourcesCount", len(received.RemovedResourcesKey),
	)
}

func resourceCount(resources core_model.ResourceList) int {
	if resources == nil {
		return 0
	}

	return len(resources.GetItems())
}
