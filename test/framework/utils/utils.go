package utils

import (
	"bytes"
	"fmt"
	"net"
	"regexp"
	"slices"
	"strings"
	"text/template"

	ginko "github.com/onsi/ginkgo/v2"
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

func TestCaseName(ginkgo ginko.FullGinkgoTInterface) string {
	nameSplit := strings.Split(ginkgo.Name(), " ")
	return nameSplit[len(nameSplit)-1]
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

// BuildIptablesGoldenFileName constructs the complete file path for a golden file used
// for storing expected iptables rules based on a specific test configuration.
//
// Args:
//
//	dir (string): The base directory containing transparent proxy tests.
//	image (string): The Docker image name used for testing.
//	cmd (string): The iptables save command used in the test.
//	flags ([]string): The list of flags used during transparent proxy
//	  installation.
//
// Returns:
//
//	[]string: A slice of strings representing the complete file path for the
//	  golden file.
//	  - The first element is the provided base directory.
//	  - The second element is always "testdata", a subdirectory for storing
//	    golden files.
//	  - The third element is the actual golden file name based on the sanitized
//	    image name and flags joined with hyphens with iptables cmd suffix.
//
// Example:
//
//	BuildIptablesGoldenFileName(
//	  "install",
//	  "Ubuntu 22.04",
//	  "iptables-save",
//	  []string{"--redirect-dns"},
//	) # Returns [
//	  "install",
//	  "testdata",
//	  "ubuntu-22-04-iptables-redirect-dns.iptables.golden",
//	]
func BuildIptablesGoldenFileName(
	dir string,
	image string,
	cmd string,
	flags []string,
) []string {
	// Construct the golden file name by combining the sanitized image name,
	// cleaned flag names joined with hyphens, and the optional suffix.
	// The final file name has the format"<sanitized-image>-<flags>-<suffix>.golden".
	fileName := fmt.Sprintf(
		"%s.%s.golden",
		joinNonEmptyWithHyphen(
			// Sanitize the Docker image name (e.g., "ubuntu:22.04" becomes
			// "ubuntu-22-04").
			cleanName(image),
			// Sanitize and join the flag names with hyphens (e.g.,
			// ["--redirect-dns"] becomes "redirect-dns").
			cleanName(strings.Join(flags, "-")),
		),
		// Remove the "-save" suffix from the command.
		strings.TrimSuffix(cmd, "-save"),
	)

	return []string{dir, "testdata", fileName}
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
