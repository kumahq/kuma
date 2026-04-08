package readiness_test

import (
	"context"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v2/app/kuma-dp/pkg/dataplane/readiness"
	core_xds "github.com/kumahq/kuma/v2/pkg/core/xds"
)

var _ = Describe("Readiness Reporter", func() {
	var (
		reporter *readiness.Reporter
		client   *http.Client
		stopCh   chan struct{}
		tmpDir   string
	)

	startReporter := func(adminSocketPath string) {
		stopCh = make(chan struct{})
		var err error
		tmpDir, err = os.MkdirTemp("", "readiness-test-*")
		Expect(err).ToNot(HaveOccurred())

		socketPath := core_xds.ReadinessReporterSocketName(tmpDir)
		client = &http.Client{
			Transport: &http.Transport{
				DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
					return (&net.Dialer{}).DialContext(ctx, "unix", socketPath)
				},
			},
		}
		reporter = readiness.NewReporter(tmpDir, adminSocketPath)
		go func() {
			defer GinkgoRecover()
			_ = reporter.Start(stopCh)
		}()
		Eventually(func() error {
			conn, err := net.Dial("unix", socketPath)
			if err != nil {
				return err
			}
			conn.Close()
			return nil
		}, 5*time.Second, 50*time.Millisecond).Should(Succeed())
	}

	doGet := func(path string) (*http.Response, error) {
		return client.Get("http://unix" + path)
	}

	AfterEach(func() {
		if stopCh != nil {
			close(stopCh)
		}
		if tmpDir != "" {
			_ = os.RemoveAll(tmpDir)
		}
	})

	Context("without admin socket", func() {
		BeforeEach(func() {
			startReporter("")
		})

		It("returns READY on /ready", func() {
			resp, err := doGet("/ready")
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

			resp, err := doGet("/ready")
			Expect(err).ToNot(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusServiceUnavailable))
			body, err := io.ReadAll(resp.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(body)).To(Equal("TERMINATING"))
		})

		It("returns 404 for non-ready paths", func() {
			for _, path := range []string{"/stats", "/config_dump", "/clusters", "/quitquitquit", "/logging"} {
				resp, err := doGet(path)
				Expect(err).ToNot(HaveOccurred())
				resp.Body.Close()
				Expect(resp.StatusCode).To(Equal(http.StatusNotFound), "path %s should return 404", path)
			}
		})
	})

	Context("with admin socket", func() {
		BeforeEach(func() {
			startReporter("")
		})

		It("does not expose admin endpoints", func() {
			resp, err := doGet("/stats")
			Expect(err).ToNot(HaveOccurred())
			resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
		})
	})
})
