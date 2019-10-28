package cmd

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"sigs.k8s.io/testing_frameworks/integration/addr"
)

var _ = Describe("run", func() {

	var backupSetupSignalHandler func() <-chan struct{}

	BeforeEach(func() {
		backupSetupSignalHandler = setupSignalHandler
	})

	AfterEach(func() {
		setupSignalHandler = backupSetupSignalHandler
	})

	var stopCh chan struct{}

	BeforeEach(func() {
		stopCh = make(chan struct{})

		setupSignalHandler = func() <-chan struct{} {
			return stopCh
		}
	})

	var configFile *os.File

	BeforeEach(func() {
		var err error
		configFile, err = ioutil.TempFile("", "*")
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if configFile != nil {
			err := os.Remove(configFile.Name())
			Expect(err).ToNot(HaveOccurred())
		}
	})

	var port int

	BeforeEach(func() {
		var err error
		port, _, err = addr.Suggest()
		Expect(err).NotTo(HaveOccurred())

		var certDir string
		certDir, err = filepath.Abs("testdata")
		Expect(err).NotTo(HaveOccurred())

		config := fmt.Sprintf(`
        webHookServer:
          port: %d
          certDir: %s
`, port, certDir)
		_, err = configFile.WriteString(config)
		Expect(err).ToNot(HaveOccurred())
	})

	var client *http.Client
	BeforeEach(func() {
		client = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		}
	})

	It("should be possible to run `kuma-injector run`", func(done Done) {
		// given
		cmd := newRootCmd()
		cmd.SetArgs([]string{"run", "--config-file", configFile.Name()})

		// when
		By("starting the Kuma Injector")
		errCh := make(chan error)
		go func() {
			defer close(errCh)
			errCh <- cmd.Execute()
		}()

		// then
		By("waiting for Kuma Injector to become ready")
		Eventually(func() bool {
			resp, err := client.Get(fmt.Sprintf("https://localhost:%d/inject-sidecar", port))
			if err != nil {
				return false
			}
			defer resp.Body.Close()
			return resp.StatusCode == http.StatusOK
		}, "10s", "10ms").Should(BeTrue())

		// when
		By("signalling Kuma Injector to stop")
		close(stopCh)

		// then
		err := <-errCh
		Expect(err).ToNot(HaveOccurred())

		// complete
		close(done)
	}, 15)
})
