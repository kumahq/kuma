package client

import (
	"context"
	std_errors "errors"
	"io"
	"time"

	envoy_sd "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
)

type UpstreamResponse struct {
	ControlPlaneId      string
	Type                core_model.ResourceType
	AddedResources      core_model.ResourceList
	InvalidResourcesKey []core_model.ResourceKey
	RemovedResourcesKey []core_model.ResourceKey
	IsInitialRequest    bool
}

func (u *UpstreamResponse) Validate() error {
	if u.AddedResources == nil {
		return nil
	}
	var err error
	for _, res := range u.AddedResources.GetItems() {
		if validationErr := core_model.Validate(res); validationErr != nil {
			err = std_errors.Join(err, validationErr)
			u.InvalidResourcesKey = append(u.InvalidResourcesKey, core_model.MetaToResourceKey(res.GetMeta()))
		}
	}
	return err
}

type Callbacks struct {
	OnResourcesReceived func(upstream UpstreamResponse) (error, error)
}

// All methods other than Receive() are non-blocking. It does not wait until the peer CP receives the message.
type DeltaKDSStream interface {
	SendMsg(*envoy_sd.DeltaDiscoveryRequest) error
	BuildDeltaSubScribeRequest(resourceType core_model.ResourceType) *envoy_sd.DeltaDiscoveryRequest
	Receive() (UpstreamResponse, error)
	BuildACKRequest(resourceType core_model.ResourceType) *envoy_sd.DeltaDiscoveryRequest
	BuildNACKRequest(resourceType core_model.ResourceType, err error) *envoy_sd.DeltaDiscoveryRequest
	MarkInitialRequestDone(resourceType core_model.ResourceType)
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
	type wrappedReq struct {
		deltaReq *envoy_sd.DeltaDiscoveryRequest
		cbFunc   func()
	}

	wrappedReqCh := make(chan wrappedReq)
	errCh := make(chan error)
	sendErrChannelFn := func(err error) {
		select {
		case errCh <- err:
		default:
			s.log.Error(err, "failed to send error to closed channel")
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	// DeltaDiscoveryRequest signals
	go func() {
		for _, typ := range s.resourceTypes {
			s.log.V(1).Info("sending DeltaDiscoveryRequest", "type", typ)
			subScribeRequest := s.kdsStream.BuildDeltaSubScribeRequest(typ)

			select {
			case <-ctx.Done():
				return
			case wrappedReqCh <- wrappedReq{
				deltaReq: subScribeRequest,
				cbFunc:   nil,
			}:
			}
		}
	}()

	// Sending messages through grpc stream
	go func() {
		defer cancel()

		for {
			select {
			case <-ctx.Done():
				return

			case req := <-wrappedReqCh:
				err := s.kdsStream.SendMsg(req.deltaReq)
				if err != nil {
					if err == io.EOF {
						s.log.V(1).Info("stream ended")
					} else {
						sendErrChannelFn(errors.Wrap(err, "failed to send request"))
					}
					return
				}
				if req.cbFunc != nil {
					req.cbFunc()
				}
			}
		}
	}()

	// receive from grpc stream
	go func() {
		defer cancel()

		s.log.V(1).Info("start to receive discovery responses")
		for {
			select {
			case <-ctx.Done():
				return

			default:
				received, err := s.kdsStream.Receive()
				if err != nil {
					if err == io.EOF {
						s.log.V(1).Info("stream ended")
					} else {
						sendErrChannelFn(errors.Wrap(err, "failed to receive a discovery response"))
					}
					return
				}

				s.log.V(1).Info("DeltaDiscoveryResponse received", "response", received)
				validationErrors := received.Validate()
				if validationErrors != nil {
					s.log.Info("received resource is invalid, sending NACK", "err", validationErrors)
					nackRequest := s.kdsStream.BuildNACKRequest(received.Type, validationErrors)
					if nackRequest == nil {
						continue
					}
					wrappedReqCh <- wrappedReq{
						deltaReq: nackRequest,
						cbFunc: func() {
							s.kdsStream.MarkInitialRequestDone(received.Type)
						},
					}
					continue
				}

				if s.callbacks != nil {
					err, nackError := s.callbacks.OnResourcesReceived(received)
					if err != nil {
						sendErrChannelFn(err)
						return
					}
					if nackError != nil {
						nackRequest := s.kdsStream.BuildNACKRequest(received.Type, nackError)
						if nackRequest == nil {
							continue
						}
						wrappedReqCh <- wrappedReq{
							deltaReq: nackRequest,
							cbFunc: func() {
								s.kdsStream.MarkInitialRequestDone(received.Type)
							},
						}
						continue
					}
				}

				if !received.IsInitialRequest {
					// Execute backoff only on subsequent request.
					// When client first connects, the server sends empty DeltaDiscoveryResponse for every resource type.
					time.Sleep(s.responseBackoff)
				}

				s.log.V(1).Info("sending ACK", "type", received.Type)
				ackRequest := s.kdsStream.BuildACKRequest(received.Type)
				if ackRequest == nil {
					continue
				}
				wrappedReqCh <- wrappedReq{
					deltaReq: ackRequest,
					cbFunc: func() {
						s.kdsStream.MarkInitialRequestDone(received.Type)
					},
				}
			}
		}
	}()

	err := <-errCh
	return err
}
