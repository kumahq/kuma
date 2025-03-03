package utils

import (
	"bytes"
	"fmt"
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
