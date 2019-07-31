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

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core"
)

var _ = Describe("run", func() {

	var backupSetupSignalHandler func() <-chan struct{}

	BeforeEach(func() {
		backupSetupSignalHandler = core.SetupSignalHandler
	})
	AfterEach(func() {
		core.SetupSignalHandler = backupSetupSignalHandler
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
			os.RemoveAll(configDir)
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
			os.Setenv(parts[0], parts[1])
		}
	})

	It("should be possible to start dataplane (Envoy) using `konvoy-dataplane run`", func(done Done) {
		// setup
		pidFile := filepath.Join(configDir, "konvoy-mock.pid")

		// and
		env := map[string]string{
			"KONVOY_CONTROL_PLANE_XDS_SERVER_ADDRESS": "konvoy-control-plane.internal",
			"KONVOY_CONTROL_PLANE_XDS_SERVER_PORT":    "1234",
			"KONVOY_DATAPLANE_ID":                     "example",
			"KONVOY_DATAPLANE_SERVICE":                "demo",
			"KONVOY_DATAPLANE_ADMIN_PORT":             "2345",
			"KONVOY_DATAPLANE_RUNTIME_BINARY_PATH":    filepath.Join("testdata", "envoy-mock.sleep.sh"),
			"KONVOY_DATAPLANE_RUNTIME_CONFIG_DIR":     configDir,
			"TEST_SCRIPT_PID_FILE":                    pidFile,
		}
		for key, value := range env {
			os.Setenv(key, value)
		}

		// given
		cmd := newRootCmd()
		cmd.SetArgs([]string{"run"})
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

		// complete
		close(done)
	}, 10)
})
