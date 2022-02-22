//go:build !windows
// +build !windows

package envoy_test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/app/kuma-dp/pkg/dataplane/envoy"
	kuma_dp "github.com/kumahq/kuma/pkg/config/app/kuma-dp"
	"github.com/kumahq/kuma/pkg/test"
)

var _ = Describe("Envoy", func() {

	var configDir string

	BeforeEach(func() {
		var err error
		configDir, err = os.MkdirTemp("", "")
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if configDir != "" {
			// when
			err := os.RemoveAll(configDir)
			// then
			Expect(err).ToNot(HaveOccurred())
		}
	})

	var outReader *os.File
	var outWriter, errWriter *os.File

	BeforeEach(func() {
		var err error
		outReader, outWriter, err = os.Pipe()
		Expect(err).ToNot(HaveOccurred())
		_, errWriter, err = os.Pipe()
		Expect(err).ToNot(HaveOccurred())
	})

	var stopCh chan struct{}
	var errCh chan error

	BeforeEach(func() {
		stopCh = make(chan struct{})
		errCh = make(chan error)
	})

	RunMockEnvoy := func(dataplane *envoy.Envoy) {
		go func() {
			errCh <- dataplane.Start(stopCh)
		}()

		Eventually(func() bool {
			select {
			case err := <-errCh:
				Expect(err).ToNot(HaveOccurred())
				return true
			default:
				return false
			}
		}, "5s", "10ms").Should(BeTrue())

		err := outWriter.Close()
		Expect(err).ToNot(HaveOccurred())
	}

	Describe("Run(..)", func() {
		It("should generate bootstrap config file and start Envoy", test.Within(10*time.Second, func() {
			// given
			cfg := kuma_dp.Config{
				Dataplane: kuma_dp.Dataplane{
					DrainTime: 15 * time.Second,
				},
				DataplaneRuntime: kuma_dp.DataplaneRuntime{
					BinaryPath: filepath.Join("testdata", "envoy-mock.exit-0.sh"),
					ConfigDir:  configDir,
				},
			}
			expectedConfigFile := filepath.Join(configDir, "bootstrap.yaml")

			By("starting a mock dataplane")
			// when
			dataplane, err := envoy.New(envoy.Opts{
				Config: cfg,
				BootstrapConfig: []byte(`node:
  id: example`),
				Stdout: outWriter,
				Stderr: errWriter,
			})
			Expect(err).To(Succeed())

			RunMockEnvoy(dataplane)

			Expect(err).ToNot(HaveOccurred())

			By("verifying the output of mock dataplane")
			// when
			var buf bytes.Buffer
			_, err = buf.ReadFrom(outReader)
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			if runtime.GOOS == "linux" {
				Expect(strings.TrimSpace(buf.String())).To(Equal(
					fmt.Sprintf("--config-path %s --drain-time-s 15 --disable-hot-restart --log-level off --cpuset-threads",
						expectedConfigFile)),
				)
			} else {
				Expect(strings.TrimSpace(buf.String())).To(Equal(
					fmt.Sprintf("--config-path %s --drain-time-s 15 --disable-hot-restart --log-level off",
						expectedConfigFile)),
				)
			}

			By("verifying the contents Envoy config file")
			// when
			actual, err := os.ReadFile(expectedConfigFile)
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(actual).To(MatchYAML(`
            node:
              id: example
`))
		}))

		It("should pass the concurrency Envoy", test.Within(10*time.Second, func() {
			// given
			cfg := kuma_dp.Config{
				Dataplane: kuma_dp.Dataplane{
					DrainTime: 15 * time.Second,
				},
				DataplaneRuntime: kuma_dp.DataplaneRuntime{
					BinaryPath:  filepath.Join("testdata", "envoy-mock.exit-0.sh"),
					ConfigDir:   configDir,
					Concurrency: 9,
				},
			}

			expectedConfigFile := filepath.Join(configDir, "bootstrap.yaml")

			By("starting a mock dataplane")
			// when
			dataplane, err := envoy.New(envoy.Opts{
				Config: cfg,
				BootstrapConfig: []byte(`node:
  id: example`),
				Stdout: outWriter,
				Stderr: errWriter,
			})
			Expect(err).To(Succeed())

			RunMockEnvoy(dataplane)

			Expect(err).ToNot(HaveOccurred())

			By("verifying the output of mock dataplane")
			// when
			var buf bytes.Buffer
			_, err = buf.ReadFrom(outReader)
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(strings.TrimSpace(buf.String())).To(Equal(
				fmt.Sprintf("--config-path %s --drain-time-s 15 --disable-hot-restart --log-level off --concurrency 9",
					expectedConfigFile)),
			)
		}))

		It("should return an error if Envoy crashes", test.Within(10*time.Second, func() {
			// given
			cfg := kuma_dp.Config{
				DataplaneRuntime: kuma_dp.DataplaneRuntime{
					BinaryPath: filepath.Join("testdata", "envoy-mock.exit-1.sh"),
					ConfigDir:  configDir,
				},
			}

			By("starting a mock dataplane")
			// when
			dataplane, err := envoy.New(envoy.Opts{
				Config:          cfg,
				BootstrapConfig: []byte{},
				Stdout:          &bytes.Buffer{},
				Stderr:          &bytes.Buffer{},
			})
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			go func() {
				errCh <- dataplane.Start(stopCh)
			}()

			By("waiting for mock dataplane to complete")
			// when
			err = <-errCh
			// then
			Expect(err).To(BeAssignableToTypeOf(&exec.ExitError{}))

			// when
			exitError := err.(*exec.ExitError)
			// then
			Expect(exitError.ProcessState.ExitCode()).To(Equal(1))
		}))

		It("should return an error if Envoy binary path is not found", test.Within(10*time.Second, func() {
			// given
			cfg := kuma_dp.Config{
				DataplaneRuntime: kuma_dp.DataplaneRuntime{
					BinaryPath: "testdata",
					ConfigDir:  configDir,
				},
			}

			By("starting a mock dataplane")
			// when
			dataplane, err := envoy.New(envoy.Opts{
				Config:          cfg,
				BootstrapConfig: []byte{},
				Stdout:          &bytes.Buffer{},
				Stderr:          &bytes.Buffer{},
			})
			// then
			Expect(dataplane).To(BeNil())
			// and
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(("could not find binary in any of the following paths")))
		}))
	})

	Describe("Parse version", func() {
		It("should properly read envoy version for unix-based systems", func() {
			// when
			version, err := envoy.GetEnvoyVersion(filepath.Join("testdata", "envoy-mock.exit-0.sh"))

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(version.Version).To(Equal("1.15.0"))
			Expect(version.Build).To(Equal("50ef0945fa2c5da4bff7627c3abf41fdd3b7cffd/1.15.0/clean-getenvoy-2aa564b-envoy/RELEASE/BoringSSL"))
		})

		It("should properly read envoy version with label for unix-based systems", func() {
			// when
			version, err := envoy.GetEnvoyVersion(filepath.Join("testdata", "envoy-mock.with-label.sh"))

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(version.Version).To(Equal("1.20.0-dev"))
			Expect(version.Build).To(Equal("50ef0945fa2c5da4bff7627c3abf41fdd3b7cffd/1.20.0-dev/clean-getenvoy-2aa564b-envoy/RELEASE/BoringSSL"))
		})

		It("should properly read envoy version for windows", func() {
			// when
			version, err := envoy.GetEnvoyVersion(filepath.Join("testdata", "envoy-mock-windows.exit-0.sh"))

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(version.Version).To(Equal("1.19.0"))
			Expect(version.Build).To(Equal("68fe53a889416fd8570506232052b06f5a531541/1.19.0/Modified/RELEASE/BoringSSL"))
		})
	})
})
