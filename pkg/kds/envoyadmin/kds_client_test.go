package envoyadmin_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/config"
	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	store_config "github.com/kumahq/kuma/pkg/config/core/resources/store"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_system "github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/kds/envoyadmin"
	"github.com/kumahq/kuma/pkg/kds/service"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("KDS client", func() {
	zoneInsight := func(storeType store_config.StoreType) model.Resource {
		zoneInsight := core_system.NewZoneInsightResource()
		t1, _ := time.Parse(time.RFC3339, "2017-07-17T17:07:47+00:00")
		cfg := kuma_cp.DefaultConfig()
		cfg.Store.Type = storeType
		displayCfg, _ := config.ConfigForDisplay(&cfg)
		zoneInsight.Spec.Subscriptions = []*v1alpha1.KDSSubscription{
			{
				ConnectTime: util_proto.MustTimestampProto(t1),
				Config:      displayCfg,
			},
		}
		return zoneInsight
	}

	Context("Universal", func() {
		rpcs := service.NewEnvoyAdminRPCs()
		resStore := memory.NewStore()
		resManager := manager.NewResourceManager(resStore)
		client := envoyadmin.NewClient(rpcs, resManager)

		zoneName := "zone-1"
		tenantZoneID := service.TenantZoneClientIDFromCtx(context.Background(), zoneName)
		var stream *mockStream

		BeforeEach(func() {
			stream = &mockStream{
				receivedRequests: make(chan *mesh_proto.XDSConfigRequest, 1),
			}
			rpcs.XDSConfigDump.ClientConnected(tenantZoneID.String(), stream)
		})

		It("should execute config dump for legacy", func() {
			// given
			dpRes := core_mesh.NewDataplaneResource()
			dpRes.SetMeta(&test_model.ResourceMeta{
				Mesh: "default",
				Name: "zone-1.dp-1",
				Labels: map[string]string{
					mesh_proto.DisplayName: "dp-1",
				},
			})
			configContent := []byte("config")

			// when
			respCh := make(chan []byte)
			go func() {
				defer GinkgoRecover()
				resp, err := client.ConfigDump(context.Background(), dpRes, false)
				Expect(err).To(Succeed())
				respCh <- resp
			}()

			// and
			request := <-stream.receivedRequests
			Expect(request.ResourceName).To(Equal("dp-1"))

			Eventually(func() error {
				return rpcs.XDSConfigDump.ResponseReceived(tenantZoneID.String(), &mesh_proto.XDSConfigResponse{
					RequestId: request.RequestId,
					Result: &mesh_proto.XDSConfigResponse_Config{
						Config: configContent,
					},
				})
			}, "10s", "100ms").Should(Succeed())

			// then
			Eventually(respCh).Should(Receive(Equal(configContent)))
		})

		It("should execute config dump", func() {
			// given
			zoneInsight := zoneInsight(store_config.PgxStore)
			Expect(resStore.Create(context.Background(), zoneInsight, store.CreateByKey("zone-1", model.NoMesh))).To(Succeed())

			// and
			dpRes := core_mesh.NewDataplaneResource()
			dpRes.SetMeta(&test_model.ResourceMeta{
				Mesh: "default",
				Name: "dp-1-1234",
				Labels: map[string]string{
					mesh_proto.DisplayName: "dp-1",
					mesh_proto.ZoneTag:     "zone-1",
				},
			})
			configContent := []byte("config")

			// when
			respCh := make(chan []byte)
			go func() {
				defer GinkgoRecover()
				resp, err := client.ConfigDump(context.Background(), dpRes, false)
				Expect(err).To(Succeed())
				respCh <- resp
			}()

			// and
			request := <-stream.receivedRequests
			Expect(request.ResourceName).To(Equal("dp-1"))

			Eventually(func() error {
				return rpcs.XDSConfigDump.ResponseReceived(tenantZoneID.String(), &mesh_proto.XDSConfigResponse{
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
			_, err := client.ConfigDump(context.Background(), dpRes, false)

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
			_, err := client.ConfigDump(ctx, dpRes, false)

			// then
			Expect(err).To(MatchError(context.DeadlineExceeded))
		})

		It("should rethrow error from zone CP", func() {
			// given
			dpRes := core_mesh.NewDataplaneResource()
			dpRes.SetMeta(&test_model.ResourceMeta{
				Mesh: "default",
				Name: zoneName + ".dp-1",
				Labels: map[string]string{
					mesh_proto.DisplayName: "dp-1",
				},
			})

			// when
			errCh := make(chan error)
			go func() {
				defer GinkgoRecover()
				_, err := client.ConfigDump(context.Background(), dpRes, false)
				errCh <- err
			}()

			// and
			request := <-stream.receivedRequests
			Expect(request.ResourceName).To(Equal("dp-1"))

			Eventually(func() error {
				return rpcs.XDSConfigDump.ResponseReceived(tenantZoneID.String(), &mesh_proto.XDSConfigResponse{
					RequestId: request.RequestId,
					Result: &mesh_proto.XDSConfigResponse_Error{
						Error: "failed",
					},
				})
			}, "10s", "100ms").Should(Succeed())

			// then
			Eventually(errCh).Should(Receive(MatchError("could not send XDSConfigRequest: failed")))
		})
	})

	Context("Kubernetes", func() {
		streams := service.NewEnvoyAdminRPCs()
		resStore := memory.NewStore()
		resManager := manager.NewResourceManager(resStore)
		client := envoyadmin.NewClient(streams, resManager)

		zoneName := "zone-1"
		tenantZoneID := service.TenantZoneClientIDFromCtx(context.Background(), zoneName)
		var stream *mockStream

		BeforeEach(func() {
			stream = &mockStream{
				receivedRequests: make(chan *mesh_proto.XDSConfigRequest, 1),
			}
			streams.XDSConfigDump.ClientConnected(tenantZoneID.String(), stream)
		})

		It("should execute config dump for legacy endpoint", func() {
			// given
			dpRes := core_mesh.NewDataplaneResource()
			dpRes.SetMeta(&test_model.ResourceMeta{
				Mesh: "default",
				Name: "zone-1.dp-1.my-namespace",
				Labels: map[string]string{
					mesh_proto.DisplayName: "dp-1.my-namespace",
				},
			})
			configContent := []byte("config")

			// when
			respCh := make(chan []byte)
			go func() {
				defer GinkgoRecover()
				resp, err := client.ConfigDump(context.Background(), dpRes, false)
				Expect(err).To(Succeed())
				respCh <- resp
			}()

			// and
			request := <-stream.receivedRequests
			Expect(request.ResourceName).To(Equal("dp-1.my-namespace"))

			Eventually(func() error {
				return streams.XDSConfigDump.ResponseReceived(tenantZoneID.String(), &mesh_proto.XDSConfigResponse{
					RequestId: request.RequestId,
					Result: &mesh_proto.XDSConfigResponse_Config{
						Config: configContent,
					},
				})
			}, "10s", "100ms").Should(Succeed())

			// then
			Eventually(respCh).Should(Receive(Equal(configContent)))
		})

		It("should execute config dump", func() {
			// given
			zoneInsight := zoneInsight(store_config.KubernetesStore)
			Expect(resStore.Create(context.Background(), zoneInsight, store.CreateByKey("zone-1", model.NoMesh))).To(Succeed())

			// and
			dpRes := core_mesh.NewDataplaneResource()
			dpRes.SetMeta(&test_model.ResourceMeta{
				Mesh: "default",
				Name: "dp-1-1234",
				Labels: map[string]string{
					mesh_proto.DisplayName:      "dp-1",
					mesh_proto.KubeNamespaceTag: "kuma-demo",
					mesh_proto.ZoneTag:          "zone-1",
				},
			})
			configContent := []byte("config")

			// when
			respCh := make(chan []byte)
			go func() {
				defer GinkgoRecover()
				resp, err := client.ConfigDump(context.Background(), dpRes, false)
				Expect(err).To(Succeed())
				respCh <- resp
			}()

			// and
			request := <-stream.receivedRequests
			Expect(request.ResourceName).To(Equal("dp-1.kuma-demo"))

			Eventually(func() error {
				return streams.XDSConfigDump.ResponseReceived(tenantZoneID.String(), &mesh_proto.XDSConfigResponse{
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
