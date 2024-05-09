package completion_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/app/kumactl/pkg/test"
	. "github.com/kumahq/kuma/pkg/test/matchers"
)

var _ = Describe("kumactl completion", func() {
	type testCase struct {
		extraArgs  []string
		goldenFile string
	}

	DescribeTable("should generate completion code",
		func(given testCase) {
			// given
			args := append([]string{"completion"}, given.extraArgs...)
			stdout, stderr, rootCmd := test.DefaultTestingRootCmd(args...)

			// when
			err := rootCmd.Execute()

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(stderr.String()).To(BeEmpty())

			// and
			actual := stdout.Bytes()
			Expect(actual).To(MatchGoldenEqual("testdata", given.goldenFile))
		},
		Entry("should generate bash completion code", testCase{
			extraArgs: []string{
				"bash",
			},
			goldenFile: "bash.golden",
		}),
		Entry("should generate fish completion code", testCase{
			extraArgs: []string{
				"fish",
			},
			goldenFile: "fish.golden",
		}),
		Entry("should generate zsh completion code", testCase{
			extraArgs: []string{
				"zsh",
			},
			goldenFile: "zsh.golden",
		}),
	)
})
