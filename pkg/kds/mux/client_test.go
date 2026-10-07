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

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	kuma_cp "github.com/kumahq/kuma/v3/pkg/config/app/kuma-cp"
	"github.com/kumahq/kuma/v3/pkg/core"
	"github.com/kumahq/kuma/v3/pkg/core/resources/store"
	"github.com/kumahq/kuma/v3/pkg/core/runtime/component"
	"github.com/kumahq/kuma/v3/pkg/kds/mux"
	"github.com/kumahq/kuma/v3/pkg/kds/service"
	kds_sync_store "github.com/kumahq/kuma/v3/pkg/kds/store"
	core_metrics "github.com/kumahq/kuma/v3/pkg/metrics"
	"github.com/kumahq/kuma/v3/pkg/plugins/resources/memory"
	kds_setup "github.com/kumahq/kuma/v3/pkg/test/kds/setup"
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
	// Regression test: when a GlobalToZoneSync gRPC stream is terminated by
	// the server, the mux client must trigger a full reconnect via
	// ResilientComponent.
	//
	// Before the fix, startGlobalToZoneSync silently exited on nil error
	// (io.EOF) without sending to errorCh. Because Start() only returns
	// when it receives from errorCh (or stop), the mux client stayed alive
	// — healthchecks and ZoneToGlobal kept working — but GlobalToZone was
	// permanently dead.
	//
	// In production this was triggered by a global CP restart behind a load
	// balancer: the LB closed the GlobalToZone HTTP/2 stream with a clean
	// TCP FIN (io.EOF) instead of a gRPC error.
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

			globalStore := memory.NewStore()
			cfg := kuma_cp.DefaultConfig()
			cfg.Multizone.Zone.Name = "zone-1"
			rt := kds_setup.NewTestRuntime(context.Background(), cfg, globalStore)

			metrics, err := core_metrics.NewMetrics("")
			Expect(err).ToNot(HaveOccurred())

			zoneStore := memory.NewStore()
			resourceSyncer, err := kds_sync_store.NewResourceSyncer(
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
				metrics,
				service.NewEnvoyAdminProcessor(rt.ReadOnlyResourceManager(), rt.EnvoyAdminClient(), rt.Config().Multizone.Zone.KDS.MaxMsgSize),
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
		},
		Entry("server returns nil (io.EOF)", testCase{
			description: "LB or global CP closes stream cleanly",
			errCode:     codes.OK, // handler returns nil
		}),
		Entry("server returns Canceled", testCase{
			description: "global CP explicitly cancels the stream",
			errCode:     codes.Canceled,
		}),
	)
})

// envoyAdminErrServer simulates a Global CP whose Envoy admin rpc fails with
// ResourceExhausted - a message past its receive limit - while the resource
// sync streams stay healthy.
type envoyAdminErrServer struct {
	mesh_proto.UnimplementedKDSSyncServiceServer
	mesh_proto.UnimplementedGlobalKDSServiceServer
	mu                sync.Mutex
	globalToZoneConns int
	xdsConfigsConns   int
}

func (s *envoyAdminErrServer) GlobalToZoneSync(stream mesh_proto.KDSSyncService_GlobalToZoneSyncServer) error {
	s.mu.Lock()
	s.globalToZoneConns++
	s.mu.Unlock()
	<-stream.Context().Done()
	return nil
}

func (s *envoyAdminErrServer) ZoneToGlobalSync(stream mesh_proto.KDSSyncService_ZoneToGlobalSyncServer) error {
	<-stream.Context().Done()
	return nil
}

func (s *envoyAdminErrServer) HealthCheck(_ context.Context, _ *mesh_proto.ZoneHealthCheckRequest) (*mesh_proto.ZoneHealthCheckResponse, error) {
	return &mesh_proto.ZoneHealthCheckResponse{
		Interval: durationpb.New(time.Minute),
	}, nil
}

func (s *envoyAdminErrServer) StreamXDSConfigs(_ mesh_proto.GlobalKDSService_StreamXDSConfigsServer) error {
	s.mu.Lock()
	s.xdsConfigsConns++
	s.mu.Unlock()
	return status.Error(codes.ResourceExhausted, "could not receive a message: grpc: received message after decompression larger than max (20000000 vs. 10485760)")
}

func (s *envoyAdminErrServer) StreamStats(stream mesh_proto.GlobalKDSService_StreamStatsServer) error {
	<-stream.Context().Done()
	return nil
}

func (s *envoyAdminErrServer) StreamClusters(stream mesh_proto.GlobalKDSService_StreamClustersServer) error {
	<-stream.Context().Done()
	return nil
}

func (s *envoyAdminErrServer) connections() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.globalToZoneConns
}

var _ = Describe("Client", func() {
	// An Envoy admin rpc that dies with ResourceExhausted must not take the
	// KDS multiplex down. The global CP rejects the message because it is
	// over its receive limit, and a retry would resend exactly the same
	// message, so restarting would loop with resource sync down every time.
	It("does not restart the mux when an Envoy admin rpc exceeds the message size", func() {
		svc := &envoyAdminErrServer{}
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
		resourceSyncer, err := kds_sync_store.NewResourceSyncer(
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
			metrics,
			service.NewEnvoyAdminProcessor(rt.ReadOnlyResourceManager(), rt.EnvoyAdminClient(), rt.Config().Multizone.Zone.KDS.MaxMsgSize),
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

		// Retrying the oversized message would resend it in a loop, so the stream must be opened exactly once. EXC:FILE011:documents-a-non-obvious-invariant
		Eventually(func() int {
			svc.mu.Lock()
			defer svc.mu.Unlock()
			return svc.xdsConfigsConns
		}, "10s", "100ms").Should(Equal(1))
		Consistently(func() int {
			svc.mu.Lock()
			defer svc.mu.Unlock()
			return svc.xdsConfigsConns
		}, "3s", "100ms").Should(Equal(1))
		Eventually(svc.connections, "10s", "100ms").Should(Equal(1))
		// The failing rpc stays down, everything else keeps running on the
		// same connection.
		Consistently(svc.connections, "2s", "100ms").Should(Equal(1))
	})
})

// diagnosticResetServer simulates a Global CP whose diagnostic rpc stream
// gets reset by a proxy (RST_STREAM), while resource sync stays healthy.
type diagnosticResetServer struct {
	mesh_proto.UnimplementedKDSSyncServiceServer
	mesh_proto.UnimplementedGlobalKDSServiceServer
	reconnectedOnce     sync.Once
	mu                  sync.Mutex
	globalToZoneConns   int
	clustersConns       int
	clustersCallTimes   []time.Time
	clustersReconnected chan struct{}
}

func (s *diagnosticResetServer) GlobalToZoneSync(stream mesh_proto.KDSSyncService_GlobalToZoneSyncServer) error {
	s.mu.Lock()
	s.globalToZoneConns++
	s.mu.Unlock()
	<-stream.Context().Done()
	return nil
}

func (s *diagnosticResetServer) ZoneToGlobalSync(stream mesh_proto.KDSSyncService_ZoneToGlobalSyncServer) error {
	<-stream.Context().Done()
	return nil
}

func (s *diagnosticResetServer) HealthCheck(_ context.Context, _ *mesh_proto.ZoneHealthCheckRequest) (*mesh_proto.ZoneHealthCheckResponse, error) {
	return &mesh_proto.ZoneHealthCheckResponse{
		Interval: durationpb.New(time.Minute),
	}, nil
}

func (s *diagnosticResetServer) StreamXDSConfigs(stream mesh_proto.GlobalKDSService_StreamXDSConfigsServer) error {
	<-stream.Context().Done()
	return nil
}

func (s *diagnosticResetServer) StreamStats(stream mesh_proto.GlobalKDSService_StreamStatsServer) error {
	<-stream.Context().Done()
	return nil
}

func (s *diagnosticResetServer) StreamClusters(stream mesh_proto.GlobalKDSService_StreamClustersServer) error {
	s.mu.Lock()
	count := s.clustersConns
	s.clustersConns++
	s.clustersCallTimes = append(s.clustersCallTimes, time.Now())
	s.mu.Unlock()
	if count == 0 {
		return status.Error(codes.Internal, "stream terminated by RST_STREAM with error code: PROTOCOL_ERROR")
	}
	s.reconnectedOnce.Do(func() { close(s.clustersReconnected) })
	<-stream.Context().Done()
	return nil
}

func (s *diagnosticResetServer) globalToZoneConnections() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.globalToZoneConns
}

type diagnosticUnimplementedServer struct {
	mesh_proto.UnimplementedKDSSyncServiceServer
	mesh_proto.UnimplementedGlobalKDSServiceServer
	mu                sync.Mutex
	globalToZoneConns int
	clustersConns     int
}

func (s *diagnosticUnimplementedServer) GlobalToZoneSync(stream mesh_proto.KDSSyncService_GlobalToZoneSyncServer) error {
	s.mu.Lock()
	s.globalToZoneConns++
	s.mu.Unlock()
	<-stream.Context().Done()
	return nil
}

func (s *diagnosticUnimplementedServer) ZoneToGlobalSync(stream mesh_proto.KDSSyncService_ZoneToGlobalSyncServer) error {
	<-stream.Context().Done()
	return nil
}

func (s *diagnosticUnimplementedServer) HealthCheck(_ context.Context, _ *mesh_proto.ZoneHealthCheckRequest) (*mesh_proto.ZoneHealthCheckResponse, error) {
	return &mesh_proto.ZoneHealthCheckResponse{
		Interval: durationpb.New(time.Minute),
	}, nil
}

func (s *diagnosticUnimplementedServer) StreamXDSConfigs(stream mesh_proto.GlobalKDSService_StreamXDSConfigsServer) error {
	<-stream.Context().Done()
	return nil
}

func (s *diagnosticUnimplementedServer) StreamStats(stream mesh_proto.GlobalKDSService_StreamStatsServer) error {
	<-stream.Context().Done()
	return nil
}

func (s *diagnosticUnimplementedServer) StreamClusters(_ mesh_proto.GlobalKDSService_StreamClustersServer) error {
	s.mu.Lock()
	s.clustersConns++
	s.mu.Unlock()
	return status.Error(codes.Unimplemented, "unknown method StreamClusters")
}

func (s *diagnosticUnimplementedServer) globalToZoneConnections() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.globalToZoneConns
}

func (s *diagnosticUnimplementedServer) clustersConnections() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.clustersConns
}

var _ = Describe("Client", func() {
	// A diagnostic rpc (clusters, stats, XDS configs) that dies with a transport-level error, e.g. RST_STREAM from a proxy idle timeout, must be reopened on its own: before the fix the error terminated the whole KDS multiplex, so a proxy resetting idle streams flapped the zone Online/Offline in a loop while resource sync was healthy. EXC:FILE011:documents-a-non-obvious-invariant
	It("reopens only the failed diagnostic rpc and keeps the mux running", func() {
		svc := &diagnosticResetServer{
			clustersReconnected: make(chan struct{}),
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
		resourceSyncer, err := kds_sync_store.NewResourceSyncer(
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
			metrics,
			service.NewEnvoyAdminProcessor(rt.ReadOnlyResourceManager(), rt.EnvoyAdminClient(), rt.Config().Multizone.Zone.KDS.MaxMsgSize),
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

		// The failed clusters stream is reopened by the client itself, while resource sync is never torn down. EXC:FILE011:documents-a-non-obvious-invariant
		var first, second time.Time
		Eventually(func() bool {
			svc.mu.Lock()
			defer svc.mu.Unlock()
			if len(svc.clustersCallTimes) < 2 {
				return false
			}
			first, second = svc.clustersCallTimes[0], svc.clustersCallTimes[1]
			return true
		}, "10s", "100ms").Should(BeTrue())
		// The reopen honors the initial backoff instead of hot-retrying. EXC:FILE011:documents-a-non-obvious-invariant
		Expect(second.Sub(first)).To(BeNumerically(">=", 900*time.Millisecond))
		Eventually(svc.clustersReconnected, "10s", "100ms").Should(BeClosed())
		Consistently(svc.globalToZoneConnections, "3s", "100ms").Should(Equal(1))
	})

	// An old Global CP answers Unimplemented for a diagnostic rpc: the stream must stay down without retrying and without restarting the KDS multiplex. EXC:FILE011:documents-a-non-obvious-invariant
	It("does not restart the mux or retry an unimplemented diagnostic rpc", func() {
		svc := &diagnosticUnimplementedServer{}
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
		resourceSyncer, err := kds_sync_store.NewResourceSyncer(
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
			metrics,
			service.NewEnvoyAdminProcessor(rt.ReadOnlyResourceManager(), rt.EnvoyAdminClient(), rt.Config().Multizone.Zone.KDS.MaxMsgSize),
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

		Eventually(svc.clustersConnections, "10s", "100ms").Should(Equal(1))
		Consistently(svc.clustersConnections, "5s", "100ms").Should(Equal(1))
		Consistently(svc.globalToZoneConnections, "5s", "100ms").Should(Equal(1))
	})
})
