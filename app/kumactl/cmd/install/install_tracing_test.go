package install_test

import (
	"bytes"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/app/kumactl/cmd"
)

var _ = Describe("kumactl install tracing", func() {

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
			rootCmd.SetArgs(append([]string{"install", "tracing"}, given.extraArgs...))
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
			goldenFile: "install-tracing.defaults.golden.yaml",
		}),
		Entry("should generate Kubernetes resources with custom settings", testCase{
			extraArgs: []string{
				"--namespace", "kuma",
			},
			goldenFile: "install-tracing.overrides.golden.yaml",
		}),
	)
})
