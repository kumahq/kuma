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

var _ = Describe("kumactl install observability", func() {
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
			rootCmd.SetArgs(append([]string{"install", "observability"}, given.extraArgs...))
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
			goldenFile: "install-observability.defaults.golden.yaml",
		}),
		Entry("should generate Kubernetes resources with custom settings", testCase{
			extraArgs: []string{
				"--namespace", "kuma",
				"--mesh", "mesh-1",
				"--kuma-cp-address", "http://kuma.local:5681",
			},
			goldenFile: "install-observability.overrides.golden.yaml",
		}),
		Entry("should generate Kubernetes resources without prometheus", testCase{
			extraArgs: []string{
				"--components", "grafana,jaeger,loki",
			},
			goldenFile: "install-observability.no-prometheus.golden.yaml",
		}),
		Entry("should generate Kubernetes resources without grafana", testCase{
			extraArgs: []string{
				"--components", "prometheus,jaeger,loki",
			},
			goldenFile: "install-observability.no-grafana.golden.yaml",
		}),
		Entry("should generate Kubernetes resources without loki", testCase{
			extraArgs: []string{
				"--components", "prometheus,jaeger,grafana",
			},
			goldenFile: "install-observability.no-loki.golden.yaml",
		}),
		Entry("should generate Kubernetes resources without jaeger", testCase{
			extraArgs: []string{
				"--components", "prometheus,grafana,loki",
			},
			goldenFile: "install-observability.no-jaeger.golden.yaml",
		}),
	)
})
