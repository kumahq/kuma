package uninstall_test

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/util/test"
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

	DescribeTable("should install transparent proxy",
		func(given testCase) {
			// given
			rootCmd := test.DefaultTestingRootCmd()
			rootCmd.SetArgs(append([]string{"uninstall", "transparent-proxy", "--dry-run"}, given.extraArgs...))
			rootCmd.SetOut(stdout)
			rootCmd.SetErr(stderr)

			// when
			err := rootCmd.Execute()
			// then
			Expect(err).To(HaveOccurred())
			// and
			Expect(stderr.String()).To(Equal("Error: transparent proxy cleanup failed: cleanup is not supported\n"))

			// TODO once delete works again to something similar to what we do for `install_transparent_proxy_test.go` with Transform.
		},
		Entry("should generate defaults with username", testCase{
			extraArgs:  nil,
			goldenFile: "uninstall-transparent-proxy.defaults.golden.txt",
		}),
	)
})
