package install

import (
	"context"
	"fmt"
	"io"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/exec"

	"github.com/kumahq/kuma/v2/pkg/test/matchers"
	"github.com/kumahq/kuma/v2/pkg/util/pointer"
	test_container "github.com/kumahq/kuma/v2/test/framework/container"
	"github.com/kumahq/kuma/v2/test/framework/utils"
	. "github.com/kumahq/kuma/v2/test/transparentproxy"
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

type testCase struct {
	name             string
	image            string
	postStart        [][]string
	goldenFileSuffix string
	additionalFlags  []string
}

func Install() {
	DescribeTable(
		"kumactl install transparent-proxy inside Docker container",
		func(tc testCase) {
			Expect(Config.KumactlLinuxBin).NotTo(BeEmpty())

			container, err := test_container.NewContainerSetup().
				WithImage(tc.image).
				WithKumactlBinary(Config.KumactlLinuxBin).
				WithPostStart(tc.postStart).
				WithPrivileged(true).
				Start(context.Background())

			Expect(err).ToNot(HaveOccurred())

			DeferCleanup(func() {
				Expect(container.Terminate(context.Background())).To(Succeed())
			})

			EnsureInstallSuccessful(container, tc.additionalFlags)
			EnsureGoldenFiles(container, tc)
		},
		EntriesForImages(Config.DockerImagesToTest),
	)
}

func EnsureInstallSuccessful(container testcontainers.Container, flags []string) {
	GinkgoHelper()

	exitCode, reader, err := container.Exec(
		context.Background(),
		append(
			[]string{
				"kumactl",
				"install",
				"transparent-proxy",
				"--kuma-dp-user",
				"kuma-dp",
			},
			flags...,
		),
		exec.Multiplexed(),
	)
	Expect(err).NotTo(HaveOccurred())

	if exitCode != 0 {
		buf := new(strings.Builder)
		_, _ = io.Copy(buf, reader)

		Fail(fmt.Sprintf("installation ended with code %d: %s", exitCode, buf))
	}
}

func EnsureGoldenFiles(container testcontainers.Container, tc testCase) {
	GinkgoHelper()

	saveCmds := []string{
		"iptables-save",
		"iptables-legacy-save",
		"iptables-nft-save",
	}

	if Config.IPV6 {
		saveCmds = append(
			saveCmds,
			"ip6tables-save",
			"ip6tables-legacy-save",
			"ip6tables-nft-save",
		)
	}

	for _, cmd := range saveCmds {
		golden := utils.BuildIptablesGoldenFileName(
			"install",
			tc.name,
			cmd,
			tc.goldenFileSuffix,
		)

		exitCode, reader, err := container.Exec(
			context.Background(),
			[]string{cmd},
			exec.Multiplexed(),
		)
		Expect(err).NotTo(HaveOccurred())

		buf := new(strings.Builder)
		Expect(io.Copy(buf, reader)).Error().NotTo(HaveOccurred())

		if exitCode != 0 {
			if !strings.Contains(buf.String(), "executable file not found") {
				Fail(fmt.Sprintf(
					"installation ended with code %d: %s",
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
//
// Args:
//
//	images (map[string]string): A map where the keys are descriptive names of
//	  the Docker images, and the values are the actual Docker image names to use
//	  for testing.
//	entry (func(description interface{}, args ...interface{}) TableEntry):
//	  A function used to create Ginkgo TableEntry objects. Use PEntry or XEntry
//	  for creating paused or excluded entries, or Entry for regular ones.
//	decorators (...interface{}): Optional decorators to apply to each TableEntry
//	  for additional customization or configuration.
//
// Returns:
//
//	[]TableEntry: A slice of Ginkgo test entries, each representing a unique
//	  test case for a combination of Docker image and optional flags.
//
// This function performs the following steps:
//  1. Iterates through the provided map of Docker images.
//  2. For each image, it calls the genEntriesForImage function to generate
//     individual test entries. The genEntriesForImage function creates test
//     entries based on the base image and all possible combinations of
//     additional flags defined in the configuration.
//  3. Combines all generated test entries into a single slice and returns it.
func genEntriesForImages(
	images map[string]string,
	entry func(description interface{}, args ...interface{}) TableEntry,
	decorators ...interface{},
) []TableEntry {
	var entries []TableEntry

	for name, image := range images {
		entries = append(
			entries,
			genEntriesForImage(
				name,
				image,
				Config.InstallFlagsToTest,
				entry,
				decorators...,
			)...,
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
	case strings.Contains(image, "redhat/ubi"), strings.Contains(image, "centos"):
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

// EntriesForImages generates Ginkgo test entries for various Transparent Proxy
// installation scenarios on a given set of Docker images.
//
// Note:
//   - Container lifecycle hooks can sometimes fail silently, especially during
//     critical operations like installing iptables. Handling these failures is
//     challenging and can complicate the tests. To mitigate this, the function
//     includes FlakeAttempts(3) to retry the test up to three times. This
//     approach ensures robustness without significantly increasing test
//     complexity. It is extremaly unlikely that tproxy installation would be
//     flaky, so this method should provide a reliable solution.
func EntriesForImages(images map[string]string) []TableEntry {
	return genEntriesForImages(images, Entry, FlakeAttempts(3))
}
