package config

import (
	"bufio"
	"bytes"
	"context"
	std_errors "errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/pkg/errors"
	k8s_version "k8s.io/apimachinery/pkg/util/version"

	. "github.com/kumahq/kuma/pkg/transparentproxy/iptables/consts"
)

const (
	WarningDryRunNoValidIptablesFound = "[dry-run]: no valid iptables executables found; the generated iptables rules may differ from those generated in an environment with valid iptables executables"
)

type Version struct {
	k8s_version.Version

	Mode IptablesMode
}

func getIptablesVersion(ctx context.Context, path string) (Version, error) {
	isVersionMissing := func(output string) bool {
		return strings.Contains(output, fmt.Sprintf("unrecognized option '%s'", FlagVersion))
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	// #nosec G204
	cmd := exec.CommandContext(ctx, path, FlagVersion)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	// Older iptables versions (e.g., 1.4.21, 1.6.1) may not support the `--version`
	// flag. Depending on the version, this may result in:
	//   - The command exiting with a non-zero code and a warning written to stderr
	//   - The command exiting with code 0 but still outputting a warning to stderr
	// In these cases, the function assumes the iptables mode is legacy
	switch {
	case err != nil && isVersionMissing(err.Error()):
		return Version{Mode: IptablesModeLegacy}, nil
	case stderr.Len() > 0 && isVersionMissing(stderr.String()):
		return Version{Mode: IptablesModeLegacy}, nil
	case err != nil:
		return Version{}, err
	}

	matched := IptablesModeRegex.FindStringSubmatch(stdout.String())
	if len(matched) < 2 {
		return Version{}, errors.Errorf("unable to parse iptables version in: '%s'", stdout.String())
	}

	version, err := k8s_version.ParseGeneric(matched[1])
	if err != nil {
		return Version{}, errors.Wrapf(err, "invalid iptables version string: '%s'", matched[1])
	}

	if len(matched) < 3 {
		return Version{Version: *version, Mode: IptablesModeLegacy}, nil
	}

	return Version{Version: *version, Mode: IptablesModeMap[matched[2]]}, nil
}

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
	l Logger,
	cniMode bool,
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
			if v, err := getIptablesVersion(ctx, path); err == nil && v.Mode == c.mode {
				return InitializedExecutable{
					Path:    path,
					logger:  l,
					name:    c.name,
					prefix:  c.prefix,
					cniMode: cniMode,
					version: v,
					args:    args,
				}, nil
			}
		}
	}

	return InitializedExecutable{}, errors.Errorf("failed to find executable %s", nameWithMode)
}

type InitializedExecutable struct {
	Path string

	logger  Logger
	name    string
	prefix  string
	cniMode bool
	version Version

	// args holds a set of default parameters or flags that are automatically
	// added to every execution of this executable. These parameters are
	// prepended to any additional arguments provided in the Exec method. This
	// ensures that certain flags or options are always applied whenever the
	// executable is run.
	args []string
}

func (c InitializedExecutable) NeedLock() bool {
	// iptables-nft does not use the xtables lock, so no lock is needed for this
	// mode
	if c.version.Mode == IptablesModeNft {
		return false
	}

	// Only iptables and iptables-restore executables need a lock because they
	// perform write operations
	switch c.name {
	case "", "restore":
		return true
	}

	// For all other cases (such as iptables-save), no lock is needed as they
	// don't perform write operations
	return false
}

type ExecutablesIPvX struct {
	Iptables        Executable
	IptablesSave    Executable
	IptablesRestore Executable
}

func NewExecutablesIPvX(ipv6 bool, mode IptablesMode) ExecutablesIPvX {
	newExecutable := func(name string) Executable {
		return Executable{
			name:   name,
			mode:   mode,
			prefix: IptablesCommandByFamily[ipv6],
		}
	}

	return ExecutablesIPvX{
		Iptables:        newExecutable(""),
		IptablesSave:    newExecutable("save"),
		IptablesRestore: newExecutable("restore"),
	}
}

func (c ExecutablesIPvX) Initialize(
	ctx context.Context,
	l Logger,
	cfg Config,
) (InitializedExecutablesIPvX, error) {
	var errs []error

	iptables, err := c.Iptables.Initialize(ctx, l, cfg.CNIMode, nil)
	if err != nil {
		errs = append(errs, err)
	}

	iptablesSave, err := c.IptablesSave.Initialize(ctx, l, cfg.CNIMode, nil)
	if err != nil {
		errs = append(errs, err)
	}

	restoreArgs := buildRestoreArgs(cfg, c.IptablesRestore.mode)
	iptablesRestore, err := c.IptablesRestore.Initialize(ctx, l, cfg.CNIMode, restoreArgs)
	if err != nil {
		errs = append(errs, err)
	}

	if len(errs) != 0 {
		return InitializedExecutablesIPvX{}, errors.Wrap(
			std_errors.Join(errs...),
			"failed to initialize executables",
		)
	}

	functionality, err := verifyFunctionality(ctx, iptables, iptablesSave)
	if err != nil {
		return InitializedExecutablesIPvX{}, errors.Wrap(
			err,
			"failed to verify functionality",
		)
	}

	return InitializedExecutablesIPvX{
		Iptables:        iptables,
		IptablesSave:    iptablesSave,
		IptablesRestore: iptablesRestore,
		Functionality:   functionality,

		retry:  cfg.Retry,
		logger: l,
	}, nil
}

type InitializedExecutablesIPvX struct {
	Iptables        InitializedExecutable
	IptablesSave    InitializedExecutable
	IptablesRestore InitializedExecutable
	Functionality   Functionality

	retry  Retry
	logger Logger
}

// restore executes the iptables-restore command with the given rules and
// additional arguments. It writes the rules to a temporary file and tries
// to restore the iptables rules from this file. If the command fails, it
// retries the specified number of times.
func (c InitializedExecutablesIPvX) restore(
	ctx context.Context,
	f *os.File,
	quiet bool,
	args ...string,
) (string, error) {
	args = append(args, f.Name())

	argsAll := slices.DeleteFunc(
		slices.Concat(c.IptablesRestore.args, args),
		func(s string) bool { return s == "" },
	)

	for i := 0; i <= c.retry.MaxRetries; i++ {
		c.logger.try = i + 1

		if !quiet {
			c.logger.InfoTry(c.IptablesRestore.Path, strings.Join(argsAll, " "))
		}

		stdout, _, err := c.IptablesRestore.Exec(ctx, args...)
		if err == nil {
			return stdout.String(), nil
		}

		c.logger.ErrorTry(err, "restoring failed:")

		if i < c.retry.MaxRetries {
			if !quiet {
				c.logger.InfoTry("will try again in", c.retry.SleepBetweenRetries)
			}

			time.Sleep(c.retry.SleepBetweenRetries.Duration)
		}
	}

	return "", errors.Errorf("%s failed", c.IptablesRestore.Path)
}

// Restore executes the iptables-restore command with the given rules and the
// --noflush flag to ensure that the current rules are not flushed before
// restoring. This function is a wrapper around the restore function with the
// --noflush flag.
func (c InitializedExecutablesIPvX) Restore(
	ctx context.Context,
	rules string,
	quiet bool,
) (string, error) {
	f, err := createTempFile(c.IptablesRestore.prefix)
	if err != nil {
		return "", err
	}
	defer f.Close()
	defer os.Remove(f.Name())

	// Log the file name and the rules to be written.
	if !quiet {
		c.logger.Info("writing the following rules to file:", f.Name())
		c.logger.InfoWithoutPrefix(strings.TrimSpace(rules))
	}

	if err := writeToFile(rules, f); err != nil {
		return "", err
	}

	return c.restore(ctx, f, quiet, FlagNoFlush)
}

// RestoreWithFlush executes the iptables-restore command with the given rules,
// allowing the current rules to be flushed before restoring. This function is
// a wrapper around the restore function without the --noflush flag.
func (c InitializedExecutablesIPvX) RestoreWithFlush(
	ctx context.Context,
	rules string,
	quiet bool,
) (string, error) {
	// Create a backup file for existing iptables rules.
	backupFile, err := createBackupFile(c.IptablesRestore.prefix)
	if err != nil {
		return "", errors.Wrap(err, "failed to create backup file for iptables rules")
	}
	defer backupFile.Close()

	// Save the current iptables rules to the backup file.
	stdout, _, err := c.IptablesSave.Exec(ctx)
	if err != nil {
		return "", errors.Wrap(err, "failed to execute iptables-save command")
	}

	if err := writeToFile(stdout.String(), backupFile); err != nil {
		return "", errors.Wrap(err, "failed to write current iptables rules to backup file")
	}

	// Create a temporary file for the new iptables rules.
	restoreFile, err := createTempFile(c.IptablesRestore.prefix)
	if err != nil {
		return "", errors.Wrap(err, "failed to create temporary file for new iptables rules")
	}
	defer restoreFile.Close()
	defer os.Remove(restoreFile.Name())

	// Write the new iptables rules to the temporary file.
	if err := writeToFile(rules, restoreFile); err != nil {
		return "", errors.Wrap(err, "failed to write new iptables rules to temporary file")
	}

	// Attempt to restore the new iptables rules from the temporary file.
	output, err := c.restore(ctx, restoreFile, quiet)
	if err != nil {
		c.logger.Errorf("restoring backup file: %s", backupFile.Name())

		if _, err := c.restore(ctx, backupFile, quiet); err != nil {
			c.logger.Warnf("restoring backup failed: %s", err)
		}

		return "", errors.Wrapf(err, "failed to restore rules from file: %s", restoreFile.Name())
	}

	return output, nil
}

// RestoreTest runs iptables-restore with the --test flag to validate the
// iptables rules without applying them.
//
// This function calls the internal `restore` method with the --test flag to
// ensure that the iptables rules specified in the `rules` string are valid. If
// the rules are valid, it returns the output from iptables-restore. If there is
// an error, it returns the error message.
func (c InitializedExecutablesIPvX) RestoreTest(
	ctx context.Context,
	rules string,
) (string, error) {
	f, err := createTempFile(c.IptablesRestore.prefix)
	if err != nil {
		return "", err
	}
	defer f.Close()
	defer os.Remove(f.Name())

	if err := writeToFile(rules, f); err != nil {
		return "", err
	}

	stdout, _, err := c.IptablesRestore.Exec(ctx, FlagTest, f.Name())
	if err != nil {
		// There is an existing bug which occurs on Ubuntu 20.04
		// ref. https://bugs.debian.org/cgi-bin/bugreport.cgi?bug=960003
		if strings.Contains(strings.ToLower(err.Error()), "segmentation fault") {
			c.logger.Warnf(
				`cannot confirm rules are valid because "%s %s" is returning unexpected error: %q. See https://bugs.debian.org/cgi-bin/bugreport.cgi?bug=960003 for more details`,
				c.IptablesRestore.Path,
				FlagTest,
				err,
			)
			return "", nil
		}

		return "", errors.Wrap(err, "rules are invalid")
	}

	return stdout.String(), nil
}

type Executables struct {
	NftIPv4    ExecutablesIPvX
	NftIPv6    ExecutablesIPvX
	LegacyIPv4 ExecutablesIPvX
	LegacyIPv6 ExecutablesIPvX
}

func NewExecutables() Executables {
	return Executables{
		NftIPv4:    NewExecutablesIPvX(false, IptablesModeNft),
		NftIPv6:    NewExecutablesIPvX(true, IptablesModeNft),
		LegacyIPv4: NewExecutablesIPvX(false, IptablesModeLegacy),
		LegacyIPv6: NewExecutablesIPvX(true, IptablesModeLegacy),
	}
}

func (c Executables) InitializeIPv4(
	ctx context.Context,
	l Logger,
	cfg Config,
) (InitializedExecutablesIPvX, ExecutablesIPvX, error) {
	var errs []error

	nft, nftErr := c.NftIPv4.Initialize(ctx, l, cfg)
	if nftErr != nil {
		errs = append(errs, nftErr)
	}

	legacy, legacyErr := c.LegacyIPv4.Initialize(ctx, l, cfg)
	if legacyErr != nil {
		errs = append(errs, legacyErr)
	}

	switch {
	case len(errs) == 2 && cfg.DryRun:
		l.Warn(WarningDryRunNoValidIptablesFound)
		return InitializedExecutablesIPvX{}, ExecutablesIPvX{}, nil
	case len(errs) == 2:
		return InitializedExecutablesIPvX{}, ExecutablesIPvX{}, errors.Wrap(std_errors.Join(errs...), "failed to find valid nft or legacy executables")
	case legacyErr != nil:
		return nft, c.NftIPv6, nil
	case nftErr != nil:
		return legacy, c.LegacyIPv6, nil
	case nft.Functionality.Rules.ExistingRules && legacy.Functionality.Rules.ExistingRules:
		switch {
		case nft.Functionality.Chains.DockerOutput && legacy.Functionality.Chains.DockerOutput:
			fallthrough
		case !nft.Functionality.Chains.DockerOutput && !legacy.Functionality.Chains.DockerOutput:
			l.Warn("conflicting iptables modes detected; both iptables-nft and iptables-legacy have existing rules and/or custom chains. To avoid potential conflicts, iptables-legacy will be ignored, and iptables-nft will be used")
			return nft, c.NftIPv6, nil
		case legacy.Functionality.Chains.DockerOutput:
			return legacy, c.LegacyIPv6, nil
		default:
			return nft, c.NftIPv6, nil
		}
	case legacy.Functionality.Rules.ExistingRules:
		return legacy, c.LegacyIPv6, nil
	default:
		return nft, c.NftIPv6, nil
	}
}

// buildRestoreArgs constructs a slice of flags for restoring iptables rules
// based on the provided configuration and iptables mode.
//
// This function generates a list of command-line flags to be used with
// iptables-restore, tailored to the given parameters:
//   - For non-legacy iptables mode, it returns an empty slice, as no additional
//     flags are required.
//   - For legacy mode, it conditionally adds the `--wait` and `--wait-interval`
//     flags based on the provided configuration values.
func buildRestoreArgs(cfg Config, mode IptablesMode) []string {
	var flags []string

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

// createBackupFile generates a backup file with a specified prefix and a
// timestamp suffix. The file is created in the system's temporary directory and
// is used to store iptables rules for backup purposes.
func createBackupFile(prefix string) (*os.File, error) {
	// Generate a timestamp suffix for the backup file name.
	dateSuffix := time.Now().Format("2006-01-02-150405")

	// Construct the backup file name using the provided prefix and the
	// timestamp suffix.
	fileName := fmt.Sprintf("%s-rules.%s.txt.backup", prefix, dateSuffix)
	filePath := filepath.Join(os.TempDir(), fileName)

	// Create the backup file in the system's temporary directory.
	f, err := os.Create(filePath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create backup file: %s", filePath)
	}

	return f, nil
}

// createTempFile generates a temporary file with a specified prefix. The file
// is created in the system's default temporary directory and is used for
// storing iptables rules temporarily.
//
// This function performs the following steps:
//  1. Constructs a template for the temporary file name using the provided
//     prefix.
//  2. Creates the temporary file in the system's default temporary directory.
func createTempFile(prefix string) (*os.File, error) {
	// Construct a template for the temporary file name using the provided prefix.
	nameTemplate := fmt.Sprintf("%s-rules.*.txt", prefix)

	// Create the temporary file in the system's default temporary directory.
	f, err := os.CreateTemp(os.TempDir(), nameTemplate)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create temporary file with template: %s", nameTemplate)
	}

	return f, nil
}

// writeToFile writes the provided content to the specified file using a
// buffered writer. It ensures that all data is written to the file by flushing
// the buffer.
//
// This function performs the following steps:
//  1. Writes the content to the file using a buffered writer for efficiency.
//  2. Flushes the buffer to ensure all data is written to the file.
func writeToFile(content string, f *os.File) error {
	// Write the content to the file using a buffered writer.
	writer := bufio.NewWriter(f)
	if _, err := writer.WriteString(content); err != nil {
		return errors.Wrapf(err, "failed to write to file: %s", f.Name())
	}

	// Flush the buffer to ensure all data is written.
	if err := writer.Flush(); err != nil {
		return errors.Wrapf(err, "failed to flush the buffered writer for file: %s", f.Name())
	}

	return nil
}

func handleRunError(err error, stderr *bytes.Buffer) error {
	if stderr.Len() > 0 {
		stderrTrimmed := strings.TrimSpace(stderr.String())
		stderrLines := strings.Split(stderrTrimmed, "\n")
		stderrFormated := strings.Join(stderrLines, ", ")

		return errors.Errorf("%s: %s", err, stderrFormated)
	}

	return err
}
