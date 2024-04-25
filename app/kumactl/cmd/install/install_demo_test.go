package install_test

import (
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/app/kumactl/pkg/test"
)

var _ = Context("kumactl install demo", func() {
	type testCase struct {
		extraArgs  []string
		goldenFile string
	}

	DescribeTable("should generate Kubernetes resources",
		func(given testCase) {
			// given
			args := append([]string{"install", "demo"}, given.extraArgs...)
			stdout, stderr, rootCmd := test.DefaultTestingRootCmd(args...)

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
			goldenFile: "install-demo.defaults.golden.yaml",
		}),
		Entry("should generate Kubernetes resources with custom settings", testCase{
			extraArgs: []string{
				"--zone", "aws", "--namespace", "not-kuma-demo",
			},
			goldenFile: "install-demo.overrides.golden.yaml",
		}),
		Entry("should respect --without-gateway", testCase{
			extraArgs: []string{
				"--without-gateway",
			},
			goldenFile: "install-demo.without-gateway.golden.yaml",
		}),
	)
})
