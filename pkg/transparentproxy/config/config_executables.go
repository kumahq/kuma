package config

import (
	"bytes"
	"context"
	std_errors "errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/pkg/errors"

	. "github.com/kumahq/kuma/pkg/transparentproxy/iptables/consts"
)

type Executable struct {
	name string
	mode IptablesMode
	ipv6 bool
}

func (c Executable) Initialize(
	ctx context.Context,
) (InitializedExecutable, error) {
	prefix := IptablesCommandByFamily[c.ipv6]

	// ip{6}tables-{nft|legacy}, ip{6}tables-{nft|legacy}-save,
	// ip{6}tables-{nft|legacy}-restore
	nameWithMode := joinNonEmptyWithHyphen(prefix, string(c.mode), c.name)
	// ip{6}tables, ip{6}tables-save, ip{6}tables-restore
	nameWithoutMode := joinNonEmptyWithHyphen(prefix, c.name)

	paths := getPathsToSearchForExecutable(nameWithMode, nameWithoutMode)

	for _, path := range paths {
		if found := findPath(path); found != "" {
			if verifyIptablesMode(ctx, path, c.mode) {
				return InitializedExecutable{Executable: c, Path: path}, nil
			}
		}
	}

	return InitializedExecutable{}, errors.Errorf(
		"failed to find executable %s",
		nameWithMode,
	)
}

type InitializedExecutable struct {
	Executable
	Path string
}

func (c InitializedExecutable) Exec(
	ctx context.Context,
	args ...string,
) (*bytes.Buffer, *bytes.Buffer, error) {
	return execCmd(ctx, c.Path, args...)
}

type Executables struct {
	Iptables        Executable
	IptablesSave    Executable
	IptablesRestore Executable
}

func NewExecutables(ipv6 bool, mode IptablesMode) Executables {
	newExecutable := func(name string) Executable {
		return Executable{
			name: name,
			mode: mode,
			ipv6: ipv6,
		}
	}

	return Executables{
		Iptables:        newExecutable(""),
		IptablesSave:    newExecutable("save"),
		IptablesRestore: newExecutable("restore"),
	}
}

func (c Executables) Initialize(
	ctx context.Context,
) (InitializedExecutables, error) {
	var errs []error

	iptables, err := c.Iptables.Initialize(ctx)
	if err != nil {
		errs = append(errs, err)
	}

	iptablesSave, err := c.IptablesSave.Initialize(ctx)
	if err != nil {
		errs = append(errs, err)
	}

	iptablesRestore, err := c.IptablesRestore.Initialize(ctx)
	if err != nil {
		errs = append(errs, err)
	}

	if len(errs) != 0 {
		return InitializedExecutables{}, errors.Wrap(
			std_errors.Join(errs...),
			"failed to initialize executables",
		)
	}

	functionality, err := verifyFunctionality(ctx, iptables, iptablesSave)
	if err != nil {
		return InitializedExecutables{}, errors.Wrap(
			err,
			"failed to verify functionality",
		)
	}

	return InitializedExecutables{
		Iptables:        iptables,
		IptablesSave:    iptablesSave,
		IptablesRestore: iptablesRestore,
		Functionality:   functionality,
	}, nil
}

type InitializedExecutables struct {
	Iptables        InitializedExecutable
	IptablesSave    InitializedExecutable
	IptablesRestore InitializedExecutable
	Functionality   Functionality
}

type ExecutablesIPvX struct {
	IPv4 Executables
	IPv6 Executables
	Mode IptablesMode
}

func NewExecutablesIPvX(mode IptablesMode) ExecutablesIPvX {
	return ExecutablesIPvX{
		IPv4: NewExecutables(false, mode),
		IPv6: NewExecutables(true, mode),
		Mode: mode,
	}
}

func (c ExecutablesIPvX) Initialize(
	ctx context.Context,
) (InitializedExecutablesIPvX, error) {
	var errs []error

	ipv4, ipv4Err := c.IPv4.Initialize(ctx)
	if ipv4Err != nil {
		errs = append(errs, ipv4Err)
	}

	ipv6, ipv6Err := c.IPv6.Initialize(ctx)
	if ipv6Err != nil {
		errs = append(errs, ipv6Err)
	}

	if len(errs) == 2 {
		return InitializedExecutablesIPvX{}, errors.Wrap(
			std_errors.Join(errs...),
			"failed to find valid IPv4 or IPv6 executables",
		)
	}

	return InitializedExecutablesIPvX{
		ExecutablesIPvX: c,
		IPv4:            ipv4,
		IPv6:            ipv6,
	}, nil
}

type InitializedExecutablesIPvX struct {
	ExecutablesIPvX
	IPv4 InitializedExecutables
	IPv6 InitializedExecutables
}

func (c InitializedExecutablesIPvX) hasDockerOutputChain() bool {
	return c.IPv4.Functionality.Chains.DockerOutput ||
		c.IPv6.Functionality.Chains.DockerOutput
}

type ExecutablesNftLegacy struct {
	Nft    ExecutablesIPvX
	Legacy ExecutablesIPvX
}

func NewExecutablesNftLegacy() ExecutablesNftLegacy {
	return ExecutablesNftLegacy{
		Nft:    NewExecutablesIPvX(IptablesModeNft),
		Legacy: NewExecutablesIPvX(IptablesModeLegacy),
	}
}

func (c ExecutablesNftLegacy) Initialize(
	ctx context.Context,
	cfg Config,
) (InitializedExecutablesIPvX, error) {
	var errs []error

	// When dry run there is no need continue initialization
	if cfg.DryRun {
		return InitializedExecutablesIPvX{}, nil
	}

	nft, nftErr := c.Nft.Initialize(ctx)
	if nftErr != nil {
		errs = append(errs, nftErr)
	}

	legacy, legacyErr := c.Legacy.Initialize(ctx)
	if legacyErr != nil {
		errs = append(errs, legacyErr)
	}

	switch {
	case len(errs) == 2:
		return InitializedExecutablesIPvX{}, errors.Wrap(
			std_errors.Join(errs...),
			"failed to find valid nft or legacy executables",
		)
	// No valid legacy executables
	case legacyErr != nil:
		return nft, nil
	// No valid nft executables
	case nftErr != nil:
		return legacy, nil
	// Both types of executables contain custom DOCKER_OUTPUT chain in nat
	// table. We are prioritizing nft
	case nft.hasDockerOutputChain() && legacy.hasDockerOutputChain():
		fmt.Fprintln(
			cfg.RuntimeStderr,
			"[WARNING] conflicting iptables modes detected. Two iptables"+
				" versions (iptables-nft and iptables-legacy) were found."+
				" Both contain a nat table with a chain named 'DOCKER_OUTPUT'."+
				" To avoid potential conflicts, iptables-legacy will be"+
				" ignored and iptables-nft will be used.",
		)
		return nft, nil
	case legacy.hasDockerOutputChain():
		return legacy, nil
	default:
		return nft, nil
	}
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
func verifyIptablesMode(
	ctx context.Context,
	path string,
	mode IptablesMode,
) bool {
	isVersionMissing := func(output string) bool {
		return strings.Contains(
			output,
			fmt.Sprintf("unrecognized option '%s'", FlagVersion),
		)
	}

	stdout, stderr, err := execCmd(ctx, path, FlagVersion)
	if err != nil {
		return isVersionMissing(err.Error()) && mode == IptablesModeLegacy
	}

	if stderr != nil && stderr.Len() > 0 && isVersionMissing(stderr.String()) {
		return mode == IptablesModeLegacy
	}

	matched := IptablesModeRegex.FindStringSubmatch(stdout.String())
	if len(matched) == 2 {
		return slices.Contains(IptablesModeMap[mode], matched[1])
	}

	return false
}

// getPathsToSearchForExecutable generates a list of potential paths for the
// given executable considering both versions with and without the mode suffix.
//
// This function prioritizes finding the executable with the mode information
// embedded in the name (e.g., iptables-nft) for faster mode verification.
// It achieves this by:
//  1. Adding the `nameWithMode` (e.g., iptables-nft) as the first potential
//     path.
//  2. Appending paths formed by joining `FallbackExecutablesSearchLocations`
//     with `nameWithMode` (e.g., /usr/sbin/iptables-nft, /sbin/iptables-nft).
//  3. After checking paths with the mode suffix, it adds the `nameWithoutMode`
//     (e.g., iptables) as a fallback.
//  4. Similar to step 2, it appends paths formed by joining
//     `FallbackExecutablesSearchLocations` with `nameWithoutMode`.
//
// Finally, the function returns the combined list of potential paths for the
// executable.
func getPathsToSearchForExecutable(
	nameWithMode string,
	nameWithoutMode string,
) []string {
	var paths []string

	paths = append(paths, nameWithMode)
	for _, fallbackPath := range FallbackExecutablesSearchLocations {
		paths = append(paths, filepath.Join(fallbackPath, nameWithMode))
	}

	paths = append(paths, nameWithoutMode)
	for _, fallbackPath := range FallbackExecutablesSearchLocations {
		paths = append(paths, filepath.Join(fallbackPath, nameWithoutMode))
	}

	return paths
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
