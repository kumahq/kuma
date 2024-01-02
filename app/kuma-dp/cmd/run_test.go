//go:build !windows
// +build !windows

package cmd

import (
	"context"
	"fmt"
	"io"
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
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/bootstrap/types"
)

var _ = Describe("run", func() {
	var cancel func()
	var ctx context.Context
	opts := kuma_cmd.RunCmdOpts{
		SetupSignalHandler: func() (context.Context, context.Context) {
			return ctx, ctx
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

	type testCase struct {
		envVars      map[string]string
		args         []string
		expectedFile string
	}
	DescribeTable("should be possible to start dataplane (Envoy) using `kuma-dp run`",
		func(givenFunc func() testCase) {
			given := givenFunc()

			// setup
			envoyPidFile := filepath.Join(tmpDir, "envoy-mock.pid")
			envoyCmdlineFile := filepath.Join(tmpDir, "envoy-mock.cmdline")
			corednsPidFile := filepath.Join(tmpDir, "coredns-mock.pid")
			corednsCmdlineFile := filepath.Join(tmpDir, "coredns-mock.cmdline")

			// and
			env := given.envVars
			env["ENVOY_MOCK_PID_FILE"] = envoyPidFile
			env["ENVOY_MOCK_CMDLINE_FILE"] = envoyCmdlineFile
			env["COREDNS_MOCK_PID_FILE"] = corednsPidFile
			env["COREDNS_MOCK_CMDLINE_FILE"] = corednsCmdlineFile
			for key, value := range env {
				Expect(os.Setenv(key, value)).To(Succeed())
			}

			// given
			rootCtx := DefaultRootContext()
			rootCtx.BootstrapGenerator = func(_ context.Context, _ string, cfg kumadp.Config, _ envoy.BootstrapParams) (*envoy_bootstrap_v3.Bootstrap, *types.KumaSidecarConfiguration, error) {
				respBytes, err := os.ReadFile(filepath.Join("testdata", "bootstrap-config.golden.yaml"))
				Expect(err).ToNot(HaveOccurred())
				bootstrap := &envoy_bootstrap_v3.Bootstrap{}
				if err := util_proto.FromYAML(respBytes, bootstrap); err != nil {
					return nil, nil, err
				}
				return bootstrap, &types.KumaSidecarConfiguration{}, nil
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
			By("starting the Kuma DP")
			errCh := make(chan error)
			go func() {
				defer close(errCh)
				errCh <- cmd.Execute()
			}()

			// then
			var actualConfigFile string
			envoyPid := verifyComponentProcess("Envoy", envoyPidFile, envoyCmdlineFile, func(actualArgs []string) {
				Expect(actualArgs[0]).To(Equal("--version"))
				Expect(actualArgs[1]).To(Equal("--config-path"))
				actualConfigFile = actualArgs[2]
				Expect(actualConfigFile).To(BeARegularFile())
				if given.expectedFile != "" {
					Expect(actualArgs[2]).To(Equal(given.expectedFile))
				}
			})

			corednsPid := verifyComponentProcess("coredns", corednsPidFile, corednsCmdlineFile, func(actualArgs []string) {
				Expect(actualArgs).To(HaveLen(3))
				Expect(actualArgs[0]).To(Equal("-conf"))
				Expect(actualArgs[2]).To(Equal("-quiet"))
			})

			// when
			By("signaling the dataplane manager to stop")
			// we need to close writer, otherwise Cmd#Wait will never finish.
			Expect(writer.Close()).To(Succeed())
			cancel()

			// then
			err := <-errCh
			Expect(err).ToNot(HaveOccurred())

			By("waiting for dataplane (Envoy) to get stopped")
			Eventually(func() bool {
				// send sig 0 to check whether Envoy process still exists
				err := syscall.Kill(int(envoyPid), syscall.Signal(0))
				// we expect Envoy process to get killed by now
				return err != nil
			}, "5s", "100ms").Should(BeTrue())
			By("waiting for dataplane (coredns) to get stopped")
			Eventually(func() bool {
				// send sig 0 to check whether Envoy process still exists
				err := syscall.Kill(int(corednsPid), syscall.Signal(0))
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
		Entry("can be launched with given coredns configuration path", func() testCase {
			corefileTemplate := filepath.Join(tmpDir, "Corefile")
			_ = os.WriteFile(corefileTemplate, []byte("abcd"), 0o600)
			return testCase{
				envVars: map[string]string{
					"KUMA_CONTROL_PLANE_API_SERVER_URL":      "http://localhost:1234",
					"KUMA_DATAPLANE_NAME":                    "example",
					"KUMA_DATAPLANE_MESH":                    "default",
					"KUMA_DATAPLANE_RUNTIME_BINARY_PATH":     filepath.Join("testdata", "envoy-mock.sleep.sh"),
					"KUMA_DATAPLANE_RUNTIME_CONFIG_DIR":      tmpDir,
					"KUMA_DNS_CORE_DNS_BINARY_PATH":          filepath.Join("testdata", "coredns-mock.sleep.sh"),
					"KUMA_DNS_CORE_DNS_CONFIG_TEMPLATE_PATH": corefileTemplate,
				},
				args:         []string{},
				expectedFile: filepath.Join(tmpDir, "bootstrap.yaml"),
			}
		}),
	)

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
}, Ordered)

func verifyComponentProcess(processDescription, pidfile string, cmdlinefile string, argsVerifier func(expectedArgs []string)) int64 {
	var pid int64
	By(fmt.Sprintf("waiting for dataplane (%s) to get started", processDescription))
	Eventually(func() bool {
		data, err := os.ReadFile(pidfile)
		if err != nil {
			return false
		}
		pid, err = strconv.ParseInt(strings.TrimSpace(string(data)), 10, 32)
		return err == nil
	}, "5s", "100ms").Should(BeTrue())
	Expect(pid).ToNot(BeZero())

	By(fmt.Sprintf("verifying the arguments %s was launched with", processDescription))
	// when
	cmdline, err := os.ReadFile(cmdlinefile)

	// then
	Expect(err).ToNot(HaveOccurred())
	// and
	if argsVerifier != nil {
		actualArgs := strings.FieldsFunc(string(cmdline), func(c rune) bool {
			return c == '\n'
		})
		argsVerifier(actualArgs)
	}
	return pid
}
