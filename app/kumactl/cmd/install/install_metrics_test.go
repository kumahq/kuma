package install_test

import (
	"bytes"
	"io/ioutil"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/app/kumactl/cmd"
	"github.com/kumahq/kuma/app/kumactl/cmd/install"
	"github.com/kumahq/kuma/app/kumactl/pkg/install/data"
	kuma_version "github.com/kumahq/kuma/pkg/version"
)

var _ = Describe("kumactl install metrics", func() {

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

	BeforeEach(func() {
		kuma_version.Build = kuma_version.BuildInfo{
			Version:   "0.0.1",
			GitTag:    "v0.0.1",
			GitCommit: "91ce236824a9d875601679aa80c63783fb0e8725",
			BuildDate: "2019-08-07T11:26:06Z",
		}
		install.DefaultMetricsTemplateArgs.KumaPrometheusSdVersion = "0.0.1"
	})

	DescribeTable("should generate Kubernetes resources",
		func(given testCase) {
			// given
			rootCmd := cmd.DefaultRootCmd()
			rootCmd.SetArgs(append([]string{"install", "metrics"}, given.extraArgs...))
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
			goldenFile: "install-metrics.defaults.golden.yaml",
		}),
		Entry("should generate Kubernetes resources with custom settings", testCase{
			extraArgs: []string{
				"--namespace", "kuma",
				"--mesh", "mesh-1",
				"--kuma-prometheus-sd-image", "kuma-ci/kuma-prometheus-sd",
				"--kuma-prometheus-sd-version", "greatest",
				"--kuma-cp-address", "http://kuma.local:5681",
			},
			goldenFile: "install-metrics.overrides.golden.yaml",
		}),
		Entry("should generate Kubernetes resources without prometheus", testCase{
			extraArgs: []string{
				"--without-prometheus",
			},
			goldenFile: "install-metrics.no-prometheus.golden.yaml",
		}),
		Entry("should generate Kubernetes resources without grafana", testCase{
			extraArgs: []string{
				"--without-grafana",
			},
			goldenFile: "install-metrics.no-grafana.golden.yaml",
		}),
	)
})
