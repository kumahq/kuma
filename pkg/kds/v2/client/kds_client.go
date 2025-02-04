package client

import (
	"io"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core/resources/model"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
)

type UpstreamResponse struct {
	ControlPlaneId      string
	Type                model.ResourceType
	AddedResources      model.ResourceList
	RemovedResourcesKey []model.ResourceKey
	IsInitialRequest    bool
}

func (u *UpstreamResponse) Validate() error {
	if u.AddedResources == nil {
		return nil
	}
	for _, res := range u.AddedResources.GetItems() {
		if err := model.Validate(res); err != nil {
			return err
		}
	}
	return nil
}

type Callbacks struct {
	OnResourcesReceived func(upstream UpstreamResponse) error
}

// All methods other than Receive() are non-blocking. It does not wait until the peer CP receives the message.
type DeltaKDSStream interface {
	DeltaDiscoveryRequest(resourceType model.ResourceType) error
	Receive() (UpstreamResponse, error)
	ACK(resourceType model.ResourceType) error
	NACK(resourceType model.ResourceType, err error) error
}

type KDSSyncClient interface {
	Receive() error
}

type kdsSyncClient struct {
	log             logr.Logger
	resourceTypes   []core_model.ResourceType
	callbacks       *Callbacks
	kdsStream       DeltaKDSStream
	responseBackoff time.Duration
}

func NewKDSSyncClient(
	log logr.Logger,
	rt []core_model.ResourceType,
	kdsStream DeltaKDSStream,
	cb *Callbacks,
	responseBackoff time.Duration,
) KDSSyncClient {
	return &kdsSyncClient{
		log:             log,
		resourceTypes:   rt,
		kdsStream:       kdsStream,
		callbacks:       cb,
		responseBackoff: responseBackoff,
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
			if err == io.EOF {
				return nil
			}
			return errors.Wrap(err, "failed to receive a discovery response")
		}
		s.log.V(1).Info("DeltaDiscoveryResponse received", "response", received)

		if err := received.Validate(); err != nil {
			s.log.Info("received resource is invalid, sending NACK", "err", err)
			if err := s.kdsStream.NACK(received.Type, err); err != nil {
				if err == io.EOF {
					return nil
				}
				return errors.Wrap(err, "failed to NACK a discovery response")
			}
			continue
		}

		if s.callbacks == nil {
			s.log.Info("no callback set, sending ACK", "type", string(received.Type))
			if err := s.kdsStream.ACK(received.Type); err != nil {
				if err == io.EOF {
					return nil
				}
				return errors.Wrap(err, "failed to ACK a discovery response")
			}
			continue
		}
		err = s.callbacks.OnResourcesReceived(received)
		if err != nil {
			if store.IsResourceAlreadyExists(err) {
				s.log.Info("received resource already exists, sending NACK", "err", err)
				if err := s.kdsStream.NACK(received.Type, err); err != nil {
					if err == io.EOF {
						return nil
					}
					return errors.Wrap(err, "failed to NACK a discovery response")
				}
				s.log.V(1).Info("NACK", "response", received)
				continue
			} else {
				return errors.Wrapf(err, "failed to store %s resources", received.Type)
			}
		}
		if !received.IsInitialRequest {
			// Execute backoff only on subsequent request.
			// When client first connects, the server sends empty DeltaDiscoveryResponse for every resource type.
			time.Sleep(s.responseBackoff)
		}
		s.log.V(1).Info("sending ACK", "type", received.Type)
		if err := s.kdsStream.ACK(received.Type); err != nil {
			if err == io.EOF {
				return nil
			}
			return errors.Wrap(err, "failed to ACK a discovery response")
		}
	}
}
