package mux_test

import (
	"context"
	"io"
	"strings"

	envoy_sd "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	stream_v3 "github.com/envoyproxy/go-control-plane/pkg/server/stream/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/kumahq/kuma/v2/pkg/kds/mux"
)

type fakeDeltaStream struct {
	sendErr error
	recvErr error
}

func (f *fakeDeltaStream) Send(*envoy_sd.DeltaDiscoveryResponse) error { return f.sendErr }
func (f *fakeDeltaStream) Recv() (*envoy_sd.DeltaDiscoveryRequest, error) {
	return &envoy_sd.DeltaDiscoveryRequest{}, f.recvErr
}
func (f *fakeDeltaStream) SetHeader(metadata.MD) error  { return nil }
func (f *fakeDeltaStream) SendHeader(metadata.MD) error { return nil }
func (f *fakeDeltaStream) SetTrailer(metadata.MD)       {}
func (f *fakeDeltaStream) Context() context.Context     { return context.Background() }
func (f *fakeDeltaStream) SendMsg(any) error            { return nil }
func (f *fakeDeltaStream) RecvMsg(any) error            { return nil }

var _ stream_v3.DeltaStream = &fakeDeltaStream{}

var _ = Describe("ErrorRecorderStream", func() {
	tooLarge := status.Error(codes.ResourceExhausted, "grpc: received message after decompression larger than max 10485760")

	response := func(typeURL string, size int) *envoy_sd.DeltaDiscoveryResponse {
		return &envoy_sd.DeltaDiscoveryResponse{
			TypeUrl:   typeURL,
			Resources: []*envoy_sd.Resource{{Name: "r", Version: strings.Repeat("a", size)}},
		}
	}

	It("should name the largest response sent when the peer rejects a message on Recv", func() {
		// given
		delegate := &fakeDeltaStream{recvErr: tooLarge}
		s := mux.NewErrorRecorderStream(delegate)
		Expect(s.Send(response("Dataplane", 10))).To(Succeed())
		Expect(s.Send(response("DataplaneInsight", 500))).To(Succeed())
		Expect(s.Send(response("Dataplane", 10))).To(Succeed())

		// when
		_, err := s.Recv()

		// then
		Expect(err).To(MatchError(ContainSubstring("type=DataplaneInsight size=")))
		Expect(err).To(MatchError(ContainSubstring("resources=1")))
		Expect(err).To(MatchError(ContainSubstring("larger than max 10485760")))
		Expect(status.Code(err)).To(Equal(codes.ResourceExhausted))
		Expect(s.Err()).To(Equal(err))
	})

	It("should name the response when the local transport rejects it on Send", func() {
		// given
		s := mux.NewErrorRecorderStream(&fakeDeltaStream{sendErr: tooLarge})

		// when
		err := s.Send(response("Dataplane", 10))

		// then
		Expect(err).To(MatchError(ContainSubstring("type=Dataplane size=")))
		Expect(status.Code(err)).To(Equal(codes.ResourceExhausted))
	})

	It("should leave other errors untouched", func() {
		// given
		internal := status.Error(codes.Internal, "stream failed")
		s := mux.NewErrorRecorderStream(&fakeDeltaStream{recvErr: internal})
		Expect(s.Send(response("Dataplane", 10))).To(Succeed())

		// when
		_, err := s.Recv()

		// then
		Expect(err).To(Equal(internal))
		Expect(s.Err()).To(Equal(internal))
	})

	It("should not record end of stream as an error", func() {
		// given
		s := mux.NewErrorRecorderStream(&fakeDeltaStream{recvErr: io.EOF})

		// when
		_, err := s.Recv()

		// then
		Expect(err).To(Equal(io.EOF))
		Expect(s.Err()).ToNot(HaveOccurred())
	})
})
