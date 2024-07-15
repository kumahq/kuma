package transparentproxy

import (
	"context"
	std_errors "errors"
	"fmt"
	"io"
	"slices"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/exec"

	"github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/util/pointer"
	test_container "github.com/kumahq/kuma/test/framework/container"
	"github.com/kumahq/kuma/test/framework/utils"
)

var (
	// Generic
	userAdd = []string{"useradd", "-u", "5678", "kuma-dp"}
	// Debian, Ubuntu
	aptUpdate          = []string{"apt-get", "update", "-y"}
	aptInstallIptables = []string{"apt-get", "install", "-y", "iptables"}
	// RHEL, Amazon Linux, Centos
	yumUpdate          = []string{"yum", "update", "-y"}
	yumInstallIptables = []string{"yum", "install", "-y", "iptables"}
	// On Amazon Linux, the `useradd` command is not available by default.
	// The `shadow-utils` package, which provides user management tools,
	// includes `useradd`.
	yumInstallShadowUtils = []string{"yum", "install", "-y", "shadow-utils"}
	// Alpine
	addUser        = []string{"adduser", "-u", "5678", "kuma-dp", "-D"}
	apkAddIptables = []string{"apk", "add", "iptables"}
)

// customIptablesRules defines a set of custom iptables rules that are used to
// ensure our cleanup process does not remove non-transparent-proxy related
// rules. These rules create custom chains and add rules to various tables
// (nat, raw, mangle) to confirm that the cleanup process will only remove
// rules related to the transparent proxy, leaving other custom rules intact.
//
// The rules include:
//   - Creating custom chains in the nat, raw, and mangle tables.
//   - Adding rules to the OUTPUT and PREROUTING chains that direct traffic to
//     the custom chains.
var (
	customIptablesRules = [][]string{
		{"-t", "nat", "-N", "CUSTOM_CHAIN_NAT"},
		{"-t", "raw", "-N", "CUSTOM_CHAIN_RAW"},
		{"-t", "mangle", "-N", "CUSTOM_CHAIN_MANGLE"},
		{"-t", "nat", "-A", "OUTPUT", "-p", "tcp", "-j", "CUSTOM_CHAIN_NAT"},
		{"-t", "nat", "-A", "PREROUTING", "-p", "udp", "-j", "CUSTOM_CHAIN_NAT"},
		{"-t", "raw", "-A", "OUTPUT", "-p", "tcp", "--dport", "53", "-j", "CUSTOM_CHAIN_RAW"},
		{"-t", "raw", "-A", "PREROUTING", "-j", "CUSTOM_CHAIN_RAW"},
		{"-t", "mangle", "-A", "OUTPUT", "-p", "udp", "--sport", "53", "-j", "CUSTOM_CHAIN_MANGLE"},
		{"-t", "mangle", "-A", "PREROUTING", "-j", "CUSTOM_CHAIN_MANGLE"},
	}
)

// The following variables are used in tests to manage and verify iptables rules
// across different iptables implementations and versions. These lists include
// commands for installing custom rules (iptablesCmds) and for saving the rules
// to check against expected outcomes (ipv4SaveCmds, ipv6SaveCmds,
// iptablesSaveCmds).
var (
	iptablesCmds = []string{
		"iptables",
		"iptables-nft",
		"iptables-legacy",
		"ip6tables",
		"ip6tables-nft",
		"ip6tables-legacy",
	}
	ipv4SaveCmds = []string{
		"iptables-save",
		"iptables-nft-save",
		"iptables-legacy-save",
	}
	ipv6SaveCmds = []string{
		"ip6tables-save",
		"ip6tables-nft-save",
		"ip6tables-legacy-save",
	}
	iptablesSaveCmds = slices.Concat(slices.Clone(ipv4SaveCmds), ipv6SaveCmds)
)

type testCase struct {
	name             string
	image            string
	postStart        [][]string
	goldenFileSuffix string
	additionalFlags  []string
}

var _ = Describe("Transparent Proxy", func() {
	DescribeTable(
		"install in container",
		func(tc testCase) {
			ctx := context.Background()

			// Given the kumactl binary path is not empty
			Expect(Config.KumactlLinuxBin).NotTo(BeEmpty())

			// Given a container setup with specified image and settings
			c, err := test_container.NewContainerSetup().
				WithImage(tc.image).
				WithKumactlBinary(Config.KumactlLinuxBin).
				WithPostStart(tc.postStart).
				WithPrivileged(true).
				Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Clean up the container after the test
			DeferCleanup(func() {
				Expect(c.Terminate(ctx)).To(Succeed())
			})

			// When the transparent proxy is installed successfully
			EnsureInstallSuccessful(ctx, c, tc.additionalFlags)

			// Then the golden files should match the expected output
			EnsureGoldenFiles(ctx, c, tc)
		},
		// Generate entries for each Docker image to test
		genEntriesForImages(Config.DockerImagesToTest, Entry, FlakeAttempts(3)),
	)

	DescribeTable(
		"uninstall in container",
		func(tc testCase) {
			ctx := context.Background()

			// Given the kumactl binary path is not empty
			Expect(Config.KumactlLinuxBin).NotTo(BeEmpty())

			// Given a container setup with specified image and settings
			c, err := test_container.NewContainerSetup().
				WithImage(tc.image).
				WithKumactlBinary(Config.KumactlLinuxBin).
				WithPostStart(tc.postStart).
				WithPrivileged(true).
				Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Clean up the container after the test
			DeferCleanup(func() {
				Expect(c.Terminate(ctx)).To(Succeed())
			})

			// Given custom iptables rules are added to the container
			Expect(addCustomIptablesRules(ctx, c)).To(Succeed())

			// Given the iptables-save output before installing the proxy
			before := getIptablesSaveOutput(ctx, c)

			// When the transparent proxy is installed successfully
			EnsureInstallSuccessful(ctx, c, tc.additionalFlags)

			// When the transparent proxy is uninstalled successfully
			EnsureUninstallSuccessful(ctx, c)

			// Then the iptables-save output after uninstall should match the
			// output before install
			after := getIptablesSaveOutput(ctx, c)

			Expect(before).To(HaveLen(len(after)))

			for cmd, output := range before {
				Expect(after[cmd]).To(Equal(output))
			}
		},
		// Generate entries for each Docker image to test
		genEntriesForImages(Config.DockerImagesToTest, Entry, FlakeAttempts(3)),
	)
})

// EnsureInstallSuccessful installs the transparent proxy in a test container
// using the provided flags, and ensures the installation completes
// successfully.
//
// This function performs the following steps:
//   - Constructs the kumactl install command with the given flags.
//   - Executes the command in the specified container context.
//   - Asserts that no error occurred during execution.
//   - If the command does not exit successfully (exit code != 0), it reads the
//     command output and fails the test with an appropriate message.
//
// Args:
//   - ctx (context.Context): The context for controlling the command execution.
//   - c (testcontainers.Container): The test container where the command will
//     be executed.
//   - flags ([]string): Additional flags to pass to the kumactl install
//     command.
func EnsureInstallSuccessful(
	ctx context.Context,
	c testcontainers.Container,
	flags []string,
) {
	GinkgoHelper()

	cmd := slices.Concat(
		[]string{
			"kumactl",
			"install",
			"transparent-proxy",
			"--kuma-dp-user",
			"kuma-dp",
		},
		flags,
	)

	exitCode, reader, err := c.Exec(ctx, cmd, exec.Multiplexed())
	Expect(err).NotTo(HaveOccurred())

	if exitCode != 0 {
		buf := new(strings.Builder)
		Expect(io.Copy(buf, reader)).Error().NotTo(HaveOccurred())
		Fail(fmt.Sprintf("installation ended with code %d: %s", exitCode, buf))
	}
}

// EnsureUninstallSuccessful runs the `kumactl uninstall transparent-proxy`
// command in the specified container and checks if the uninstallation is
// successful.
func EnsureUninstallSuccessful(ctx context.Context, c testcontainers.Container) {
	GinkgoHelper()

	exitCode, reader, err := c.Exec(
		ctx,
		[]string{"kumactl", "uninstall", "transparent-proxy"},
		exec.Multiplexed(),
	)
	Expect(err).NotTo(HaveOccurred())

	if exitCode != 0 {
		buf := new(strings.Builder)
		Expect(io.Copy(buf, reader)).Error().NotTo(HaveOccurred())
		Fail(fmt.Sprintf("uninstall ended with code %d: %s", exitCode, buf))
	}
}

// EnsureGoldenFiles validates the current iptables rules in a test container
// against predefined golden files, ensuring the rules match the expected
// configuration.
//
// This function performs the following steps:
//   - Clones the list of IPv4 save commands and optionally includes IPv6 save
//     commands if configured.
//   - Iterates through each save command, executing it in the specified
//     container context.
//   - Constructs the golden file name based on the test case name, command, and
//     suffix.
//   - Reads the command output and compares it against the corresponding golden
//     file.
//   - If the command exits unsuccessfully, it checks if the error is due to the
//     executable not being found and handles it accordingly.
func EnsureGoldenFiles(
	ctx context.Context,
	c testcontainers.Container,
	tc testCase,
) {
	GinkgoHelper()

	saveCmds := slices.Clone(ipv4SaveCmds)

	if Config.IPV6 {
		saveCmds = slices.Concat(saveCmds, ipv6SaveCmds)
	}

	for _, cmd := range saveCmds {
		golden := utils.BuildIptablesGoldenFileName(
			tc.name,
			cmd,
			tc.goldenFileSuffix,
		)

		exitCode, reader, err := c.Exec(ctx, []string{cmd}, exec.Multiplexed())
		Expect(err).NotTo(HaveOccurred())

		buf := new(strings.Builder)
		Expect(io.Copy(buf, reader)).Error().NotTo(HaveOccurred())

		if exitCode != 0 {
			if !strings.Contains(buf.String(), "executable file not found") {
				Fail(fmt.Sprintf(
					"command ended with code %d: %s",
					exitCode,
					buf.String(),
				))
			}

			Expect("executable not found\n").
				To(matchers.MatchGoldenEqual(golden...))

			continue
		}

		Expect(utils.CleanIptablesSaveOutput(buf.String())).
			To(matchers.MatchGoldenEqual(golden...))
	}
}

// genEntriesForImages generates Ginkgo test entries for Transparent Proxy
// (tproxy) installation scenarios across a set of Docker images.
func genEntriesForImages(
	images map[string]string,
	entry func(description interface{}, args ...interface{}) TableEntry,
	decorators ...interface{},
) []TableEntry {
	var entries []TableEntry

	for name, image := range images {
		entries = slices.Concat(
			entries,
			genEntriesForImage(
				name,
				image,
				Config.InstallFlagsToTest,
				entry,
				decorators...,
			),
		)
	}

	return entries
}

// genEntriesForImage generates Ginkgo test entries for various scenarios
// involving Transparent Proxy (tproxy) installation within a Docker image.
//
// Args:
//   - name (string): The shortened name of the Docker image for easier
//     reference.
//   - image (string): The base Docker image name to use for testing.
//   - additionalFlagsToTest (*FlagsMap): A map of optional flags to include
//     during tproxy installation. Each key is a string suffix for the golden
//     file, and each value is a slice of flags for a single test case (e.g.,
//     {"--redirect-all-dns-traffic"}).
//   - entry (func(description interface{}, args ...interface{}) TableEntry):
//     A function used to create Ginkgo TableEntry objects. Use PEntry or XEntry
//     for creating paused or excluded entries.
//   - decorators (...interface{}): Optional decorators to apply to each
//     TableEntry for additional customization or configuration.
//
// Returns:
//   - []TableEntry: A slice of Ginkgo TableEntry objects, each representing a
//     unique test case with the following configuration:
//   - Image name: The Docker image name used for the test (may include
//     additional flags).
//   - testCase: A struct containing detailed test case parameters:
//   - name: The shortened name of the Docker image.
//   - image: The base Docker image name.
//   - postStart: Commands to execute after starting the container (e.g.,
//     adding a user, updating package lists, installing iptables).
//   - additionalFlags (optional): Additional flags to pass during tproxy
//     installation.
//
// This function generates entries for the following test variations:
//   - Base installation on different Docker images (Ubuntu, Debian, Alpine,
//     RHEL variants, Amazon Linux).
//   - Installation with additional flags specified in `additionalFlagsToTest`.
//
// Note:
//   - The function tailors commands (e.g., package managers, iptables binaries)
//     based on the image being used to ensure compatibility with the package
//     management system and other environment-specific requirements.
func genEntriesForImage(
	name string,
	image string,
	additionalFlagsToTest *FlagsMap,
	entry func(description interface{}, args ...interface{}) TableEntry,
	decorators ...interface{},
) []TableEntry {
	var postStart [][]string

	image = strings.ToLower(image)

	switch {
	case strings.Contains(image, "debian"), strings.Contains(image, "ubuntu"):
		postStart = [][]string{aptUpdate, aptInstallIptables, userAdd}
	case strings.Contains(image, "alpine"):
		postStart = [][]string{apkAddIptables, addUser}
	case strings.Contains(image, "redhat/ubi"),
		strings.Contains(image, "centos"),
		strings.Contains(image, "fedora"):
		postStart = [][]string{yumUpdate, yumInstallIptables, userAdd}
	case strings.Contains(image, "amazonlinux"):
		postStart = [][]string{
			yumUpdate,
			yumInstallIptables,
			yumInstallShadowUtils,
			userAdd,
		}
	}

	// buildArgs is a helper function that combines test case parameters with
	// any additional decorators.
	buildArgs := func(
		goldenFileSuffix string,
		flags []string,
		decorators []interface{},
	) []interface{} {
		var args []interface{}

		args = append(args, testCase{
			name:             name,
			image:            image,
			postStart:        postStart,
			goldenFileSuffix: goldenFileSuffix,
			additionalFlags:  flags,
		})
		args = append(args, decorators...)

		return args
	}

	entries := []TableEntry{
		entry(
			fmt.Sprintf("%s (%s)", name, image),
			buildArgs("", nil, decorators)...,
		),
	}

	for goldenFileSuffix, flags := range pointer.Deref(additionalFlagsToTest) {
		func(goldenFileSuffix string, flags []string) {
			entries = append(entries, entry(
				fmt.Sprintf(
					"%s (%s) with flags: %s",
					name,
					image,
					strings.Join(flags, " "),
				),
				buildArgs(goldenFileSuffix, flags, decorators)...,
			))
		}(goldenFileSuffix, flags)
	}

	return entries
}

// getIptablesSaveOutput retrieves the output of iptables-save commands from a
// given container.
//
// This function executes all the defined iptables-save commands inside the
// provided container and collects their outputs. The outputs are cleaned and
// returned in a map where the keys are the command names and the values are
// their respective outputs.
func getIptablesSaveOutput(
	ctx context.Context,
	container testcontainers.Container,
) map[string]string {
	GinkgoHelper()

	output := map[string]string{}

	for _, cmd := range iptablesSaveCmds {
		if exitCode, reader, err := container.Exec(
			ctx,
			[]string{cmd},
			exec.Multiplexed(),
		); exitCode == 0 && err == nil {
			buf := new(strings.Builder)
			Expect(io.Copy(buf, reader)).Error().NotTo(HaveOccurred())
			output[cmd] = utils.CleanIptablesSaveOutput(buf.String())
		}
	}

	return output
}

// addCustomIptablesRules adds a set of custom iptables rules to a given
// container.
//
// This function iterates over the predefined iptables commands and custom
// iptables rules, attempting to add each rule using each iptables command.
// If all attempts to add the rules using all iptables commands fail, it
// returns an error detailing the failures.
func addCustomIptablesRules(
	ctx context.Context,
	c testcontainers.Container,
) error {
	var errs []error

	for _, iptables := range iptablesCmds {
		var cmdErrs []error

		for _, rule := range customIptablesRules {
			if exitCode, _, err := c.Exec(
				ctx,
				slices.Concat([]string{iptables}, rule),
				exec.Multiplexed(),
			); err != nil || exitCode != 0 {
				cmdErrs = append(
					cmdErrs,
					errors.Wrapf(err, "exit code %d", exitCode),
				)
			}
		}

		if len(cmdErrs) > 0 {
			errs = append(errs, std_errors.Join(cmdErrs...))
		}
	}

	if len(errs) == len(iptablesCmds) {
		return errors.Wrap(
			std_errors.Join(errs...),
			"all iptables commands used to add custom iptables rules failed",
		)
	}

	return nil
}
