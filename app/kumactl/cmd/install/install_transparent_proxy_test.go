package install_test

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
			rootCmd.SetArgs(append([]string{"install", "transparent-proxy", "--dry-run"}, given.extraArgs...))
			rootCmd.SetOut(stdout)
			rootCmd.SetErr(stderr)

			// when
			err := rootCmd.Execute()
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(stderr.Bytes()).To(BeNil())

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
			extraArgs: []string{
				"--kuma-dp-user", "root",
				"--kuma-cp-ip", "1.2.3.4",
			},
			goldenFile: "install-transparent-proxy.defaults.golden.txt",
		}),
		Entry("should generate defaults with user id", testCase{
			extraArgs: []string{
				"--kuma-dp-uid", "0",
				"--kuma-cp-ip", "1.2.3.4",
			},
			goldenFile: "install-transparent-proxy.defaults.golden.txt",
		}),
		Entry("should generate defaults with overrides", testCase{
			extraArgs: []string{
				"--kuma-dp-user", "root",
				"--kuma-cp-ip", "1.2.3.4",
				"--redirect-outbound-port", "12345",
				"--redirect-inbound-port", "12346",
				"--exclude-outbound-ports", "2000,2001",
				"--exclude-inbound-ports", "1000,1001",
			},
			goldenFile: "install-transparent-proxy.overrides.golden.txt",
		}),
	)
})
