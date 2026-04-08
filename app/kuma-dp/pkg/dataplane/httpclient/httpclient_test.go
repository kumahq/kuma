package httpclient_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v2/app/kuma-dp/pkg/dataplane/httpclient"
)

var _ = Describe("httpclient", func() {
	Describe("NewUDS", func() {
		It("should create a client that dials a Unix domain socket", func() {
			socketPath := filepath.Join(os.TempDir(), fmt.Sprintf("httpclient-test-%d.sock", GinkgoParallelProcess()))
			defer os.Remove(socketPath)

			lis, err := net.Listen("unix", socketPath)
			Expect(err).ToNot(HaveOccurred())
			defer lis.Close()

			srv := &http.Server{
				Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusOK)
				}),
				ReadHeaderTimeout: time.Second,
			}
			go func() { _ = srv.Serve(lis) }()
			defer func() { Expect(srv.Shutdown(context.Background())).To(Succeed()) }()

			client := httpclient.NewUDS(socketPath, 2*time.Second, 5*time.Second)
			resp, err := client.Get("http://localhost/test")
			Expect(err).ToNot(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
		})

		It("should set the client timeout", func() {
			client := httpclient.NewUDS("/nonexistent.sock", 1*time.Second, 3*time.Second)
			Expect(client.Timeout).To(Equal(3 * time.Second))
		})
	})

	Describe("NewTCPOrUDS", func() {
		It("should return a plain TCP client when socketPath is empty", func() {
			client := httpclient.NewTCPOrUDS("", 1*time.Second, 5*time.Second)
			Expect(client.Transport).To(BeNil())
			Expect(client.Timeout).To(Equal(5 * time.Second))
		})

		It("should return a UDS client when socketPath is non-empty", func() {
			client := httpclient.NewTCPOrUDS("/some/socket.sock", 1*time.Second, 5*time.Second)
			Expect(client.Transport).ToNot(BeNil())
			Expect(client.Timeout).To(Equal(5 * time.Second))
		})
	})
})
