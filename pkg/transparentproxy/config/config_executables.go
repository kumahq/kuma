package config

import (
	"bytes"
	"context"
	std_errors "errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/exp/maps"
	k8s_version "k8s.io/apimachinery/pkg/util/version"

	. "github.com/kumahq/kuma/pkg/transparentproxy/consts"
	util_maps "github.com/kumahq/kuma/pkg/util/maps"
)

const (
	WarningDryRunNoValidIptablesFound = "[dry-run]: no valid iptables executables found; the generated iptables rules may differ from those generated in an environment with valid iptables executables"
)

type Executable struct {
	name string
	path string
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
	initialized := InitializedExecutable{
		logger:  l,
		name:    c.name,
		prefix:  c.prefix,
		cniMode: cniMode,
		args:    args,
	}

	// ip{6}tables-{nft|legacy}, ip{6}tables-{nft|legacy}-save,
	// ip{6}tables-{nft|legacy}-restore
	nameWithMode := joinNonEmptyWithHyphen(c.prefix, string(c.mode), c.name)
	// ip{6}tables, ip{6}tables-save, ip{6}tables-restore
	nameWithoutMode := joinNonEmptyWithHyphen(c.prefix, c.name)

	if c.path != "" {
		if found := findPath(c.path); found != "" {
			v, err := getIptablesVersion(ctx, found)
			if err != nil {
				return InitializedExecutable{}, errors.Wrapf(err, "invalid executable at specified path '%s' for '%s'", c.path, nameWithoutMode)
			}

			initialized.Path = found
			initialized.version = v

			return initialized, nil
		}

		return InitializedExecutable{}, errors.Errorf("specified path '%s' for executable '%s' does not exist", c.path, nameWithoutMode)
	}

	for _, path := range getPathsToSearchForExecutable(nameWithMode, nameWithoutMode) {
		if found := findPath(path); found != "" {
			if v, err := getIptablesVersion(ctx, found); err == nil && v.Mode == c.mode {
				initialized.Path = found
				initialized.version = v
				return initialized, nil
			}
		}
	}

	return InitializedExecutable{}, errors.Errorf("could not locate executable '%s' with mode '%s'", nameWithoutMode, c.mode)
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

	mode IptablesMode
}

func (c ExecutablesIPvX) WithPaths(iptables, iptablesSave, iptablesRestore string) ExecutablesIPvX {
	newExecutable := func(e Executable, path string) Executable {
		return Executable{
			name:   e.name,
			mode:   e.mode,
			prefix: e.prefix,
			path:   path,
		}
	}

	return ExecutablesIPvX{
		Iptables:        newExecutable(c.Iptables, iptables),
		IptablesSave:    newExecutable(c.IptablesSave, iptablesSave),
		IptablesRestore: newExecutable(c.IptablesRestore, iptablesRestore),
		mode:            c.mode,
	}
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
		mode:            mode,
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
		return InitializedExecutablesIPvX{}, errors.Wrap(std_errors.Join(errs...), "initialization of one or more executables failed")
	}

	mode, err := inferIptablesMode(iptables, iptablesSave, iptablesRestore)
	if err != nil {
		return InitializedExecutablesIPvX{}, errors.Wrap(err, "failed to infer consistent iptables mode")
	}

	functionality, err := verifyFunctionality(ctx, iptables, iptablesSave)
	if err != nil {
		return InitializedExecutablesIPvX{}, errors.Wrap(err, "functionality verification failed")
	}

	return InitializedExecutablesIPvX{
		Iptables:        iptables,
		IptablesSave:    iptablesSave,
		IptablesRestore: iptablesRestore,
		Functionality:   functionality,

		mode: mode,

		retry:  cfg.Retry,
		logger: l,
	}, nil
}

type InitializedExecutablesIPvX struct {
	Iptables        InitializedExecutable
	IptablesSave    InitializedExecutable
	IptablesRestore InitializedExecutable
	Functionality   Functionality

	mode IptablesMode

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
	// Embedded structs to allow unmarshalling executable paths from a flat configuration file
	// instead of requiring nested objects
	ExecutablesPathsIPv4
	ExecutablesPathsIPv6
	nftIPv4    ExecutablesIPvX
	nftIPv6    ExecutablesIPvX
	legacyIPv4 ExecutablesIPvX
	legacyIPv6 ExecutablesIPvX
}

func NewExecutables() Executables {
	return Executables{
		nftIPv4:    NewExecutablesIPvX(false, IptablesModeNft),
		nftIPv6:    NewExecutablesIPvX(true, IptablesModeNft),
		legacyIPv4: NewExecutablesIPvX(false, IptablesModeLegacy),
		legacyIPv6: NewExecutablesIPvX(true, IptablesModeLegacy),
	}
}

func (c *Executables) InitializeIPv4(
	ctx context.Context,
	l Logger,
	cfg Config,
) (InitializedExecutablesIPvX, error) {
	var errs []error

	if initialized, ok, err := tryInitializeExecutablePaths(ctx, l, cfg, c.ExecutablesPathsIPv4); err != nil {
		l.Warn(err)
	} else if ok {
		return initialized, nil
	}

	nft, nftErr := c.nftIPv4.Initialize(ctx, l, cfg)
	if nftErr != nil {
		errs = append(errs, nftErr)
	}

	legacy, legacyErr := c.legacyIPv4.Initialize(ctx, l, cfg)
	if legacyErr != nil {
		errs = append(errs, legacyErr)
	}

	switch {
	case len(errs) == 2 && cfg.DryRun:
		l.Warn(WarningDryRunNoValidIptablesFound)
		return InitializedExecutablesIPvX{}, nil
	case len(errs) == 2:
		return InitializedExecutablesIPvX{}, errors.Wrap(std_errors.Join(errs...), "failed to find valid nft or legacy executables")
	case legacyErr != nil:
		return nft, nil
	case nftErr != nil:
		return legacy, nil
	case nft.Functionality.Rules.ExistingRules && legacy.Functionality.Rules.ExistingRules:
		switch {
		case nft.Functionality.Chains.DockerOutput && legacy.Functionality.Chains.DockerOutput:
			fallthrough
		case !nft.Functionality.Chains.DockerOutput && !legacy.Functionality.Chains.DockerOutput:
			l.Warn("conflicting iptables modes detected; both iptables-nft and iptables-legacy have existing rules and/or custom chains. To avoid potential conflicts, iptables-legacy will be ignored, and iptables-nft will be used")
			return nft, nil
		case legacy.Functionality.Chains.DockerOutput:
			return legacy, nil
		default:
			return nft, nil
		}
	case legacy.Functionality.Rules.ExistingRules:
		return legacy, nil
	default:
		return nft, nil
	}
}

func (c *Executables) InitializeIPv6(
	ctx context.Context,
	l Logger,
	cfg Config,
	modeIPv4 IptablesMode,
) (InitializedExecutablesIPvX, error) {
	if initialized, ok, err := tryInitializeExecutablePaths(ctx, l, cfg, c.ExecutablesPathsIPv6); err != nil {
		l.Warn(err)
	} else if ok {
		return initialized, nil
	}

	switch modeIPv4 {
	case IptablesModeNft:
		return c.nftIPv6.Initialize(ctx, l, cfg)
	case IptablesModeLegacy:
		return c.legacyIPv6.Initialize(ctx, l, cfg)
	default:
		return InitializedExecutablesIPvX{}, errors.Errorf("unknown iptables mode '%s'", modeIPv4)
	}
}

func (c *Executables) Set(s string) error {
	var errs []error

	if s = strings.TrimSpace(s); s == "" {
		return nil
	}

	for _, block := range removeEmptyStrings(strings.Split(s, ",")) {
		name, path, found := strings.Cut(block, ":")
		if !found {
			errs = append(
				errs,
				errors.Errorf("invalid format in '%s': expected '<name>:<path>' (e.g., 'iptables:/usr/sbin/iptables' or 'ip6tables-save:/usr/sbin/ip6tables-save')", block),
			)
			continue
		}

		cleanPath := filepath.Clean(path)

		switch name {
		case Iptables:
			c.ExecutablesPathsIPv4.Iptables = cleanPath
		case IptablesSave:
			c.ExecutablesPathsIPv4.IptablesSave = cleanPath
		case IptablesRestore:
			c.ExecutablesPathsIPv4.IptablesRestore = cleanPath
		case Ip6tables:
			c.ExecutablesPathsIPv6.Ip6tables = cleanPath
		case Ip6tablesSave:
			c.ExecutablesPathsIPv6.Ip6tablesSave = cleanPath
		case Ip6tablesRestore:
			c.ExecutablesPathsIPv6.Ip6tablesRestore = cleanPath
		default:
			errs = append(
				errs,
				errors.Errorf("unsupported executable name '%s': valid names are %s", name, getNamesString(c.ExecutablesPathsIPv4, c.ExecutablesPathsIPv6)),
			)
		}
	}

	return std_errors.Join(errs...)
}

func (c *Executables) String() string {
	var result []string

	for name, path := range getPathsMap(c.ExecutablesPathsIPv4, c.ExecutablesPathsIPv6) {
		if path != "" {
			result = append(result, fmt.Sprintf("%s:%s", name, path))
		}
	}

	return strings.Join(result, ",")
}

func (c *Executables) Type() string {
	return "name:path[,name:path...]"
}

type executablesPaths interface {
	getPathsMap() map[string]string
	convert() ExecutablesIPvX
}

var _ executablesPaths = ExecutablesPathsIPv4{}

type ExecutablesPathsIPv4 struct {
	Iptables        string `json:"iptables"`
	IptablesSave    string `json:"iptables-save"`
	IptablesRestore string `json:"iptables-restore"`
}

func (c ExecutablesPathsIPv4) getPathsMap() map[string]string {
	return map[string]string{
		Iptables:        c.Iptables,
		IptablesSave:    c.IptablesSave,
		IptablesRestore: c.IptablesRestore,
	}
}

func (c ExecutablesPathsIPv4) convert() ExecutablesIPvX {
	return NewExecutablesIPvX(false, IptablesModeUnknown).
		WithPaths(c.Iptables, c.IptablesSave, c.IptablesRestore)
}

var _ executablesPaths = ExecutablesPathsIPv6{}

type ExecutablesPathsIPv6 struct {
	Ip6tables        string `json:"ip6tables"`
	Ip6tablesSave    string `json:"ip6tables-save"`
	Ip6tablesRestore string `json:"ip6tables-restore"`
}

func (c ExecutablesPathsIPv6) getPathsMap() map[string]string {
	return map[string]string{
		Ip6tables:        c.Ip6tables,
		Ip6tablesSave:    c.Ip6tablesSave,
		Ip6tablesRestore: c.Ip6tablesRestore,
	}
}

func (c ExecutablesPathsIPv6) convert() ExecutablesIPvX {
	return NewExecutablesIPvX(true, IptablesModeUnknown).
		WithPaths(c.Ip6tables, c.Ip6tablesSave, c.Ip6tablesRestore)
}

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
		return Version{}, formatIptablesVersionErrorf(err.Error())
	}

	matched := IptablesModeRegex.FindStringSubmatch(stdout.String())
	if len(matched) < 2 {
		return Version{}, errors.Wrap(formatIptablesVersionErrorf(stdout.String()), "unable to parse iptables version")
	}

	version, err := k8s_version.ParseGeneric(matched[1])
	if err != nil {
		return Version{}, errors.Wrapf(formatIptablesVersionErrorf(err.Error()), "invalid iptables version string: '%s'", matched[1])
	}

	if len(matched) < 3 {
		return Version{Version: *version, Mode: IptablesModeLegacy}, nil
	}

	return Version{Version: *version, Mode: IptablesModeMap[matched[2]]}, nil
}

func getExecutablesModesString(executables ...InitializedExecutable) string {
	var result []string

	for _, e := range executables {
		result = append(
			result,
			fmt.Sprintf("%s: %s (%s)", joinNonEmptyWithHyphen(e.prefix, e.name), e.Path, e.version.Mode),
		)
	}

	slices.Sort(result)

	return strings.Join(result, ", ")
}

func inferIptablesMode(executables ...InitializedExecutable) (IptablesMode, error) {
	modesSet := make(map[IptablesMode]struct{}, len(executables))

	for _, executable := range executables {
		modesSet[executable.version.Mode] = struct{}{}
	}

	modes := maps.Keys(modesSet)

	if len(modes) != 1 {
		return IptablesModeUnknown, errors.Errorf(
			"executables are of mixed types; all must be of the same type ('%s' or '%s') [%s]",
			IptablesModeNft,
			IptablesModeLegacy,
			getExecutablesModesString(executables...),
		)
	}

	return modes[0], nil
}

func tryInitializeExecutablePaths(
	ctx context.Context,
	l Logger,
	cfg Config,
	ep executablesPaths,
) (InitializedExecutablesIPvX, bool, error) {
	if reflect.ValueOf(ep).IsZero() {
		return InitializedExecutablesIPvX{}, false, nil
	}

	if paths := getNonEmptyPaths(ep); len(paths) != 0 && len(paths) != 3 {
		return InitializedExecutablesIPvX{}, false, errors.Errorf(
			"provided incomplete executables configuration: %s must all be specified together; provided paths (%s) will be ignored and automatic executables detection will proceed",
			getNamesString(ep),
			getNamesWithPathsString(ep),
		)
	}

	initialized, err := ep.convert().Initialize(ctx, l, cfg)
	if err != nil {
		return InitializedExecutablesIPvX{}, false, errors.Wrap(
			err,
			"failed to initialize executables from the provided paths; automatic detection will proceed",
		)
	}

	l.Infof("provided executables will be used (%s)", getNamesWithPathsString(ep))

	return initialized, true, nil
}

func getNonEmptyPaths(ep executablesPaths) []string {
	return removeEmptyStrings(maps.Values(ep.getPathsMap()))
}

func removeEmptyStrings(strngs []string) []string {
	return slices.DeleteFunc(
		strngs,
		func(s string) bool {
			return strings.TrimSpace(s) == ""
		},
	)
}

func getNamesWithPathsString(eps ...executablesPaths) string {
	var result []string

	pathsMap := getPathsMap(eps...)
	for _, name := range util_maps.SortedKeys(pathsMap) {
		result = append(result, fmt.Sprintf("%s: '%s'", name, pathsMap[name]))
	}

	return strings.Join(result, ", ")
}

func getNamesString(eps ...executablesPaths) string {
	var result []string

	pathsMap := getPathsMap(eps...)
	for i, name := range util_maps.SortedKeys(pathsMap) {
		if i == len(pathsMap)-1 {
			result = append(result, fmt.Sprintf("and '%s'", name))
		} else {
			result = append(result, fmt.Sprintf("'%s'", name))
		}
	}

	return strings.Join(result, ", ")
}

func getPathsMap(eps ...executablesPaths) map[string]string {
	result := map[string]string{}

	for _, ep := range eps {
		for name, path := range ep.getPathsMap() {
			result[name] = path
		}
	}

	return result
}

func formatIptablesVersionErrorf(format string, a ...any) error {
	msg := fmt.Sprintf(format, a...)
	msgWithoutNewLines := strings.ReplaceAll(msg, "\n", " ")
	msgWithoutDuplicatedSpaces := strings.ReplaceAll(msgWithoutNewLines, "  ", " ")
	msgWithoutDotSuffix := strings.TrimRight(msgWithoutDuplicatedSpaces, ".")

	if len(msgWithoutDotSuffix) > 500 {
		msgWithoutDotSuffix = fmt.Sprintf("%.500s...", msgWithoutDotSuffix)
	}

	return errors.New(msgWithoutDotSuffix)
}
