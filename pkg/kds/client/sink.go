package client

import (
	"io"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
)

type Callbacks struct {
	OnResourcesReceived func(clusterID string, rs model.ResourceList) error
}

type kdsSink struct {
	log           logr.Logger
	resourceTypes []model.ResourceType
	callbacks     *Callbacks
	kdsStream     KDSStream
}

func NewKDSSink(log logr.Logger, rt []model.ResourceType, kdsStream KDSStream, cb *Callbacks) component.Component {
	return &kdsSink{
		log:           log,
		resourceTypes: rt,
		kdsStream:     kdsStream,
		callbacks:     cb,
	}
}

func (s *kdsSink) Start(stop <-chan struct{}) (errs error) {
	for _, typ := range s.resourceTypes {
		s.log.V(1).Info("sending DiscoveryRequest", "type", typ)
		if err := s.kdsStream.DiscoveryRequest(typ); err != nil {
			return errors.Wrap(err, "discovering failed")
		}
	}

	for {
		select {
		case <-stop:
			return nil
		default:
		}

		clusterID, rs, err := s.kdsStream.Receive()
		if err != nil {
			if err == io.EOF {
				break
			}
			return errors.Wrap(err, "failed to receive a discovery response")
		}
		s.log.V(1).Info("DiscoveryResponse received", "response", rs)

		if s.callbacks == nil {
			s.log.Info("no callback set, sending ACK", "type", string(rs.GetItemType()))
			if err := s.kdsStream.ACK(string(rs.GetItemType())); err != nil {
				if err == io.EOF {
					break
				}
				return errors.Wrap(err, "failed to ACK a discovery response")
			}
			continue
		}
		if err := s.callbacks.OnResourcesReceived(clusterID, rs); err != nil {
			s.log.Info("error during callback received, sending NACK", "err", err)
			if err := s.kdsStream.NACK(string(rs.GetItemType()), err); err != nil {
				if err == io.EOF {
					break
				}
				return errors.Wrap(err, "failed to NACK a discovery response")
			}
		} else {
			s.log.V(1).Info("sending ACK", "type", string(rs.GetItemType()))
			if err := s.kdsStream.ACK(string(rs.GetItemType())); err != nil {
				if err == io.EOF {
					break
				}
				return errors.Wrap(err, "failed to ACK a discovery response")
			}
		}
	}
	return nil
}

func (s *kdsSink) NeedLeaderElection() bool {
	return false
}
