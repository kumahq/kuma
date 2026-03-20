package otelreceiver

import (
	"context"
	"fmt"
	"net"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	logspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	tracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	commonv1 "go.opentelemetry.io/proto/otlp/common/v1"
	logsv1 "go.opentelemetry.io/proto/otlp/logs/v1"
	tracev1 "go.opentelemetry.io/proto/otlp/trace/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"

	"github.com/kumahq/kuma/v2/app/kuma-dp/pkg/dataplane/otelenv"
	motb_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshopentelemetrybackend/api/v1alpha1"
	core_xds "github.com/kumahq/kuma/v2/pkg/core/xds"
)

func testSocketPath(prefix string) string {
	return filepath.Join(GinkgoT().TempDir(), fmt.Sprintf("%s.sock", prefix))
}

func traceBackend(socketPath, endpoint string) core_xds.OtelPipeBackend {
	return core_xds.OtelPipeBackend{
		SocketPath: socketPath,
		Endpoint:   endpoint,
		EnvPolicy: &core_xds.OtelResolvedEnvPolicy{
			Mode:                 motb_api.EnvModeOptional,
			Precedence:           motb_api.EnvPrecedenceExplicitFirst,
			AllowSignalOverrides: true,
		},
		Traces: &core_xds.OtelSignalRuntimePlan{
			Enabled: true,
		},
	}
}

func dialUnixGRPC(socketPath string) (*grpc.ClientConn, error) {
	return grpc.NewClient(
		"passthrough:///otel-local",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return net.Dial("unix", socketPath)
		}),
	)
}

var _ = Describe("Manager reconcile", func() {
	It("should add and remove backends", func() {
		manager := NewManager(otelenv.Config{})
		DeferCleanup(manager.stopAll)

		socketA := testSocketPath("a")
		socketB := testSocketPath("b")
		backendA := traceBackend(socketA, "collector-a:4317")
		backendB := traceBackend(socketB, "collector-b:4317")

		Expect(manager.reconcile([]core_xds.OtelPipeBackend{backendA, backendB})).To(Succeed())
		Expect(manager.running).To(HaveLen(2))
		Expect(manager.running).To(HaveKey(socketA))
		Expect(manager.running).To(HaveKey(socketB))

		Expect(manager.reconcile([]core_xds.OtelPipeBackend{backendB})).To(Succeed())
		Expect(manager.running).To(HaveLen(1))
		Expect(manager.running).NotTo(HaveKey(socketA))
		Expect(manager.running).To(HaveKey(socketB))
	})

	It("should restart backend when resolved runtime changes", func() {
		manager := NewManager(otelenv.Config{
			Shared: otelenv.Layer{
				Endpoint: otelenv.FieldValue{Present: true, Value: "env-collector:4317"},
			},
		})
		DeferCleanup(manager.stopAll)

		socketPath := testSocketPath("collector")
		initialBackend := traceBackend(socketPath, "collector:4317")
		Expect(manager.reconcile([]core_xds.OtelPipeBackend{initialBackend})).To(Succeed())

		originalServer := manager.running[socketPath].server

		Expect(manager.reconcile([]core_xds.OtelPipeBackend{initialBackend})).To(Succeed())
		Expect(manager.running[socketPath].server).To(BeIdenticalTo(originalServer))

		updatedBackend := initialBackend
		envPolicyCopy := *updatedBackend.EnvPolicy
		envPolicyCopy.Precedence = motb_api.EnvPrecedenceEnvFirst
		updatedBackend.EnvPolicy = &envPolicyCopy

		Expect(manager.reconcile([]core_xds.OtelPipeBackend{updatedBackend})).To(Succeed())
		Expect(manager.running[socketPath].server).NotTo(BeIdenticalTo(originalServer))
		Expect(sameBackendRuntime(manager.running[socketPath].runtime, manager.envConfig.ResolveBackend(updatedBackend))).To(BeTrue())
	})

	It("should route trace and log traffic to different exporters when the runtime diverges", func() {
		traceCollector := &testTraceCollector{}
		traceServer := grpc.NewServer()
		tracepb.RegisterTraceServiceServer(traceServer, traceCollector)

		traceListener, err := net.Listen("tcp", "127.0.0.1:0")
		Expect(err).NotTo(HaveOccurred())
		defer traceListener.Close()

		go func() {
			defer GinkgoRecover()
			_ = traceServer.Serve(traceListener)
		}()
		defer traceServer.Stop()

		logsCollector := &testLogsCollector{}
		logsServer := grpc.NewServer()
		logspb.RegisterLogsServiceServer(logsServer, logsCollector)

		logsListener, err := net.Listen("tcp", "127.0.0.1:0")
		Expect(err).NotTo(HaveOccurred())
		defer logsListener.Close()

		go func() {
			defer GinkgoRecover()
			_ = logsServer.Serve(logsListener)
		}()
		defer logsServer.Stop()

		manager := NewManager(otelenv.Config{
			Logs: otelenv.Layer{
				Endpoint: otelenv.FieldValue{Present: true, Value: logsListener.Addr().String()},
			},
		})
		DeferCleanup(manager.stopAll)

		socketPath := testSocketPath("diverged")
		backend := traceBackend(socketPath, traceListener.Addr().String())
		backend.EnvPolicy.Precedence = motb_api.EnvPrecedenceEnvFirst
		backend.Logs = &core_xds.OtelSignalRuntimePlan{
			Enabled:         true,
			EnvInputPresent: true,
			OverrideKinds:   []string{"endpoint"},
		}

		Expect(manager.reconcile([]core_xds.OtelPipeBackend{backend})).To(Succeed())

		conn, err := dialUnixGRPC(socketPath)
		Expect(err).NotTo(HaveOccurred())
		defer func() {
			_ = conn.Close()
		}()

		traceClient := tracepb.NewTraceServiceClient(conn)
		logsClient := logspb.NewLogsServiceClient(conn)

		traceReq := &tracepb.ExportTraceServiceRequest{
			ResourceSpans: []*tracev1.ResourceSpans{
				{ScopeSpans: []*tracev1.ScopeSpans{
					{Spans: []*tracev1.Span{{Name: "trace-via-runtime"}}},
				}},
			},
		}
		_, err = traceClient.Export(context.Background(), traceReq)
		Expect(err).NotTo(HaveOccurred())

		logsReq := &logspb.ExportLogsServiceRequest{
			ResourceLogs: []*logsv1.ResourceLogs{
				{ScopeLogs: []*logsv1.ScopeLogs{
					{LogRecords: []*logsv1.LogRecord{
						{Body: &commonv1.AnyValue{Value: &commonv1.AnyValue_StringValue{StringValue: "logs-via-runtime"}}},
					}},
				}},
			},
		}
		_, err = logsClient.Export(context.Background(), logsReq)
		Expect(err).NotTo(HaveOccurred())

		Expect(proto.Equal(traceReq, traceCollector.latestRequest())).To(BeTrue())
		Expect(proto.Equal(logsReq, logsCollector.latestRequest())).To(BeTrue())
	})

	It("should return error for invalid socket path", func() {
		manager := NewManager(otelenv.Config{})
		DeferCleanup(manager.stopAll)

		backend := traceBackend("/nonexistent/dir/broken.sock", "collector:4317")

		err := manager.reconcile([]core_xds.OtelPipeBackend{backend})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("failed to start OTel receiver"))
		Expect(manager.running).To(BeEmpty())
	})
})

var _ = Describe("sameBackendRuntime", func() {
	It("should return true for identical runtimes", func() {
		runtime := otelenv.BackendRuntime{
			Traces: otelenv.SignalRuntime{
				Enabled:  true,
				HTTPPath: "/v1/traces",
				Transport: otelenv.ExporterTransport{
					Protocol: core_xds.OtelProtocolHTTPProtobuf,
					Endpoint: "collector:4318",
					Headers:  map[string]string{"authorization": "token"},
				},
			},
		}

		Expect(sameBackendRuntime(runtime, runtime)).To(BeTrue())
	})

	It("should return false when signal transport differs", func() {
		a := otelenv.BackendRuntime{
			Traces: otelenv.SignalRuntime{
				Enabled: true,
				Transport: otelenv.ExporterTransport{
					Protocol: core_xds.OtelProtocolGRPC,
					Endpoint: "collector-a:4317",
				},
			},
		}
		b := otelenv.BackendRuntime{
			Traces: otelenv.SignalRuntime{
				Enabled: true,
				Transport: otelenv.ExporterTransport{
					Protocol: core_xds.OtelProtocolGRPC,
					Endpoint: "collector-b:4317",
				},
			},
		}

		Expect(sameBackendRuntime(a, b)).To(BeFalse())
	})
})

var _ = Describe("SetOnReconcile callback", func() {
	It("should call onReconcile after reconciling backends", func() {
		var callbackBackends []core_xds.OtelPipeBackend
		manager := NewManager(otelenv.Config{})
		manager.SetOnReconcile(func(backends []core_xds.OtelPipeBackend) {
			callbackBackends = backends
		})
		DeferCleanup(manager.stopAll)

		socketA := testSocketPath("callback-a")
		backendA := traceBackend(socketA, "collector-a:4317")
		backendA.Metrics = &core_xds.OtelSignalRuntimePlan{
			Enabled:         true,
			RefreshInterval: "1m0s",
		}

		Expect(manager.reconcile([]core_xds.OtelPipeBackend{backendA})).To(Succeed())
		Expect(callbackBackends).To(HaveLen(1))
		Expect(callbackBackends[0].SocketPath).To(Equal(socketA))
		Expect(callbackBackends[0].Metrics.Enabled).To(BeTrue())
	})

	It("should pass all backends including metrics-disabled ones", func() {
		var callbackBackends []core_xds.OtelPipeBackend
		manager := NewManager(otelenv.Config{})
		manager.SetOnReconcile(func(backends []core_xds.OtelPipeBackend) {
			callbackBackends = backends
		})
		DeferCleanup(manager.stopAll)

		socketA := testSocketPath("all-a")
		socketB := testSocketPath("all-b")
		backendA := traceBackend(socketA, "collector-a:4317")
		backendB := traceBackend(socketB, "collector-b:4317")
		// backendB has no metrics signal

		Expect(manager.reconcile([]core_xds.OtelPipeBackend{backendA, backendB})).To(Succeed())
		Expect(callbackBackends).To(HaveLen(2))
	})
})

var _ = Describe("testSocketPath helper", func() {
	It("should create unique paths in temp dir", func() {
		p := testSocketPath("foo")
		Expect(p).To(ContainSubstring("foo.sock"))
		Expect(strings.HasPrefix(p, "/")).To(BeTrue())
	})
})
