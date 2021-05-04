package install_test

import (
	"bytes"
	"path/filepath"

	kuma_version "github.com/kumahq/kuma/pkg/version"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/app/kumactl/cmd"
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

	type testCaseErr struct {
		extraArgs   []string
		expectedErr string
	}

	BeforeEach(func() {
		kuma_version.Build = kuma_version.BuildInfo{
			Version:   "0.0.1",
			GitTag:    "v0.0.1",
			GitCommit: "91ce236824a9d875601679aa80c63783fb0e8725",
			BuildDate: "2019-08-07T11:26:06Z",
		}
	})

	DescribeTable("should generate error",
		func(given testCaseErr) {
			// given
			rootCmd := cmd.DefaultRootCmd()
			rootCmd.SetArgs(append([]string{"install", "gateway"}, given.extraArgs...))
			rootCmd.SetOut(stdout)
			rootCmd.SetErr(stderr)

			// when
			err := rootCmd.Execute()

			// then
			Expect(err.Error()).To(Equal(given.expectedErr))
		},
		Entry("should fail due to lack of type", testCaseErr{
			extraArgs:   nil,
			expectedErr: "required flag(s) \"type\" not set",
		}),
		Entry("should fail due to invalid type", testCaseErr{
			extraArgs: []string{
				"--type", "invalidtype",
			},
			expectedErr: "Only gateway type 'kong' currently supported",
		}),
	)

	DescribeTable("should generate Kubernetes resources",
		func(given testCase) {
			// given
			rootCmd := cmd.DefaultRootCmd()
			rootCmd.SetArgs(append([]string{"install", "gateway"}, given.extraArgs...))
			rootCmd.SetOut(stdout)
			rootCmd.SetErr(stderr)

			// when
			err := rootCmd.Execute()

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(stderr.Bytes()).To(BeNil())

			// and output matches golden files
			actual := stdout.Bytes()
			ExpectMatchesGoldenFiles(actual, filepath.Join("testdata", given.goldenFile))
		},
		Entry("should generate Kubernetes resources with default settings", testCase{
			extraArgs: []string{
				"--type", "kong",
			},
			goldenFile: "install-gateway.defaults.golden.yaml",
		}),
		Entry("should generate Kubernetes resources with custom settings", testCase{
			extraArgs: []string{
				"--type", "kong", "--namespace", "notdefault",
			},
			goldenFile: "install-gateway.overrides.golden.yaml",
		}),
	)
})
