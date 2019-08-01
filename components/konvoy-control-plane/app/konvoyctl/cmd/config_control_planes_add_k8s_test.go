package cmd

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/spf13/cobra"

	konvoyctl_k8s "github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/pkg/k8s"
	konvoyctl_k8s_fake "github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/pkg/k8s/fake"
)

var _ = Describe("konvoy config control-planes add k8s", func() {

	// overridden package variables
	var backupDetectKubeConfig func() (konvoyctl_k8s.KubeConfig, error)

	BeforeEach(func() {
		backupDetectKubeConfig = detectKubeConfig
	})

	AfterEach(func() {
		detectKubeConfig = backupDetectKubeConfig
	})

	var configFile *os.File

	BeforeEach(func() {
		var err error
		configFile, err = ioutil.TempFile("", "")
		Expect(err).ToNot(HaveOccurred())
	})
	AfterEach(func() {
		if configFile != nil {
			os.Remove(configFile.Name())
		}
	})

	var rootCmd *cobra.Command
	var buf *bytes.Buffer

	BeforeEach(func() {
		rootCmd = defaultRootCmd()
		buf = &bytes.Buffer{}
		rootCmd.SetOut(buf)
	})

	Describe("error cases", func() {

		It("should require name", func() {
			// given
			rootCmd.SetArgs([]string{"--config-file", configFile.Name(),
				"config", "control-planes", "add", "k8s"})
			// when
			err := rootCmd.Execute()
			// then
			Expect(err.Error()).To(MatchRegexp(requiredFlagNotSet("name")))
		})

		It("should require namespace to be non-empty", func() {
			// given
			rootCmd.SetArgs([]string{"--config-file", configFile.Name(),
				"config", "control-planes", "add", "k8s",
				"--name", "example",
				"--namespace", ""})
			// when
			err := rootCmd.Execute()
			// then
			Expect(err.Error()).To(ContainSubstring(`flag "namespace" must have a non-empty value`))
		})

		Describe("should fail if kubernetes environment is incomplete", func() {

			BeforeEach(func() {
				// setup
				rootCmd.SetArgs([]string{"--config-file", configFile.Name(),
					"config", "control-planes", "add", "k8s",
					"--name", "example"})
			})

			It("should fail if kubectl config not found", func() {
				// given
				detectKubeConfig = func() (konvoyctl_k8s.KubeConfig, error) {
					return nil, errors.New("kubectl config not found")
				}
				// when
				err := rootCmd.Execute()
				// then
				Expect(err.Error()).To(ContainSubstring("Failed to detect current `kubectl` context"))
			})

			It("should fail if current kubectl context is not set", func() {
				// setup
				detectKubeConfig = func() (konvoyctl_k8s.KubeConfig, error) {
					return &konvoyctl_k8s_fake.FakeKubeConfig{
						CurrentContext: "",
					}, nil
				}
				// given
				rootCmd.SetArgs([]string{"--config-file", configFile.Name(),
					"config", "control-planes", "add", "k8s",
					"--name", "example"})
				// when
				err := rootCmd.Execute()
				// then
				Expect(err.Error()).To(ContainSubstring("Failed to detect current `kubectl` context"))
			})

			It("should fail if current kubectl context is not set", func() {
				// setup
				detectKubeConfig = func() (konvoyctl_k8s.KubeConfig, error) {
					return &konvoyctl_k8s_fake.FakeKubeConfig{
						CurrentContext: "minikube",
						ClientErr:      errors.New("Failed to connect to Kubernetes API Server"),
					}, nil
				}
				// given
				rootCmd.SetArgs([]string{"--config-file", configFile.Name(),
					"config", "control-planes", "add", "k8s",
					"--name", "example"})
				// when
				err := rootCmd.Execute()
				// then
				Expect(err.Error()).To(ContainSubstring("Failed to connect to a target Kubernetes cluster (`kubectl` context \"minikube\")"))
			})

			It("should fail if can't get kubernetes Namespace", func() {
				// setup
				detectKubeConfig = func() (konvoyctl_k8s.KubeConfig, error) {
					return &konvoyctl_k8s_fake.FakeKubeConfig{
						CurrentContext: "minikube",
						Client: &konvoyctl_k8s_fake.FakeClient{
							NamespaceExistsErr: errors.New("failed to get Namespace"),
						},
					}, nil
				}
				// given
				rootCmd.SetArgs([]string{"--config-file", configFile.Name(),
					"config", "control-planes", "add", "k8s",
					"--name", "example"})
				// when
				err := rootCmd.Execute()
				// then
				Expect(err.Error()).To(ContainSubstring("Failed to determine whether a target Kubernetes cluster (`kubectl` context \"minikube\") has \"konvoy-system\" namespace"))
			})

			It("should fail if kubernetes Namespace doesn't exists", func() {
				// setup
				detectKubeConfig = func() (konvoyctl_k8s.KubeConfig, error) {
					return &konvoyctl_k8s_fake.FakeKubeConfig{
						CurrentContext: "minikube",
						Client: &konvoyctl_k8s_fake.FakeClient{
							NamespaceExists: false,
						},
					}, nil
				}
				// given
				rootCmd.SetArgs([]string{"--config-file", configFile.Name(),
					"config", "control-planes", "add", "k8s",
					"--name", "example"})
				// when
				err := rootCmd.Execute()
				// then
				Expect(err.Error()).To(ContainSubstring("There is no Control Plane installed on a target Kubernetes cluster (`kubectl` context \"minikube\")"))
			})
		})
	})

	Describe("happy path", func() {

		type testCase struct {
			extraArgs  []string
			configFile string
			goldenFile string
		}

		DescribeTable("should add a new Control Plane by name",
			func(given testCase) {
				// setup
				detectKubeConfig = func() (konvoyctl_k8s.KubeConfig, error) {
					return &konvoyctl_k8s_fake.FakeKubeConfig{
						Filename:       "/home/user/.kube/config",
						CurrentContext: "minikube",
						Client: &konvoyctl_k8s_fake.FakeClient{
							NamespaceExists: true,
						},
					}, nil
				}

				// setup
				initial, err := ioutil.ReadFile(filepath.Join("testdata", given.configFile))
				Expect(err).ToNot(HaveOccurred())
				err = ioutil.WriteFile(configFile.Name(), initial, 0600)
				Expect(err).ToNot(HaveOccurred())

				// given
				rootCmd.SetArgs(append([]string{"--config-file", configFile.Name(),
					"config", "control-planes", "add", "k8s",
					"--name", "example"}, given.extraArgs...))

				// when
				err = rootCmd.Execute()
				// then
				Expect(err).ToNot(HaveOccurred())

				// when
				expected, err := ioutil.ReadFile(filepath.Join("testdata", given.goldenFile))
				// then
				Expect(err).ToNot(HaveOccurred())

				// when
				actual, err := ioutil.ReadFile(configFile.Name())
				// then
				Expect(err).ToNot(HaveOccurred())

				// and
				Expect(actual).To(MatchYAML(expected))
			},
			Entry("should add the first Control Plane", testCase{
				configFile: "config-ontrol-planes-add-k8s.01.initial.yaml",
				goldenFile: "config-ontrol-planes-add-k8s.01.golden.yaml",
			}),
			Entry("should add the second Control Plane", testCase{
				configFile: "config-ontrol-planes-add-k8s.02.initial.yaml",
				goldenFile: "config-ontrol-planes-add-k8s.02.golden.yaml",
			}),
			Entry("should add a Control Plane in a non-default namespace", testCase{
				extraArgs:  []string{"--namespace", "konvoy"},
				configFile: "config-ontrol-planes-add-k8s.03.initial.yaml",
				goldenFile: "config-ontrol-planes-add-k8s.03.golden.yaml",
			}),
		)
	})
})
