// +build !windows

package envoy

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	kuma_dp "github.com/kumahq/kuma/pkg/config/app/kuma-dp"
	"github.com/kumahq/kuma/pkg/xds/bootstrap/types"
)

var _ = Describe("Envoy", func() {

	var configDir string

	BeforeEach(func() {
		var err error
		configDir, err = ioutil.TempDir("", "")
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

	Describe("Run(..)", func() {
		It("should generate bootstrap config file and start Envoy", func(done Done) {
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
			sampleConfig := func(string, kuma_dp.Config, BootstrapParams) ([]byte, types.BootstrapVersion, error) {
				return []byte(`node:
  id: example`), types.BootstrapV2, nil
			}
			expectedConfigFile := filepath.Join(configDir, "bootstrap.yaml")

			By("starting a mock dataplane")
			// when
			dataplane, _ := New(Opts{
				Config:    cfg,
				Generator: sampleConfig,
				Stdout:    outWriter,
				Stderr:    errWriter,
			})
			// and
			go func() {
				errCh <- dataplane.Start(stopCh)
			}()

			By("waiting for mock dataplane to complete")
			// then
			Eventually(func() bool {
				select {
				case err := <-errCh:
					Expect(err).ToNot(HaveOccurred())
					return true
				default:
					return false
				}
			}, "5s", "10ms").Should(BeTrue())

			By("closing the write side of the pipe")
			// when
			err := outWriter.Close()
			// then
			Expect(err).ToNot(HaveOccurred())

			By("verifying the output of mock dataplane")
			// when
			var buf bytes.Buffer
			_, err = buf.ReadFrom(outReader)
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(strings.TrimSpace(buf.String())).To(Equal(fmt.Sprintf("-c %s --drain-time-s 15 --disable-hot-restart --bootstrap-version 2", expectedConfigFile)))

			By("verifying the contents Envoy config file")
			// when
			actual, err := ioutil.ReadFile(expectedConfigFile)
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(actual).To(MatchYAML(`
            node:
              id: example
`))
			// complete
			close(done)
		}, 10)

		It("should return an error if Envoy crashes", func(done Done) {
			// given
			cfg := kuma_dp.Config{
				DataplaneRuntime: kuma_dp.DataplaneRuntime{
					BinaryPath: filepath.Join("testdata", "envoy-mock.exit-1.sh"),
					ConfigDir:  configDir,
				},
			}
			sampleConfig := func(string, kuma_dp.Config, BootstrapParams) ([]byte, types.BootstrapVersion, error) {
				return nil, "", nil
			}

			By("starting a mock dataplane")
			// when
			dataplane, err := New(Opts{
				Config:    cfg,
				Generator: sampleConfig,
				Stdout:    &bytes.Buffer{},
				Stderr:    &bytes.Buffer{},
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

			// complete
			close(done)
		}, 10)

		It("should return an error if Envoy binay path is not found", func(done Done) {
			// given
			cfg := kuma_dp.Config{
				DataplaneRuntime: kuma_dp.DataplaneRuntime{
					BinaryPath: filepath.Join("testdata"),
					ConfigDir:  configDir,
				},
			}
			sampleConfig := func(string, kuma_dp.Config, BootstrapParams) ([]byte, types.BootstrapVersion, error) {
				return nil, "", nil
			}

			By("starting a mock dataplane")
			// when
			dataplane, err := New(Opts{
				Config:    cfg,
				Generator: sampleConfig,
				Stdout:    &bytes.Buffer{},
				Stderr:    &bytes.Buffer{},
			})
			// then
			Expect(dataplane).To(BeNil())
			// and
			Expect(err.Error()).To(ContainSubstring(("could not find binary in any of the following paths")))

			// complete
			close(done)
		}, 10)
	})

	Describe("Parse version", func() {
		It("should properly read envoy version", func() {
			// given
			cfg := kuma_dp.Config{
				DataplaneRuntime: kuma_dp.DataplaneRuntime{
					BinaryPath: filepath.Join("testdata", "envoy-mock.exit-0.sh"),
					ConfigDir:  configDir,
				},
			}

			// when
			dataplane, err := New(Opts{
				Config: cfg,
				Stdout: &bytes.Buffer{},
				Stderr: &bytes.Buffer{},
			})
			Expect(err).ToNot(HaveOccurred())
			version, err := dataplane.version()

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(version.Version).To(Equal("1.15.0"))
			Expect(version.Build).To(Equal("50ef0945fa2c5da4bff7627c3abf41fdd3b7cffd/1.15.0/clean-getenvoy-2aa564b-envoy/RELEASE/BoringSSL"))
		})
	})
})
