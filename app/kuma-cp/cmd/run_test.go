package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/testing_frameworks/integration/addr"

	kuma_cmd "github.com/kumahq/kuma/pkg/cmd"
	"github.com/kumahq/kuma/pkg/test"
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

func RunSmokeTest(factory ConfigFactory, workdir string) {
	Describe("run", func() {
		var errCh chan error
		var configFile *os.File

		var diagnosticsPort int
		var ctx context.Context
		var cancel func()
		var opts = kuma_cmd.RunCmdOpts{
			SetupSignalHandler: func() context.Context {
				return ctx
			},
		}

		JustBeforeEach(func() {
			ctx, cancel = context.WithCancel(context.Background())
			errCh = make(chan error)

			freePort, _, err := addr.Suggest()
			Expect(err).NotTo(HaveOccurred())
			diagnosticsPort = freePort

			file, err := os.CreateTemp("", "*")
			Expect(err).ToNot(HaveOccurred())
			configFile = file
		})

		JustAfterEach(func() {
			if configFile != nil {
				err := os.Remove(configFile.Name())
				Expect(err).ToNot(HaveOccurred())
			}
			if workdir != "" {
				err := os.RemoveAll(workdir)
				Expect(err).ToNot(HaveOccurred())
			}
		})

		It("should be possible to run `kuma-cp run with default mode`", test.Within(time.Minute, func() {
			// given
			config := fmt.Sprintf(factory.GenerateConfig(), diagnosticsPort)
			_, err := configFile.WriteString(config)
			Expect(err).ToNot(HaveOccurred())
			cmd := newRunCmdWithOpts(opts)
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
			By("signaling Control Plane to stop")
			cancel()

			// then
			err = <-errCh
			Expect(err).ToNot(HaveOccurred())
		}))
	})
}
