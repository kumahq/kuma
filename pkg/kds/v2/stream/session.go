package mux

import (
	"context"
	"sync"
	"time"

	envoy_sd "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
)

type Session interface {
	ServerStream() mesh_proto.KDSSyncService_GlobalToZoneSyncServer
	PeerID() string
	Error() <-chan error
}

type session struct {
	peerID       string
	serverStream *kdsServerStream

	err       chan error
	sync.Once // protects err, so we only send the first error and close the channel
}

type KDSSyncStream interface {
	Send(*envoy_sd.DeltaDiscoveryResponse) error
	Recv() (*envoy_sd.DeltaDiscoveryRequest, error)
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
func NewSession(peerID string, stream KDSSyncStream, bufferSize uint32, sendTimeout time.Duration) Session {
	s := &session{
		peerID: peerID,
		err:    make(chan error),
		serverStream: &kdsServerStream{
			ctx:        stream.Context(),
			sendBuffer: make(chan *envoy_sd.DeltaDiscoveryResponse, bufferSize),
			recvBuffer: make(chan *envoy_sd.DeltaDiscoveryRequest, bufferSize),
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

// handleRecv polls to receive messages from the KDSSync.
// Depending on the message it dispatches to either the server receive buffer or the client receive buffer.
// It also closes both streams when an error on the recv side happens.
// We can rely on an error on recv to end the session because we're sure an error on recv will always happen, it might be io.EOF if we're just done.
func (s *session) handleRecv(stream KDSSyncStream) {
	for {
		msg, err := stream.Recv()
		if err != nil {
			s.serverStream.close()
			// Recv always finishes with either an EOF or another error
			s.setError(err)
			return
		}
		// We can safely not care about locking as we're only closing the channel from this goroutine.
		s.serverStream.recvBuffer <- msg
	}
}

// handleSend polls either sendBuffer and call send on the KDSSync.
// This call is stopped whenever either of the sendBuffer are closed (in practice they are always closed together anyway).
func (s *session) handleSend(stream KDSSyncStream, sendTimeout time.Duration) {
	for {
		var msgToSend *envoy_sd.DeltaDiscoveryResponse
		msg, more := <-s.serverStream.sendBuffer
		if !more {
			return
		}
		msgToSend = msg
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
				s.setError(errors.New("timeout while sending a message to peer"))
			}
		}()
		if err := stream.Send(msgToSend); err != nil {
			s.setError(err)
			cancel()
			return
		}
		cancel()
	}
}

func (s *session) setError(err error) {
	// execute this once so writers to this channel won't be stuck or trying to write to a close channel
	// We only care about the first error, because it results in broken session anyway.
	s.Once.Do(func() {
		s.err <- err
		close(s.err)
	})
}

func (s *session) ServerStream() mesh_proto.KDSSyncService_GlobalToZoneSyncServer {
	return s.serverStream
}

func (s *session) PeerID() string {
	return s.peerID
}

func (s *session) Error() <-chan error {
	return s.err
}
