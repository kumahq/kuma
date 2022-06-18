package client

import (
	"errors"
	"fmt"
	"io"

	"github.com/go-logr/logr"

	model "github.com/kumahq/kuma/pkg/core/resources/model"
)

type Callbacks struct {
	OnResourcesReceived func(clusterID string, rs model.ResourceList) error
}

type KDSSink interface {
	Receive() error
}

type kdsSink struct {
	log           logr.Logger
	resourceTypes []model.ResourceType
	callbacks     *Callbacks
	kdsStream     KDSStream
}

func NewKDSSink(log logr.Logger, rt []model.ResourceType, kdsStream KDSStream, cb *Callbacks) KDSSink {
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
			return fmt.Errorf("discovering failed: %w", err)
		}
	}

	for {
		clusterID, rs, err := s.kdsStream.Receive()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return fmt.Errorf("failed to receive a discovery response: %w", err)
		}
		s.log.V(1).Info("DiscoveryResponse received", "response", rs)

		if s.callbacks == nil {
			s.log.Info("no callback set, sending ACK", "type", string(rs.GetItemType()))
			if err := s.kdsStream.ACK(string(rs.GetItemType())); err != nil {
				if errors.Is(err, io.EOF) {
					return nil
				}
				return fmt.Errorf("failed to ACK a discovery response: %w", err)
			}
			continue
		}
		if err := s.callbacks.OnResourcesReceived(clusterID, rs); err != nil {
			s.log.Info("error during callback received, sending NACK", "err", err)
			if err := s.kdsStream.NACK(string(rs.GetItemType()), err); err != nil {
				if errors.Is(err, io.EOF) {
					return nil
				}
				return fmt.Errorf("failed to NACK a discovery response: %w", err)
			}
		} else {
			s.log.V(1).Info("sending ACK", "type", string(rs.GetItemType()))
			if err := s.kdsStream.ACK(string(rs.GetItemType())); err != nil {
				if errors.Is(err, io.EOF) {
					return nil
				}
				return fmt.Errorf("failed to ACK a discovery response: %w", err)
			}
		}
	}
}
