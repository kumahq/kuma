// +build !windows

package cmd

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/Kong/kuma/app/kuma-dp/pkg/dataplane/envoy"
	kumadp "github.com/Kong/kuma/pkg/config/app/kuma-dp"
	util_proto "github.com/Kong/kuma/pkg/util/proto"
	envoy_bootstrap "github.com/envoyproxy/go-control-plane/envoy/config/bootstrap/v2"
	"github.com/gogo/protobuf/proto"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/Kong/kuma/pkg/core"
)

var _ = Describe("run", func() {

	var backupSetupSignalHandler func() <-chan struct{}
	var backupBootstrapGenerator envoy.BootstrapConfigFactoryFunc

	BeforeEach(func() {
		backupSetupSignalHandler = core.SetupSignalHandler
		backupBootstrapGenerator = bootstrapGenerator
		bootstrapGenerator = func(cfg kumadp.Config) (proto.Message, error) {
			bootstrap := envoy_bootstrap.Bootstrap{}
			respBytes, err := ioutil.ReadFile(filepath.Join("testdata", "bootstrap-config.golden.yaml"))
			Expect(err).ToNot(HaveOccurred())
			err = util_proto.FromYAML(respBytes, &bootstrap)
			Expect(err).ToNot(HaveOccurred())
			return &bootstrap, nil
		}
	})
	AfterEach(func() {
		core.SetupSignalHandler = backupSetupSignalHandler
		bootstrapGenerator = backupBootstrapGenerator
	})

	var stopCh chan struct{}

	BeforeEach(func() {
		stopCh = make(chan struct{})

		core.SetupSignalHandler = func() <-chan struct{} {
			return stopCh
		}
	})

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

	var backupEnvVars []string

	BeforeEach(func() {
		backupEnvVars = os.Environ()
	})
	AfterEach(func() {
		os.Clearenv()
		for _, envVar := range backupEnvVars {
			parts := strings.SplitN(envVar, "=", 2)
			Expect(os.Setenv(parts[0], parts[1])).To(Succeed())
		}
	})

	type testCase struct {
		envVars map[string]string
		args    []string
	}
	DescribeTable("should be possible to start dataplane (Envoy) using `kuma-dp run`",
		func(given testCase) {
			// setup
			pidFile := filepath.Join(configDir, "envoy-mock.pid")

			// and
			env := given.envVars
			env["KUMA_DATAPLANE_RUNTIME_CONFIG_DIR"] = configDir
			env["ENVOY_MOCK_PID_FILE"] = pidFile
			for key, value := range env {
				Expect(os.Setenv(key, value)).To(Succeed())
			}

			// given
			cmd := newRootCmd()
			cmd.SetArgs(append([]string{"run"}, given.args...))
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})

			// when
			By("starting the dataplane manager")
			errCh := make(chan error)
			go func() {
				defer close(errCh)
				errCh <- cmd.Execute()
			}()

			// then
			var pid int64
			By("waiting for dataplane (Envoy) to get started")
			Eventually(func() bool {
				data, err := ioutil.ReadFile(pidFile)
				if err != nil {
					return false
				}
				pid, err = strconv.ParseInt(strings.TrimSpace(string(data)), 10, 32)
				return err == nil
			}, "5s", "100ms").Should(BeTrue())
			// and
			Expect(pid).ToNot(BeZero())

			// when
			By("signalling the dataplane manager to stop")
			close(stopCh)

			// then
			select {
			case err := <-errCh:
				Expect(err).ToNot(HaveOccurred())
			}

			By("waiting for dataplane (Envoy) to get stopped")
			Eventually(func() bool {
				// send sig 0 to check whether Envoy process still exists
				err := syscall.Kill(int(pid), syscall.Signal(0))
				// we expect Envoy process to get killed by now
				return err != nil
			}, "5s", "100ms").Should(BeTrue())
		},
		Entry("can be launched with env vars", testCase{
			envVars: map[string]string{
				"KUMA_CONTROL_PLANE_BOOTSTRAP_SERVER_URL": "http://localhost:1234",
				"KUMA_DATAPLANE_NAME":                     "example",
				"KUMA_DATAPLANE_ADMIN_PORT":               "2345",
				"KUMA_DATAPLANE_RUNTIME_BINARY_PATH":      filepath.Join("testdata", "envoy-mock.sleep.sh"),
			},
			args: []string{},
		}),
		Entry("can be launched with args", testCase{
			envVars: map[string]string{},
			args: []string{
				"--cp-address", "http://localhost:1234",
				"--name", "example",
				"--admin-port", "2345",
				"--binary-path", filepath.Join("testdata", "envoy-mock.sleep.sh"),
			},
		}),
	)
})
