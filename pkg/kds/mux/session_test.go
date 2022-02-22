package mux_test

import (
	"bytes"
	"context"
	"io"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	envoy_sd "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"google.golang.org/grpc/metadata"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/kds/mux"
	"github.com/kumahq/kuma/pkg/test"
)

// This is a go antipattern but it's the simplest way to check we're not running send or recv from multiple goroutines.
var getGID = func() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}

type testMultiplexStream struct {
	sync.Mutex
	input                chan *mesh_proto.Message
	output               chan *mesh_proto.Message
	countSendByGoRoutine map[uint64]bool
	countRecvByGoRoutine map[uint64]bool
	ctx                  context.Context
}

func NewTestMultiplexStream(ctx context.Context, input chan *mesh_proto.Message, output chan *mesh_proto.Message) *testMultiplexStream {
	return &testMultiplexStream{
		ctx:                  metadata.AppendToOutgoingContext(ctx, mux.KDSVersionHeaderKey, mux.KDSVersionV3),
		input:                input,
		output:               output,
		countSendByGoRoutine: map[uint64]bool{},
		countRecvByGoRoutine: map[uint64]bool{},
	}
}

func (t *testMultiplexStream) Send(msg *mesh_proto.Message) error {
	t.Lock()
	id := getGID()
	t.countSendByGoRoutine[id] = true
	t.Unlock()
	t.output <- msg
	return nil
}

func (t *testMultiplexStream) Recv() (*mesh_proto.Message, error) {
	t.Lock()
	id := getGID()
	t.countRecvByGoRoutine[id] = true
	t.Unlock()
	select {
	case res := <-t.input:
		return res, nil
	case <-t.ctx.Done():
		return nil, io.EOF
	}
}

func (t *testMultiplexStream) Context() context.Context {
	return t.ctx
}

func (t *testMultiplexStream) CheckInvariants() error {
	t.Lock()
	defer t.Unlock()
	if len(t.countSendByGoRoutine) > 1 {
		return errors.New("Send() was called from multiple goroutines this shouldn't happen")
	}
	if len(t.countRecvByGoRoutine) > 1 {
		return errors.New("Recv() was called from multiple goroutines this shouldn't happen")
	}
	return nil
}

var _ = Describe("Multiplex Session", func() {

	Context("basic Send/Recv operations", func() {
		var clientSession mux.Session
		var serverSession mux.Session

		BeforeEach(func() {
			input := make(chan *mesh_proto.Message, 1)
			output := make(chan *mesh_proto.Message, 1)
			clientSession = mux.NewSession("global", NewTestMultiplexStream(context.Background(), input, output))
			serverSession = mux.NewSession("zone-1", NewTestMultiplexStream(context.Background(), output, input))
		})

		It("should Send to clientSession's ClientStream and Recv from serverSession's ServerStream", func() {
			err := clientSession.ClientStream().Send(&envoy_sd.DiscoveryRequest{VersionInfo: "1"})
			Expect(err).ToNot(HaveOccurred())
			msg, err := serverSession.ServerStream().Recv()
			Expect(err).ToNot(HaveOccurred())
			Expect(msg.VersionInfo).To(Equal("1"))
		})
		It("should Send to serverSession's ServerStream and Recv from clientSession's ClientStream", func() {
			err := serverSession.ServerStream().Send(&envoy_sd.DiscoveryResponse{VersionInfo: "2"})
			Expect(err).ToNot(HaveOccurred())
			msg, err := clientSession.ClientStream().Recv()
			Expect(err).ToNot(HaveOccurred())
			Expect(msg.VersionInfo).To(Equal("2"))
		})
		It("should Send to clientSession's ServerStream and Recv from serverSession's ClientStream", func() {
			err := clientSession.ServerStream().Send(&envoy_sd.DiscoveryResponse{VersionInfo: "3"})
			Expect(err).ToNot(HaveOccurred())
			msg, err := serverSession.ClientStream().Recv()
			Expect(err).ToNot(HaveOccurred())
			Expect(msg.VersionInfo).To(Equal("3"))
		})
		It("should Send to serverSession's ClientStream and Recv from clientSession's ServerStream", func() {
			err := serverSession.ClientStream().Send(&envoy_sd.DiscoveryRequest{VersionInfo: "4"})
			Expect(err).ToNot(HaveOccurred())
			msg, err := clientSession.ServerStream().Recv()
			Expect(err).ToNot(HaveOccurred())
			Expect(msg.VersionInfo).To(Equal("4"))
		})
	})

	It("When context is cancelled it should stop sending and receiving", test.Within(5*time.Minute, func() {
		dummyRequest := &envoy_sd.DiscoveryRequest{VersionInfo: "v2"}
		dummyResponse := &envoy_sd.DiscoveryResponse{VersionInfo: "v2"}
		input := make(chan *mesh_proto.Message, 1)
		output := make(chan *mesh_proto.Message, 1)
		ctx, cancelCtx := context.WithCancel(context.Background())
		muxStream := NewTestMultiplexStream(ctx, input, output)
		session := mux.NewSession("dummy", muxStream)

		Expect(session.ServerStream().Send(dummyResponse)).To(Succeed())
		<-output
		Expect(session.ClientStream().Send(dummyRequest)).To(Succeed())
		<-output

		input <- &mesh_proto.Message{Value: &mesh_proto.Message_Request{Request: dummyRequest}}
		req, err := session.ServerStream().Recv()
		Expect(err).ToNot(HaveOccurred())
		Expect(req).ToNot(BeNil())

		input <- &mesh_proto.Message{Value: &mesh_proto.Message_Response{Response: dummyResponse}}
		res, err := session.ClientStream().Recv()
		Expect(err).ToNot(HaveOccurred())
		Expect(res).ToNot(BeNil())

		// Let's cancel the ctx
		cancelCtx()
		err = <-session.Error()
		Expect(err).To(Equal(io.EOF))

		Expect(session.ServerStream().Send(dummyResponse)).To(Equal(io.EOF))
		Expect(session.ClientStream().Send(dummyRequest)).To(Equal(io.EOF))
		req, err = session.ServerStream().Recv()
		Expect(err).To(Equal(io.EOF))
		Expect(req).To(BeNil())

		res, err = session.ClientStream().Recv()
		Expect(err).To(Equal(io.EOF))
		Expect(res).To(BeNil())
	}))

	Context("concurrent operations", func() {

		Context("Recv", func() {
			var clientSession mux.Session
			var serverSession mux.Session

			BeforeEach(func() {
				input := make(chan *mesh_proto.Message, 1)
				output := make(chan *mesh_proto.Message, 1)
				clientSession = mux.NewSession("global", NewTestMultiplexStream(context.Background(), input, output))
				serverSession = mux.NewSession("zone-1", NewTestMultiplexStream(context.Background(), output, input))
			})
			It("should block while proper Send called", test.Within(time.Second, func() {
				wg := sync.WaitGroup{}
				wg.Add(1)

				go func() {
					request, err := serverSession.ServerStream().Recv()
					Expect(err).ToNot(HaveOccurred())
					Expect(request.VersionInfo).To(Equal("1"))
					wg.Done()
				}()

				err := clientSession.ServerStream().Send(&envoy_sd.DiscoveryResponse{VersionInfo: "2"})
				Expect(err).ToNot(HaveOccurred())
				err = clientSession.ClientStream().Send(&envoy_sd.DiscoveryRequest{VersionInfo: "1"})
				Expect(err).ToNot(HaveOccurred())

				resp, err := serverSession.ClientStream().Recv()
				Expect(err).ToNot(HaveOccurred())
				Expect(resp.VersionInfo).To(Equal("2"))

				wg.Wait()
			}))
		})

		Context("Single sided", func() {
			var input chan *mesh_proto.Message
			var output chan *mesh_proto.Message
			var muxStream *testMultiplexStream
			var session mux.Session
			dummyRequest := &envoy_sd.DiscoveryRequest{VersionInfo: "v2"}
			dummyResponse := &envoy_sd.DiscoveryResponse{VersionInfo: "v2"}
			var cancelCtx func()

			BeforeEach(func() {
				input = make(chan *mesh_proto.Message, 1)
				output = make(chan *mesh_proto.Message, 1)
				var ctx context.Context
				ctx, cancelCtx = context.WithCancel(context.Background())
				muxStream = NewTestMultiplexStream(ctx, input, output)
				session = mux.NewSession("dummy", muxStream)
			})

			It("calls clientStream and serverStream Send() in parallel", test.Within(time.Second*5, func() {
				numSent := int32(0)
				numItems := 10
				wg := sync.WaitGroup{}
				wg.Add(3)
				go func() {
					defer wg.Done()
					for i := 0; i < numItems; i++ {
						Expect(session.ClientStream().Send(dummyRequest)).ToNot(HaveOccurred())
					}
				}()
				go func() {
					defer wg.Done()
					for i := 0; i < numItems; i++ {
						Expect(session.ServerStream().Send(dummyResponse)).ToNot(HaveOccurred())
					}
				}()
				go func() {
					defer wg.Done()
					for {
						_, ok := <-output
						if !ok {
							return
						}
						atomic.AddInt32(&numSent, 1)
					}
				}()
				Eventually(func() int {
					return int(atomic.LoadInt32(&numSent))
				}).Should(Equal(numItems * 2))
				cancelCtx()
				<-session.Error()
				Expect(muxStream.CheckInvariants()).ToNot(HaveOccurred())
				close(output)
				wg.Wait()
			}))

			It("calls clientStream and serverStream Recv() in parallel", test.Within(time.Second*5, func() {
				numRecv := int32(0)
				readersWg := sync.WaitGroup{}
				numItems := 10
				readersWg.Add(1)
				go func() {
					defer readersWg.Done()
					for i := 0; i < numItems; i++ {
						_, err := session.ClientStream().Recv()
						if err != nil {
							return
						}
						atomic.AddInt32(&numRecv, 1)
					}
				}()
				readersWg.Add(1)
				go func() {
					defer readersWg.Done()
					for i := 0; i < numItems; i++ {
						_, err := session.ServerStream().Recv()
						if err != nil {
							return
						}
						atomic.AddInt32(&numRecv, 1)
					}
				}()
				for i := 0; i < numItems; i++ {
					input <- &mesh_proto.Message{Value: &mesh_proto.Message_Request{Request: dummyRequest}}
					input <- &mesh_proto.Message{Value: &mesh_proto.Message_Response{Response: dummyResponse}}
				}
				Eventually(func() int {
					return int(atomic.LoadInt32(&numRecv))
				}).Should(Equal(numItems * 2))
				cancelCtx()
				<-session.Error()
				readersWg.Wait()
				Expect(muxStream.CheckInvariants()).ToNot(HaveOccurred())
			}))
		})
	})
})
