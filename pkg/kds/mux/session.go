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

// handleRecv polls to receive messages from the KDSStream (the actual grpc bidi-stream).
// Depending on the message it dispatches to either the server receive buffer or the client receive buffer.
// It also closes both streams when an error on the recv side happens.
// We can rely on an error on recv to end the session because we're sure an error on recv will always happen, it might be io.EOF if we're just done.
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
		// We can safely not care about locking as we're only closing the channel from this goroutine.
		switch v := msg.Value.(type) {
		case *mesh_proto.Message_LegacyRequest:
			msg = &mesh_proto.Message{Value: &mesh_proto.Message_Request{Request: DiscoveryRequestV3(v.LegacyRequest)}}
			s.serverStream.bufferStream.recvBuffer <- msg
		case *mesh_proto.Message_Request:
			s.serverStream.bufferStream.recvBuffer <- msg
		case *mesh_proto.Message_LegacyResponse:
			msg = &mesh_proto.Message{Value: &mesh_proto.Message_Response{Response: DiscoveryResponseV3(v.LegacyResponse)}}
			s.clientStream.bufferStream.recvBuffer <- msg
		case *mesh_proto.Message_Response:
			s.clientStream.bufferStream.recvBuffer <- msg
		}
	}
}

// handleSend polls either sendBuffer and call send on the KDSStream (the actual grpc bidi-stream).
// This call is stopped whenever either of the sendBuffer are closed (in practice they are always closed together anyway).
func (s *session) handleSend(stream MultiplexStream) {
	kdsVersion := KDSVersion(stream.Context())
	for {
		select {
		case item, more := <-s.serverStream.bufferStream.sendBuffer:
			if !more {
				return
			}
			r := item.msg
			if kdsVersion == KDSVersionV2 {
				r = &mesh_proto.Message{Value: &mesh_proto.Message_LegacyResponse{LegacyResponse: DiscoveryResponseV2(r.GetResponse())}}
			}
			if err := stream.Send(r); err != nil {
				s.err <- err
				return
			}
			// item.errChan <- err
		case item, more := <-s.clientStream.bufferStream.sendBuffer:
			if !more {
				return
			}
			r := item.msg
			if kdsVersion == KDSVersionV2 {
				r = &mesh_proto.Message{Value: &mesh_proto.Message_LegacyRequest{LegacyRequest: DiscoveryRequestV2(r.GetRequest())}}
			}
			if err := stream.Send(r); err != nil {
				s.err <- err
				return
			}
			// item.errChan <- err
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

	// Protects the send-buffer against writing on a closed channel, this is needed as we don't control in which goroutine `Send` will be called.
	lock   sync.Mutex
	closed bool
}

func newBufferStream() *bufferStream {
	return &bufferStream{
		sendBuffer: make(chan sendItem, 1000),
		recvBuffer: make(chan *mesh_proto.Message, 1000),
	}
}

func (k *bufferStream) Send(message *mesh_proto.Message) error {
	k.lock.Lock()
	if k.closed {
		k.lock.Unlock()
		return io.EOF
	}
	k.lock.Unlock()
	errChan := make(chan error)
	k.sendBuffer <- sendItem{msg: message, errChan: errChan}

	// r := <-errChan
	return nil
}

func (k *bufferStream) Recv() (*mesh_proto.Message, error) {
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
