// +build !windows

package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/Kong/kuma/pkg/catalog"

	"github.com/Kong/kuma/app/kuma-dp/pkg/dataplane/envoy"
	catalog_client "github.com/Kong/kuma/pkg/catalog/client"
	kumadp "github.com/Kong/kuma/pkg/config/app/kuma-dp"
	test_catalog "github.com/Kong/kuma/pkg/test/catalog"
	util_proto "github.com/Kong/kuma/pkg/util/proto"
	envoy_bootstrap "github.com/envoyproxy/go-control-plane/envoy/config/bootstrap/v2"
	"github.com/golang/protobuf/proto"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/Kong/kuma/pkg/core"

	"github.com/Kong/kuma/pkg/test"
)

var _ = Describe("run", func() {

	var backupSetupSignalHandler func() <-chan struct{}
	var backupBootstrapGenerator envoy.BootstrapConfigFactoryFunc
	var backupCatalogClientFactory CatalogClientFactory

	catalogDataplaneTokenServerEnabledFn := func(address string) (client catalog_client.CatalogClient, e error) {
		return &test_catalog.StaticCatalogClient{
			Resp: catalog.Catalog{
				Apis: catalog.Apis{
					Bootstrap: catalog.BootstrapApi{
						Url: "http://localhost:5681",
					},
					DataplaneToken: catalog.DataplaneTokenApi{
						LocalUrl: "http://localhost:5683",
					},
				},
			},
		}, nil
	}

	BeforeEach(func() {
		backupSetupSignalHandler = core.SetupSignalHandler
		backupBootstrapGenerator = bootstrapGenerator
		backupCatalogClientFactory = catalogClientFactory
		bootstrapGenerator = func(_ string, cfg kumadp.Config) (proto.Message, error) {
			bootstrap := envoy_bootstrap.Bootstrap{}
			respBytes, err := ioutil.ReadFile(filepath.Join("testdata", "bootstrap-config.golden.yaml"))
			Expect(err).ToNot(HaveOccurred())
			err = util_proto.FromYAML(respBytes, &bootstrap)
			Expect(err).ToNot(HaveOccurred())
			return &bootstrap, nil
		}
		catalogClientFactory = func(address string) (client catalog_client.CatalogClient, e error) {
			return &test_catalog.StaticCatalogClient{
				Resp: catalog.Catalog{
					Apis: catalog.Apis{
						Bootstrap: catalog.BootstrapApi{
							Url: "http://localhost:5681",
						},
						DataplaneToken: catalog.DataplaneTokenApi{
							LocalUrl: "", // dataplane token is disabled
						},
					},
				},
			}, nil
		}
	})
	AfterEach(func() {
		core.SetupSignalHandler = backupSetupSignalHandler
		bootstrapGenerator = backupBootstrapGenerator
		catalogClientFactory = backupCatalogClientFactory
	})

	var stopCh chan struct{}

	BeforeEach(func() {
		stopCh = make(chan struct{})

		core.SetupSignalHandler = func() <-chan struct{} {
			return stopCh
		}
	})

	var tmpDir string

	BeforeEach(func() {
		var err error
		tmpDir, err = ioutil.TempDir("", "")
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

			By("verifying the arguments Envoy was launched with")
			// when
			cmdline, err := ioutil.ReadFile(cmdlineFile)
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			actualArgs := strings.Split(string(cmdline), "\n")
			Expect(actualArgs[0]).To(Equal("-c"))
			actualConfigFile := actualArgs[1]
			Expect(actualConfigFile).To(BeARegularFile())

			// then
			if given.expectedFile != "" {
				Expect(actualArgs[1]).To(Equal(given.expectedFile))
			}

			// when
			By("signalling the dataplane manager to stop")
			close(stopCh)

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
					"KUMA_DATAPLANE_ADMIN_PORT":          fmt.Sprintf("%d", port),
					"KUMA_DATAPLANE_RUNTIME_BINARY_PATH": filepath.Join("testdata", "envoy-mock.sleep.sh"),
					// Notice: KUMA_DATAPLANE_RUNTIME_CONFIG_DIR is not set in order to let `kuma-dp` to create a temporary directory
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
					"KUMA_DATAPLANE_ADMIN_PORT":          fmt.Sprintf("%d", port),
					"KUMA_DATAPLANE_RUNTIME_BINARY_PATH": filepath.Join("testdata", "envoy-mock.sleep.sh"),
					"KUMA_DATAPLANE_RUNTIME_CONFIG_DIR":  tmpDir,
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
					"--admin-port", fmt.Sprintf("%d", port),
					"--binary-path", filepath.Join("testdata", "envoy-mock.sleep.sh"),
					// Notice: --config-dir is not set in order to let `kuma-dp` to create a temporary directory
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
					"--admin-port", fmt.Sprintf("%d", port),
					"--binary-path", filepath.Join("testdata", "envoy-mock.sleep.sh"),
					"--config-dir", tmpDir,
				},
				expectedFile: filepath.Join(tmpDir, "bootstrap.yaml"),
			}
		}),
		Entry("can be launched with args and dataplane token", func() testCase {
			catalogClientFactory = catalogDataplaneTokenServerEnabledFn
			return testCase{
				envVars: map[string]string{},
				args: []string{
					"--cp-address", "http://localhost:1234",
					"--name", "example",
					"--admin-port", fmt.Sprintf("%d", port),
					"--binary-path", filepath.Join("testdata", "envoy-mock.sleep.sh"),
					"--dataplane-token-file", filepath.Join("testdata", "token"),
					// Notice: --config-dir is not set in order to let `kuma-dp` to create a temporary directory
				},
				expectedFile: "",
			}
		}),
		Entry("can be launched without Envoy Admin API (env vars)", func() testCase {
			return testCase{
				envVars: map[string]string{
					"KUMA_CONTROL_PLANE_API_SERVER_URL":  "http://localhost:1234",
					"KUMA_DATAPLANE_NAME":                "example",
					"KUMA_DATAPLANE_ADMIN_PORT":          "",
					"KUMA_DATAPLANE_RUNTIME_BINARY_PATH": filepath.Join("testdata", "envoy-mock.sleep.sh"),
					// Notice: KUMA_DATAPLANE_RUNTIME_CONFIG_DIR is not set in order to let `kuma-dp` to create a temporary directory
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
					"--admin-port", "",
					"--binary-path", filepath.Join("testdata", "envoy-mock.sleep.sh"),
					// Notice: --config-dir is not set in order to let `kuma-dp` to create a temporary directory
				},
				expectedFile: "",
			}
		}),
	)

	It("should fail when dataplane token server is enabled but token is not provided", func() {
		// setup
		catalogClientFactory = catalogDataplaneTokenServerEnabledFn

		// given
		cmd := newRootCmd()
		cmd.SetArgs([]string{
			"run",
			"--cp-address", "http://localhost:1234",
			"--name", "example",
			"--admin-port", fmt.Sprintf("%d", port),
			"--binary-path", filepath.Join("testdata", "envoy-mock.sleep.sh"),
		})

		// when
		err := cmd.Execute()

		// then
		Expect(err).To(HaveOccurred())
		Expect(err).To(MatchError("Kuma CP is configured with Dataplane Token Server therefore the Dataplane Token is required. Generate token using 'kumactl generate dataplane-token > /path/file' and provide it via --dataplane-token-file=/path/file argument to Kuma DP"))
	})

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
		cmd := newRootCmd()
		cmd.SetArgs([]string{
			"run",
			"--cp-address", "http://localhost:1234",
			"--name", "example",
			"--admin-port", fmt.Sprintf("%d", port),
			"--binary-path", filepath.Join("testdata", "envoy-mock.sleep.sh"),
		})

		// when
		err = cmd.Execute()

		// then
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring(fmt.Sprintf(`unable to find a free port in the range "%d" for Envoy Admin API to listen on`, port)))
	})
})
