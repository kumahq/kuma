package install_test

import (
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v2/app/kumactl/pkg/test"
)

var _ = Context("kumactl install observability", func() {
	type testCase struct {
		extraArgs  []string
		goldenFile string
	}

	DescribeTable("should generate Kubernetes resources",
		func(given testCase) {
			// given
			args := append([]string{"install", "observability"}, given.extraArgs...)
			stdout, stderr, rootCmd := test.DefaultTestingRootCmd(args...)

			// when
			err := rootCmd.Execute()

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(stderr.String()).To(ContainSubstring("Warning: 'kumactl install observability' is deprecated and will be removed in Kuma 3.0."))

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
