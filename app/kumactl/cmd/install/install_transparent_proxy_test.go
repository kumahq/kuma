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

var _ = Describe("kumactl install transparent proxy", func() {

	var stdout *bytes.Buffer
	var stderr *bytes.Buffer

	BeforeEach(func() {
		stdout = &bytes.Buffer{}
		stderr = &bytes.Buffer{}
	})

	type testCase struct {
		extraArgs    []string
		goldenFile   string
		errorMessage string
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
			Expect(stderr.String()).To(Equal(given.errorMessage))

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
			},
			goldenFile: "install-transparent-proxy.defaults.golden.txt",
		}),
		Entry("should generate defaults with user id", testCase{
			extraArgs: []string{
				"--kuma-dp-uid", "0",
			},
			goldenFile: "install-transparent-proxy.defaults.golden.txt",
		}),
		Entry("should generate defaults with user id and DNS redirected and log deprecate", testCase{
			extraArgs: []string{
				"--kuma-dp-uid", "0",
				"--redirect-all-dns-traffic",
				"--redirect-dns-port", "12345",
				"--redirect-dns-upstream-target-chain", "DOCKER_OUTPUT",
			},
			goldenFile:   "install-transparent-proxy.dns.golden.txt",
			errorMessage: "# `--redirect-dns-upstream-target-chain` is deprecated, please avoid using it",
		}),
		Entry("should generate defaults with user id and DNS redirected without conntrack zone splitting and log deprecate", testCase{
			extraArgs: []string{
				"--kuma-dp-uid", "0",
				"--redirect-all-dns-traffic",
				"--redirect-dns-port", "12345",
				"--redirect-dns-upstream-target-chain", "DOCKER_OUTPUT",
				"--skip-dns-conntrack-zone-split",
			},
			goldenFile:   "install-transparent-proxy.dns.golden.txt",
			errorMessage: "# `--redirect-dns-upstream-target-chain` is deprecated, please avoid using it",
		}),
		Entry("should generate defaults with overrides", testCase{
			extraArgs: []string{
				"--kuma-dp-user", "root",
				"--redirect-outbound-port", "12345",
				"--redirect-inbound-port", "12346",
				"--redirect-inbound-port-v6", "123457",
				"--exclude-outbound-ports", "2000,2001",
				"--exclude-inbound-ports", "1000,1001",
			},
			goldenFile: "install-transparent-proxy.overrides.golden.txt",
		}),
		Entry("should generate defaults with outbound exclude ports", testCase{
			extraArgs: []string{
				"--kuma-dp-user", "root",
				"--experimental-transparent-proxy-engine",
				"--exclude-outbound-tcp-ports-for-uids", "1900,1902,1000-2000:103,104,106-108",
				"--exclude-outbound-tcp-ports-for-uids", "2900,2902,3000-5000:203,204,206-208",
				"--exclude-outbound-udp-ports-for-uids", "3900,3902,4000-6000:303,304,306-308",
			},
			goldenFile: "install-transparent-proxy.excludedports.txt",
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
			Expect(stderr.String()).To(ContainSubstring(given.errorMessage))
		},
		Entry("should generate defaults with username", testCase{
			extraArgs: []string{
				"--kuma-dp-user", "root",
				"--redirect-dns",
				"--redirect-all-dns-traffic",
			},
			goldenFile:   "install-transparent-proxy.defaults.golden.txt",
			errorMessage: "one of --redirect-dns or --redirect-all-dns-traffic should be specified",
		}),
		Entry("should error out on invalid port value", testCase{
			extraArgs: []string{
				"--kuma-dp-user", "root",
				"--experimental-transparent-proxy-engine",
				"--exclude-outbound-tcp-ports-for-uids", "a3000-5000:1",
			},
			errorMessage: "Error: failed to setup transparent proxy: parsing excluded outbound TCP ports for UIDs failed: values or range a3000-5000 failed validation: value a3000, is not valid uint16",
		}),
	)
})
