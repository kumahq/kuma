package mux_test

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/envoyproxy/go-control-plane/pkg/server/delta/v3"
	stream_v3 "github.com/envoyproxy/go-control-plane/pkg/server/stream/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	kuma_cp "github.com/kumahq/kuma/v2/pkg/config/app/kuma-cp"
	"github.com/kumahq/kuma/v2/pkg/core"
	"github.com/kumahq/kuma/v2/pkg/core/resources/store"
	"github.com/kumahq/kuma/v2/pkg/core/runtime/component"
	"github.com/kumahq/kuma/v2/pkg/kds/mux"
	"github.com/kumahq/kuma/v2/pkg/kds/service"
	sync_store_v2 "github.com/kumahq/kuma/v2/pkg/kds/v2/store"
	core_metrics "github.com/kumahq/kuma/v2/pkg/metrics"
	"github.com/kumahq/kuma/v2/pkg/plugins/resources/memory"
	kds_setup "github.com/kumahq/kuma/v2/pkg/test/kds/setup"
)

// reconnectTrackingServer simulates a Global CP that closes the
// GlobalToZoneSync stream cleanly on the first connection and then
// tracks whether the zone mux client re-establishes it.
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
		// First connection: close the stream after a short delay.
		// Returning nil closes the stream cleanly (zone gets io.EOF).
		// Returning a status error simulates the global CP explicitly
		// canceling the stream.
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

	// Second (or later) connection: zone successfully reconnected.
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

// testZoneDeltaServer is a no-op delta server for the zone-to-global
// direction. It blocks until the stream context is canceled.
type testZoneDeltaServer struct{}

func (s *testZoneDeltaServer) DeltaStreamHandler(str stream_v3.DeltaStream, _ string) error {
	<-str.Context().Done()
	return nil
}

var _ delta.Server = &testZoneDeltaServer{}

var _ = Describe("Client", func() {
	// Regression test: when a GlobalToZoneSync gRPC stream is closed cleanly
	// by the server (the zone receives io.EOF), the mux client must treat
	// this as an error and trigger a full reconnect via ResilientComponent.
	//
	// Before the fix, startGlobalToZoneSync silently exited on nil error
	// without sending to errorCh. Because Start() only returns when it
	// receives from errorCh (or stop), the mux client stayed alive —
	// healthchecks and ZoneToGlobal kept working — but GlobalToZone was
	// permanently dead. No new resources from the global CP would ever
	// reach the zone.
	//
	// In production this was triggered by a global CP restart behind a
	// load balancer: the LB closed the GlobalToZone HTTP/2 stream with a
	// clean TCP FIN (io.EOF) instead of a gRPC error.
	It("reconnects when globalToZone stream is closed by server (io.EOF)", func() {
		svc := &reconnectTrackingServer{
			reconnectedCh: make(chan struct{}),
		}
		lis, err := net.Listen("tcp", "127.0.0.1:0")
		Expect(err).ToNot(HaveOccurred())
		defer lis.Close()

		grpcSrv := grpc.NewServer()
		mesh_proto.RegisterKDSSyncServiceServer(grpcSrv, svc)
		mesh_proto.RegisterGlobalKDSServiceServer(grpcSrv, svc)
		go func() { _ = grpcSrv.Serve(lis) }()
		defer grpcSrv.Stop()

		globalStore := memory.NewStore()
		cfg := kuma_cp.DefaultConfig()
		cfg.Multizone.Zone.Name = "zone-1"
		rt := kds_setup.NewTestRuntime(context.Background(), cfg, globalStore)

		metrics, err := core_metrics.NewMetrics("")
		Expect(err).ToNot(HaveOccurred())

		zoneStore := memory.NewStore()
		resourceSyncer, err := sync_store_v2.NewResourceSyncer(
			core.Log.WithName("syncer"),
			zoneStore,
			store.NoTransactions{},
			metrics,
			context.Background(),
		)
		Expect(err).ToNot(HaveOccurred())

		muxClient := mux.NewClient(
			context.Background(),
			"grpc://"+lis.Addr().String(),
			"zone-1",
			*rt.Config().Multizone.Zone.KDS,
			rt.Config().Experimental,
			metrics,
			service.NewEnvoyAdminProcessor(rt.ReadOnlyResourceManager(), rt.EnvoyAdminClient()),
			resourceSyncer,
			rt,
			&testZoneDeltaServer{},
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
	})

	It("reconnects when globalToZone stream is canceled by server", func() {
		svc := &reconnectTrackingServer{
			reconnectedCh:    make(chan struct{}),
			firstConnErrCode: codes.Canceled,
		}
		lis, err := net.Listen("tcp", "127.0.0.1:0")
		Expect(err).ToNot(HaveOccurred())
		defer lis.Close()

		grpcSrv := grpc.NewServer()
		mesh_proto.RegisterKDSSyncServiceServer(grpcSrv, svc)
		mesh_proto.RegisterGlobalKDSServiceServer(grpcSrv, svc)
		go func() { _ = grpcSrv.Serve(lis) }()
		defer grpcSrv.Stop()

		globalStore := memory.NewStore()
		cfg := kuma_cp.DefaultConfig()
		cfg.Multizone.Zone.Name = "zone-1"
		rt := kds_setup.NewTestRuntime(context.Background(), cfg, globalStore)

		metrics, err := core_metrics.NewMetrics("")
		Expect(err).ToNot(HaveOccurred())

		zoneStore := memory.NewStore()
		resourceSyncer, err := sync_store_v2.NewResourceSyncer(
			core.Log.WithName("syncer"),
			zoneStore,
			store.NoTransactions{},
			metrics,
			context.Background(),
		)
		Expect(err).ToNot(HaveOccurred())

		muxClient := mux.NewClient(
			context.Background(),
			"grpc://"+lis.Addr().String(),
			"zone-1",
			*rt.Config().Multizone.Zone.KDS,
			rt.Config().Experimental,
			metrics,
			service.NewEnvoyAdminProcessor(rt.ReadOnlyResourceManager(), rt.EnvoyAdminClient()),
			resourceSyncer,
			rt,
			&testZoneDeltaServer{},
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
	})
})
