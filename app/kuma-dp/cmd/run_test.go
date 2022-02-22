//go:build !windows
// +build !windows

package cmd

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	envoy_bootstrap_v3 "github.com/envoyproxy/go-control-plane/envoy/config/bootstrap/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/app/kuma-dp/pkg/dataplane/envoy"
	kuma_cmd "github.com/kumahq/kuma/pkg/cmd"
	kumadp "github.com/kumahq/kuma/pkg/config/app/kuma-dp"
	"github.com/kumahq/kuma/pkg/test"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("run", func() {

	var cancel func()
	var ctx context.Context
	opts := kuma_cmd.RunCmdOpts{
		SetupSignalHandler: func() context.Context {
			return ctx
		},
	}

	var tmpDir string

	BeforeEach(func() {
		ctx, cancel = context.WithCancel(context.Background())
		var err error
		tmpDir, err = os.MkdirTemp("", "")
		Expect(err).ToNot(HaveOccurred())
	})
	AfterEach(func() {
		if tmpDir != "" {
			// when
			err := os.RemoveAll(tmpDir)
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

	var port int
	BeforeEach(func() {
		var err error
		port, err = test.GetFreePort()
		Expect(err).NotTo(HaveOccurred())
	})

	type testCase struct {
		envVars      map[string]string
		args         []string
		expectedFile string
	}
	DescribeTable("should be possible to start dataplane (Envoy) using `kuma-dp run`",
		func(givenFunc func() testCase) {
			given := givenFunc()

			// setup
			pidFile := filepath.Join(tmpDir, "envoy-mock.pid")
			cmdlineFile := filepath.Join(tmpDir, "envoy-mock.cmdline")

			// and
			env := given.envVars
			env["ENVOY_MOCK_PID_FILE"] = pidFile
			env["ENVOY_MOCK_CMDLINE_FILE"] = cmdlineFile
			for key, value := range env {
				Expect(os.Setenv(key, value)).To(Succeed())
			}

			// given
			rootCtx := DefaultRootContext()
			rootCtx.BootstrapGenerator = func(_ string, cfg kumadp.Config, _ envoy.BootstrapParams) (*envoy_bootstrap_v3.Bootstrap, []byte, error) {
				respBytes, err := os.ReadFile(filepath.Join("testdata", "bootstrap-config.golden.yaml"))
				Expect(err).ToNot(HaveOccurred())
				bootstrap := &envoy_bootstrap_v3.Bootstrap{}
				if err := util_proto.FromYAML(respBytes, bootstrap); err != nil {
					return nil, nil, err
				}
				return bootstrap, respBytes, nil
			}

			reader, writer := io.Pipe()
			go func() {
				defer GinkgoRecover()
				_, err := io.ReadAll(reader)
				Expect(err).ToNot(HaveOccurred())
			}()

			cmd := NewRootCmd(opts, rootCtx)
			cmd.SetArgs(append([]string{"run"}, given.args...))
			cmd.SetOut(writer)
			cmd.SetErr(writer)

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
				data, err := os.ReadFile(pidFile)
				if err != nil {
					return false
				}
				pid, err = strconv.ParseInt(strings.TrimSpace(string(data)), 10, 32)
				return err == nil
			}, "5s", "100ms").Should(BeTrue())
			// and
			Expect(pid).ToNot(BeZero())

			By("verifying the arguments Envoy was launched with")
			// when
			cmdline, err := os.ReadFile(cmdlineFile)
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			actualArgs := strings.Split(string(cmdline), "\n")
			Expect(actualArgs[0]).To(Equal("--version"))
			Expect(actualArgs[1]).To(Equal("--config-path"))
			actualConfigFile := actualArgs[2]
			Expect(actualConfigFile).To(BeARegularFile())

			// then
			if given.expectedFile != "" {
				Expect(actualArgs[2]).To(Equal(given.expectedFile))
			}

			// when
			By("signaling the dataplane manager to stop")
			cancel()

			// then
			err = <-errCh
			Expect(err).ToNot(HaveOccurred())

			By("waiting for dataplane (Envoy) to get stopped")
			Eventually(func() bool {
				// send sig 0 to check whether Envoy process still exists
				err := syscall.Kill(int(pid), syscall.Signal(0))
				// we expect Envoy process to get killed by now
				return err != nil
			}, "5s", "100ms").Should(BeTrue())

			By("verifying that temporary configuration dir gets removed")
			if given.expectedFile == "" {
				Expect(actualConfigFile).NotTo(BeAnExistingFile())
			}

			By("verifying that explicit configuration dir is not removed")
			if given.expectedFile != "" {
				Expect(given.expectedFile).To(BeAnExistingFile())
			}
		},
		Entry("can be launched with env vars", func() testCase {
			return testCase{
				envVars: map[string]string{
					"KUMA_CONTROL_PLANE_API_SERVER_URL":  "http://localhost:1234",
					"KUMA_DATAPLANE_NAME":                "example",
					"KUMA_DATAPLANE_MESH":                "default",
					"KUMA_DATAPLANE_ADMIN_PORT":          fmt.Sprintf("%d", port),
					"KUMA_DATAPLANE_RUNTIME_BINARY_PATH": filepath.Join("testdata", "envoy-mock.sleep.sh"),
					// Notice: KUMA_DATAPLANE_RUNTIME_CONFIG_DIR is not set in order to let `kuma-dp` to create a temporary directory
					"KUMA_DNS_CORE_DNS_BINARY_PATH": filepath.Join("testdata", "coredns-mock.sleep.sh"),
				},
				args:         []string{},
				expectedFile: "",
			}
		}),
		Entry("can be launched with env vars and given config dir", func() testCase {
			return testCase{
				envVars: map[string]string{
					"KUMA_CONTROL_PLANE_API_SERVER_URL":  "http://localhost:1234",
					"KUMA_DATAPLANE_NAME":                "example",
					"KUMA_DATAPLANE_MESH":                "default",
					"KUMA_DATAPLANE_ADMIN_PORT":          fmt.Sprintf("%d", port),
					"KUMA_DATAPLANE_RUNTIME_BINARY_PATH": filepath.Join("testdata", "envoy-mock.sleep.sh"),
					"KUMA_DATAPLANE_RUNTIME_CONFIG_DIR":  tmpDir,
					"KUMA_DNS_CORE_DNS_BINARY_PATH":      filepath.Join("testdata", "coredns-mock.sleep.sh"),
				},
				args:         []string{},
				expectedFile: filepath.Join(tmpDir, "bootstrap.yaml"),
			}
		}),
		Entry("can be launched with args", func() testCase {
			return testCase{
				envVars: map[string]string{},
				args: []string{
					"--cp-address", "http://localhost:1234",
					"--name", "example",
					"--mesh", "default",
					"--admin-port", fmt.Sprintf("%d", port),
					"--binary-path", filepath.Join("testdata", "envoy-mock.sleep.sh"),
					// Notice: --config-dir is not set in order to let `kuma-dp` to create a temporary directory
					"--dns-coredns-path", filepath.Join("testdata", "coredns-mock.sleep.sh"),
				},
				expectedFile: "",
			}
		}),
		Entry("can be launched with args and given config dir", func() testCase {
			return testCase{
				envVars: map[string]string{},
				args: []string{
					"--cp-address", "http://localhost:1234",
					"--name", "example",
					"--mesh", "default",
					"--admin-port", fmt.Sprintf("%d", port),
					"--binary-path", filepath.Join("testdata", "envoy-mock.sleep.sh"),
					"--config-dir", tmpDir,
					"--dns-coredns-path", filepath.Join("testdata", "coredns-mock.sleep.sh"),
				},
				expectedFile: filepath.Join(tmpDir, "bootstrap.yaml"),
			}
		}),
		Entry("can be launched with args and dataplane token", func() testCase {
			return testCase{
				envVars: map[string]string{},
				args: []string{
					"--cp-address", "http://localhost:1234",
					"--name", "example",
					"--mesh", "default",
					"--admin-port", fmt.Sprintf("%d", port),
					"--binary-path", filepath.Join("testdata", "envoy-mock.sleep.sh"),
					"--dataplane-token-file", filepath.Join("testdata", "token"),
					// Notice: --config-dir is not set in order to let `kuma-dp` to create a temporary directory
					"--dns-coredns-path", filepath.Join("testdata", "coredns-mock.sleep.sh"),
				},
				expectedFile: "",
			}
		}),
		Entry("can be launched without Envoy Admin API (env vars)", func() testCase {
			return testCase{
				envVars: map[string]string{
					"KUMA_CONTROL_PLANE_API_SERVER_URL":  "http://localhost:1234",
					"KUMA_DATAPLANE_NAME":                "example",
					"KUMA_DATAPLANE_MESH":                "default",
					"KUMA_DATAPLANE_ADMIN_PORT":          "",
					"KUMA_DATAPLANE_RUNTIME_BINARY_PATH": filepath.Join("testdata", "envoy-mock.sleep.sh"),
					// Notice: KUMA_DATAPLANE_RUNTIME_CONFIG_DIR is not set in order to let `kuma-dp` to create a temporary directory
					"KUMA_DNS_CORE_DNS_BINARY_PATH": filepath.Join("testdata", "coredns-mock.sleep.sh"),
				},
				args:         []string{},
				expectedFile: "",
			}
		}),
		Entry("can be launched without Envoy Admin API (command-line args)", func() testCase {
			return testCase{
				envVars: map[string]string{},
				args: []string{
					"--cp-address", "http://localhost:1234",
					"--name", "example",
					"--mesh", "default",
					"--admin-port", "",
					"--binary-path", filepath.Join("testdata", "envoy-mock.sleep.sh"),
					// Notice: --config-dir is not set in order to let `kuma-dp` to create a temporary directory
					"--dns-coredns-path", filepath.Join("testdata", "coredns-mock.sleep.sh"),
				},
				expectedFile: "",
			}
		}),
		Entry("can be launched with dataplane template", func() testCase {
			return testCase{
				envVars: map[string]string{},
				args: []string{
					"--cp-address", "http://localhost:1234",
					"--admin-port", fmt.Sprintf("%d", port),
					"--binary-path", filepath.Join("testdata", "envoy-mock.sleep.sh"),
					"--dataplane-token-file", filepath.Join("testdata", "token"),
					"--dataplane-file", filepath.Join("testdata", "dataplane_template.yaml"),
					"--dataplane-var", "name=example",
					"--dataplane-var", "address=127.0.0.1",
					"--dns-coredns-path", filepath.Join("testdata", "coredns-mock.sleep.sh"),
				},
				expectedFile: "",
			}
		}),
	)

	It("should fail when there are no free ports in the port range chosen for Envoy Admin API", func() {

		By("simulating another Envoy instance that already uses this port")
		// given
		address := fmt.Sprintf("%s:%d", "127.0.0.1", port)
		// when
		l, err := net.Listen("tcp", address)
		// then
		Expect(err).ToNot(HaveOccurred())
		// and
		defer l.Close()

		// given
		cmd := NewRootCmd(opts, DefaultRootContext())
		cmd.SetArgs([]string{
			"run",
			"--cp-address", "http://localhost:1234",
			"--name", "example",
			"--mesh", "default",
			"--admin-port", fmt.Sprintf("%d", port),
			"--binary-path", filepath.Join("testdata", "envoy-mock.sleep.sh"),
			"--dns-coredns-path", filepath.Join("testdata", "coredns-mock.sleep.sh"),
		})

		// when
		err = cmd.Execute()

		// then
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring(fmt.Sprintf(`unable to find a free port in the range "%d" for Envoy Admin API to listen on`, port)))
	})

	It("should fail when name and mesh is provided with dataplane definition", func() {
		// given
		cmd := NewRootCmd(opts, DefaultRootContext())
		cmd.SetArgs([]string{
			"run",
			"--cp-address", "http://localhost:1234",
			"--binary-path", filepath.Join("testdata", "envoy-mock.sleep.sh"),
			"--dataplane-file", filepath.Join("testdata", "dataplane_template.yaml"),
			"--dataplane-var", "name=example",
			"--dataplane-var", "address=127.0.0.1",
			"--name=xyz",
			"--mesh=xyz",
			"--dns-coredns-path", filepath.Join("testdata", "coredns-mock.sleep.sh"),
		})

		// when
		err := cmd.Execute()

		// then
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("--name and --mesh cannot be specified"))
	})

	It("should fail when the proxy type is unknown", func() {
		// given
		cmd := NewRootCmd(opts, DefaultRootContext())
		cmd.SetArgs([]string{
			"run",
			"--cp-address", "http://localhost:1234",
			"--binary-path", filepath.Join("testdata", "envoy-mock.sleep.sh"),
			"--dataplane-file", filepath.Join("testdata", "dataplane_template.yaml"),
			"--dns-coredns-path", filepath.Join("testdata", "coredns-mock.sleep.sh"),
			"--proxy-type", "phoney",
		})

		// when
		err := cmd.Execute()

		// then
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("invalid proxy type"))
	})

})
