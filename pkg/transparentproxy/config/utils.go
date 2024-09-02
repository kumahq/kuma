package config

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/vishvananda/netlink"

	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/consts"
)

func parsePort(s string) (Port, error) {
	u, err := parseUint16(s)

	if err != nil || u == 0 {
		return 0, errors.Errorf("value '%s' is not a valid port (uint16 in the range [1, 65535])", s)
	}

	return Port(u), nil
}

// getLoopbackInterfaceName retrieves the name of the loopback interface on the
// system. This function iterates over all network interfaces and checks if the
// 'net.FlagLoopback' flag is set. If a loopback interface is found, its name is
// returned. Otherwise, an error message indicating that no loopback interface
// was found is returned.
func getLoopbackInterfaceName() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", errors.Wrap(err, "failed to retrieve network interfaces")
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagLoopback != 0 {
			return iface.Name, nil
		}
	}

	return "", errors.New("no loopback interface found on the system")
}

// parseExcludePortsForUIDs parses a slice of strings representing port
// exclusion rules based on UIDs and returns a slice of Exclusion structs.
//
// Each input string should follow the format: <protocol:>?<ports:>?<uids>.
// This means the string can contain optional protocol and port values,
// followed by mandatory UID values. Examples of valid formats include:
//   - "tcp:22:1000-2000" (TCP protocol, port 22, UIDs from 1000 to 2000)
//   - "udp:53:1001" (UDP protocol, port 53, UID 1001)
//   - "80:1002" (Any protocol, port 80, UID 1002)
//   - "1003" (Any protocol, any port, UID 1003)
func parseExcludePortsForUIDs(exclusionRules []string) ([]Exclusion, error) {
	var result []Exclusion

	for _, elem := range exclusionRules {
		parts := strings.Split(elem, ":")
		if len(parts) == 0 || len(parts) > 3 {
			return nil, errors.Errorf(
				"invalid format for excluding ports by UIDs: '%s'. Expected format: <protocol:>?<ports:>?<uids>",
				elem,
			)
		}

		var portValuesOrRange, protocolOpts, uidValuesOrRange string

		switch len(parts) {
		case 1:
			protocolOpts = "*"
			portValuesOrRange = "*"
			uidValuesOrRange = parts[0]
		case 2:
			protocolOpts = "*"
			portValuesOrRange = parts[0]
			uidValuesOrRange = parts[1]
		case 3:
			protocolOpts = parts[0]
			portValuesOrRange = parts[1]
			uidValuesOrRange = parts[2]
		}

		if uidValuesOrRange == "*" {
			return nil, errors.New("wildcard '*' is not allowed for UIDs")
		}

		if portValuesOrRange == "*" || portValuesOrRange == "" {
			portValuesOrRange = "1-65535"
		}

		if err := validateUintValueOrRange(portValuesOrRange); err != nil {
			return nil, errors.Wrap(err, "invalid port range")
		}

		if strings.Contains(uidValuesOrRange, ",") {
			return nil, errors.Errorf(
				"invalid UID entry: '%s'. It should either be a single item or a range",
				uidValuesOrRange,
			)
		}

		if err := validateUintValueOrRange(uidValuesOrRange); err != nil {
			return nil, errors.Wrap(err, "invalid UID range")
		}

		var protocols []consts.ProtocolL4
		if protocolOpts == "" || protocolOpts == "*" {
			protocols = []consts.ProtocolL4{consts.ProtocolTCP, consts.ProtocolUDP}
		} else {
			for _, s := range strings.Split(protocolOpts, ",") {
				if p := consts.ParseProtocolL4(s); p != consts.ProtocolUndefined {
					protocols = append(protocols, p)
					continue
				}

				return nil, errors.Errorf(
					"invalid or unsupported protocol: '%s'",
					s,
				)
			}
		}

		for _, p := range protocols {
			ports := strings.ReplaceAll(portValuesOrRange, "-", ":")
			uids := strings.ReplaceAll(uidValuesOrRange, "-", ":")

			result = append(result, Exclusion{
				Ports:    ValueOrRangeList(ports),
				UIDs:     ValueOrRangeList(uids),
				Protocol: p,
			})
		}
	}

	return result, nil
}

// parseExcludePortsForIPs parses a slice of strings representing port exclusion
// rules based on IP addresses and returns a slice of IPToPorts structs.
//
// This function currently allows each exclusion rule to be a valid IPv4 or IPv6
// address, with or without a CIDR suffix. It is designed to potentially support
// more complex exclusion rules in the future.
func parseExcludePortsForIPs(
	exclusionRules []string,
	ipv6 bool,
) ([]Exclusion, error) {
	var result []Exclusion

	for _, rule := range exclusionRules {
		if rule == "" {
			return nil, errors.New(
				"invalid exclusion rule: the rule cannot be empty",
			)
		}

		for _, address := range strings.Split(rule, ",") {
			err, isExpectedIPVersion := validateIP(address, ipv6)
			if err != nil {
				return nil, errors.Wrap(err, "invalid exclusion rule")
			}

			if isExpectedIPVersion {
				result = append(result, Exclusion{Address: address})
			}
		}
	}

	return result, nil
}

// validateUintValueOrRange validates whether a given string represents a valid
// single uint16 value or a range of uint16 values. The input string can contain
// multiple comma-separated values or ranges (e.g., "80,1000-2000").
func validateUintValueOrRange(valueOrRange string) error {
	for _, element := range strings.Split(valueOrRange, ",") {
		for _, port := range strings.Split(element, "-") {
			if _, err := parseUint16(port); err != nil {
				return errors.Wrapf(
					err,
					"validation failed for value or range '%s'",
					valueOrRange,
				)
			}
		}
	}

	return nil
}

// validateIP validates an IP address or CIDR and checks if it matches the
// expected IP version (IPv4 or IPv6).
func validateIP(address string, ipv6 bool) (error, bool) {
	// Attempt to parse the address as a CIDR.
	ip, _, err := net.ParseCIDR(address)
	// If parsing as CIDR fails, attempt to parse it as a plain IP address.
	if err != nil {
		ip = net.ParseIP(address)
	}

	// If parsing as both CIDR and IP address fails, return an error with a
	// message.
	if ip == nil {
		return errors.Errorf(
			"invalid IP address: '%s'. Expected format: <ip> or <ip>/<cidr> (e.g., 10.0.0.1, 172.16.0.0/16, fe80::1, fe80::/10)",
			address,
		), false
	}

	// Check if the IP version matches the expected IP version.
	// For IPv4, ip.To4() will not be nil. For IPv6, ip.To4() will be nil.
	return nil, ipv6 == (ip.To4() == nil)
}

// parseUint16 parses a string representing a uint16 value and returns its
// uint16 representation.
func parseUint16(port string) (uint16, error) {
	parsedPort, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		return 0, errors.Errorf("invalid uint16 value: '%s'", port)
	}

	return uint16(parsedPort), nil
}

// hasLocalIPv6 checks if the local system has an active non-loopback IPv6
// address. It scans through all network interfaces to find any IPv6 address
// that is not a loopback address.
func hasLocalIPv6() (bool, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return false, err
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok &&
			!ipnet.IP.IsLoopback() &&
			ipnet.IP.To4() == nil {
			return true, nil
		}
	}

	return false, errors.New("no local IPv6 addresses detected")
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

		return errors.Wrapf(err, "failed to add IPv6 address %s to loopback interface", addr.IPNet)
	}

	return nil
}

func findUserUID(userOrUID string) (string, bool) {
	if userOrUID == "" {
		return "", false
	}

	if u, err := user.LookupId(userOrUID); err == nil {
		return u.Uid, true
	}

	if u, err := user.Lookup(userOrUID); err == nil {
		return u.Uid, true
	}

	return "", false
}

// executables

// buildRestoreArgs constructs a slice of flags for restoring iptables rules
// based on the provided configuration and iptables mode.
//
// This function generates a list of command-line flags to be used with
// iptables-restore, tailored to the given parameters:
//   - For non-legacy iptables mode, it returns an empty slice, as no additional
//     flags are required.
//   - For legacy mode, it conditionally adds the `--wait` and `--wait-interval`
//     flags based on the provided configuration values.
func buildRestoreArgs(cfg Config, mode consts.IptablesMode) []string {
	var flags []string

	if mode != consts.IptablesModeLegacy {
		return flags
	}

	if cfg.Wait > 0 {
		flags = append(flags, fmt.Sprintf("%s=%d", consts.FlagWait, cfg.Wait))
	}

	if cfg.WaitInterval > 0 {
		flags = append(flags, fmt.Sprintf(
			"%s=%d",
			consts.FlagWaitInterval,
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
	for _, fallbackPath := range consts.FallbackExecutablesSearchLocations {
		paths = append(paths, filepath.Join(fallbackPath, nameWithMode))
	}

	paths = append(paths, nameWithoutMode)
	for _, fallbackPath := range consts.FallbackExecutablesSearchLocations {
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
