package mux

import (
	"sync/atomic"

	envoy_sd "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
)

type Session interface {
	ServerStream() mesh_proto.KumaDiscoveryService_StreamKumaResourcesServer
	ClientStream() mesh_proto.KumaDiscoveryService_StreamKumaResourcesClient
	PeerID() string
	Done() <-chan struct{}
	Error() error
}

type session struct {
	peerID       string
	done         chan struct{}
	err          chan error
	serverStream *kdsServerStream
	clientStream *kdsClientStream
	closed       int32
}

func NewSession(peerID string, stream MultiplexStream, stop <-chan struct{}) Session {
	s := &session{
		peerID: peerID,
		done:   make(chan struct{}, 1),
		err:    make(chan error, 1),
		serverStream: &kdsServerStream{
			requests:        make(chan *envoy_sd.DiscoveryRequest, 1),
			MultiplexStream: stream,
		},
		clientStream: &kdsClientStream{
			responses:       make(chan *envoy_sd.DiscoveryResponse, 1),
			MultiplexStream: stream,
		},
		closed: int32(0),
	}
	go func() {
		defer s.close()
		if err := s.handle(stream, stop); err != nil {
			s.err <- err
		}
	}()
	return s
}

func (s *session) handle(stream MultiplexStream, stop <-chan struct{}) error {
	for {
		select {
		case <-stop:
			return nil
		default:
			if atomic.LoadInt32(&s.closed) == 1 {
				return nil
			}
		}

		msg, err := stream.Recv()
		if err != nil {
			return err
		}
		switch v := msg.Value.(type) {
		case *mesh_proto.Message_LegacyRequest:
			s.serverStream.put(DiscoveryRequestV3(v.LegacyRequest))
		case *mesh_proto.Message_Request:
			s.serverStream.put(v.Request)
		case *mesh_proto.Message_LegacyResponse:
			s.clientStream.put(DiscoveryResponseV3(v.LegacyResponse))
		case *mesh_proto.Message_Response:
			s.clientStream.put(v.Response)
		}
	}
}

func (s *session) ServerStream() mesh_proto.KumaDiscoveryService_StreamKumaResourcesServer {
	return s.serverStream
}

func (s *session) ClientStream() mesh_proto.KumaDiscoveryService_StreamKumaResourcesClient {
	return s.clientStream
}

func (s *session) PeerID() string {
	return s.peerID
}

func (s *session) Done() <-chan struct{} {
	return s.done
}

func (s *session) Error() error {
	select {
	case err := <-s.err:
		return err
	default:
		return nil
	}
}

func (s *session) close() {
	close(s.done)
	close(s.serverStream.requests)
	close(s.clientStream.responses)
}
