package install_test

import (
	"bytes"
	"regexp"
	"slices"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	gomega_types "github.com/onsi/gomega/types"

	"github.com/kumahq/kuma/app/kumactl/cmd/install"
	"github.com/kumahq/kuma/app/kumactl/pkg/test"
	"github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/transparentproxy/config"
)

var _ = Context("kumactl install transparent proxy", func() {
	type testCase struct {
		skip         func(stdout, stderr string) bool
		extraArgs    []string
		goldenFile   string
		errorMatcher gomega_types.GomegaMatcher
	}

	cleanStderr := func(stderr *bytes.Buffer) string {
		return strings.Join(
			slices.DeleteFunc(
				strings.Split(stderr.String(), "\n"),
				func(line string) bool {
					return strings.Contains(line, config.WarningDryRunNoValidIptablesFound) ||
						strings.Contains(line, install.WarningDryRunRunningAsNonRoot)
				},
			),
			"\n",
		)
	}

	DescribeTable("should install transparent proxy",
		func(given testCase) {
			// given
			args := append([]string{"install", "transparent-proxy", "--dry-run", "--ip-family-mode", "ipv4"}, given.extraArgs...)
			stdoutBuf, stderrBuf, rootCmd := test.DefaultTestingRootCmd(args...)

			// when
			err := rootCmd.Execute()
			stdout := stdoutBuf.String()
			stderr := cleanStderr(stderrBuf)

			if given.skip != nil && given.skip(stdout, stderr) {
				Skip("test skipped")
			}

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			if given.errorMatcher == nil {
				Expect(stderr).To(BeEmpty())
			} else {
				Expect(stderr).To(given.errorMatcher)
			}

			Expect(stdout).To(WithTransform(func(in string) string {
				// Replace some stuff that are environment dependent with placeholders
				out := regexp.MustCompile(`-o ([^ ]+)`).ReplaceAllString(in, "-o ifPlaceholder")
				out = regexp.MustCompile(`-m comment --comment ".*?" `).ReplaceAllString(out, "")
				out = regexp.MustCompile(`(?m)^-I OUTPUT (\d+) -p udp --dport 53 -m owner --uid-owner (\d+) -j (\w+)$`).
					ReplaceAllString(out, "-I OUTPUT $1 -p udp --dport 53 -m owner --uid-owner $2 -j dnsJumpTargetPlaceholder")
				return out
			}, matchers.MatchGoldenEqual("testdata", given.goldenFile)))
		},
		Entry("should generate defaults with username", testCase{
			extraArgs: []string{
				"--kuma-dp-user", "root",
			},
			goldenFile: "install-transparent-proxy.defaults.golden.txt",
		}),
		Entry("should generate defaults with user id", testCase{
			extraArgs: []string{
				"--kuma-dp-user", "0",
			},
			goldenFile: "install-transparent-proxy.defaults.golden.txt",
		}),
		Entry("should generate defaults with user id and DNS redirected when no conntrack module present", testCase{
			extraArgs: []string{
				"--kuma-dp-user", "0",
				"--redirect-all-dns-traffic",
				"--redirect-dns-port", "12345",
			},
			skip: func(stdout, stderr string) bool {
				return !strings.Contains(
					stderr,
					"conntrack zone splitting is disabled. This requires the 'conntrack' iptables module",
				)
			},
			errorMatcher: ContainSubstring("conntrack zone splitting is disabled. This requires the 'conntrack' iptables module"),
			goldenFile:   "install-transparent-proxy.dns.no-conntrack.golden.txt",
		}),
		Entry("should generate defaults with user id and DNS redirected", testCase{
			extraArgs: []string{
				"--kuma-dp-user", "0",
				"--redirect-all-dns-traffic",
				"--redirect-dns-port", "12345",
			},
			skip: func(stdout, stderr string) bool {
				return strings.Contains(
					stderr,
					"conntrack zone splitting is disabled. This requires the 'conntrack' iptables module",
				)
			},
			goldenFile: "install-transparent-proxy.dns.golden.txt",
		}),
		Entry("should generate defaults with user id and DNS redirected without conntrack zone splitting and log deprecate", testCase{
			extraArgs: []string{
				"--kuma-dp-user", "0",
				"--redirect-all-dns-traffic",
				"--redirect-dns-port", "12345",
				"--skip-dns-conntrack-zone-split",
			},
			goldenFile: "install-transparent-proxy.dns.no-conntrack.golden.txt",
		}),
		Entry("should generate defaults with overrides and log deprecate", testCase{
			extraArgs: []string{
				"--kuma-dp-user", "root",
				"--redirect-outbound-port", "12345",
				"--redirect-inbound-port", "12346",
				"--exclude-outbound-ports", "2000,2001",
				"--exclude-inbound-ports", "1000,1001",
			},
			goldenFile: "install-transparent-proxy.overrides.golden.txt",
		}),
		Entry("should generate when ipv6 disabled", testCase{
			extraArgs: []string{
				"--kuma-dp-user", "root",
				"--redirect-outbound-port", "12345",
				"--redirect-inbound-port", "12346",
				"--ip-family-mode", "ipv4",
				"--exclude-outbound-ports", "2000,2001",
				"--exclude-inbound-ports", "1000,1001",
			},
			goldenFile: "install-transparent-proxy.overrides.golden.txt",
		}),
		Entry("should generate defaults with outbound exclude ports", testCase{
			extraArgs: []string{
				"--kuma-dp-user", "root",
				"--exclude-outbound-ports-for-uids", "tcp:1900,1902,1000-2000:106-108",
				"--exclude-outbound-ports-for-uids", "tcp:2900,2902,3000-5000:203",
				"--exclude-outbound-ports-for-uids", "udp:3900,3902,4000-6000:303",
			},
			goldenFile: "install-transparent-proxy.excludedports.txt",
		}),
		Entry("should generate defaults with outbound exclude ports for uids wildcard", testCase{
			extraArgs: []string{
				"--kuma-dp-user", "root",
				"--exclude-outbound-ports-for-uids", "*:*:0",
			},
			goldenFile: "install-transparent-proxy.excludedports_wildcards.txt",
		}),
		Entry("should generate defaults with outbound exclude ports for uids no protocol means both protocols", testCase{
			extraArgs: []string{
				"--kuma-dp-user", "root",
				"--exclude-outbound-ports-for-uids", "tcp:*:0",
			},
			goldenFile: "install-transparent-proxy.excludedports_allprotos.txt",
		}),
		Entry("should generate defaults with outbound exclude ports for uids only uid", testCase{
			extraArgs: []string{
				"--kuma-dp-user", "root",
				"--exclude-outbound-ports-for-uids", "123",
			},
			goldenFile: "install-transparent-proxy.excludedports_onlyuid.txt",
		}),
		Entry("should generate defaults with outbound exclude ports", testCase{
			extraArgs: []string{
				"--kuma-dp-user", "root",
				"--exclude-outbound-ports-for-uids", "tcp:1900,1902,1000-2000:106-108",
				"--exclude-outbound-ports-for-uids", "tcp:2900,2902,3000-5000:203",
				"--exclude-outbound-ports-for-uids", "udp:3900,3902,4000-6000:303",
			},
			goldenFile: "install-transparent-proxy.excludedports_simpler.txt",
		}),
	)

	DescribeTable("should return error",
		func(given testCase) {
			// given
			args := append([]string{"install", "transparent-proxy", "--dry-run"}, given.extraArgs...)
			_, stderrBuf, rootCmd := test.DefaultTestingRootCmd(args...)

			// when
			err := rootCmd.Execute()
			stderr := cleanStderr(stderrBuf) // then
			Expect(err).To(HaveOccurred())

			// and
			Expect(stderr).To(given.errorMatcher)
		},
		Entry("should generate defaults with username", testCase{
			extraArgs: []string{
				"--kuma-dp-user", "root",
				"--redirect-dns",
				"--redirect-all-dns-traffic",
			},
			goldenFile:   "install-transparent-proxy.defaults.golden.txt",
			errorMatcher: Equal("Error: only one of '--redirect-dns' or '--redirect-all-dns-traffic' should be specified\n"),
		}),
		Entry("should error out on invalid port value", testCase{
			extraArgs: []string{
				"--kuma-dp-user", "root",
				"--exclude-outbound-ports-for-uids", "a3000-5000:1",
			},
			errorMatcher: ContainSubstring("parsing excluded outbound ports for uids failed: invalid port range: validation failed for value or range 'a3000-5000': invalid uint16 value: 'a3000'\n"),
		}),
		Entry("should error out on invalid protocol value", testCase{
			extraArgs: []string{
				"--kuma-dp-user", "root",
				"--exclude-outbound-ports-for-uids", "http:3000-5000:1",
			},
			errorMatcher: ContainSubstring("parsing excluded outbound ports for uids failed: invalid or unsupported protocol: 'http'\n"),
		}),
		Entry("should error out on wildcard on uids", testCase{
			extraArgs: []string{
				"--kuma-dp-user", "root",
				"--exclude-outbound-ports-for-uids", "3000-5000:1:*",
			},
			errorMatcher: ContainSubstring("parsing excluded outbound ports for uids failed: wildcard '*' is not allowed for UIDs\n"),
		}),
		Entry("should error out with list on uids", testCase{
			extraArgs: []string{
				"--kuma-dp-user", "root",
				"--exclude-outbound-ports-for-uids", "3000-5000:1:1,2,3",
			},
			errorMatcher: ContainSubstring("parsing excluded outbound ports for uids failed: invalid UID entry: '1,2,3'. It should either be a single item or a range\n"),
		}),
		Entry("should error out on invalid port in list", testCase{
			extraArgs: []string{
				"--kuma-dp-user", "root",
				"--exclude-outbound-ports-for-uids", "tcp,http:1:1",
			},
			errorMatcher: ContainSubstring("parsing excluded outbound ports for uids failed: invalid or unsupported protocol: 'http'\n"),
		}),
	)
})
