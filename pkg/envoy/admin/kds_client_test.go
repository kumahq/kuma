package admin_test

import (
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

	timeout := 1 * time.Second
	streams := service.NewXdsConfigStreams()
	client := admin.NewKDSEnvoyAdminClient(streams, timeout)

	zoneName := "zone-1"
	var stream *mockStream

	BeforeEach(func() {
		stream = &mockStream{
			receivedRequests: make(chan *mesh_proto.XDSConfigRequest, 1),
		}
		streams.ZoneConnected(zoneName, stream)
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
			resp, err := client.ConfigDump(dpRes)
			Expect(err).To(Succeed())
			respCh <- resp
		}()

		// and
		request := <-stream.receivedRequests
		err := streams.ResponseReceived(zoneName, &mesh_proto.XDSConfigResponse{
			RequestId: request.RequestId,
			Result: &mesh_proto.XDSConfigResponse_Config{
				Config: configContent,
			},
		})

		// then
		Expect(err).ToNot(HaveOccurred())
		resp := <-respCh
		Expect(resp).To(Equal(configContent))
	})

	It("should fail when zone is not connected", func() {
		// given
		dpRes := core_mesh.NewDataplaneResource()
		dpRes.SetMeta(&test_model.ResourceMeta{
			Mesh: "default",
			Name: "not-connected.dp-1",
		})

		// when
		_, err := client.ConfigDump(dpRes)

		// then
		Expect(err).To(MatchError("could not send XDSConfigRequest: zone not-connected is not connected"))
	})

	It("should time out after X seconds", func() {
		// given
		dpRes := core_mesh.NewDataplaneResource()
		dpRes.SetMeta(&test_model.ResourceMeta{
			Mesh: "default",
			Name: zoneName + ".dp-1",
		})

		// when
		_, err := client.ConfigDump(dpRes)

		// then
		Expect(err).To(MatchError("timeout. Did not receive the response within 1s"))
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
			_, err := client.ConfigDump(dpRes)
			errCh <- err
		}()

		// and
		request := <-stream.receivedRequests
		err := streams.ResponseReceived(zoneName, &mesh_proto.XDSConfigResponse{
			RequestId: request.RequestId,
			Result: &mesh_proto.XDSConfigResponse_Error{
				Error: "failed",
			},
		})

		// then
		Expect(err).ToNot(HaveOccurred())
		err = <-errCh
		Expect(err).To(MatchError("error response from Zone CP: failed"))
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

func (m *mockStream) Recv() (*mesh_proto.XDSConfigResponse, error) {
	return nil, nil
}

var _ mesh_proto.GlobalKDSService_StreamXDSConfigsServer = &mockStream{}
