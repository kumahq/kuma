package install_test

import (
	"bytes"
	"regexp"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	gomega_types "github.com/onsi/gomega/types"

	"github.com/kumahq/kuma/pkg/test/matchers"
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
		skip         func(stdout, stderr *bytes.Buffer) bool
		extraArgs    []string
		goldenFile   string
		errorMatcher gomega_types.GomegaMatcher
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

			if given.skip != nil && !given.skip(stdout, stderr) {
				Skip("test skipped")
			}

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			if given.errorMatcher == nil {
				Expect(stderr.String()).To(BeEmpty())
			} else {
				Expect(stderr.String()).To(given.errorMatcher)
			}

			Expect(stdout.String()).To(WithTransform(func(in string) string {
				// Replace some stuff that are environment dependent with placeholders
				out := regexp.MustCompile(`-o ([^ ]+)`).ReplaceAllString(in, "-o ifPlaceholder")
				out = regexp.MustCompile(`-([sd]) ([^ ]+)`).ReplaceAllString(out, "-$1 subnetPlaceholder/mask")
				out = regexp.MustCompile(`(?m)^-I OUTPUT (\d+) -p udp --dport 53 -m owner --uid-owner (\d+) -j (\w+)$`).
					ReplaceAllString(out, "-I OUTPUT $1 -p udp --dport 53 -m owner --uid-owner $2 -j dnsJumpTargetPlaceholder")
				out = strings.ReplaceAll(out, "15006", "inboundPort")
				out = strings.ReplaceAll(out, "15010", "inboundPort")
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
				"--kuma-dp-uid", "0",
			},
			goldenFile: "install-transparent-proxy.defaults.golden.txt",
		}),
		Entry("should generate defaults with user id and DNS redirected when no conntrack module present", testCase{
			extraArgs: []string{
				"--kuma-dp-uid", "0",
				"--redirect-all-dns-traffic",
				"--redirect-dns-port", "12345",
			},
			skip: func(stdout, stderr *bytes.Buffer) bool {
				return strings.HasPrefix(
					stderr.String(),
					"# [WARNING] error occurred when validating if 'conntrack' iptables module is present. Rules for DNS conntrack zone splitting won't be applied:",
				)
			},
			errorMatcher: HavePrefix("# [WARNING] error occurred when validating if 'conntrack' iptables module is present. Rules for DNS conntrack zone splitting won't be applied:"),
			goldenFile:   "install-transparent-proxy.dns.no-conntrack.golden.txt",
		}),
		Entry("should generate defaults with user id and DNS redirected", testCase{
			extraArgs: []string{
				"--kuma-dp-uid", "0",
				"--redirect-all-dns-traffic",
				"--redirect-dns-port", "12345",
			},
			skip: func(stdout, stderr *bytes.Buffer) bool {
				return !strings.HasPrefix(
					stderr.String(),
					"# [WARNING] error occurred when validating if 'conntrack' iptables module is present. Rules for DNS conntrack zone splitting won't be applied:",
				)
			},
			goldenFile: "install-transparent-proxy.dns.golden.txt",
		}),
		Entry("should generate defaults with user id and DNS redirected without conntrack zone splitting and log deprecate", testCase{
			extraArgs: []string{
				"--kuma-dp-uid", "0",
				"--redirect-all-dns-traffic",
				"--redirect-dns-port", "12345",
				"--skip-dns-conntrack-zone-split",
			},
			goldenFile: "install-transparent-proxy.dns.no-conntrack.golden.txt",
		}),
		Entry("should generate defaults with overrides", testCase{
			extraArgs: []string{
				"--kuma-dp-user", "root",
				"--redirect-outbound-port", "12345",
				"--redirect-inbound-port", "12346",
				"--redirect-inbound-port-v6", "12346",
				"--exclude-outbound-ports", "2000,2001",
				"--exclude-inbound-ports", "1000,1001",
			},
			goldenFile: "install-transparent-proxy.overrides.golden.txt",
		}),
		Entry("should generate defaults with outbound exclude ports", testCase{
			extraArgs: []string{
				"--kuma-dp-user", "root",
				"--exclude-outbound-tcp-ports-for-uids", "1900,1902,1000-2000:106-108",
				"--exclude-outbound-tcp-ports-for-uids", "2900,2902,3000-5000:203",
				"--exclude-outbound-udp-ports-for-uids", "3900,3902,4000-6000:303",
			},
			errorMatcher: Equal("# [WARNING] flag --exclude-outbound-tcp-ports-for-uids is deprecated use --exclude-outbound-ports-for-uids instead\n# [WARNING] flag --exclude-outbound-udp-ports-for-uids is deprecated use --exclude-outbound-ports-for-uids instead\n"),
			goldenFile:   "install-transparent-proxy.excludedports.txt",
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
			rootCmd := test.DefaultTestingRootCmd()
			rootCmd.SetArgs(append([]string{"install", "transparent-proxy", "--dry-run"}, given.extraArgs...))
			rootCmd.SetOut(stdout)
			rootCmd.SetErr(stderr)

			// when
			err := rootCmd.Execute()
			// then
			Expect(err).To(HaveOccurred())
			// and
			Expect(stderr.String()).To(given.errorMatcher)
		},
		Entry("should generate defaults with username", testCase{
			extraArgs: []string{
				"--kuma-dp-user", "root",
				"--redirect-dns",
				"--redirect-all-dns-traffic",
			},
			goldenFile:   "install-transparent-proxy.defaults.golden.txt",
			errorMatcher: Equal("Error: one of --redirect-dns or --redirect-all-dns-traffic should be specified\n"),
		}),
		Entry("should error out on invalid port value", testCase{
			extraArgs: []string{
				"--kuma-dp-user", "root",
				"--exclude-outbound-ports-for-uids", "a3000-5000:1",
			},
			errorMatcher: Equal("Error: failed to setup transparent proxy: parsing excluded outbound ports for uids failed: values or range 'a3000-5000' failed validation: value 'a3000', is not valid uint16\n"),
		}),
		Entry("should error out on invalid protocol value", testCase{
			extraArgs: []string{
				"--kuma-dp-user", "root",
				"--exclude-outbound-ports-for-uids", "http:3000-5000:1",
			},
			errorMatcher: Equal("Error: failed to setup transparent proxy: parsing excluded outbound ports for uids failed: protocol 'http' is invalid or unsupported\n"),
		}),
		Entry("should error out on wildcard on uids", testCase{
			extraArgs: []string{
				"--kuma-dp-user", "root",
				"--exclude-outbound-ports-for-uids", "3000-5000:1:*",
			},
			errorMatcher: Equal("Error: failed to setup transparent proxy: parsing excluded outbound ports for uids failed: can't use wildcard '*' for uids\n"),
		}),
		Entry("should error out with list on uids", testCase{
			extraArgs: []string{
				"--kuma-dp-user", "root",
				"--exclude-outbound-ports-for-uids", "3000-5000:1:1,2,3",
			},
			errorMatcher: Equal("Error: failed to setup transparent proxy: parsing excluded outbound ports for uids failed: uid entry invalid:'1,2,3', it should either be a single item or a range\n"),
		}),
		Entry("should error out on invalid port in list", testCase{
			extraArgs: []string{
				"--kuma-dp-user", "root",
				"--exclude-outbound-ports-for-uids", "tcp,http:1:1",
			},
			errorMatcher: Equal("Error: failed to setup transparent proxy: parsing excluded outbound ports for uids failed: protocol 'http' is invalid or unsupported\n"),
		}),
	)
})
