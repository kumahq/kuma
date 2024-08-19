package probes_test

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/bakito/go-log-logr-adapter/adapter"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	err_pkg "github.com/pkg/errors"
	"google.golang.org/grpc"
	grpchealth "google.golang.org/grpc/health/grpc_health_v1"

	"github.com/kumahq/kuma/app/kuma-dp/pkg/dataplane/probes"
	kuma_probes "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/probes"
	kuma_tls "github.com/kumahq/kuma/pkg/tls"
	kuma_srv "github.com/kumahq/kuma/pkg/util/http/server"
)

const podIP = "127.0.0.1"

var _ = Describe("Virtual Probes", func() {
	probes.LocalAddrIPv4 = &net.TCPAddr{IP: net.ParseIP("127.0.0.1")}
	probes.LocalAddrIPv6 = &net.TCPAddr{IP: net.ParseIP("::1")}

	Describe("Virtual Probe Listener", func() {
		It("should start and stop the listener", func() {
			stopCh := make(chan struct{})
			errCh := make(chan error)
			prober := probes.NewProber(podIP, 9001)

			go func() {
				errCh <- prober.Start(stopCh)
			}()

			time.Sleep(2 * time.Second)
			close(stopCh)

			var err error
			select {
			case err = <-errCh:
			default:
			}

			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("HTTP Probes", func() {
		var stopCh chan struct{}
		var errCh chan error
		vProbePort := uint32(9002)

		BeforeAll(func() {
			stopCh = make(chan struct{})
			errCh = make(chan error)
			prober := probes.NewProber(podIP, vProbePort)
			mockApp := &mockApplication{
				HTTP: &mockHTTPServerConfig{
					ListenPort:        8080,
					Path:              "/healthz",
					ReturnStatusCode:  200,
					FailWhenHeader:    "x-custom-header-triggers-failure",
					ExecutionDuration: time.Duration(3) * time.Second,
				},
			}

			go func() {
				errCh <- prober.Start(stopCh)
			}()
			go func() {
				errCh <- mockApp.Start(stopCh)
			}()
			// wait a short period of time for the servers to be ready
			<-time.After(500 * time.Millisecond)
		})
		AfterAll(func() {
			close(stopCh)

			var err error
			select {
			case err = <-errCh:
			default:
			}

			Expect(err).ToNot(HaveOccurred())
		})

		It("should probe HTTP upstream when it's healthy", func() {
			probeReq, err := http.NewRequest("GET", virtualProbesURL("/8080/healthz", vProbePort), nil)
			Expect(err).ToNot(HaveOccurred())
			probeReq.Header.Set(kuma_probes.HeaderNameTimeout, "5")

			response, err := http.DefaultClient.Do(probeReq)

			Expect(err).ToNot(HaveOccurred())
			Expect(response.StatusCode).To(Equal(200))
		})

		It("should probe HTTP upstream when the port is not listening", func() {
			probeReq, err := http.NewRequest("GET", virtualProbesURL("/8081/healthz", vProbePort), nil)
			Expect(err).ToNot(HaveOccurred())

			response, err := http.DefaultClient.Do(probeReq)

			Expect(err).ToNot(HaveOccurred())
			Expect(response.StatusCode).To(Equal(503))
		})

		It("should probe HTTP upstream when the application reports a failure and return application status code", func() {
			// given a header set to trigger a failure
			probeReq, err := http.NewRequest("GET", virtualProbesURL("/8080/healthz", vProbePort), nil)
			Expect(err).ToNot(HaveOccurred())
			probeReq.Header.Set("x-custom-header-triggers-failure", "present")
			probeReq.Header.Set(kuma_probes.HeaderNameTimeout, "5") // 5s is longer than the execution duration	(3s)

			response, err := http.DefaultClient.Do(probeReq)

			Expect(err).ToNot(HaveOccurred())
			Expect(response.StatusCode).To(Equal(400))
		})

		It("should probe HTTP upstream when path does not match", func() {
			// given a header set to trigger a failure
			probeReq, err := http.NewRequest("GET", virtualProbesURL("/8080/bad-path", vProbePort), nil)
			Expect(err).ToNot(HaveOccurred())
			probeReq.Header.Set(kuma_probes.HeaderNameTimeout, "5") // 5s is longer than the execution duration	(3s)

			response, err := http.DefaultClient.Do(probeReq)

			Expect(err).ToNot(HaveOccurred())
			Expect(response.StatusCode).To(Equal(404))
		})

		It("should fail with short timeout when probing", func() {
			// given a timeout shorter than the execution duration
			probeReq, err := http.NewRequest("GET", virtualProbesURL("/8080/healthz", vProbePort), nil)
			Expect(err).ToNot(HaveOccurred())
			probeReq.Header.Set(kuma_probes.HeaderNameTimeout, "2") // 2s is shorter than the execution duration (3s)

			response, err := http.DefaultClient.Do(probeReq)

			Expect(err).ToNot(HaveOccurred())
			Expect(response.StatusCode).To(Equal(503))
		})
	}, Ordered)

	Describe("HTTPS Probes", func() {
		vProbePort := uint32(9003)
		var stopCh chan struct{}
		var errCh chan error

		BeforeAll(func() {
			stopCh = make(chan struct{})
			errCh = make(chan error)
			prober := probes.NewProber(podIP, vProbePort)
			mockApp := &mockApplication{
				HTTP: &mockHTTPServerConfig{
					Path:             "/healthz?scheme=https",
					HTTPS:            true,
					ListenPort:       18443,
					ReturnStatusCode: 200,
				},
			}

			go func() {
				errCh <- prober.Start(stopCh)
			}()
			go func() {
				errCh <- mockApp.Start(stopCh)
			}()
			// wait a short period of time for the servers to be ready
			<-time.After(500 * time.Millisecond)
		})
		AfterAll(func() {
			close(stopCh)

			var err error
			select {
			case err = <-errCh:
			default:
			}

			Expect(err).ToNot(HaveOccurred())
		})

		It("should probe HTTPS upstream without verifying server certificates and keep query", func() {
			// time.Sleep(100 * time.Second)
			probeReq, err := http.NewRequest("GET", virtualProbesURL("/18443/healthz?scheme=https", vProbePort), nil)
			Expect(err).ToNot(HaveOccurred())
			probeReq.Header.Set(kuma_probes.HeaderNameScheme, "HTTPS")

			response, err := http.DefaultClient.Do(probeReq)

			Expect(err).ToNot(HaveOccurred())
			Expect(response.StatusCode).To(Equal(200))
		})
	}, Ordered)

	Describe("TCP Probes", func() {
		var stopCh chan struct{}
		var errCh chan error
		vProbePort := uint32(9004)

		BeforeAll(func() {
			stopCh = make(chan struct{})
			errCh = make(chan error)
			prober := probes.NewProber(podIP, vProbePort)
			go func() {
				errCh <- prober.Start(stopCh)
			}()

			mockApp := &mockApplication{
				TCP: &mockTCPServerConfig{
					ListenPort: 6379,
				},
			}

			go func() {
				errCh <- mockApp.Start(stopCh)
			}()
			// wait a short period of time for the server to be ready
			<-time.After(500 * time.Millisecond)
		})
		AfterAll(func() {
			close(stopCh)

			var err error
			select {
			case err = <-errCh:
			default:
			}

			Expect(err).ToNot(HaveOccurred())
		})

		It("should probe TCP server when it's healthy", func() {
			probeReq, err := http.NewRequest("GET", virtualProbesURL("/tcp/6379", vProbePort), nil)
			Expect(err).ToNot(HaveOccurred())

			response, err := http.DefaultClient.Do(probeReq)

			Expect(err).ToNot(HaveOccurred())
			Expect(response.StatusCode).To(Equal(200))
		})

		It("should probe TCP server when the port is not listening", func() {
			probeReq, err := http.NewRequest("GET", virtualProbesURL("/tcp/6000", vProbePort), nil)
			Expect(err).ToNot(HaveOccurred())
			probeReq.Header.Set(kuma_probes.HeaderNameTimeout, "3")

			response, err := http.DefaultClient.Do(probeReq)

			Expect(err).ToNot(HaveOccurred())
			Expect(response.StatusCode).To(Equal(503))
		})
	}, Ordered)

	Describe("GRPC Probes", func() {
		var stopCh chan struct{}
		var errCh chan error
		vProbePort := uint32(9005)

		BeforeAll(func() {
			stopCh = make(chan struct{})
			errCh = make(chan error)
			prober := probes.NewProber(podIP, vProbePort)
			go func() {
				errCh <- prober.Start(stopCh)
			}()

			mockApp := &mockApplication{
				GRPC: &mockGRPCServerConfig{
					ListenPort:        5678,
					ServiceName:       "liveness",
					IsHealthy:         true,
					ExecutionDuration: time.Duration(3) * time.Second,
				},
			}

			go func() {
				errCh <- mockApp.Start(stopCh)
			}()
			// wait a short period of time for the server to be ready
			<-time.After(500 * time.Millisecond)
		})
		AfterAll(func() {
			close(stopCh)

			var err error
			select {
			case err = <-errCh:
			default:
			}

			Expect(err).ToNot(HaveOccurred())
		})

		It("should probe gRPC server when it's healthy", func() {
			probeReq, err := http.NewRequest("GET", virtualProbesURL("/grpc/5678", vProbePort), nil)
			Expect(err).ToNot(HaveOccurred())
			probeReq.Header.Set(kuma_probes.HeaderNameGRPCService, "liveness")
			probeReq.Header.Set(kuma_probes.HeaderNameTimeout, "5")

			response, err := http.DefaultClient.Do(probeReq)

			Expect(err).ToNot(HaveOccurred())
			Expect(response.StatusCode).To(Equal(200))
		})

		It("should fail with a short timeout when probing", func() {
			probeReq, err := http.NewRequest("GET", virtualProbesURL("/grpc/5678", vProbePort), nil)
			Expect(err).ToNot(HaveOccurred())
			probeReq.Header.Set(kuma_probes.HeaderNameGRPCService, "liveness")
			probeReq.Header.Set(kuma_probes.HeaderNameTimeout, "2")

			response, err := http.DefaultClient.Do(probeReq)

			Expect(err).ToNot(HaveOccurred())
			Expect(response.StatusCode).To(Equal(503))
		})

		It("should probe gRPC server when the port is not listening", func() {
			probeReq, err := http.NewRequest("GET", virtualProbesURL("/grpc/5656", vProbePort), nil)
			Expect(err).ToNot(HaveOccurred())
			response, err := http.DefaultClient.Do(probeReq)

			Expect(err).ToNot(HaveOccurred())
			Expect(response.StatusCode).To(Equal(503))
		})

		It("should probe gRPC server when the application reports a failure", func() {
			probeReq, err := http.NewRequest("GET", virtualProbesURL("/grpc/5678", vProbePort), nil)
			Expect(err).ToNot(HaveOccurred())
			probeReq.Header.Set(kuma_probes.HeaderNameGRPCService, "readiness")

			response, err := http.DefaultClient.Do(probeReq)

			Expect(err).ToNot(HaveOccurred())
			Expect(response.StatusCode).To(Equal(503))
		})
	}, Ordered)
})

type mockHTTPServerConfig struct {
	ListenPort        uint32
	HTTPS             bool
	ReturnStatusCode  int
	Path              string
	ExecutionDuration time.Duration
	FailWhenHeader    string
}

type mockTCPServerConfig struct {
	ListenPort uint32
}

type mockGRPCServerConfig struct {
	ListenPort        uint32
	ServiceName       string
	IsHealthy         bool
	ExecutionDuration time.Duration
}

type mockApplication struct {
	HTTP *mockHTTPServerConfig
	TCP  *mockTCPServerConfig
	GRPC *mockGRPCServerConfig

	grpchealth.UnimplementedHealthServer
}

func (m *mockApplication) Start(stop <-chan struct{}) error {
	errCh := make(chan error)

	switch {
	case m.HTTP != nil:
		go func() {
			errCh <- m.startHTTPServer(stop)
		}()
	case m.TCP != nil:
		go func() {
			errCh <- m.startTCPServer(stop)
		}()
	case m.GRPC != nil:
		go func() {
			errCh <- m.startGRPCServer(stop)
		}()
	}

	return <-errCh
}

func (m *mockApplication) startHTTPServer(stop <-chan struct{}) error {
	if m.HTTP == nil {
		return nil
	}

	var httpReady atomic.Bool
	server := &http.Server{
		ReadHeaderTimeout: time.Second,
		Addr:              fmt.Sprintf(":%d", m.HTTP.ListenPort),
		Handler:           m,
		ErrorLog:          adapter.ToStd(GinkgoLogr),
	}
	if m.HTTP.HTTPS {
		tlsConfig, err := configureSelfSignedServerTLS("mock-application")
		if err != nil {
			return err_pkg.Wrap(err, "could not configure self-signed server TLS for the mock HTTPS server")
		}

		server.TLSConfig = tlsConfig
	}

	return startServer(func(stopper chan func()) error {
		stopper <- func() {
			GinkgoLogr.Info("stopping the mock HTTP Server")
			httpReady.Store(false)
			_ = server.Shutdown(context.Background())
		}

		errCh := make(chan error)
		GinkgoLogr.Info("starting mock HTTP Server", "address", server.Addr)
		if err := kuma_srv.StartServer(GinkgoLogr, server, &httpReady, errCh); err != nil {
			return err
		}
		return <-errCh
	}, stop)
}

func (m *mockApplication) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	if m.HTTP.ExecutionDuration > 0 {
		<-time.After(m.HTTP.ExecutionDuration)
	}

	if m.HTTP.Path != "" && m.HTTP.Path != req.RequestURI {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	if m.HTTP.FailWhenHeader != "" && req.Header.Get(m.HTTP.FailWhenHeader) != "" {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	writer.WriteHeader(m.HTTP.ReturnStatusCode)
	_, _ = writer.Write([]byte("mock HTTP server response"))
}

func configureSelfSignedServerTLS(commonName string) (*tls.Config, error) {
	kp, err := kuma_tls.NewSelfSignedCert("server", kuma_tls.DefaultKeyType, commonName)
	if err != nil {
		return nil, err
	}

	cert, err := tls.X509KeyPair(kp.CertPEM, kp.KeyPEM)
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
		MaxVersion:   tls.VersionTLS13,
	}, nil
}

func (m *mockApplication) startTCPServer(stop <-chan struct{}) error {
	if m.TCP == nil {
		return nil
	}
	return startServer(func(stopper chan func()) error {
		config := &net.ListenConfig{}
		addr := fmt.Sprintf(":%d", m.TCP.ListenPort)
		GinkgoLogr.Info("starting the mock TCP server", "address", addr)
		l, err := config.Listen(context.Background(), "tcp", addr)
		if err != nil {
			return err
		}
		stopper <- func() {
			GinkgoLogr.Info("stopping the mock TCP Server")
			_ = l.Close()
		}

		errCh := make(chan error)
		go m.handleTcpConnections(l, stop, errCh)
		return <-errCh
	}, stop)
}

func (m *mockApplication) handleTcpConnections(l net.Listener, cExit <-chan struct{}, cErr chan<- error) {
	for {
		conn, err := l.Accept()
		if err != nil {
			cErr <- err
			return
		}

		_, _ = conn.Write([]byte("connection to mock TCP server has successfully established"))
		_ = conn.Close()

		select {
		case <-cExit:
			return
		default:
		}
	}
}

func (m *mockApplication) startGRPCServer(stop <-chan struct{}) error {
	if m.GRPC == nil {
		return nil
	}

	grpcS := grpc.NewServer()
	grpchealth.RegisterHealthServer(grpcS, m)

	return startServer(func(stopper chan func()) error {
		stopper <- func() {
			GinkgoLogr.Info("stopping the mock gRPC Server")
			grpcS.Stop()
		}

		addr := fmt.Sprintf(":%d", m.GRPC.ListenPort)
		GinkgoLogr.Info("starting the mock gRPC server", "address", addr)
		lis, err := net.Listen("tcp", addr)
		if err != nil {
			return err_pkg.Wrap(err, "unable to listen the mock gRPC server")
		}
		if err := grpcS.Serve(lis); err != nil {
			return err
		}
		return nil
	}, stop)
}

func (m *mockApplication) Check(ctx context.Context, req *grpchealth.HealthCheckRequest) (*grpchealth.HealthCheckResponse, error) {
	status := grpchealth.HealthCheckResponse_NOT_SERVING
	if m.GRPC.IsHealthy {
		status = grpchealth.HealthCheckResponse_SERVING
	}

	if req.GetService() == m.GRPC.ServiceName {
		if m.GRPC.ExecutionDuration > 0 {
			<-time.After(m.GRPC.ExecutionDuration)
		}
		return &grpchealth.HealthCheckResponse{Status: status}, nil
	}

	return &grpchealth.HealthCheckResponse{Status: grpchealth.HealthCheckResponse_NOT_SERVING}, nil
}

func startServer(starter func(chan func()) error, stop <-chan struct{}) error {
	sReady := make(chan struct{}, 1)
	stopGetter := make(chan func(), 1)
	sError := make(chan error, 1)
	go func() {
		err := starter(stopGetter)
		if err != nil {
			sError <- err
			close(sReady)
		}
	}()

	<-sReady
	select {
	case serverErr := <-sError:
		return serverErr
	case <-stop:
		stopper := <-stopGetter
		if stopper != nil {
			stopper()
		}
		return nil
	}
}

func virtualProbesURL(path string, virtProbePort uint32) string {
	return fmt.Sprintf("http://%s%s", net.JoinHostPort(podIP, strconv.Itoa(int(virtProbePort))), path)
}
