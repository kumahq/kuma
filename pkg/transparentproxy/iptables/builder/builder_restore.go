package builder

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/v2/pkg/transparentproxy/config"
	. "github.com/kumahq/kuma/v2/pkg/transparentproxy/iptables/consts"
	. "github.com/kumahq/kuma/v2/pkg/transparentproxy/iptables/parameters"
)

var dockerOutputChainRegex = regexp.MustCompile(`(?m)^:DOCKER_OUTPUT`)

var fallbackPaths = []string{
	"/usr/sbin",
	"/sbin",
	"/usr/bin",
	"/bin",
}

func buildRestoreParameters(cfg config.Config, rulesFile *os.File, restoreLegacy bool) []string {
	return NewParameters().
		AppendIf(restoreLegacy, Wait(cfg.Wait), WaitInterval(cfg.WaitInterval)).
		Append(NoFlush()).
		Build(cfg.Verbose, rulesFile.Name())
}

func findExecutable(prefix string, mode string, name string) Executable {
	// ip{6}tables-{nft|legacy}, ip{6}tables-{nft|legacy}-save,
	// ip{6}tables-{nft|legacy}-restore
	nameWithMode := joinNonEmptyWithHyphen(prefix, mode, name)
	// ip{6}tables, ip{6}tables-save, ip{6}tables-restore
	nameWithoutMode := joinNonEmptyWithHyphen(prefix, name)

	paths := getPathsToSearchForExecutable(nameWithMode, nameWithoutMode)

	for _, path := range paths {
		if found := findPath(path); found != "" {
			if verifyIptablesMode(path, mode) {
				return newExecutable(nameWithMode, found)
			}
		}
	}

	return Executable{Name: nameWithMode}
}

type Executable struct {
	Name string
	Path string
}

func newExecutable(name string, path string) Executable {
	return Executable{
		Name: name,
		Path: path,
	}
}

func (e Executable) exec(ctx context.Context, args ...string) (*bytes.Buffer, error) {
	stdout, _, err := execCmd(ctx, e.Path, args...)
	if err != nil {
		return nil, err
	}

	return stdout, nil
}

type Executables struct {
	Iptables               Executable
	Save                   Executable
	Restore                Executable
	fallback               *Executables
	mode                   string
	foundDockerOutputChain bool
}

func newExecutables(ipv6 bool, mode string) *Executables {
	prefix := iptables
	if ipv6 {
		prefix = ip6tables
	}

	return &Executables{
		Iptables: findExecutable(prefix, mode, ""),
		Save:     findExecutable(prefix, mode, "save"),
		Restore:  findExecutable(prefix, mode, "restore"),
		mode:     mode,
	}
}

var necessaryMatchExtensions = []string{
	"owner",
	"tcp",
	"udp",
}

func (e *Executables) legacy() bool {
	return e.mode == "legacy"
}

func (e *Executables) verify(ctx context.Context, cfg config.Config) (*Executables, error) {
	var missing []string

	if e.Save.Path == "" {
		missing = append(missing, e.Save.Name)
	}

	if e.Restore.Path == "" {
		missing = append(missing, e.Restore.Name)
	}

	if len(missing) > 0 {
		return nil, errors.Errorf("couldn't find %s executables: [%s]", e.mode, strings.Join(missing, ", "))
	}

	// We always need to have access to the "nat" table
	if stdout, err := e.Save.exec(ctx, "-t", "nat"); err != nil {
		return nil, errors.Wrap(err, "couldn't verify if table: 'nat' is available")
	} else if cfg.ShouldRedirectDNS() || cfg.ShouldCaptureAllDNS() {
		e.foundDockerOutputChain = dockerOutputChainRegex.Match(stdout.Bytes())
	}

	// It seems in some cases (GKE with ContainerOS), even if "iptables-nft" is available
	// there are some kernel modules with iptables match extensions missing.
	for _, matchExtension := range necessaryMatchExtensions {
		if _, err := e.Iptables.exec(ctx, "-m", matchExtension, "--help"); err != nil {
			return nil, errors.Wrapf(err, "verification if match: %q exist failed", matchExtension)
		}
	}

	if cfg.ShouldConntrackZoneSplit(e.Iptables.Path) {
		if _, err := e.Save.exec(ctx, "-t", "raw"); err != nil {
			return nil, errors.Wrap(err, "couldn't verify if table: 'raw' is available")
		}
	}

	return e, nil
}

func (e *Executables) withFallback(fallback *Executables) *Executables {
	if fallback != nil {
		e.fallback = fallback
	}

	return e
}

func DetectIptablesExecutables(
	ctx context.Context,
	cfg config.Config,
	ipv6 bool,
) (*Executables, error) {
	nft, nftVerifyErr := newExecutables(ipv6, "nft").verify(ctx, cfg)          //nolint:contextcheck
	legacy, legacyVerifyErr := newExecutables(ipv6, "legacy").verify(ctx, cfg) //nolint:contextcheck

	if nftVerifyErr != nil && legacyVerifyErr != nil {
		return nil, fmt.Errorf("no valid iptables executable found: %s, %s", nftVerifyErr, legacyVerifyErr)
	}

	if nftVerifyErr != nil {
		return legacy, nil
	}

	// Found DOCKER_OUTPUT chain in iptables-nft
	if nft.foundDockerOutputChain {
		return nft.withFallback(legacy), nil
	}

	if legacyVerifyErr != nil {
		return nft, nil
	}

	// Found DOCKER_OUTPUT chain in iptables-legacy
	if legacy.foundDockerOutputChain {
		return legacy.withFallback(nft), nil
	}

	return nft.withFallback(legacy), nil
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

// findPath attempts to locate the executable named by 'path' on the system.
//
// This function uses exec.LookPath to search for the executable based on the
// following logic:
//   - If 'path' contains a slash (/), it's considered an absolute path and
//     searched for directly.
//   - If 'path' doesn't contain a slash:
//   - LookPath searches for the executable in directories listed in the
//     system's PATH environment variable.
//   - In Go versions before 1.19, a relative path to the current working
//     directory could be returned for non-absolute paths. In Go 1.19 and
//     later, such cases will result in an exec.ErrDot error with the relative
//     path.
//
// The function handles these cases as follows:
// - If no error occurs, the absolute path found by exec.LookPath is returned.
// - If exec.ErrDot is encountered:
//   - The current working directory is retrieved using os.Getwd().
//     If successful:
//   - The relative path found by exec.LookPath is prepended with the current
//     working directory using filepath.Join to create an absolute path.
//   - If getting the current working directory fails:
//   - The original relative path found by LookPath is returned as a fallback
//
// If no path is found or an unexpected error occurs, an empty string is
// returned.
func findPath(path string) string {
	found, err := exec.LookPath(path)
	switch {
	case err == nil:
		return found
	case errors.Is(err, exec.ErrDot):
		// Go 1.19+ behavior: relative path found. Try to prepend the current
		// working directory.
		if pwd, err := os.Getwd(); err == nil {
			return filepath.Join(pwd, found)
		}

		// Couldn't get the current working directory, fallback to the relative
		// path.
		return found
	}

	return ""
}

// getPathsToSearchForExecutable generates a list of potential paths for the
// given executable considering both versions with and without the mode suffix.
//
// This function prioritizes finding the executable with the mode information
// embedded in the name (e.g., iptables-nft) for faster mode verification.
// It achieves this by:
//  1. Adding the nameWithMode (e.g., iptables-nft) as the first potential path.
//  2. Appending paths formed by joining fallbackPaths with nameWithMode (e.g.,
//     /usr/sbin/iptables-nft, /sbin/iptables-nft).
//  3. After checking paths with the mode suffix, it adds the nameWithoutMode
//     (e.g., iptables) as a fallback.
//  4. Similar to step 2, it appends paths formed by joining fallbackPaths with
//     nameWithoutMode.
//
// Finally, the function returns the combined list of potential paths for the
// executable.
func getPathsToSearchForExecutable(
	nameWithMode string,
	nameWithoutMode string,
) []string {
	var paths []string

	paths = append(paths, nameWithMode)
	for _, fallbackPath := range fallbackPaths {
		paths = append(paths, filepath.Join(fallbackPath, nameWithMode))
	}

	paths = append(paths, nameWithoutMode)
	for _, fallbackPath := range fallbackPaths {
		paths = append(paths, filepath.Join(fallbackPath, nameWithoutMode))
	}

	return paths
}

// verifyIptablesMode checks if the provided 'path' corresponds to an iptables
// executable operating in the expected mode.
//
// This function verifies the mode by:
//  1. Executing the iptables command specified by 'path' with the `--version`
//     argument to obtain the version output.
//  2. Parsing the standard output using the `consts.IptablesModeRegex`.
//     - The regex is designed to extract the mode string from the output (e.g.,
//     "legacy" or "nf_tables").
//     - If a match is found, the extracted mode is compared with the expected
//     mode (`mode`) using the `consts.IptablesModeMap`.
//  3. Returning:
//     - `true` if the extracted mode matches the expected mode.
//     - `false` if the command execution fails, parsing fails, or the extracted
//     mode doesn't match the expected mode.
//
// Special Considerations:
// Older iptables versions (e.g., 1.4.21, 1.6.1) may not support the `--version`
// flag and exhibit the following behaviors:
//   - The command exits with a non-zero code and a warning is written to
//     stderr.
//   - A warning is written to stderr but the command exits with code 0.
//
// In these cases, the function assumes the iptables mode is legacy
// (`consts.IptablesModeLegacy`) due to the age of these versions.
func verifyIptablesMode(path string, mode string) bool {
	isVersionMissing := func(output string) bool {
		return strings.Contains(output, "unrecognized option '--version'")
	}

	stdout, stderr, err := execCmd(context.Background(), path, "--version")
	if err != nil {
		return isVersionMissing(err.Error()) && mode == "legacy"
	}

	if stderr != nil && stderr.Len() > 0 && isVersionMissing(stderr.String()) {
		return mode == "legacy"
	}

	matched := IptablesModeRegex.FindStringSubmatch(stdout.String())
	if len(matched) == 2 {
		return slices.Contains(IptablesModeMap[mode], matched[1])
	}

	return false
}

// execCmd executes a command specified by 'path' and its arguments ('args')
// within the provided context ('ctx').
//   - Success: If the command executes successfully, the captured standard
//     output is returned as a bytes.Buffer and nil error.
//   - Error with stderr output: If the command execution encounters an error
//     and there's content captured in the standard error buffer, the error
//     includes the stderr content. The original error is wrapped with a
//     formatted stderr message and a nil buffer is returned.
//   - Error without stderr output: If the command execution encounters an error
//     but there's no captured standard error, the original error is simply
//     returned with a nil buffer.
func execCmd(
	ctx context.Context,
	path string,
	args ...string,
) (*bytes.Buffer, *bytes.Buffer, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	// #nosec G204
	cmd := exec.CommandContext(ctx, path, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if stderr.Len() > 0 {
			stderrTrimmed := strings.TrimSpace(stderr.String())
			stderrLines := strings.Split(stderrTrimmed, "\n")
			stderrFormated := strings.Join(stderrLines, ", ")

			return nil, nil, errors.Errorf("%s: %s", err, stderrFormated)
		}

		return nil, nil, err
	}

	return &stdout, &stderr, nil
}
