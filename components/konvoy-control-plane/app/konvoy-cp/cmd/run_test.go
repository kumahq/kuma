package cmd

import (
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"sigs.k8s.io/testing_frameworks/integration/addr"
)

var _ = Describe("run", func() {

	var stopCh chan struct{}
	var errCh chan error

	var diagnosticsPort int

	BeforeEach(func() {
		Expect(testEnv.Config).NotTo(BeNil())

		stopCh = make(chan struct{})
		errCh = make(chan error)

		freePort, _, err := addr.Suggest()
		Expect(err).NotTo(HaveOccurred())
		diagnosticsPort = freePort
	})

	It("should be possible to run `konvoy-cp run`", func(done Done) {
		// given
		cmd := newRunCmdWithOpts(runCmdOpts{
			SetupSignalHandler: func() <-chan struct{} {
				return stopCh
			},
		})
		cmd.SetArgs([]string{"--grpc-port=0", "--http-port=0", fmt.Sprintf("--diagnostics-port=%d", diagnosticsPort), "--metrics-port=0"})

		// when
		By("starting the Control Plane")
		go func() {
			defer close(errCh)
			errCh <- cmd.Execute()
		}()

		// then
		By("waiting for Control Plane to become healthy")
		Eventually(func() bool {
			resp, err := http.Get(fmt.Sprintf("http://localhost:%d/healthy", diagnosticsPort))
			if err != nil {
				return false
			}
			defer resp.Body.Close()
			return resp.StatusCode == http.StatusOK
		}, "10s", "10ms").Should(BeTrue())

		// then
		By("waiting for Control Plane to become ready")
		Eventually(func() bool {
			resp, err := http.Get(fmt.Sprintf("http://localhost:%d/ready", diagnosticsPort))
			if err != nil {
				return false
			}
			defer resp.Body.Close()
			return resp.StatusCode == http.StatusOK
		}, "10s", "10ms").Should(BeTrue())

		// when
		By("signalling Control Plane to stop")
		close(stopCh)

		// then
		select {
		case err := <-errCh:
			Expect(err).ToNot(HaveOccurred())
		}

		// complete
		close(done)
	}, 15)
})
