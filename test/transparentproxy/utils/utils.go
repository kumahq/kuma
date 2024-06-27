package utils

import (
	"context"
	"fmt"
	"io"
	"regexp"
	"slices"
	"strings"

	"github.com/pkg/errors"
	"github.com/testcontainers/testcontainers-go"
)

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
	out := strings.TrimSpace(
		strings.Join(
			slices.DeleteFunc(
				strings.Split(output, "\n"),
				func(line string) bool {
					return strings.HasPrefix(line, "#")
				},
			),
			"\n",
		),
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

// BuildGoldenFileName constructs the complete file path for a golden file used
// for storing expected iptables rules based on a specific test configuration.
//
// Args:
//
//	dir (string): The base directory containing transparent proxy tests.
//	image (string): The Docker image name used for testing.
//	flags ([]string, optional): The list of flags used during transparent proxy
//	  installation.
//	sufix (string, optional): An additional suffix to be appended to the
//	  filename for further distinction between golden files (e.g., for different
//	  test purposes).
//
// Returns:
//
//	[]string: A slice of strings representing the complete file path for the
//	  golden file.
//	  - The first element is the provided base directory.
//	  - The second element is always "testdata", a subdirectory for storing
//	    golden files.
//	  - The third element is the actual golden file name based on the sanitized
//	    image name, flags, and optional suffix, joined with hyphens.
//
// Example:
//
//	BuildGoldenFileName(
//	  "install",
//	  "Ubuntu 22.04",
//	  []string{"--redirect-dns"},
//	  "ipv6",
//	) # Returns ["install", "testdata", "ubuntu-22-04-redirect-dns-ipv6.golden"]
func BuildGoldenFileName(
	dir string,
	image string,
	flags []string,
	sufix string,
) []string {
	// Construct the golden file name by combining the sanitized image name,
	// cleaned flag names joined with hyphens, and the optional suffix.
	// The final file name has the format"<sanitized-image>-<flags>-<suffix>.golden".
	fileName := fmt.Sprintf(
		"%s.golden",
		joinNonEmptyWithHyphen(
			// Sanitize the Docker image name (e.g., "ubuntu:22.04" becomes
			// "ubuntu-22-04").
			cleanName(image),
			// Sanitize and join the flag names with hyphens (e.g.,
			// ["--redirect-dns"] becomes "redirect-dns").
			cleanName(strings.Join(flags, "-")),
			// Append the optional suffix directly.
			sufix,
		),
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

// buildContainerHook constructs a testcontainers.ContainerHook that executes a
// provided command within a test container.
//
// This function is used to create hooks for running specific commands inside a
// test container. The hook captures the command's exit code and standard
// output. If the command exits with a non-zero status code, an error is
// returned with details including the command, exit code, and standard output.
func buildContainerHook(cmd []string) testcontainers.ContainerHook {
	return func(ctx context.Context, container testcontainers.Container) error {
		status, reader, err := container.Exec(ctx, cmd)
		if err != nil {
			return err
		}

		if status != 0 {
			buf := new(strings.Builder)
			if _, err := io.Copy(buf, reader); err != nil {
				return err
			}

			return errors.Errorf(
				"%s failed (exit code: %d): %s",
				strings.Join(cmd, " "),
				status,
				buf.String(),
			)
		}

		return nil
	}
}

func BuildContainerHooks(cmds [][]string) []testcontainers.ContainerHook {
	var hooks []testcontainers.ContainerHook

	for _, cmd := range cmds {
		hooks = append(hooks, buildContainerHook(cmd))
	}

	return hooks
}
