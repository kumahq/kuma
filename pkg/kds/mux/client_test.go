package mux_test

import (
	"context"
	"net"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/kds/mux"
	"github.com/kumahq/kuma/pkg/kds/service"
	kds_client_v2 "github.com/kumahq/kuma/pkg/kds/v2/client"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
)

// reconnectTrackingServer simulates a Global CP that closes the
// GlobalToZoneSync stream on the first connection and then tracks
// whether the zone mux client re-establishes it.
type reconnectTrackingServer struct {
	mesh_proto.UnimplementedKDSSyncServiceServer
	mesh_proto.UnimplementedGlobalKDSServiceServer
	mu                sync.Mutex
	globalToZoneConns int
	reconnectedOnce   sync.Once
	reconnectedCh     chan struct{} // closed on second GlobalToZoneSync call
	firstConnErrCode  codes.Code    // if non-zero, return this status code on first connection instead of nil
}

func (s *reconnectTrackingServer) GlobalToZoneSync(stream mesh_proto.KDSSyncService_GlobalToZoneSyncServer) error {
	s.mu.Lock()
	s.globalToZoneConns++
	count := s.globalToZoneConns
	s.mu.Unlock()

	if count == 1 {
		select {
		case <-time.After(300 * time.Millisecond):
		case <-stream.Context().Done():
			return nil
		}
		if s.firstConnErrCode != codes.OK {
			return status.Error(s.firstConnErrCode, "stream terminated by global")
		}
		return nil
	}

	s.reconnectedOnce.Do(func() { close(s.reconnectedCh) })
	<-stream.Context().Done()
	return nil
}

func (s *reconnectTrackingServer) ZoneToGlobalSync(stream mesh_proto.KDSSyncService_ZoneToGlobalSyncServer) error {
	<-stream.Context().Done()
	return nil
}

func (s *reconnectTrackingServer) HealthCheck(_ context.Context, _ *mesh_proto.ZoneHealthCheckRequest) (*mesh_proto.ZoneHealthCheckResponse, error) {
	return &mesh_proto.ZoneHealthCheckResponse{
		Interval: durationpb.New(time.Minute),
	}, nil
}

func (s *reconnectTrackingServer) StreamXDSConfigs(stream mesh_proto.GlobalKDSService_StreamXDSConfigsServer) error {
	<-stream.Context().Done()
	return nil
}

func (s *reconnectTrackingServer) StreamStats(stream mesh_proto.GlobalKDSService_StreamStatsServer) error {
	<-stream.Context().Done()
	return nil
}

func (s *reconnectTrackingServer) StreamClusters(stream mesh_proto.GlobalKDSService_StreamClustersServer) error {
	<-stream.Context().Done()
	return nil
}

var _ = Describe("Client", func() {
	type testCase struct {
		description string
		errCode     codes.Code
	}

	DescribeTable("reconnects when globalToZone stream is terminated by server",
		func(tc testCase) {
			svc := &reconnectTrackingServer{
				reconnectedCh:    make(chan struct{}),
				firstConnErrCode: tc.errCode,
			}
			lis, err := net.Listen("tcp", "127.0.0.1:0")
			Expect(err).ToNot(HaveOccurred())
			defer lis.Close()

			grpcSrv := grpc.NewServer()
			mesh_proto.RegisterKDSSyncServiceServer(grpcSrv, svc)
			mesh_proto.RegisterGlobalKDSServiceServer(grpcSrv, svc)
			go func() { _ = grpcSrv.Serve(lis) }()
			defer grpcSrv.Stop()

			cfg := kuma_cp.DefaultConfig()
			cfg.Multizone.Zone.Name = "zone-1"

			metrics, err := core_metrics.NewMetrics("")
			Expect(err).ToNot(HaveOccurred())

			typesSentByGlobal := registry.Global().ObjectTypes(model.HasKDSFlag(model.GlobalToZoneSelector))
			runtimeInfo := core_runtime.NewRuntimeInfo("zone-1", config_core.Zone)

			globalToZoneCb := mux.OnGlobalToZoneSyncStartedFunc(func(stream mesh_proto.KDSSyncService_GlobalToZoneSyncClient, errChan chan error) {
				kdsStream := kds_client_v2.NewDeltaKDSStream(stream, "zone-1", runtimeInfo, "", len(typesSentByGlobal))
				syncClient := kds_client_v2.NewKDSSyncClient(
					core.Log.WithName("test-g2z"),
					typesSentByGlobal,
					kdsStream,
					nil,
					0,
				)
				go func() {
					err := syncClient.Receive()
					if err != nil && !errors.Is(err, context.Canceled) {
						select {
						case errChan <- errors.Wrap(err, "GlobalToZoneSyncClient finished with an error"):
						default:
						}
					}
				}()
			})

			zoneToGlobalCb := mux.OnZoneToGlobalSyncStartedFunc(func(stream mesh_proto.KDSSyncService_ZoneToGlobalSyncClient, errChan chan error) {
				go func() {
					<-stream.Context().Done()
				}()
			})

			muxClient := mux.NewClient(
				context.Background(),
				"grpc://"+lis.Addr().String(),
				"zone-1",
				globalToZoneCb,
				zoneToGlobalCb,
				*cfg.Multizone.Zone.KDS,
				cfg.Experimental,
				metrics,
				service.NewEnvoyAdminProcessor(nil, nil),
			)

			resilient := component.NewResilientComponent(
				core.Log.WithName("test-resilient"),
				muxClient,
				1*time.Millisecond,
				10*time.Millisecond,
			)

			stop := make(chan struct{})
			go func() { _ = resilient.Start(stop) }()
			defer close(stop)

			Eventually(svc.reconnectedCh, "10s", "100ms").Should(BeClosed())
		},
		Entry("server returns nil (io.EOF)", testCase{
			description: "LB or global CP closes stream cleanly",
			errCode:     codes.OK,
		}),
		Entry("server returns Canceled", testCase{
			description: "global CP explicitly cancels the stream",
			errCode:     codes.Canceled,
		}),
	)
})
