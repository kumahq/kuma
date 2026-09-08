package readiness_test

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v3/app/kuma-dp/pkg/dataplane/readiness"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
)

var _ = Describe("Readiness Reporter", func() {
	var (
		reporter *readiness.Reporter
		port     uint32
		baseURL  string
		stopCh   chan struct{}
	)

	startReporter := func(adminSocketPath string) {
		stopCh = make(chan struct{})
		// Find a free port
		lis, err := net.Listen("tcp", "127.0.0.1:0")
		Expect(err).ToNot(HaveOccurred())
		port = uint32(lis.Addr().(*net.TCPAddr).Port)
		Expect(lis.Close()).To(Succeed())

		baseURL = fmt.Sprintf("http://127.0.0.1:%d", port)
		reporter = readiness.NewReporter("127.0.0.1", port, readiness.EnvoyAdmin{SocketPath: adminSocketPath}, nil)
		go func() {
			defer GinkgoRecover()
			_ = reporter.Start(stopCh)
		}()
		Eventually(func() error {
			resp, err := http.Get(baseURL + "/ready")
			if err != nil {
				return err
			}
			resp.Body.Close()
			return nil
		}, 5*time.Second, 50*time.Millisecond).Should(Succeed())
	}

	AfterEach(func() {
		if stopCh != nil {
			close(stopCh)
		}
	})

	Context("with DNS config gate", func() {
		var dnsReady chan struct{}

		BeforeEach(func() {
			stopCh = make(chan struct{})
			lis, err := net.Listen("tcp", "127.0.0.1:0")
			Expect(err).ToNot(HaveOccurred())
			port = uint32(lis.Addr().(*net.TCPAddr).Port)
			Expect(lis.Close()).To(Succeed())

			baseURL = fmt.Sprintf("http://127.0.0.1:%d", port)
			dnsReady = make(chan struct{})
			reporter = readiness.NewReporter("127.0.0.1", port, readiness.EnvoyAdmin{}, dnsReady)
			go func() {
				defer GinkgoRecover()
				_ = reporter.Start(stopCh)
			}()
			Eventually(func() error {
				resp, err := http.Get(baseURL + "/ready")
				if err != nil {
					return err
				}
				resp.Body.Close()
				return nil
			}, 5*time.Second, 50*time.Millisecond).Should(Succeed())
		})

		It("returns NOT_READY until DNS config channel is closed", func() {
			resp, err := http.Get(baseURL + "/ready")
			Expect(err).ToNot(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusServiceUnavailable))
			body, err := io.ReadAll(resp.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(body)).To(Equal("NOT_READY"))
		})

		It("returns READY after DNS config channel is closed", func() {
			close(dnsReady)
			Eventually(func() int {
				resp, err := http.Get(baseURL + "/ready")
				if err != nil {
					return 0
				}
				resp.Body.Close()
				return resp.StatusCode
			}, 5*time.Second, 50*time.Millisecond).Should(Equal(http.StatusOK))
		})
	})

	Context("with DNS config gate timeout", func() {
		BeforeEach(func() {
			stopCh = make(chan struct{})
			lis, err := net.Listen("tcp", "127.0.0.1:0")
			Expect(err).ToNot(HaveOccurred())
			port = uint32(lis.Addr().(*net.TCPAddr).Port)
			Expect(lis.Close()).To(Succeed())

			baseURL = fmt.Sprintf("http://127.0.0.1:%d", port)
			dnsReady := make(chan struct{})
			reporter = readiness.NewReporterWithDeadline("127.0.0.1", port, readiness.EnvoyAdmin{}, dnsReady, time.Now().Add(-time.Second))
			go func() {
				defer GinkgoRecover()
				_ = reporter.Start(stopCh)
			}()
			Eventually(func() error {
				resp, err := http.Get(baseURL + "/ready")
				if err != nil {
					return err
				}
				resp.Body.Close()
				return nil
			}, 5*time.Second, 50*time.Millisecond).Should(Succeed())
		})

		It("bypasses gate and returns READY when deadline is already past", func() {
			resp, err := http.Get(baseURL + "/ready")
			Expect(err).ToNot(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			body, err := io.ReadAll(resp.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(body)).To(Equal("READY"))
		})
	})

	Context("without admin socket", func() {
		BeforeEach(func() {
			startReporter("")
		})

		It("returns READY on /ready", func() {
			resp, err := http.Get(baseURL + "/ready")
			Expect(err).ToNot(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			body, err := io.ReadAll(resp.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(body)).To(Equal("READY"))
			Expect(resp.Header.Get("x-powered-by")).To(Equal("kuma-dp"))
		})

		It("returns TERMINATING after Terminating()", func() {
			reporter.Terminating()

			resp, err := http.Get(baseURL + "/ready")
			Expect(err).ToNot(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusServiceUnavailable))
			body, err := io.ReadAll(resp.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(body)).To(Equal("TERMINATING"))
		})

		It("returns 404 for non-ready paths", func() {
			for _, path := range []string{"/stats", "/config_dump", "/clusters", "/quitquitquit", "/logging"} {
				resp, err := http.Get(baseURL + path)
				Expect(err).ToNot(HaveOccurred())
				resp.Body.Close()
				Expect(resp.StatusCode).To(Equal(http.StatusNotFound), "path %s should return 404", path)
			}
		})
	})

	Context("with admin socket", func() {
		var (
			adminListener net.Listener
			adminServer   *http.Server
		)

		BeforeEach(func() {
			// Fake Envoy admin on a real UDS. It serves /ready AND the
			// sensitive admin endpoints, so that a 404 from the reporter
			// proves the reporter refuses to proxy them - not merely that
			// the backend lacks the route.
			dir, err := os.MkdirTemp("", "rdy")
			Expect(err).ToNot(HaveOccurred())
			DeferCleanup(func() { _ = os.RemoveAll(dir) })
			sockPath := filepath.Join(dir, "admin.sock")

			adminListener, err = net.Listen("unix", sockPath)
			Expect(err).ToNot(HaveOccurred())

			adminMux := http.NewServeMux()
			adminMux.HandleFunc("/ready", func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("LIVE\n"))
			})
			for _, p := range []string{"/config_dump", "/stats", "/clusters", "/certs", "/logging", "/quitquitquit"} {
				adminMux.HandleFunc(p, func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte("SENSITIVE"))
				})
			}
			adminServer = &http.Server{
				Handler:           adminMux,
				ReadHeaderTimeout: time.Second,
			}
			go func() { _ = adminServer.Serve(adminListener) }()

			startReporter(sockPath)
		})

		AfterEach(func() {
			if adminServer != nil {
				_ = adminServer.Close()
			}
		})

		It("proxies /ready to Envoy admin over the UDS", func() {
			resp, err := http.Get(baseURL + "/ready")
			Expect(err).ToNot(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			body, err := io.ReadAll(resp.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(body)).To(Equal("LIVE\n"))
		})

		It("does not expose any admin endpoint on the readiness port", func() {
			for _, path := range []string{"/config_dump", "/stats", "/clusters", "/certs", "/logging", "/quitquitquit"} {
				resp, err := http.Get(baseURL + path)
				Expect(err).ToNot(HaveOccurred())
				body, err := io.ReadAll(resp.Body)
				resp.Body.Close()
				Expect(err).ToNot(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(http.StatusNotFound), "path %s should return 404", path)
				Expect(string(body)).ToNot(ContainSubstring("SENSITIVE"), "path %s must not be proxied", path)
			}
		})

		It("serves only the exact /ready path, not subpaths", func() {
			// /ready is registered without a trailing slash, so net/http.ServeMux
			// matches it exactly. Subpaths must 404 instead of falling through to
			// the readiness handler (a trailing-slash pattern would match the
			// whole subtree).
			for _, path := range []string{"/ready/", "/ready/config_dump", "/readyz"} {
				resp, err := http.Get(baseURL + path)
				Expect(err).ToNot(HaveOccurred())
				body, err := io.ReadAll(resp.Body)
				resp.Body.Close()
				Expect(err).ToNot(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(http.StatusNotFound), "path %s should return 404", path)
				Expect(string(body)).ToNot(Equal("LIVE\n"), "path %s must not hit the readiness handler", path)
			}
		})
	})

	Context("with IPv6 wildcard listener", func() {
		BeforeEach(func() {
			probe, err := net.Listen("tcp6", "[::1]:0")
			if err != nil {
				Skip("IPv6 loopback not available: " + err.Error())
			}
			Expect(probe.Close()).To(Succeed())

			stopCh = make(chan struct{})
			lis, err := net.Listen("tcp", "[::]:0")
			Expect(err).ToNot(HaveOccurred())
			port = uint32(lis.Addr().(*net.TCPAddr).Port)
			Expect(lis.Close()).To(Succeed())

			reporter = readiness.NewReporter("::", port, readiness.EnvoyAdmin{}, nil)
			go func() {
				defer GinkgoRecover()
				_ = reporter.Start(stopCh)
			}()
			baseURL = fmt.Sprintf("http://[::1]:%d", port)
			Eventually(func() error {
				resp, err := http.Get(baseURL + "/ready")
				if err != nil {
					return err
				}
				resp.Body.Close()
				return nil
			}, 5*time.Second, 50*time.Millisecond).Should(Succeed())
		})

		It("accepts probes on IPv6 loopback", func() {
			resp, err := http.Get(fmt.Sprintf("http://[::1]:%d/ready", port))
			Expect(err).ToNot(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
		})

		It("accepts probes on IPv4 loopback (dual-stack)", func() {
			resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/ready", port))
			Expect(err).ToNot(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
		})
	})
})

var _ = Describe("IdentityGate", func() {
	var (
		gate     *readiness.IdentityGate
		listener net.Listener
		server   *http.Server
	)

	BeforeEach(func() {
		socketPath := filepath.Join(GinkgoT().TempDir(), "identity.sock")
		certificate, err := tls.LoadX509KeyPair(
			filepath.Join("..", "..", "..", "..", "..", "test", "certs", "server-cert.pem"),
			filepath.Join("..", "..", "..", "..", "..", "test", "certs", "server-key.pem"),
		)
		Expect(err).ToNot(HaveOccurred())
		rawListener, err := net.Listen("unix", socketPath)
		Expect(err).ToNot(HaveOccurred())
		listener = tls.NewListener(rawListener, &tls.Config{
			Certificates: []tls.Certificate{certificate},
			MinVersion:   tls.VersionTLS12,
		})
		server = &http.Server{
			ReadHeaderTimeout: time.Second,
			Handler: http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
				_, _ = writer.Write([]byte("ready"))
			}),
		}
		go func() { _ = server.Serve(listener) }()
		gate = readiness.NewIdentityGate(socketPath)
	})

	AfterEach(func() {
		Expect(server.Close()).To(Succeed())
	})

	setConfig := func(config core_xds.IdentityReadinessConfig) {
		bytes, err := json.Marshal(config)
		Expect(err).ToNot(HaveOccurred())
		Expect(gate.OnChange(context.Background(), strings.NewReader(string(bytes)))).To(Succeed())
	}

	It("fails closed until dynamic configuration arrives", func() {
		ready, err := gate.Ready(context.Background())
		Expect(err).ToNot(HaveOccurred())
		Expect(ready).To(BeFalse())
	})

	It("is ready when workload identity is not required", func() {
		setConfig(core_xds.IdentityReadinessConfig{Required: false})
		ready, err := gate.Ready(context.Background())
		Expect(err).ToNot(HaveOccurred())
		Expect(ready).To(BeTrue())
	})

	It("requires the current identity certificate", func() {
		certificate, err := tls.LoadX509KeyPair(
			filepath.Join("..", "..", "..", "..", "..", "test", "certs", "server-cert.pem"),
			filepath.Join("..", "..", "..", "..", "..", "test", "certs", "server-key.pem"),
		)
		Expect(err).ToNot(HaveOccurred())
		hash := sha256.Sum256(certificate.Certificate[0])
		setConfig(core_xds.IdentityReadinessConfig{Required: true, CertificateHash: fmt.Sprintf("%x", hash)})
		ready, err := gate.Ready(context.Background())
		Expect(err).ToNot(HaveOccurred())
		Expect(ready).To(BeTrue())

		setConfig(core_xds.IdentityReadinessConfig{Required: true, CertificateHash: "stale"})
		ready, err = gate.Ready(context.Background())
		Expect(err).To(MatchError("identity listener presented a stale certificate"))
		Expect(ready).To(BeFalse())
	})

	It("rejects an expired identity", func() {
		expiration := time.Now().Add(-time.Second)
		setConfig(core_xds.IdentityReadinessConfig{
			Required:       true,
			ExpirationTime: &expiration,
		})
		ready, err := gate.Ready(context.Background())
		Expect(err).To(MatchError("identity certificate expired"))
		Expect(ready).To(BeFalse())
	})
})

var _ = Describe("Readiness Reporter with Envoy admin on TCP", func() {
	var (
		reporter  *readiness.Reporter
		baseURL   string
		stopCh    chan struct{}
		adminSrv  *http.Server
		adminCode int
	)

	BeforeEach(func() {
		adminCode = http.StatusOK

		adminLis, err := net.Listen("tcp", "127.0.0.1:0")
		Expect(err).ToNot(HaveOccurred())
		adminAddr := adminLis.Addr().(*net.TCPAddr)

		mux := http.NewServeMux()
		mux.HandleFunc("/ready", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(adminCode)
			_, _ = w.Write([]byte("LIVE"))
		})
		adminSrv = &http.Server{Handler: mux, ReadHeaderTimeout: time.Second}
		go func() {
			defer GinkgoRecover()
			_ = adminSrv.Serve(adminLis)
		}()

		lis, err := net.Listen("tcp", "127.0.0.1:0")
		Expect(err).ToNot(HaveOccurred())
		port := uint32(lis.Addr().(*net.TCPAddr).Port)
		Expect(lis.Close()).To(Succeed())

		baseURL = fmt.Sprintf("http://127.0.0.1:%d", port)
		stopCh = make(chan struct{})
		reporter = readiness.NewReporter("127.0.0.1", port, readiness.EnvoyAdmin{
			Address: "127.0.0.1",
			Port:    uint32(adminAddr.Port),
		}, nil)
		go func() {
			defer GinkgoRecover()
			_ = reporter.Start(stopCh)
		}()
		Eventually(func() error {
			resp, err := http.Get(baseURL + "/ready")
			if err != nil {
				return err
			}
			resp.Body.Close()
			return nil
		}, 5*time.Second, 50*time.Millisecond).Should(Succeed())
	})

	AfterEach(func() {
		close(stopCh)
		Expect(adminSrv.Close()).To(Succeed())
	})

	It("should report ready when Envoy admin is ready", func() {
		resp, err := http.Get(baseURL + "/ready")
		Expect(err).ToNot(HaveOccurred())
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(http.StatusOK))
		Expect(string(body)).To(Equal("LIVE"))
	})

	It("should report not ready when Envoy admin is not ready", func() {
		adminCode = http.StatusServiceUnavailable

		resp, err := http.Get(baseURL + "/ready")
		Expect(err).ToNot(HaveOccurred())
		defer resp.Body.Close()

		Expect(resp.StatusCode).To(Equal(http.StatusServiceUnavailable))
	})
})
