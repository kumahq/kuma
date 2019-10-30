package cmd

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"sigs.k8s.io/testing_frameworks/integration/addr"
)

type ConfigFactory interface {
	GenerateConfig() string
}

type StaticConfig string

func (c StaticConfig) GenerateConfig() string {
	return string(c)
}

type ConfigFactoryFunc func() string

func (f ConfigFactoryFunc) GenerateConfig() string {
	return f()
}

func RunSmokeTest(factory ConfigFactory) {
	Describe("run", func() {
		var stopCh chan struct{}
		var errCh chan error
		var configFile *os.File

		var diagnosticsPort int

		JustBeforeEach(func() {
			stopCh = make(chan struct{})
			errCh = make(chan error)

			freePort, _, err := addr.Suggest()
			Expect(err).NotTo(HaveOccurred())
			diagnosticsPort = freePort

			file, err := ioutil.TempFile("", "*")
			Expect(err).ToNot(HaveOccurred())
			configFile = file
		})

		JustAfterEach(func() {
			if configFile != nil {
				err := os.Remove(configFile.Name())
				Expect(err).ToNot(HaveOccurred())
			}
		})

		It("should be possible to run `kuma-cp run`", func(done Done) {
			// given
			config := fmt.Sprintf(factory.GenerateConfig(), diagnosticsPort)
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
			err = <-errCh
			Expect(err).ToNot(HaveOccurred())

			// complete
			close(done)
		}, 15)
	})
}
