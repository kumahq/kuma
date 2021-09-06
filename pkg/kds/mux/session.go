package mux

import (
	"context"
	"io"
	"sync"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
)

type Session interface {
	ServerStream() mesh_proto.KumaDiscoveryService_StreamKumaResourcesServer
	ClientStream() mesh_proto.KumaDiscoveryService_StreamKumaResourcesClient
	PeerID() string
	Error() <-chan error
}

type session struct {
	peerID       string
	err          chan error
	serverStream *kdsServerStream
	clientStream *kdsClientStream
}

type MultiplexStream interface {
	Send(*mesh_proto.Message) error
	Recv() (*mesh_proto.Message, error)
	Context() context.Context
}

func NewSession(peerID string, stream MultiplexStream) Session {
	s := &session{
		peerID: peerID,
		err:    make(chan error),
		serverStream: &kdsServerStream{
			ctx:          stream.Context(),
			bufferStream: newBufferStream(),
		},
		clientStream: &kdsClientStream{
			ctx:          stream.Context(),
			bufferStream: newBufferStream(),
		},
	}
	go func() {
		s.handleSend(stream)
	}()
	go func() {
		s.handleRecv(stream)
	}()
	return s
}

func (s *session) handleRecv(stream MultiplexStream) {
	for {
		msg, err := stream.Recv()
		if err != nil {
			s.clientStream.bufferStream.close()
			s.serverStream.bufferStream.close()
			// Recv always finishes with either an EOF or another error
			s.err <- err
			return
		}
		switch v := msg.Value.(type) {
		case *mesh_proto.Message_LegacyRequest:
			msg = &mesh_proto.Message{Value: &mesh_proto.Message_Request{Request: DiscoveryRequestV3(v.LegacyRequest)}}
			s.serverStream.bufferStream.put(msg)
		case *mesh_proto.Message_Request:
			s.serverStream.bufferStream.put(msg)
		case *mesh_proto.Message_LegacyResponse:
			msg = &mesh_proto.Message{Value: &mesh_proto.Message_Response{Response: DiscoveryResponseV3(v.LegacyResponse)}}
			s.clientStream.bufferStream.put(msg)
		case *mesh_proto.Message_Response:
			s.clientStream.bufferStream.put(msg)
		}
	}
}

func (s *session) handleSend(stream MultiplexStream) {
	kdsVersion := KDSVersion(stream.Context())
	for {
		select {
		case itm, more := <-s.serverStream.bufferStream.sendBuffer:
			if !more {
				return
			}
			r := itm.msg
			if kdsVersion == KDSVersionV2 && r != nil {
				r = &mesh_proto.Message{Value: &mesh_proto.Message_LegacyResponse{LegacyResponse: DiscoveryResponseV2(r.GetResponse())}}
			}
			err := stream.Send(r)
			itm.errChan <- err
		case itm, more := <-s.clientStream.bufferStream.sendBuffer:
			if !more {
				return
			}
			r := itm.msg
			if kdsVersion == KDSVersionV2 && r != nil {
				r = &mesh_proto.Message{Value: &mesh_proto.Message_LegacyRequest{LegacyRequest: DiscoveryRequestV2(r.GetRequest())}}
			}
			err := stream.Send(r)
			itm.errChan <- err
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

func (s *session) Error() <-chan error {
	return s.err
}

type sendItem struct {
	msg     *mesh_proto.Message
	errChan chan error
}

type bufferStream struct {
	sendBuffer chan sendItem
	recvBuffer chan *mesh_proto.Message

	lock   sync.Mutex // Protects the write side of the buffer, not the read
	closed bool
}

func newBufferStream() *bufferStream {
	return &bufferStream{
		sendBuffer: make(chan sendItem, 1),
		recvBuffer: make(chan *mesh_proto.Message, 1),
	}
}

func (k *bufferStream) put(message *mesh_proto.Message) {
	k.recvBuffer <- message
}

func (k *bufferStream) Send(message *mesh_proto.Message) error {
	k.lock.Lock()
	if k.closed {
		k.lock.Unlock()
		return io.EOF
	}
	errChan := make(chan error)
	k.sendBuffer <- sendItem{msg: message, errChan: errChan}

	k.lock.Unlock()
	r := <-errChan
	return r
}

func (k *bufferStream) Recv() (*mesh_proto.Message, error) {
	k.lock.Lock()
	if k.closed {
		k.lock.Unlock()
		return nil, io.EOF
	}

	k.lock.Unlock()
	r, more := <-k.recvBuffer
	if !more {
		return nil, io.EOF
	}
	return r, nil
}

func (k *bufferStream) close() {
	k.lock.Lock()
	defer k.lock.Unlock()

	k.closed = true
	close(k.sendBuffer)
	close(k.recvBuffer)
}
