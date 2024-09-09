package config

import (
	"bufio"
	"bytes"
	"context"
	std_errors "errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/vishvananda/netlink"

	. "github.com/kumahq/kuma/pkg/transparentproxy/iptables/consts"
)

type Executable struct {
	name string
	mode IptablesMode
	// prefix represents the prefix used for iptables executables, which varies
	// based on whether it's for IPv4 or IPv6 operations. For IPv4, it can be
	// `iptables` (for binaries such as `iptables`, `iptables-restore`, or
	// `iptables-save`). For IPv6, it can be `ip6tables` (for binaries such as
	// `ip6tables`, `ip6tables-restore`, or `ip6tables-save`). This property
	// facilitates constructing the full executable names, generating temporary
	// file names for iptables rules, and other related operations.
	prefix string
}

func (c Executable) Initialize(
	ctx context.Context,
	args []string,
) (InitializedExecutable, error) {
	// ip{6}tables-{nft|legacy}, ip{6}tables-{nft|legacy}-save,
	// ip{6}tables-{nft|legacy}-restore
	nameWithMode := joinNonEmptyWithHyphen(c.prefix, string(c.mode), c.name)
	// ip{6}tables, ip{6}tables-save, ip{6}tables-restore
	nameWithoutMode := joinNonEmptyWithHyphen(c.prefix, c.name)

	paths := getPathsToSearchForExecutable(nameWithMode, nameWithoutMode)

	for _, path := range paths {
		if found := findPath(path); found != "" {
			if verifyIptablesMode(ctx, path, c.mode) {
				return InitializedExecutable{
					Executable: c,
					Path:       path,
					args:       args,
				}, nil
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

	// args holds a set of default parameters or flags that are automatically
	// added to every execution of this executable. These parameters are
	// prepended to any additional arguments provided in the Exec method. This
	// ensures that certain flags or options are always applied whenever the
	// executable is run.
	args []string
}

func (c InitializedExecutable) Exec(
	ctx context.Context,
	args ...string,
) (*bytes.Buffer, *bytes.Buffer, error) {
	return execCmd(ctx, c.Path, append(c.args, args...)...)
}

type Executables struct {
	Iptables        Executable
	IptablesSave    Executable
	IptablesRestore Executable
}

func NewExecutables(ipv6 bool, mode IptablesMode) Executables {
	newExecutable := func(name string) Executable {
		return Executable{
			name:   name,
			mode:   mode,
			prefix: IptablesCommandByFamily[ipv6],
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
	l Logger,
	cfg Config,
) (InitializedExecutables, error) {
	var errs []error

	iptables, err := c.Iptables.Initialize(ctx, nil)
	if err != nil {
		errs = append(errs, err)
	}

	iptablesSave, err := c.IptablesSave.Initialize(ctx, nil)
	if err != nil {
		errs = append(errs, err)
	}

	restoreArgs := buildRestoreArgs(cfg, c.IptablesRestore.mode)
	iptablesRestore, err := c.IptablesRestore.Initialize(ctx, restoreArgs)
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

		retry:  cfg.Retry,
		logger: l,
	}, nil
}

type InitializedExecutables struct {
	Iptables        InitializedExecutable
	IptablesSave    InitializedExecutable
	IptablesRestore InitializedExecutable
	Functionality   Functionality

	retry  RetryConfig
	logger Logger
}

// writeRulesToFile writes the provided iptables rules to a temporary file and
// returns the file path.
//
// This method performs the following steps:
//  1. Creates a temporary file with a name based on the
//     `IptablesRestore.prefix` field of the `InitializedExecutables` struct.
//  2. Logs the contents that will be written to the file.
//  3. Writes the rules to the file using a buffered writer for efficiency.
//  4. Flushes the buffer to ensure all data is written to the file.
//  5. Returns the file path if successful, or an error if any step fails.
//
// Args:
//
//	rules (string): The iptables rules to be written to the temporary file.
//
// Returns:
//
//	string: The path to the temporary file containing the iptables rules.
//	error: An error if the file creation, writing, or flushing fails.
func (c InitializedExecutables) writeRulesToFile(rules string) (string, error) {
	// Create a temporary file with a name template based on the IptablesRestore
	// prefix.
	nameTemplate := fmt.Sprintf("%s-rules.*.txt", c.IptablesRestore.prefix)
	f, err := os.CreateTemp("", nameTemplate)
	if err != nil {
		return "", errors.Wrapf(
			err,
			"failed to create temporary file: %s",
			nameTemplate,
		)
	}
	defer f.Close()

	// Remove the temporary file if an error occurs after creation.
	defer func() {
		if err != nil {
			os.Remove(f.Name())
		}
	}()

	// Log the file name and the rules to be written.
	c.logger.Info("writing the following rules to file:", f.Name())
	c.logger.InfoWithoutPrefix(strings.TrimSpace(rules))

	// Write the rules to the file using a buffered writer.
	writer := bufio.NewWriter(f)
	if _, err = writer.WriteString(rules); err != nil {
		return "", errors.Wrapf(
			err,
			"failed to write rules to the temporary file: %s",
			f.Name(),
		)
	}

	// Flush the buffer to ensure all data is written.
	if err = writer.Flush(); err != nil {
		return "", errors.Wrapf(
			err,
			"failed to flush the buffered writer for file: %s",
			f.Name(),
		)
	}

	// Return the file path.
	return f.Name(), nil
}

func (c InitializedExecutables) Restore(
	ctx context.Context,
	rules string,
) (string, error) {
	fileName, err := c.writeRulesToFile(rules)
	if err != nil {
		return "", err
	}
	defer os.Remove(fileName)

	for i := 0; i <= c.retry.MaxRetries; i++ {
		c.logger.try = i + 1

		c.logger.InfoTry(
			c.IptablesRestore.Path,
			strings.Join(c.IptablesRestore.args, " "),
			fileName,
		)

		stdout, _, err := c.IptablesRestore.Exec(ctx, fileName)
		if err == nil {
			return stdout.String(), nil
		}

		c.logger.ErrorTry(err, "restoring failed:")

		if i < c.retry.MaxRetries {
			c.logger.InfoTry("will try again in", c.retry.SleepBetweenReties)
			time.Sleep(c.retry.SleepBetweenReties)
		}
	}

	return "", errors.Errorf("%s failed", c.IptablesRestore.Path)
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

// Initialize attempts to initialize both IPv4 and IPv6 executables within the
// given context. It ensures proper configuration for IPv6 if necessary.
//
// This method performs the following steps:
//  1. Attempts to initialize the IPv4 executables with the provided context,
//     Config, and Logger. If an error occurs, it returns an error indicating
//     the failure to initialize IPv4 executables.
//  2. If IPv6 is enabled in the configuration, it attempts to initialize the
//     IPv6 executables with the provided context, Config, and Logger. If an
//     error occurs, it returns an error indicating the failure to initialize
//     IPv6 executables.
//  3. If IPv6 initialization is successful, it attempts to configure the IPv6
//     outbound address. If this configuration fails, a warning is logged, and
//     IPv6 rules will be skipped.
//
// Args:
// - ctx (context.Context): The context for managing request lifetime.
// - cfg (Config): Configuration settings for initializing the executables.
// - l (Logger): Logger for logging initialization steps and errors.
//
// Returns:
//   - InitializedExecutablesIPvX: Struct containing the initialized executables
//     for both IPv4 and IPv6.
//   - error: Error indicating the failure of either IPv4 or IPv6
//     initialization.
func (c ExecutablesIPvX) Initialize(
	ctx context.Context,
	l Logger,
	cfg Config,
) (InitializedExecutablesIPvX, error) {
	var err error

	initialized := InitializedExecutablesIPvX{ExecutablesIPvX: c}

	initialized.IPv4, err = c.IPv4.Initialize(ctx, l, cfg)
	if err != nil {
		return InitializedExecutablesIPvX{}, errors.Wrap(
			err,
			"failed to initialize IPv4 executables",
		)
	}

	if cfg.IPv6 {
		initialized.IPv6, err = c.IPv6.Initialize(ctx, l, cfg)
		if err != nil {
			return InitializedExecutablesIPvX{}, errors.Wrap(
				err,
				"failed to initialize IPv6 executables",
			)
		}

		if err := configureIPv6OutboundAddress(); err != nil {
			initialized.IPv6 = InitializedExecutables{}
			l.Warn("failed to configure IPv6 outbound address. IPv6 rules "+
				"will be skipped:", err)
		}
	}

	return initialized, nil
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
	l Logger,
	cfg Config,
) (InitializedExecutablesIPvX, error) {
	var errs []error

	nft, nftErr := c.Nft.Initialize(ctx, l, cfg)
	if nftErr != nil {
		errs = append(errs, nftErr)
	}

	legacy, legacyErr := c.Legacy.Initialize(ctx, l, cfg)
	if legacyErr != nil {
		errs = append(errs, legacyErr)
	}

	switch {
	// Dry-run mode when no valid iptables executables are found.
	case len(errs) == 2 && cfg.DryRun:
		l.Warn("dry-run mode: No valid iptables executables found. The " +
			"generated iptables rules may differ from those generated in an " +
			"environment with valid iptables executables")
		return InitializedExecutablesIPvX{}, nil
	// Regular mode when no vaild iptables executables are found
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
		l.Warn("conflicting iptables modes detected. Two iptables " +
			"versions (iptables-nft and iptables-legacy) were found. " +
			"Both contain a nat table with a chain named 'DOCKER_OUTPUT'. " +
			"To avoid potential conflicts, iptables-legacy will be ignored " +
			"and iptables-nft will be used")
		return nft, nil
	case legacy.hasDockerOutputChain():
		return legacy, nil
	default:
		return nft, nil
	}
}

// buildRestoreArgs constructs a slice of flags for restoring iptables rules
// based on the provided wait time, wait interval, and iptables mode.
//
// This function generates a list of command-line flags to be used with
// iptables-restore, tailored to the given parameters:
//   - By default, it includes the `--noflush` flag to prevent flushing of
//     existing rules.
//   - If the iptables mode is not legacy, it returns only the `--noflush` flag.
//   - For legacy mode, it conditionally adds the `--wait` and `--wait-interval`
//     flags based on the provided wait time and waitInterval.
//
// Args:
//
//	wait (uint): The wait time in seconds for iptables-restore to wait for the
//	 xtables lock before aborting.
//	interval (uint): The wait interval in seconds between attempts to acquire
//	 the xtables lock.
//	mode (IptablesMode): The mode of iptables in use, determining the
//	 applicability of the wait flags.
//
// Returns:
//
//	[]string: A slice of strings representing the constructed flags for
//	 iptables-restore.
//
// Example:
//
//	buildRestoreArgs(5, 2, IptablesModeNft) # Returns: []string{"--noflush"}
//	buildRestoreArgs(5, 2, IptablesModeLegacy)
//	 # Returns: []string{"--noflush", "--wait=5", "--wait-interval=2"}
func buildRestoreArgs(cfg Config, mode IptablesMode) []string {
	flags := []string{FlagNoFlush}

	if mode != IptablesModeLegacy {
		return flags
	}

	if cfg.Wait > 0 {
		flags = append(flags, fmt.Sprintf("%s=%d", FlagWait, cfg.Wait))
	}

	if cfg.WaitInterval > 0 {
		flags = append(flags, fmt.Sprintf(
			"%s=%d",
			FlagWaitInterval,
			cfg.WaitInterval,
		))
	}

	return flags
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
<<<<<<< HEAD
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
=======
	case err != nil && isVersionMissing(err.Error()):
		return Version{Mode: consts.IptablesModeLegacy}, nil
	case stderr.Len() > 0 && isVersionMissing(stderr.String()):
		return Version{Mode: consts.IptablesModeLegacy}, nil
	case err != nil:
		return Version{}, formatIptablesVersionError(err.Error())
	}

	matched := consts.IptablesModeRegex.FindStringSubmatch(stdout.String())
	if len(matched) < 2 {
		return Version{}, errors.Wrap(formatIptablesVersionError(stdout.String()), "unable to parse iptables version")
	}

	version, err := k8s_version.ParseGeneric(matched[1])
	if err != nil {
		return Version{}, errors.Wrapf(formatIptablesVersionError(err.Error()), "invalid iptables version string: '%s'", matched[1])
	}

	if len(matched) < 3 {
		return Version{Version: *version, Mode: consts.IptablesModeLegacy}, nil
	}

	return Version{Version: *version, Mode: consts.IptablesModeMap[matched[2]]}, nil
>>>>>>> 23ecef9db (chore(deps): bump golangci-lint to v1.60.3 (#11362))
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

// configureIPv6OutboundAddress sets up a dedicated IPv6 address (::6) on the
// loopback interface ("lo") for our transparent proxy functionality.
//
// Background:
//   - The default IPv6 configuration (prefix length 128) only allows binding to
//     the loopback address (::1).
//   - Our transparent proxy requires a distinct IPv6 address (::6 in this case)
//     to identify traffic processed by the kuma-dp sidecar.
//   - This identification allows for further processing and avoids redirection
//     loops.
//
// This function is equivalent to running the command:
// `ip -6 addr add "::6/128" dev lo`
func configureIPv6OutboundAddress() error {
	link, err := netlink.LinkByName("lo")
	if err != nil {
		return errors.Wrap(err, "failed to find loopback interface ('lo')")
	}

	// Equivalent to "::6/128"
	addr := &netlink.Addr{
		IPNet: &net.IPNet{
			IP:   net.ParseIP("::6"),
			Mask: net.CIDRMask(128, 128),
		},
	}

	if err := netlink.AddrAdd(link, addr); err != nil {
		// Address already exists, ignore error and continue
		if strings.Contains(strings.ToLower(err.Error()), "file exists") {
			return nil
		}

		return errors.Wrapf(
			err,
			"failed to add IPv6 address %s to loopback interface",
			addr.IPNet,
		)
	}

<<<<<<< HEAD
	return nil
=======
	return result
}

func formatIptablesVersionError(msg string) error {
	msgWithoutNewLines := strings.ReplaceAll(msg, "\n", " ")
	msgWithoutDuplicatedSpaces := strings.ReplaceAll(msgWithoutNewLines, "  ", " ")
	msgWithoutDotSuffix := strings.TrimRight(msgWithoutDuplicatedSpaces, ".")

	if len(msgWithoutDotSuffix) > 500 {
		msgWithoutDotSuffix = fmt.Sprintf("%.500s...", msgWithoutDotSuffix)
	}

	return errors.New(msgWithoutDotSuffix)
>>>>>>> 23ecef9db (chore(deps): bump golangci-lint to v1.60.3 (#11362))
}
