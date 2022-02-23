package install_test

import (
	"bytes"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/app/kumactl/cmd"
	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	"github.com/kumahq/kuma/pkg/util/test"
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
	})

	DescribeTable("should generate Kubernetes resources",
		func(given testCase) {
			// given
			rootCtx := kumactl_cmd.DefaultRootContext()
			rootCtx.Runtime.NewAPIServerClient = test.GetMockNewAPIServerClient()
			rootCmd := cmd.NewRootCmd(rootCtx)
			rootCmd.SetArgs(append([]string{"install", "metrics"}, given.extraArgs...))
			rootCmd.SetOut(stdout)
			rootCmd.SetErr(stderr)

			// when
			err := rootCmd.Execute()

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(stderr.String()).To(BeEmpty())

			// and output matches golden files
			actual := stdout.Bytes()
			ExpectMatchesGoldenFiles(actual, filepath.Join("testdata", given.goldenFile))
		},
		Entry("should generate Kubernetes resources with default settings", testCase{
			extraArgs:  nil,
			goldenFile: "install-metrics.defaults.golden.yaml",
		}),
		Entry("should generate Kubernetes resources with custom settings", testCase{
			extraArgs: []string{
				"--namespace", "kuma",
				"--mesh", "mesh-1",
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
