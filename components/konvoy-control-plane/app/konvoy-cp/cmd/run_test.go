package cmd

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"net/http"
	"os"

	"sigs.k8s.io/testing_frameworks/integration/addr"
)

var _ = Describe("run", func() {

	var stopCh chan struct{}
	var errCh chan error
	var configFile *os.File

	var diagnosticsPort int

	BeforeEach(func() {
		Expect(testEnv.Config).NotTo(BeNil())

		stopCh = make(chan struct{})
		errCh = make(chan error)

		freePort, _, err := addr.Suggest()
		Expect(err).NotTo(HaveOccurred())
		diagnosticsPort = freePort

		file, err := ioutil.TempFile("", "*")
		Expect(err).ToNot(HaveOccurred())
		configFile = file
	})

	AfterEach(func() {
		if configFile != nil {
			err := os.Remove(configFile.Name())
			Expect(err).ToNot(HaveOccurred())
		}
	})

	It("should be possible to run `konvoy-cp run`", func(done Done) {
		// given
		config := fmt.Sprintf(`
xdsServer:
  grpcPort: 0
  httpPort: 0
  diagnosticsPort: %d
environment: kubernetes
store:
  type: kubernetes
`, diagnosticsPort)
		_, err := configFile.WriteString(config)
		Expect(err).ToNot(HaveOccurred())

		cmd := newRunCmdWithOpts(runCmdOpts{
			SetupSignalHandler: func() <-chan struct{} {
				return stopCh
			},
		})
		cmd.SetArgs([]string{"--config-file=" + configFile.Name()})

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
