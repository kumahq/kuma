package uninstall_test

import (
<<<<<<< HEAD
	"bytes"
	"os"
	"path/filepath"
	"regexp"
=======
	"strings"
>>>>>>> f732b34e9 (refactor(transparent-proxy): move executables to config (#10619))

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/util/test"
)

<<<<<<< HEAD
var _ = Describe("kumactl install tracing", func() {
	var stdout *bytes.Buffer
	var stderr *bytes.Buffer

	BeforeEach(func() {
		stdout = &bytes.Buffer{}
		stderr = &bytes.Buffer{}
	})

=======
var _ = Describe("kumactl uninstall transparent-proxy", func() {
>>>>>>> f732b34e9 (refactor(transparent-proxy): move executables to config (#10619))
	type testCase struct {
		extraArgs  []string
		goldenFile string
	}

	DescribeTable("should uninstall transparent proxy",
		func(given testCase) {
			// given
			rootCmd := test.DefaultTestingRootCmd()
			rootCmd.SetArgs(append([]string{"uninstall", "transparent-proxy", "--dry-run"}, given.extraArgs...))
			rootCmd.SetOut(stdout)
			rootCmd.SetErr(stderr)

			// when
			err := rootCmd.Execute()
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
<<<<<<< HEAD
			Expect(stderr.String()).To(BeEmpty())
=======
			Expect(stderr.String()).To(WithTransform(func(in string) string {
				return strings.ReplaceAll(
					in,
					"# [WARNING]: dry-run mode: No valid iptables executables found. The generated iptables rules may differ from those generated in an environment with valid iptables executables\n", "",
				)
			}, Equal("Error: transparent proxy cleanup failed: cleanup is not supported\n")))
>>>>>>> f732b34e9 (refactor(transparent-proxy): move executables to config (#10619))

			// when
			regex, err := os.ReadFile(filepath.Join("testdata", given.goldenFile))
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			r, err := regexp.Compile(string(regex))
			Expect(err).ToNot(HaveOccurred())

			// then
			Expect(r.Find(stdout.Bytes())).ToNot(BeEmpty())
		},
		Entry("should generate defaults with username", testCase{
			extraArgs:  nil,
			goldenFile: "uninstall-transparent-proxy.defaults.golden.txt",
		}),
	)
})
