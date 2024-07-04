package uninstall_test

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	test_kumactl "github.com/kumahq/kuma/app/kumactl/pkg/test"
)

var _ = Describe("kumactl uninstall transparent-proxy", func() {
	type testCase struct {
		extraArgs  []string
		goldenFile string
	}

	DescribeTable("should uninstall transparent proxy",
		func(given testCase) {
			// given
			args := append([]string{"uninstall", "transparent-proxy", "--dry-run"}, given.extraArgs...)
			_, stderr, rootCmd := test_kumactl.DefaultTestingRootCmd(args...)

			// when
			err := rootCmd.Execute()
			// then
			Expect(err).To(HaveOccurred())
			// and
			Expect(stderr.String()).To(WithTransform(func(in string) string {
				return strings.ReplaceAll(
					in,
					"# [WARNING]: dry-run mode: No valid iptables executables found. The generated iptables rules may differ from those generated in an environment with valid iptables executables\n", "",
				)
			}, Equal("Error: transparent proxy cleanup failed: cleanup is not supported\n")))

			// TODO once delete works again to something similar to what we do for `install_transparent_proxy_test.go` with Transform.
		},
		Entry("should generate defaults with username", testCase{
			extraArgs:  nil,
			goldenFile: "uninstall-transparent-proxy.defaults.golden.txt",
		}),
	)
})
