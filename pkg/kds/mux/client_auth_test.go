package mux_test

import (
	"context"
	"net"
	"sync"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	kuma_cp "github.com/kumahq/kuma/v3/pkg/config/app/kuma-cp"
	"github.com/kumahq/kuma/v3/pkg/core"
	"github.com/kumahq/kuma/v3/pkg/core/resources/store"
	kds_auth "github.com/kumahq/kuma/v3/pkg/kds/auth"
	"github.com/kumahq/kuma/v3/pkg/kds/mux"
	"github.com/kumahq/kuma/v3/pkg/kds/service"
	kds_sync_store "github.com/kumahq/kuma/v3/pkg/kds/store"
	core_metrics "github.com/kumahq/kuma/v3/pkg/metrics"
	"github.com/kumahq/kuma/v3/pkg/plugins/resources/memory"
	kds_setup "github.com/kumahq/kuma/v3/pkg/test/kds/setup"
)

type staticTokenAuthenticator struct {
	zone  string
	token string
}

func (s staticTokenAuthenticator) Authenticate(_ context.Context, md metadata.MD, zone string) error {
	if tokens := md.Get(kds_auth.TokenHeader); zone != s.zone || len(tokens) != 1 || tokens[0] != s.token {
		return status.Error(codes.Unauthenticated, "invalid token")
	}
	return nil
}

type authenticatedMethods struct {
	sync.Mutex
	methods map[string]struct{}
}

func (a *authenticatedMethods) add(method string) {
	a.Lock()
	defer a.Unlock()
	a.methods[method] = struct{}{}
}

func (a *authenticatedMethods) get() []string {
	a.Lock()
	defer a.Unlock()
	var methods []string
	for method := range a.methods {
		methods = append(methods, method)
	}
	return methods
}

var _ = Describe("Client authentication", func() {
	var authenticated *authenticatedMethods
	var globalAddress string

	BeforeEach(func() {
		authenticated = &authenticatedMethods{methods: map[string]struct{}{}}
		authenticator := staticTokenAuthenticator{zone: "zone-1", token: "zone-1-token"}

		lis, err := net.Listen("tcp", "127.0.0.1:0")
		Expect(err).ToNot(HaveOccurred())
		globalAddress = "grpc://" + lis.Addr().String()

		svc := &reconnectTrackingServer{reconnectedCh: make(chan struct{})}
		grpcSrv := grpc.NewServer(
			grpc.ChainStreamInterceptor(
				kds_auth.StreamServerInterceptor(authenticator),
				func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
					authenticated.add(info.FullMethod)
					return handler(srv, ss)
				},
			),
			grpc.ChainUnaryInterceptor(
				kds_auth.UnaryServerInterceptor(authenticator),
				func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
					authenticated.add(info.FullMethod)
					return handler(ctx, req)
				},
			),
		)
		mesh_proto.RegisterKDSSyncServiceServer(grpcSrv, svc)
		mesh_proto.RegisterGlobalKDSServiceServer(grpcSrv, svc)
		go func() { _ = grpcSrv.Serve(lis) }()
		DeferCleanup(grpcSrv.Stop)
	})

	startClient := func(token string) <-chan error {
		cfg := kuma_cp.DefaultConfig()
		cfg.Multizone.Zone.Name = "zone-1"
		rt := kds_setup.NewTestRuntime(context.Background(), cfg, memory.NewStore())
		if token != "" {
			rt.KDSContext().ZoneCredentials = kds_auth.NewTokenCredentials(func() (string, error) { return token, nil }, false)
		}

		metrics, err := core_metrics.NewMetrics("")
		Expect(err).ToNot(HaveOccurred())
		resourceSyncer, err := kds_sync_store.NewResourceSyncer(core.Log.WithName("syncer"), memory.NewStore(), store.NoTransactions{}, metrics, context.Background())
		Expect(err).ToNot(HaveOccurred())

		muxClient := mux.NewClient(
			context.Background(),
			globalAddress,
			"zone-1",
			*rt.Config().Multizone.Zone.KDS,
			metrics,
			service.NewEnvoyAdminProcessor(rt.ReadOnlyResourceManager(), rt.EnvoyAdminClient(), rt.Config().Multizone.Zone.KDS.MaxMsgSize),
			resourceSyncer,
			rt,
			&testZoneDeltaServer{},
		)

		stop := make(chan struct{})
		DeferCleanup(func() { close(stop) })
		errCh := make(chan error, 1)
		go func() { errCh <- muxClient.Start(stop) }()
		return errCh
	}

	It("should send the token on every KDS RPC", func() {
		startClient("zone-1-token")

		Eventually(authenticated.get, "10s", "100ms").Should(ConsistOf(
			mesh_proto.GlobalKDSService_HealthCheck_FullMethodName,
			mesh_proto.GlobalKDSService_StreamXDSConfigs_FullMethodName,
			mesh_proto.GlobalKDSService_StreamStats_FullMethodName,
			mesh_proto.GlobalKDSService_StreamClusters_FullMethodName,
			mesh_proto.KDSSyncService_GlobalToZoneSync_FullMethodName,
			mesh_proto.KDSSyncService_ZoneToGlobalSync_FullMethodName,
		))
	})

	DescribeTable("should be rejected",
		func(token string) {
			errCh := startClient(token)

			// the first error depends on which RPC loses the race, it is not always the Unauthenticated status
			Eventually(errCh, "10s", "100ms").Should(Receive(HaveOccurred()))
			Expect(authenticated.get()).To(BeEmpty())
		},
		Entry("with an invalid token", "zone-2-token"),
		Entry("without a token", ""),
	)
})
