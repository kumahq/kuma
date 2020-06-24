package client

import (
	model "github.com/Kong/kuma/pkg/core/resources/model"
	"github.com/Kong/kuma/pkg/core/runtime/component"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"io"
)

type ClientFactory func() (KDSClient, error)

type Callbacks struct {
	OnResourcesReceived func(rs model.ResourceList) error
}

type kdsSink struct {
	log           logr.Logger
	resourceTypes []model.ResourceType
	clientFactory ClientFactory
	callbacks     *Callbacks
}

func NewKDSSink(log logr.Logger, rt []model.ResourceType, factory ClientFactory, cb *Callbacks) component.Component {
	return &kdsSink{
		log:           log,
		resourceTypes: rt,
		clientFactory: factory,
		callbacks:     cb,
	}
}

func (r *kdsSink) Start(stop <-chan struct{}) (errs error) {
	r.log.Info("creating a gRPC client for Kuma Discovery Service (KDS) server ...")
	client, err := r.clientFactory()
	if err != nil {
		return errors.Wrap(err, "failed to connect to gRPC server")
	}
	defer func() {
		r.log.Info("closing a connection ...")
		if err := client.Close(); err != nil {
			errs = multierr.Append(errs, errors.Wrapf(err, "failed to close a connection"))
		}
	}()

	r.log.Info("starting an KDS stream ...")
	stream, err := client.StartStream()
	if err != nil {
		return errors.Wrap(err, "failed to start an KDS stream")
	}
	defer func() {
		r.log.Info("closing an KDS stream ...")
		if err := stream.Close(); err != nil {
			errs = multierr.Append(errs, errors.Wrapf(err, "failed to close an xDS stream"))
		}
	}()

	for _, typ := range r.resourceTypes {
		r.log.Info("sending DiscoveryRequest ...", "type", typ)
		if err := stream.DiscoveryRequest("", typ); err != nil {
			return errors.Wrap(err, "discovering failed ...")
		}
	}

	for {
		select {
		case <-stop:
			return nil
		default:
		}

		rs, err := stream.Receive()
		if err != nil {
			return errors.Wrap(err, "failed to receive a discovery response")
		}
		r.log.Info("DiscoveryResponse received")

		if r.callbacks == nil {
			r.log.Info("sending ACK ...", "type", string(rs.GetItemType()))
			if err := stream.ACK(string(rs.GetItemType())); err != nil {
				if err == io.EOF {
					break
				}
				return errors.Wrap(err, "failed to ACK a discovery response")
			}
			continue
		}
		if err := r.callbacks.OnResourcesReceived(rs); err != nil {
			r.log.Info("error during callback received, sending NACK ...", "err", err)
			if err := stream.NACK(string(rs.GetItemType()), err); err != nil {
				if err == io.EOF {
					break
				}
				return errors.Wrap(err, "failed to NACK a discovery response")
			}
		} else {
			r.log.Info("sending ACK ...", "type", string(rs.GetItemType()))
			if err := stream.ACK(string(rs.GetItemType())); err != nil {
				if err == io.EOF {
					break
				}
				return errors.Wrap(err, "failed to ACK a discovery response")
			}
		}
	}
	return nil
}
