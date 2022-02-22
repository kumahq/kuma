package install_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

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
			rootCmd.SetArgs(append([]string{"install", "transparent-proxy", "--dry-run"}, given.extraArgs...))
			rootCmd.SetOut(stdout)
			rootCmd.SetErr(stderr)

			// when
			err := rootCmd.Execute()
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(stderr.String()).To(BeEmpty())

			// when
			regex, err := os.ReadFile(filepath.Join("testdata", given.goldenFile))
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			r, err := regexp.Compile(string(regex))
			Expect(err).ToNot(HaveOccurred())

			// then
			Expect(r.Find(stdout.Bytes())).ToNot(BeEmpty(), fmt.Sprintf("%q\n-----\n%q\n", stdout.String(), stderr.String()))
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
		Entry("should generate defaults with user id and DNS redirected", testCase{
			extraArgs: []string{
				"--kuma-dp-uid", "0",
				"--kuma-cp-ip", "1.2.3.4",
				"--skip-resolv-conf",
				"--redirect-all-dns-traffic",
				"--redirect-dns-port", "12345",
				"--redirect-dns-upstream-target-chain", "DOCKER_OUTPUT",
			},
			goldenFile: "install-transparent-proxy.dns.golden.txt",
		}),
		Entry("should generate defaults with user id and DNS redirected without conntrack zone splitting", testCase{
			extraArgs: []string{
				"--kuma-dp-uid", "0",
				"--kuma-cp-ip", "1.2.3.4",
				"--skip-resolv-conf",
				"--redirect-all-dns-traffic",
				"--redirect-dns-port", "12345",
				"--redirect-dns-upstream-target-chain", "DOCKER_OUTPUT",
				"--skip-dns-conntrack-zone-split",
			},
			goldenFile: "install-transparent-proxy.dns.golden.txt",
		}),
		Entry("should generate defaults with overrides", testCase{
			extraArgs: []string{
				"--kuma-dp-user", "root",
				"--kuma-cp-ip", "1.2.3.4",
				"--redirect-outbound-port", "12345",
				"--redirect-inbound-port", "12346",
				"--redirect-inbound-port-v6", "123457",
				"--exclude-outbound-ports", "2000,2001",
				"--exclude-inbound-ports", "1000,1001",
			},
			goldenFile: "install-transparent-proxy.overrides.golden.txt",
		}),
	)

	DescribeTable("should return error",
		func(given testCase) {
			// given
			rootCmd := test.DefaultTestingRootCmd()
			rootCmd.SetArgs(append([]string{"install", "transparent-proxy", "--dry-run"}, given.extraArgs...))
			rootCmd.SetOut(stdout)
			rootCmd.SetErr(stderr)

			// when
			err := rootCmd.Execute()
			// then
			Expect(err).To(HaveOccurred())
			// and
			Expect(stderr.String()).To(ContainSubstring("one of --redirect-dns or --redirect-all-dns-traffic should be specified"))
		},
		Entry("should generate defaults with username", testCase{
			extraArgs: []string{
				"--kuma-dp-user", "root",
				"--redirect-dns",
				"--redirect-all-dns-traffic",
			},
			goldenFile: "install-transparent-proxy.defaults.golden.txt",
		}),
	)
})
