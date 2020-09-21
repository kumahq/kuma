package install_test

import (
	"bytes"
	"io/ioutil"
	"path/filepath"

	"github.com/kumahq/kuma/app/kumactl/cmd"
	"github.com/kumahq/kuma/app/kumactl/cmd/install"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/app/kumactl/pkg/install/data"
	"github.com/kumahq/kuma/pkg/tls"
	kuma_version "github.com/kumahq/kuma/pkg/version"
)

var _ = Describe("kumactl install control-plane", func() {

	var backupNewSelfSignedCert func(string, tls.CertType, ...string) (tls.KeyPair, error)
	BeforeEach(func() {
		backupNewSelfSignedCert = install.NewSelfSignedCert
	})
	AfterEach(func() {
		install.NewSelfSignedCert = backupNewSelfSignedCert
	})

	BeforeEach(func() {
		install.NewSelfSignedCert = func(string, tls.CertType, ...string) (tls.KeyPair, error) {
			return tls.KeyPair{
				CertPEM: []byte("CERT"),
				KeyPEM:  []byte("KEY"),
			}, nil
		}
	})

	var backupBuildInfo kuma_version.BuildInfo
	BeforeEach(func() {
		backupBuildInfo = kuma_version.Build
	})
	AfterEach(func() {
		kuma_version.Build = backupBuildInfo
	})

	BeforeEach(func() {
		kuma_version.Build = kuma_version.BuildInfo{
			Version:   "0.0.1",
			GitTag:    "v0.0.1",
			GitCommit: "91ce236824a9d875601679aa80c63783fb0e8725",
			BuildDate: "2019-08-07T11:26:06Z",
		}
		install.DefaultInstallControlPlaneArgs.ControlPlane_image_tag = "0.0.1"
		install.DefaultInstallControlPlaneArgs.DataPlane_image_tag = "0.0.1"
		install.DefaultInstallControlPlaneArgs.DataPlane_initImage_tag = "0.0.1"
	})

	var stdout *bytes.Buffer
	var stderr *bytes.Buffer

	BeforeEach(func() {
		stdout = &bytes.Buffer{}
		stderr = &bytes.Buffer{}
	})

	type testCase struct {
		extraArgs  []string
		goldenFile string
	}

	DescribeTable("should generate Kubernetes resources",
		func(given testCase) {
			// given
			rootCmd := cmd.DefaultRootCmd()
			rootCmd.SetArgs(append([]string{"install", "control-plane"}, given.extraArgs...))
			rootCmd.SetOut(stdout)
			rootCmd.SetErr(stderr)

			// when
			err := rootCmd.Execute()
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(stderr.Bytes()).To(BeNil())

			// when
			expected, err := ioutil.ReadFile(filepath.Join("testdata", given.goldenFile))
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			expectedManifests := data.SplitYAML(data.File{Data: expected})

			// when
			actual := stdout.Bytes()
			// then
			Expect(actual).To(MatchYAML(expected))
			// and
			actualManifests := data.SplitYAML(data.File{Data: actual})

			// and
			Expect(len(actualManifests)).To(Equal(len(expectedManifests)))
			// and
			for i := range expectedManifests {
				Expect(actualManifests[i]).To(MatchYAML(expectedManifests[i]))
			}
		},
		Entry("should generate Kubernetes resources with default settings", testCase{
			extraArgs:  nil,
			goldenFile: "install-control-plane.defaults.golden.yaml",
		}),
		Entry("should generate Kubernetes resources with custom settings", testCase{
			extraArgs: []string{
				"--namespace", "kuma",
				"--image-pull-policy", "Never",
				"--control-plane-version", "greatest",
				"--control-plane-registry", "kuma-ci",
				"--control-plane-service-name", "kuma-ctrl-plane",
				"--injector-failure-policy", "Crash",
				"--dataplane-registry", "kuma-ci",
				"--dataplane-version", "greatest",
				"--dataplane-init-registry", "kuma-ci",
				"--dataplane-init-version", "greatest",
				"--tls-cert", "Cert",
				"--tls-key", "Key",
				"--mode", "remote",
				"--kds-global-address", "grpcs://192.168.0.1:5685",
				"--zone", "zone-1",
				"--use-node-port",
			},
			goldenFile: "install-control-plane.overrides.golden.yaml",
		}),
		Entry("should generate Kubernetes resources with CNI plugin", testCase{
			extraArgs: []string{
				"--cni-enabled",
			},
			goldenFile: "install-control-plane.cni-enabled.golden.yaml",
		}),
		Entry("should generate Kubernetes resources for Global", testCase{
			extraArgs: []string{
				"--mode", "global",
			},
			goldenFile: "install-control-plane.global.golden.yaml",
		}),
		Entry("should generate Kubernetes resources for Remote", testCase{
			extraArgs: []string{
				"--mode", "remote",
				"--zone", "zone-1",
				"--kds-global-address", "grpcs://192.168.0.1:5685",
			},
			goldenFile: "install-control-plane.remote.golden.yaml",
		}),
	)

	type errTestCase struct {
		extraArgs []string
		errorMsg  string
	}
	DescribeTable("should fail to install control plane",
		func(given errTestCase) {
			// given
			rootCmd := cmd.DefaultRootCmd()
			rootCmd.SetArgs(append([]string{"install", "control-plane"}, given.extraArgs...))
			rootCmd.SetOut(stdout)
			rootCmd.SetErr(stderr)

			//when
			err := rootCmd.Execute()

			// then
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(given.errorMsg))
		},
		Entry("--mode is unknown", errTestCase{
			extraArgs: []string{"--mode", "test"},
			errorMsg:  "invalid mode. Available modes: standalone, remote, global",
		}),
		Entry("--kds-global-address is missing when installing remote", errTestCase{
			extraArgs: []string{"--mode", "remote", "--zone", "zone-1"},
			errorMsg:  "--kds-global-address is mandatory with `remote` mode",
		}),
		Entry("--kds-global-address is not valid URL", errTestCase{
			extraArgs: []string{"--kds-global-address", "192.168.0.1:1234", "--mode", "remote", "--zone", "zone-1"},
			errorMsg:  "--kds-global-address is not valid URL. The allowed format is grpcs://hostname:port",
		}),
		Entry("--kds-global-address has no grpcs scheme", errTestCase{
			extraArgs: []string{"--kds-global-address", "http://192.168.0.1:1234", "--mode", "remote", "--zone", "zone-1"},
			errorMsg:  "--kds-global-address should start with grpcs://",
		}),
		Entry("--kds-global-address is used with standalone", errTestCase{
			extraArgs: []string{"--kds-global-address", "192.168.0.1:1234", "--mode", "standalone"},
			errorMsg:  "--kds-global-address can only be used when --mode=remote",
		}),
		Entry("--tls-cert without --tls-key", errTestCase{
			extraArgs: []string{"--tls-cert", "cert.pem"},
			errorMsg:  "both --tls-cert and --tls-key must be provided at the same time",
		}),
		Entry("--tls-key without --tls-cert", errTestCase{
			extraArgs: []string{"--tls-key", "key.pem"},
			errorMsg:  "both --tls-cert and --tls-key must be provided at the same time",
		}),
	)
})
