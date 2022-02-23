package install_test

import (
	"bytes"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/util/test"
	kuma_version "github.com/kumahq/kuma/pkg/version"
)

var _ = Describe("kumactl install gateway", func() {

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
			rootCmd := test.DefaultTestingRootCmd()
			rootCmd.SetArgs(append([]string{"install", "gateway"}, given.extraArgs...))
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
			extraArgs: []string{
				"kong",
			},
			goldenFile: "install-gateway.defaults.golden.yaml",
		}),
		Entry("should generate Kubernetes resources with custom settings", testCase{
			extraArgs: []string{
				"kong", "--namespace", "notdefault",
			},
			goldenFile: "install-gateway.overrides.golden.yaml",
		}),
		Entry("should generate Kubernetes resources with default settings for enterprise", testCase{
			extraArgs: []string{
				"kong-enterprise", "--license-path", filepath.Join("testdata", "/license"),
			},
			goldenFile: "install-gateway-enterprise.defaults.golden.yaml",
		}),
		Entry("should generate Kubernetes resources with custom settings for enterprise", testCase{
			extraArgs: []string{
				"kong-enterprise", "--license-path", filepath.Join("testdata", "/license"), "--namespace", "notdefault",
			},
			goldenFile: "install-gateway-enterprise.overrides.golden.yaml",
		}),
	)
})
