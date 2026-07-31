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
	xdsChurnMeshRe              = regexp.MustCompile(`"mesh":\s*"([^"]*)"`)
	xdsChurnMeshFallbackRe      = regexp.MustCompile(`\bmesh=(\S+)`)
	xdsChurnVersionsRe          = regexp.MustCompile(`"versions":\s*\[([^\]]*)\]`)
	xdsChurnTypedVersionRe      = regexp.MustCompile(`"([^"]+)=([0-9a-f]{16})"`)
	xdsChurnHashRe              = regexp.MustCompile(`[0-9a-f]{16}`)
)

// xdsChurnKey identifies a proxy by both mesh and name. In e2e, dataplane
// names (e.g. demo-client, test-server) are reused across meshes, so keying
// churn counts by name alone would merge unrelated proxies and could report
// spurious churn.
type xdsChurnKey struct {
	mesh  string
	proxy string
}

type xdsChurnVersion struct {
	typeURL string
	hash    string
}

type xdsChurnState struct {
	hash   string
	streak int
}

// xdsChurnThreshold is the number of times a proxy must regenerate the exact
// same content hash before it is flagged as non-deterministic xDS
// generation. A legitimate config change followed by a revert (A -> B -> A)
// already produces 2 occurrences of hash A, so the threshold must be at
// least one more repetition to avoid flagging normal config evolution.
const xdsChurnThreshold = 3

// DetectXdsChurn parses kuma-cp logs and flags proxies for which the CP
// regenerated a byte-identical xDS config (proven by a repeated content
// hash logged by pkg/xds/server/v3.reconciler.Reconcile in its "config has
// changed" log line) at least xdsChurnThreshold times without an
// intervening different hash, which indicates non-deterministic xDS
// generation (e.g. unordered map iteration when building a resource).
//
// Each "config has changed" log line is one regeneration event per changed
// resource type. The CP logs only changed types, so a type's current streak
// must survive intervening lines that mention other types. Within one line, a
// duplicated type/hash pair still counts only once.
func DetectXdsChurn(logs string) []string {
	maxStreaks := map[xdsChurnKey]map[xdsChurnVersion]int{}
	currentStates := map[xdsChurnKey]map[string]xdsChurnState{}

	for line := range strings.SplitSeq(logs, "\n") {
		if !strings.Contains(line, "config has changed") {
			continue
		}
		proxyName := xdsChurnProxyName(line)
		if proxyName == "" {
			continue
		}
		key := xdsChurnKey{mesh: xdsChurnMesh(line), proxy: proxyName}
		versions := xdsChurnVersionsRe.FindStringSubmatch(line)
		if versions == nil {
			continue
		}
		seenInLine := xdsChurnVersions(versions[1])
		if len(seenInLine) == 0 {
			continue
		}
		if currentStates[key] == nil {
			currentStates[key] = map[string]xdsChurnState{}
		}
		if maxStreaks[key] == nil {
			maxStreaks[key] = map[xdsChurnVersion]int{}
		}
		for _, version := range seenInLine {
			streak := 1
			if prev, ok := currentStates[key][version.typeURL]; ok && prev.hash == version.hash {
				streak = prev.streak + 1
			}
			currentStates[key][version.typeURL] = xdsChurnState{
				hash:   version.hash,
				streak: streak,
			}
			if streak > maxStreaks[key][version] {
				maxStreaks[key][version] = streak
			}
		}
	}

	var reports []string
	keys := slices.SortedFunc(maps.Keys(maxStreaks), func(a, b xdsChurnKey) int {
		if a.mesh != b.mesh {
			return strings.Compare(a.mesh, b.mesh)
		}
		return strings.Compare(a.proxy, b.proxy)
	})
	for _, key := range keys {
		versions := slices.SortedFunc(maps.Keys(maxStreaks[key]), func(a, b xdsChurnVersion) int {
			if a.hash != b.hash {
				return strings.Compare(a.hash, b.hash)
			}
			return strings.Compare(a.typeURL, b.typeURL)
		})
		for _, version := range versions {
			count := maxStreaks[key][version]
			if count >= xdsChurnThreshold {
				reports = append(reports, fmt.Sprintf(
					"proxy %s in mesh %s regenerated identical config %d times (hash %s) — non-deterministic xDS",
					key.proxy, key.mesh, count, version.hash,
				))
			}
		}
	}
	return reports
}

func xdsChurnVersions(raw string) []xdsChurnVersion {
	seen := map[xdsChurnVersion]bool{}
	var versions []xdsChurnVersion

	for _, match := range xdsChurnTypedVersionRe.FindAllStringSubmatch(raw, -1) {
		version := xdsChurnVersion{
			typeURL: match[1],
			hash:    match[2],
		}
		if seen[version] {
			continue
		}
		seen[version] = true
		versions = append(versions, version)
	}
	if len(versions) != 0 {
		return versions
	}

	for _, hash := range xdsChurnHashRe.FindAllString(raw, -1) {
		version := xdsChurnVersion{hash: hash}
		if seen[version] {
			continue
		}
		seen[version] = true
		versions = append(versions, version)
	}
	return versions
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

// xdsChurnMesh extracts the mesh field from a CP log line, matching the
// encoders handled by xdsChurnProxyName. An empty result (e.g. mesh-less
// zone ingress/egress proxies) is a valid key on its own.
func xdsChurnMesh(line string) string {
	if m := xdsChurnMeshRe.FindStringSubmatch(line); m != nil {
		return m[1]
	}
	if m := xdsChurnMeshFallbackRe.FindStringSubmatch(line); m != nil {
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
