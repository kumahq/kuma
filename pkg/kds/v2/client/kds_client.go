package client

import (
	"context"
	std_errors "errors"
	"fmt"
	"io"
	"time"

	envoy_sd "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/go-logr/logr"
	"golang.org/x/sync/errgroup"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	kds_util "github.com/kumahq/kuma/pkg/kds/v2/util"
)

type KDSSyncClient interface {
	Receive(ctx context.Context, group *errgroup.Group) error
}

type kdsSyncClient struct {
	log             logr.Logger
	resourceTypes   []core_model.ResourceType
	callbacks       *kds_util.Callbacks
	kdsStream       DeltaKDSStream
	responseBackoff time.Duration
}

func NewKDSSyncClient(
	log logr.Logger,
	rt []core_model.ResourceType,
	kdsStream DeltaKDSStream,
	cb *kds_util.Callbacks,
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

func (s *kdsSyncClient) Receive(ctx context.Context, group *errgroup.Group) error {
	type wrappedReq struct {
		deltaReq *envoy_sd.DeltaDiscoveryRequest
		cbFunc   func()
	}

	// the buffer size 2 is to avoid channel blocking.
	// there's a case where receiver goroutine is finished but senders are still trying to send requests to the channel
	wrappedReqCh := make(chan wrappedReq, 2)

	group.Go(func() error {
		for _, typ := range s.resourceTypes {
			req := wrappedReq{
				deltaReq: s.kdsStream.BuildDeltaSubScribeRequest(typ),
				cbFunc:   nil,
			}

			select {
			case <-ctx.Done():
				s.log.V(1).Info("stopping sending initial DeltaDiscoveryRequest signals")
				return nil
			case wrappedReqCh <- req:
				s.log.V(1).Info("sending DeltaDiscoveryRequest signal", "type", typ)
			}
		}
		return nil
	})

	group.Go(func() error {
		s.log.V(1).Info("start to send messages through grpc stream")
		for {
			select {
			case <-ctx.Done():
				s.log.V(1).Info("stopping sending grpc messages")
				return nil
			case req := <-wrappedReqCh:
				err := s.kdsStream.SendMsg(req.deltaReq)
				if err != nil {
					if std_errors.Is(err, io.EOF) {
						s.log.V(1).Info("stream ended")
						return nil
					}
					return fmt.Errorf("failed to send request: %w", err)
				}

				if req.cbFunc != nil {
					req.cbFunc()
				}
			}
		}
	})

	group.Go(func() error {
		s.log.V(1).Info("start to receive messages from grpc stream")
		for {
			select {
			case <-ctx.Done():
				s.log.V(1).Info("stopping receiving grpc messages")
			default:
				received, err := s.kdsStream.Receive()
				if err != nil {
					return fmt.Errorf("failed to receive a response: %w", err)
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
						return err
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
	})

	// group.Wait will call cancel function once all goroutines return,
	// then the grpc stream would be closed.
	err := group.Wait()
	if err != nil && !std_errors.Is(err, io.EOF) {
		return err
	}

	return nil
}
