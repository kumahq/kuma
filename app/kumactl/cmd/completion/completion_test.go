package completion_test

import (
	"bytes"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/app/kumactl/cmd"
	. "github.com/kumahq/kuma/pkg/test/matchers"
)

var _ = Describe("kumactl completion", func() {

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

	DescribeTable("should generate completion code",
		func(given testCase) {
			// given
			rootCmd := cmd.DefaultRootCmd()
			rootCmd.SetArgs(append([]string{"completion"}, given.extraArgs...))
			rootCmd.SetOut(stdout)
			rootCmd.SetErr(stderr)

			// when
			err := rootCmd.Execute()

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(stderr.String()).To(BeEmpty())

			// and
			actual := stdout.Bytes()
			Expect(actual).To(MatchGoldenEqual(filepath.Join("testdata", given.goldenFile)))
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
