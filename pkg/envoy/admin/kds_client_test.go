package admin_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/envoy/admin"
	"github.com/kumahq/kuma/pkg/kds/service"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
)

var _ = Describe("KDS client", func() {
	Context("Universal", func() {
		rpcs := service.NewEnvoyAdminRPCs()
		client := admin.NewKDSEnvoyAdminClient(rpcs, false)

		zoneName := "zone-1"
		var stream *mockStream

		BeforeEach(func() {
			stream = &mockStream{
				receivedRequests: make(chan *mesh_proto.XDSConfigRequest, 1),
			}
			rpcs.XDSConfigDump.ClientConnected(zoneName, stream)
		})

		It("should execute config dump", func() {
			// given
			dpRes := core_mesh.NewDataplaneResource()
			dpRes.SetMeta(&test_model.ResourceMeta{
				Mesh: "default",
				Name: "zone-1.dp-1",
			})
			configContent := []byte("config")

			// when
			respCh := make(chan []byte)
			go func() {
				defer GinkgoRecover()
				resp, err := client.ConfigDump(context.Background(), dpRes)
				Expect(err).To(Succeed())
				respCh <- resp
			}()

			// and
			request := <-stream.receivedRequests
			Expect(request.ResourceName).To(Equal("dp-1"))

			Eventually(func() error {
				return rpcs.XDSConfigDump.ResponseReceived(zoneName, &mesh_proto.XDSConfigResponse{
					RequestId: request.RequestId,
					Result: &mesh_proto.XDSConfigResponse_Config{
						Config: configContent,
					},
				})
			}, "10s", "100ms").Should(Succeed())

			// then
			Eventually(respCh).Should(Receive(Equal(configContent)))
		})

		It("should fail when zone is not connected", func() {
			// given
			dpRes := core_mesh.NewDataplaneResource()
			dpRes.SetMeta(&test_model.ResourceMeta{
				Mesh: "default",
				Name: "not-connected.dp-1",
			})

			// when
			_, err := client.ConfigDump(context.Background(), dpRes)

			// then
			Expect(err).To(MatchError("could not send XDSConfigRequest: client not-connected is not connected"))
		})

		It("should time out after X seconds", func() {
			// given
			dpRes := core_mesh.NewDataplaneResource()
			dpRes.SetMeta(&test_model.ResourceMeta{
				Mesh: "default",
				Name: zoneName + ".dp-1",
			})

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			// when
			_, err := client.ConfigDump(ctx, dpRes)

			// then
			Expect(err).To(MatchError(context.DeadlineExceeded))
		})

		It("should rethrow error from zone CP", func() {
			// given
			dpRes := core_mesh.NewDataplaneResource()
			dpRes.SetMeta(&test_model.ResourceMeta{
				Mesh: "default",
				Name: zoneName + ".dp-1",
			})

			// when
			errCh := make(chan error)
			go func() {
				defer GinkgoRecover()
				_, err := client.ConfigDump(context.Background(), dpRes)
				errCh <- err
			}()

			// and
			request := <-stream.receivedRequests
			Expect(request.ResourceName).To(Equal("dp-1"))

			Eventually(func() error {
				return rpcs.XDSConfigDump.ResponseReceived(zoneName, &mesh_proto.XDSConfigResponse{
					RequestId: request.RequestId,
					Result: &mesh_proto.XDSConfigResponse_Error{
						Error: "failed",
					},
				})
			}, "10s", "100ms").Should(Succeed())

			// then
			Eventually(errCh).Should(Receive(MatchError("error response from Zone CP: failed")))
		})
	})

	Context("Kubernetes", func() {
		streams := service.NewEnvoyAdminRPCs()
		client := admin.NewKDSEnvoyAdminClient(streams, true)

		zoneName := "zone-1"
		var stream *mockStream

		BeforeEach(func() {
			stream = &mockStream{
				receivedRequests: make(chan *mesh_proto.XDSConfigRequest, 1),
			}
			streams.XDSConfigDump.ClientConnected(zoneName, stream)
		})

		It("should execute config dump", func() {
			// given
			dpRes := core_mesh.NewDataplaneResource()
			dpRes.SetMeta(&test_model.ResourceMeta{
				Mesh: "default",
				Name: "zone-1.dp-1.my-namespace",
			})
			configContent := []byte("config")

			// when
			respCh := make(chan []byte)
			go func() {
				defer GinkgoRecover()
				resp, err := client.ConfigDump(context.Background(), dpRes)
				Expect(err).To(Succeed())
				respCh <- resp
			}()

			// and
			request := <-stream.receivedRequests
			Expect(request.ResourceName).To(Equal("dp-1"))

			Eventually(func() error {
				return streams.XDSConfigDump.ResponseReceived(zoneName, &mesh_proto.XDSConfigResponse{
					RequestId: request.RequestId,
					Result: &mesh_proto.XDSConfigResponse_Config{
						Config: configContent,
					},
				})
			}, "10s", "100ms").Should(Succeed())

			// then
			Eventually(respCh).Should(Receive(Equal(configContent)))
		})
	})
})

type mockStream struct {
	receivedRequests  chan *mesh_proto.XDSConfigRequest
	grpc.ServerStream // nil to implement methods
}

func (m *mockStream) Send(request *mesh_proto.XDSConfigRequest) error {
	m.receivedRequests <- request
	return nil
}

func (m *mockStream) SendMsg(request interface{}) error {
	m.receivedRequests <- request.(*mesh_proto.XDSConfigRequest)
	return nil
}

func (m *mockStream) Recv() (*mesh_proto.XDSConfigResponse, error) {
	return nil, nil
}

var _ mesh_proto.GlobalKDSService_StreamXDSConfigsServer = &mockStream{}
