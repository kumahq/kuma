package server_test

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/grpc"

	envoy "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_auth "github.com/envoyproxy/go-control-plane/envoy/api/v2/auth"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/Kong/kuma/pkg/sds/server"

	test_logr "github.com/Kong/kuma/pkg/test/logr"
)

var _ = Describe("Sds", func() {

	var stream *mockStream

	BeforeEach(func() {
		stream = newMockStream()
	})

	It("should support streams without a single SDS request", func(done Done) {
		// given
		sds := NewServer(nil, nil, test_logr.NewTestLogger(GinkgoT()))

		// when
		errCh := make(chan error)
		go func() {
			defer GinkgoRecover()

			errCh <- sds.StreamSecrets(stream)
		}()

		// when
		close(stream.in)
		// then
		err := <-errCh
		Expect(err).ToNot(HaveOccurred())

		// finally
		close(done)
	})

	It("should support SDS requests with an empty list of resource names", func(done Done) {
		// given
		sds := NewServer(nil, nil, test_logr.NewTestLogger(GinkgoT()))

		// when
		errCh := make(chan error)
		go func() {
			defer GinkgoRecover()

			errCh <- sds.StreamSecrets(stream)
		}()

		// when
		stream.in <- &envoy.DiscoveryRequest{
			// In practice, such requests can be observed when Envoy is removing
			// Listeners and Clusters with TLS configuration that refers to SDS.
			ResourceNames: []string{},
		}
		// then
		select {
		case <-stream.out:
			Fail("SDS server should not respond to SDS requests with an empty list of resource names")
		case <-time.After(200 * time.Millisecond):
			// this is expected behaviour since SDS server should ignore
			// SDS requests with an empty list of resource names
		}

		// when
		close(stream.in)
		// then
		err := <-errCh
		Expect(err).ToNot(HaveOccurred())

		// finally
		close(done)
	})

	It("should support valid SDS requests", func(done Done) {
		// given
		handler := SecretDiscoveryHandlerFunc(func(ctx context.Context, req envoy.DiscoveryRequest) (*envoy_auth.Secret, error) {
			return &envoy_auth.Secret{}, nil
		})
		sds := NewServer(handler, nil, test_logr.NewTestLogger(GinkgoT()))

		// when
		errCh := make(chan error)
		go func() {
			defer GinkgoRecover()

			errCh <- sds.StreamSecrets(stream)
		}()

		// when
		stream.in <- &envoy.DiscoveryRequest{
			ResourceNames: []string{"mesh_ca"},
		}
		// then
		resp := <-stream.out
		Expect(resp).ToNot(BeNil())

		// when
		close(stream.in)
		// then
		err := <-errCh
		Expect(err).ToNot(HaveOccurred())

		// finally
		close(done)
	})
})

func newMockStream() *mockStream {
	return &mockStream{
		ctx: context.Background(),
		in:  make(chan *envoy.DiscoveryRequest, 3),
		out: make(chan *envoy.DiscoveryResponse, 3),
	}
}

type mockStream struct {
	ctx context.Context
	in  chan *envoy.DiscoveryRequest
	out chan *envoy.DiscoveryResponse
	grpc.ServerStream
}

func (s *mockStream) Context() context.Context {
	return s.ctx
}

func (s *mockStream) Recv() (*envoy.DiscoveryRequest, error) {
	req, more := <-s.in
	if !more {
		return nil, errors.New("gRPC stream closed by client")
	}
	return req, nil
}

func (s *mockStream) Send(resp *envoy.DiscoveryResponse) error {
	s.out <- resp
	return nil
}
