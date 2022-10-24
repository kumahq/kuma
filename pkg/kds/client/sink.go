package client

import (
	"io"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
)

type Callbacks struct {
	OnResourcesReceived func(clusterID string, rs core_model.ResourceList) error
}

type KDSSink interface {
	Receive() error
}

type kdsSink struct {
	log           logr.Logger
	resourceTypes []core_model.ResourceType
	callbacks     *Callbacks
	kdsStream     KDSStream
}

func NewKDSSink(log logr.Logger, rt []core_model.ResourceType, kdsStream KDSStream, cb *Callbacks) KDSSink {
	return &kdsSink{
		log:           log,
		resourceTypes: rt,
		kdsStream:     kdsStream,
		callbacks:     cb,
	}
}

func (s *kdsSink) Receive() error {
	for _, typ := range s.resourceTypes {
		s.log.V(1).Info("sending DiscoveryRequest", "type", typ)
		if err := s.kdsStream.DiscoveryRequest(typ); err != nil {
			return errors.Wrap(err, "discovering failed")
		}
	}

	for {
		clusterID, rs, err := s.kdsStream.Receive()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return errors.Wrap(err, "failed to receive a discovery response")
		}
		s.log.V(1).Info("DiscoveryResponse received", "response", rs)

		if s.callbacks == nil {
			s.log.Info("no callback set, sending ACK", "type", string(rs.GetItemType()))
			if err := s.kdsStream.ACK(string(rs.GetItemType())); err != nil {
				if err == io.EOF {
					return nil
				}
				return errors.Wrap(err, "failed to ACK a discovery response")
			}
			continue
		}
		if err := s.callbacks.OnResourcesReceived(clusterID, rs); err != nil {
			s.log.Info("error during callback received, sending NACK", "err", err)
			if err := s.kdsStream.NACK(string(rs.GetItemType()), err); err != nil {
				if err == io.EOF {
					return nil
				}
				return errors.Wrap(err, "failed to NACK a discovery response")
			}
		} else {
			s.log.V(1).Info("sending ACK", "type", string(rs.GetItemType()))
			if err := s.kdsStream.ACK(string(rs.GetItemType())); err != nil {
				if err == io.EOF {
					return nil
				}
				return errors.Wrap(err, "failed to ACK a discovery response")
			}
		}
	}
}
