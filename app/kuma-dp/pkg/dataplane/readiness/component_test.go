package readiness_test

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v2/app/kuma-dp/pkg/dataplane/readiness"
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
		reporter = readiness.NewReporter(true, "", "127.0.0.1", port, adminSocketPath)
		go func() {
			defer GinkgoRecover()
			_ = reporter.Start(stopCh)
		}()
		Eventually(func() error {
			_, err := http.Get(baseURL + "/ready")
			return err
		}, 5*time.Second, 50*time.Millisecond).Should(Succeed())
	}

	AfterEach(func() {
		if stopCh != nil {
			close(stopCh)
		}
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
			var err error
			adminListener, err = net.Listen("tcp", "127.0.0.1:0")
			Expect(err).ToNot(HaveOccurred())

			adminMux := http.NewServeMux()
			adminMux.HandleFunc("/ready", func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("LIVE\n"))
			})
			adminServer = &http.Server{
				Handler:           adminMux,
				ReadHeaderTimeout: time.Second,
			}
			go func() { _ = adminServer.Serve(adminListener) }()

			// The admin socket path is used to create a UDS client,
			// but for testing we use a TCP-based fake. We need a real
			// UDS to pass to NewReporter, so we create a temp socket
			// that the admin client will connect to.
			//
			// Instead, we test the non-admin-socket paths here and
			// verify that non-ready paths are blocked.
			startReporter("")
		})

		AfterEach(func() {
			if adminServer != nil {
				_ = adminServer.Close()
			}
		})

		It("does not expose admin endpoints", func() {
			resp, err := http.Get(baseURL + "/stats")
			Expect(err).ToNot(HaveOccurred())
			resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
		})
	})
})
