package client

import (
	"io"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
)

type Callbacks struct {
	OnResourcesReceived func(upstream UpstreamResponse) error
}

type KDSSyncClient interface {
	Receive() error
}

type kdsSyncClient struct {
	log           logr.Logger
	resourceTypes []core_model.ResourceType
	callbacks     *Callbacks
	kdsStream     KDSStream
}

func NewKDSSyncClient(log logr.Logger, rt []core_model.ResourceType, kdsStream KDSStream, cb *Callbacks) KDSSyncClient {
	return &kdsSyncClient{
		log:           log,
		resourceTypes: rt,
		kdsStream:     kdsStream,
		callbacks:     cb,
	}
}

func (s *kdsSyncClient) Receive() error {
	for _, typ := range s.resourceTypes {
		s.log.V(1).Info("sending DiscoveryRequest", "type", typ)
		if err := s.kdsStream.DiscoveryRequest(typ); err != nil {
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
		s.log.V(1).Info("DiscoveryResponse received", "response", received)

		if s.callbacks == nil {
			s.log.Info("no callback set, sending ACK", "type", string(received.Type))
			if err := s.kdsStream.ACK(string(received.Type)); err != nil {
				if err == io.EOF {
					return nil
				}
				return errors.Wrap(err, "failed to ACK a discovery response")
			}
			continue
		}
		if err := s.callbacks.OnResourcesReceived(received); err != nil {
			s.log.Info("error during callback received, sending NACK", "err", err)
			if err := s.kdsStream.NACK(string(received.Type), err); err != nil {
				if err == io.EOF {
					return nil
				}
				return errors.Wrap(err, "failed to NACK a discovery response")
			}
		} else {
			s.log.V(1).Info("sending ACK", "type", received.Type)
			if err := s.kdsStream.ACK(string(received.Type)); err != nil {
				if err == io.EOF {
					return nil
				}
				return errors.Wrap(err, "failed to ACK a discovery response")
			}
		}
	}
}
