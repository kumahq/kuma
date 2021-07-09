package uninstall_test

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"regexp"

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

	DescribeTable("should install transparent proxy",
		func(given testCase) {
			// given
			rootCmd := cmd.DefaultRootCmd()
			rootCmd.SetArgs(append([]string{"uninstall", "transparent-proxy", "--dry-run"}, given.extraArgs...))
			rootCmd.SetOut(stdout)
			rootCmd.SetErr(stderr)

			// when
			err := rootCmd.Execute()
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(stderr.String()).To(BeEmpty())

			// when
			regex, err := ioutil.ReadFile(filepath.Join("testdata", given.goldenFile))
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
