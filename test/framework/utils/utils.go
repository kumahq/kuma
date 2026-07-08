package utils

import (
	"bytes"
	"fmt"
	"maps"
	"net"
	"regexp"
	"slices"
	"strings"
	"text/template"

	"github.com/onsi/gomega"
)

func ShellEscape(arg string) string {
	return fmt.Sprintf("'%s'", strings.ReplaceAll(arg, "'", "\\'"))
}

func GetFreePort() (int, error) {
	address, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	listener, err := net.ListenTCP("tcp", address)
	if err != nil {
		return 0, err
	}
	defer listener.Close()

	return listener.Addr().(*net.TCPAddr).Port, nil
}

func FromTemplate(g gomega.Gomega, tmpl string, data any) string {
	t, err := template.New("tmpl").Parse(tmpl)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	b := bytes.Buffer{}
	g.Expect(t.Execute(&b, data)).To(gomega.Succeed())
	return b.String()
}

func HasPanicInCpLogs(logs string) bool {
	return strings.Contains(logs, "runtime.gopanic") || strings.Contains(logs, "panic:")
}

var (
	xdsChurnProxyNameRe         = regexp.MustCompile(`"proxyName":\s*"([^"]+)"`)
	xdsChurnProxyNameFallbackRe = regexp.MustCompile(`proxyName=(\S+)`)
	xdsChurnVersionsRe          = regexp.MustCompile(`"versions":\s*\[([^\]]*)\]`)
	xdsChurnHashRe              = regexp.MustCompile(`[0-9a-f]{16}`)
)

// xdsChurnThreshold is the number of times a proxy must regenerate the exact
// same content hash before it is flagged as non-deterministic xDS
// generation. A legitimate config change followed by a revert (A -> B -> A)
// already produces 2 occurrences of hash A, so the threshold must be at
// least one more repetition to avoid flagging normal config evolution.
const xdsChurnThreshold = 3

// emptyResourcesVersionHash is the constant version string emitted by
// pkg/xds/server/v3.emptyResourcesVersion() whenever a resource type
// transitions from populated to empty. It is derived from a fixed nil
// payload, so it is identical across proxies, resource types and
// reconciliations. Counting it towards churn would flag unrelated one-time
// clears (e.g. a route policy removed, then later mTLS disabled) as if they
// were the same config being repeatedly regenerated.
const emptyResourcesVersionHash = "34c96acdcadb1bbb"

// DetectXdsChurn parses kuma-cp logs and flags proxies for which the CP
// regenerated a byte-identical xDS config (proven by a repeated content
// hash logged by pkg/xds/server/v3.reconciler.Reconcile in its "config has
// changed" log line) at least xdsChurnThreshold times without an
// intervening different hash, which indicates non-deterministic xDS
// generation (e.g. unordered map iteration when building a resource).
func DetectXdsChurn(logs string) []string {
	counts := map[string]map[string]int{}

	for line := range strings.SplitSeq(logs, "\n") {
		if !strings.Contains(line, "config has changed") {
			continue
		}
		proxyName := xdsChurnProxyName(line)
		if proxyName == "" {
			continue
		}
		versions := xdsChurnVersionsRe.FindStringSubmatch(line)
		if versions == nil {
			continue
		}
		for _, hash := range xdsChurnHashRe.FindAllString(versions[1], -1) {
			if hash == emptyResourcesVersionHash {
				continue
			}
			if counts[proxyName] == nil {
				counts[proxyName] = map[string]int{}
			}
			counts[proxyName][hash]++
		}
	}

	var reports []string
	for _, proxyName := range slices.Sorted(maps.Keys(counts)) {
		var maxHash string
		var maxCount int
		for _, hash := range slices.Sorted(maps.Keys(counts[proxyName])) {
			if count := counts[proxyName][hash]; count > maxCount {
				maxCount = count
				maxHash = hash
			}
		}
		if maxCount >= xdsChurnThreshold {
			reports = append(reports, fmt.Sprintf(
				"proxy %s regenerated identical config %d times (hash %s) — non-deterministic xDS",
				proxyName, maxCount, maxHash,
			))
		}
	}
	return reports
}

// xdsChurnProxyName extracts the proxyName field from a CP log line. The
// console encoder renders extra fields as trailing compact JSON; the
// logfmt-style fallback covers other encoders that may render fields as
// key=value pairs.
func xdsChurnProxyName(line string) string {
	if m := xdsChurnProxyNameRe.FindStringSubmatch(line); m != nil {
		return m[1]
	}
	if m := xdsChurnProxyNameFallbackRe.FindStringSubmatch(line); m != nil {
		return m[1]
	}
	return ""
}

// CleanIptablesSaveOutput processes the provided iptables-save output by
// removing comment lines and normalizing variable content for consistent
// comparison with golden files.
//
// Specifically, the function performs the following steps:
//  1. Removes comment lines that start with "#" to eliminate lines containing
//     dynamic data such as timestamps or identifiers that vary between runs.
//  2. Uses regular expressions to replace certain patterns that may differ
//     between runs, ensuring the output is standardized. For example, interface
//     names are replaced with a placeholder to facilitate consistent
//     comparison.
//
// The resulting cleaned output is more stable for comparison with golden files
// by removing elements that can change between runs, allowing for accurate
// regression testing.
func CleanIptablesSaveOutput(output string) string {
	out := strings.Join(
		slices.DeleteFunc(
			strings.Split(output, "\n"),
			func(line string) bool {
				return strings.HasPrefix(line, "#")
			},
		),
		"\n",
	)
	out = regexp.MustCompile(`-o ([^ ]+)`).
		ReplaceAllString(out, "-o ifPlaceholder")
	out = regexp.MustCompile(`(?m)^:([^ ]+) ([^ ]+) .*?$`).
		ReplaceAllString(out, ":$1 $2")

	return out
}

// cleanName sanitizes a string intended for use in file name construction for
// golden files.
func cleanName(name string) string {
	return strings.NewReplacer(
		"--", "",
		"=", "-",
		":", "-",
		".", "-",
		"/", "-",
		" ", "-",
	).Replace(strings.ToLower(name))
}

// BuildIptablesGoldenFileName constructs the complete file path for a golden
// file used for storing expected iptables rules based on a specific test
// configuration.
//
// Args:
//   - image (string): The Docker image name used for testing.
//   - cmd (string): The iptables save command used in the test.
//   - suffix (string): The optional suffix to be added to the file name.
//
// Returns:
//   - []string: A slice of strings representing the complete file path for the
//     golden file. The first element is the subdirectory "testdata", and the
//     second element is the actual golden file name based on the sanitized
//     image name and flags joined with hyphens, ending with the command and
//     ".golden" suffix.
//
// Example:
//
//	BuildIptablesGoldenFileName("RHEL 8", "iptables-save", "redirect-dns")
//	# Returns ["testdata", "rhel-8-redirect-dns.iptables.golden"]
func BuildIptablesGoldenFileName(image, cmd, suffix string) []string {
	// Construct the golden file name by combining the sanitized image name,
	// cleaned suffix joined with hyphens, and the trimmed command.
	// The final file name has the format
	// "<sanitized-image>-<suffix>.<cmd>.golden".
	fileName := fmt.Sprintf(
		"%s.%s.golden",
		joinNonEmptyWithHyphen(
			// Sanitize the Docker image name (e.g., "ubuntu:22.04" becomes
			// "ubuntu-22-04").
			cleanName(image),
			suffix,
		),
		// Remove the "-save" suffix from the command.
		strings.TrimSuffix(cmd, "-save"),
	)

	return []string{"testdata", fileName}
}

// joinNonEmptyWithHyphen joins a slice of strings with hyphens (-) as
// separators, omitting any empty strings from the joined result.
func joinNonEmptyWithHyphen(elems ...string) string {
	return strings.Join(
		slices.DeleteFunc(
			elems,
			func(s string) bool {
				return s == ""
			},
		),
		"-",
	)
}

func Indent(pem string, padding int) string {
	pad := strings.Repeat(" ", padding)
	return pad + strings.ReplaceAll(pem, "\n", "\n"+pad)
}
