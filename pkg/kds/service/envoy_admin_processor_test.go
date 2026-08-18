package service_test

import (
	"context"
	"io"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
	grpc_codes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/v3/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/v3/pkg/core/resources/store"
	"github.com/kumahq/kuma/v3/pkg/kds/service"
	"github.com/kumahq/kuma/v3/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/v3/pkg/test/resources/samples"
)

// staticAdminClient answers every Envoy admin call with a fixed payload, so a
// test can control the size of the response the processor has to send back.
type staticAdminClient struct {
	payload []byte
}

func (c *staticAdminClient) PostQuit(context.Context, *core_mesh.DataplaneResource) error {
	return nil
}

func (c *staticAdminClient) Stats(context.Context, core_model.ResourceWithAddress, mesh_proto.AdminOutputFormat, bool) ([]byte, error) {
	return c.payload, nil
}

func (c *staticAdminClient) Clusters(context.Context, core_model.ResourceWithAddress, mesh_proto.AdminOutputFormat) ([]byte, error) {
	return c.payload, nil
}

func (c *staticAdminClient) ConfigDump(context.Context, core_model.ResourceWithAddress, bool) ([]byte, error) {
	return c.payload, nil
}

// fakeStream is a bidirectional KDS stream driven by the test: requests are fed
// in through the requests channel and responses the processor sends are
// captured on the sent channel. Send rejects oversized messages the same way
// gRPC does, which aborts the stream for good.
type fakeStream[Resp any, Req any] struct {
	grpc.ClientStream
	requests   chan *Req
	sent       chan *Resp
	maxMsgSize int
}

func newFakeStream[Resp any, Req any](maxMsgSize int) *fakeStream[Resp, Req] {
	return &fakeStream[Resp, Req]{
		requests:   make(chan *Req, 1),
		sent:       make(chan *Resp, 1),
		maxMsgSize: maxMsgSize,
	}
}

func (s *fakeStream[Resp, Req]) Context() context.Context { return context.Background() }

func (s *fakeStream[Resp, Req]) Recv() (*Req, error) {
	req, ok := <-s.requests
	if !ok {
		return nil, io.EOF
	}
	return req, nil
}

func (s *fakeStream[Resp, Req]) Send(resp *Resp) error {
	if size := proto.Size(any(resp).(proto.Message)); size > s.maxMsgSize {
		return status.Errorf(grpc_codes.ResourceExhausted, "trying to send message larger than max (%d vs. %d)", size, s.maxMsgSize)
	}
	s.sent <- resp
	return nil
}

var _ = Describe("Envoy Admin Processor", func() {
	// The default maxMsgSize is 10MiB. Keeping it small in tests lets us build
	// an oversized payload without allocating megabytes.
	const maxMsgSize = 1024

	var resManager core_manager.ResourceManager

	newProcessor := func(payloadSize int) service.EnvoyAdminProcessor {
		return service.NewEnvoyAdminProcessor(
			resManager,
			&staticAdminClient{payload: make([]byte, payloadSize)},
			maxMsgSize,
		)
	}

	BeforeEach(func() {
		resManager = core_manager.NewResourceManager(memory.NewStore())
		Expect(resManager.Create(context.Background(), samples.MeshDefault(), core_store.CreateByKey(core_model.DefaultMesh, core_model.NoMesh))).To(Succeed())
		Expect(resManager.Create(context.Background(), samples.DataplaneBackend(), core_store.CreateByKey("backend-1", core_model.DefaultMesh))).To(Succeed())
	})

	Describe("XDS config dump", func() {
		It("should send the config dump when it fits within maxMsgSize", func() {
			// given
			stream := newFakeStream[mesh_proto.XDSConfigResponse, mesh_proto.XDSConfigRequest](maxMsgSize)
			errorCh := make(chan error, 1)
			go newProcessor(64).StartProcessingXDSConfigs(stream, errorCh)

			// when
			stream.requests <- &mesh_proto.XDSConfigRequest{
				RequestId:    "req-1",
				ResourceType: string(core_mesh.DataplaneType),
				ResourceName: "backend-1",
				ResourceMesh: "default",
			}

			// then
			var resp *mesh_proto.XDSConfigResponse
			Eventually(stream.sent).Should(Receive(&resp))
			Expect(resp.RequestId).To(Equal("req-1"))
			Expect(resp.GetError()).To(BeEmpty())
			Expect(resp.GetConfig()).To(HaveLen(64))
			Expect(errorCh).ToNot(Receive())
		})

		It("should reply with an error and keep the stream alive when the config dump exceeds maxMsgSize", func() {
			// given
			stream := newFakeStream[mesh_proto.XDSConfigResponse, mesh_proto.XDSConfigRequest](maxMsgSize)
			errorCh := make(chan error, 1)
			go newProcessor(2*maxMsgSize).StartProcessingXDSConfigs(stream, errorCh)

			// when
			stream.requests <- &mesh_proto.XDSConfigRequest{
				RequestId:    "req-1",
				ResourceType: string(core_mesh.DataplaneType),
				ResourceName: "backend-1",
				ResourceMesh: "default",
			}

			// then the requester gets a meaningful error instead of hanging
			var resp *mesh_proto.XDSConfigResponse
			Eventually(stream.sent).Should(Receive(&resp))
			Expect(resp.RequestId).To(Equal("req-1"))
			Expect(resp.GetConfig()).To(BeEmpty())
			Expect(resp.GetError()).To(ContainSubstring("exceeds the maximum KDS message size of 1024 bytes"))

			// and the stream is not torn down
			Expect(errorCh).ToNot(Receive())

			// and it keeps serving subsequent requests
			stream.requests <- &mesh_proto.XDSConfigRequest{
				RequestId:    "req-2",
				ResourceType: string(core_mesh.DataplaneType),
				ResourceName: "backend-1",
				ResourceMesh: "default",
			}
			Eventually(stream.sent).Should(Receive(&resp))
			Expect(resp.RequestId).To(Equal("req-2"))
		})
	})

	Describe("stats", func() {
		It("should reply with an error and keep the stream alive when stats exceed maxMsgSize", func() {
			// given
			stream := newFakeStream[mesh_proto.StatsResponse, mesh_proto.StatsRequest](maxMsgSize)
			errorCh := make(chan error, 1)
			go newProcessor(2*maxMsgSize).StartProcessingStats(stream, errorCh)

			// when
			stream.requests <- &mesh_proto.StatsRequest{
				RequestId:    "req-1",
				ResourceType: string(core_mesh.DataplaneType),
				ResourceName: "backend-1",
				ResourceMesh: "default",
			}

			// then
			var resp *mesh_proto.StatsResponse
			Eventually(stream.sent).Should(Receive(&resp))
			Expect(resp.GetStats()).To(BeEmpty())
			Expect(resp.GetError()).To(ContainSubstring("exceeds the maximum KDS message size of 1024 bytes"))
			Expect(errorCh).ToNot(Receive())
		})
	})

	Describe("clusters", func() {
		It("should reply with an error and keep the stream alive when clusters exceed maxMsgSize", func() {
			// given
			stream := newFakeStream[mesh_proto.ClustersResponse, mesh_proto.ClustersRequest](maxMsgSize)
			errorCh := make(chan error, 1)
			go newProcessor(2*maxMsgSize).StartProcessingClusters(stream, errorCh)

			// when
			stream.requests <- &mesh_proto.ClustersRequest{
				RequestId:    "req-1",
				ResourceType: string(core_mesh.DataplaneType),
				ResourceName: "backend-1",
				ResourceMesh: "default",
			}

			// then
			var resp *mesh_proto.ClustersResponse
			Eventually(stream.sent).Should(Receive(&resp))
			Expect(resp.GetClusters()).To(BeEmpty())
			Expect(resp.GetError()).To(ContainSubstring("exceeds the maximum KDS message size of 1024 bytes"))
			Expect(errorCh).ToNot(Receive())
		})
	})
})
