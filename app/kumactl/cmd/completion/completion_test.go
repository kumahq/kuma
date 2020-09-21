package completion_test

import (
	"bytes"
	"io/ioutil"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/app/kumactl/cmd"
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
			// and
			Expect(stderr.Bytes()).To(BeNil())

			// when
			expected, err := ioutil.ReadFile(filepath.Join("testdata", given.goldenFile))
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			actual := stdout.Bytes()
			// then
			Expect(len(actual)).To(Equal(len(expected)), "run kumactl completion bash/fish/zsh, copy the output and override testdata/*.golden")

			// and
			Expect(string(actual)).To(Equal(string(expected)))
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
