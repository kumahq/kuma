package mux

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
)

type Session interface {
	ServerStream() mesh_proto.KumaDiscoveryService_StreamKumaResourcesServer
	ClientStream() mesh_proto.KumaDiscoveryService_StreamKumaResourcesClient
	// PeerID is also known as the client id.
	PeerID() string
	Error() <-chan error
	SetError(err error)
}

type session struct {
	peerID       string
	serverStream *kdsServerStream
	clientStream *kdsClientStream

	err       chan error
	sync.Once // protects err, so we only send the first error and close the channel
}

type MultiplexStream interface {
	Send(*mesh_proto.Message) error
	Recv() (*mesh_proto.Message, error)
	Context() context.Context
}

// NewSession creates a multiplexed session for KDS so both CP sends and receives resources.
//
// Buffer settings recommendations:
// The buffer size should be of a size of all inflight request, so we are never blocked on the buffer channels.
// The buffer is separate for each direction (send/receive) on each multiplexed stream (ex. global acting as server/global acting as client).
// The maximum number of inflight requests are the number of synced resources, because:
//   - A CP never sends multiple DiscoveryRequests for one resource type.
//   - A CP never sends multiple DiscoveryResponses for one resource type (it waits until peer answers with ACK/NACK for the previous number)
//
// We could carefully count which resources are synced each way,
// but for the simplicity it's recommended to a set a buffer to number of resources in CP leaving a bit of buffer for resources unknown to the other side.
func NewSession(peerID string, stream MultiplexStream, bufferSize uint32, sendTimeout time.Duration) Session {
	s := &session{
		peerID: peerID,
		err:    make(chan error),
		serverStream: &kdsServerStream{
			ctx:          stream.Context(),
			bufferStream: newBufferStream(bufferSize),
		},
		clientStream: &kdsClientStream{
			ctx:          stream.Context(),
			bufferStream: newBufferStream(bufferSize),
		},
	}
	go func() {
		s.handleSend(stream, sendTimeout)
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
			s.SetError(err)
			return
		}
		// convert legacy messages
		switch v := msg.Value.(type) {
		case *mesh_proto.Message_LegacyRequest:
			msg = &mesh_proto.Message{Value: &mesh_proto.Message_Request{Request: DiscoveryRequestV3(v.LegacyRequest)}}
		case *mesh_proto.Message_LegacyResponse:
			msg = &mesh_proto.Message{Value: &mesh_proto.Message_Response{Response: DiscoveryResponseV3(v.LegacyResponse)}}
		}
		// We can safely not care about locking as we're only closing the channel from this goroutine.
		switch msg.Value.(type) {
		case *mesh_proto.Message_Request:
			s.serverStream.bufferStream.recvBuffer <- msg
		case *mesh_proto.Message_Response:
			s.clientStream.bufferStream.recvBuffer <- msg
		}
	}
}

// handleSend polls either sendBuffer and call send on the KDSStream (the actual grpc bidi-stream).
// This call is stopped whenever either of the sendBuffer are closed (in practice they are always closed together anyway).
func (s *session) handleSend(stream MultiplexStream, sendTimeout time.Duration) {
	kdsVersion := KDSVersion(stream.Context())
	for {
		var msgToSend *mesh_proto.Message
		select {
		case msg, more := <-s.serverStream.bufferStream.sendBuffer:
			if !more {
				return
			}
			if kdsVersion == KDSVersionV2 {
				msg = &mesh_proto.Message{Value: &mesh_proto.Message_LegacyResponse{LegacyResponse: DiscoveryResponseV2(msg.GetResponse())}}
			}
			msgToSend = msg
		case msg, more := <-s.clientStream.bufferStream.sendBuffer:
			if !more {
				return
			}
			if kdsVersion == KDSVersionV2 {
				msg = &mesh_proto.Message{Value: &mesh_proto.Message_LegacyRequest{LegacyRequest: DiscoveryRequestV2(msg.GetRequest())}}
			}
			msgToSend = msg
		}
		ctx, cancel := context.WithTimeout(context.Background(), sendTimeout)
		go func() {
			<-ctx.Done()
			if ctx.Err() == context.DeadlineExceeded {
				// This is very unlikely to happen, but it was introduced as a last resort protection from a gRPC streaming deadlock.
				// gRPC streaming deadlock may happen if both peers are stuck on Send() operation without calling Recv() often enough.
				// In this case, if data is big enough, both parties may wait for WINDOW_UPDATE on HTTP/2 stream.
				// We fixed the deadlock by increasing buffer size which is larger that all possible inflight request.
				// If the connection is broken and send is stuck, it's more likely for gRPC keep alive to catch such case.
				// If you still hit the timeout without deadlock, you may increase it. However, there are two possible scenarios
				// 1) This is a malicious client reading stream byte by byte. In this case it's actually better to end the stream
				// 2) A client is such overwhelmed that it cannot even let the server know that it's ready to receive more data.
				//    In this case it's recommended to scale number of instances.
				s.SetError(errors.New("timeout while sending a message to peer"))
			}
		}()
		if err := stream.Send(msgToSend); err != nil {
			s.SetError(err)
			cancel()
			return
		}
		cancel()
	}
}

func (s *session) SetError(err error) {
	// execute this once so writers to this channel won't be stuck or trying to write to a close channel
	// We only care about the first error, because it results in broken session anyway.
	s.Once.Do(func() {
		s.err <- err
		close(s.err)
	})
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

type bufferStream struct {
	sendBuffer chan *mesh_proto.Message
	recvBuffer chan *mesh_proto.Message

	// Protects the send-buffer against writing on a closed channel, this is needed as we don't control in which goroutine `Send` will be called.
	lock   sync.Mutex
	closed bool
}

func newBufferStream(bufferSize uint32) *bufferStream {
	return &bufferStream{
		sendBuffer: make(chan *mesh_proto.Message, bufferSize),
		recvBuffer: make(chan *mesh_proto.Message, bufferSize),
	}
}

func (k *bufferStream) Send(message *mesh_proto.Message) error {
	k.lock.Lock()
	defer k.lock.Unlock()
	if k.closed {
		return io.EOF
	}
	k.sendBuffer <- message
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
