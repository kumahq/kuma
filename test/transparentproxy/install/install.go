package install

import (
	"bytes"
	"context"
	std_errors "errors"
	"fmt"
	"io"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/exec"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/kumahq/kuma/pkg/test/matchers"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/transparentproxy/utils"
)

var (
	// Generic
	cmdUserAdd = []string{"useradd", "-u", "5678", "kuma-dp"}
	// Debian, Ubuntu
	cmdAptUpdate          = []string{"apt-get", "update", "-y"}
	cmdAptInstallIptables = []string{"apt-get", "install", "-y", "iptables"}
	// RHEL, Amazon Linux
	cmdYumUpdate          = []string{"yum", "update", "-y"}
	cmdYumInstallIptables = []string{"yum", "install", "-y", "iptables"}
	// On Amazon Linux, the `useradd` command is not available by default.
	// The `shadow-utils` package, which provides user management tools,
	// includes `useradd`.
	cmdYumInstallShadowUtils = []string{"yum", "install", "-y", "shadow-utils"}
	// Alpine
	cmdAddUser        = []string{"adduser", "-u", "5678", "kuma-dp", "-D"}
	cmdApkAddIptables = []string{"apk", "add", "iptables"}
)

type testCase struct {
	name            string
	image           string
	postStart       [][]string
	additionalFlags []string
}

func Install() {
	DescribeTable(
		"kumactl install transparent-proxy inside Docker container",
		func(tc testCase) {
			Expect(TProxyConfig.KumactlLinuxBin).NotTo(BeEmpty())

			container, err := testcontainers.GenericContainer(
				context.Background(),
				testcontainers.GenericContainerRequest{
					ContainerRequest: testcontainers.ContainerRequest{
						Image:      tc.image,
						Privileged: true,
						Files: []testcontainers.ContainerFile{{
							HostFilePath:      TProxyConfig.KumactlLinuxBin,
							ContainerFilePath: "/usr/local/bin/kumactl",
							FileMode:          0o700,
						}},
						Cmd: []string{"sleep", "infinity"},
						LifecycleHooks: []testcontainers.ContainerLifecycleHooks{
							{PostStarts: utils.BuildContainerHooks(tc.postStart)},
						},
						WaitingFor: wait.ForExec([]string{"kumactl", "version"}).
							WithStartupTimeout(time.Second * 10).
							WithExitCodeMatcher(func(exitCode int) bool {
								return exitCode == 0
							}).
							WithResponseMatcher(func(body io.Reader) bool {
								data, _ := io.ReadAll(body)
								return bytes.Contains(data, []byte("Client: "))
							}),
					},
					Started: true,
				},
			)
			Expect(err).ToNot(HaveOccurred())

			DeferCleanup(func() {
				Expect(container.Terminate(context.Background())).To(Succeed())
			})

			EnsureInstallSuccessful(container, tc.additionalFlags)
			EnsureGoldenFiles(container, tc)
		},
		EntriesForImages(TProxyConfig.DockerImagesToTest),
		XEntriesForImages(TProxyConfig.DockerImagesToTestPaused),
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

	var errs []error
	outputMap := map[string][]string{}

	saveCmds := map[string][]string{
		"ipv4": {
			"iptables-save",
			"iptables-legacy-save",
			"iptables-nft-save",
		},
	}

	if TProxyConfig.IPV6 {
		saveCmds["ipv6"] = []string{
			"ip6tables-save",
			"ip6tables-legacy-save",
			"ip6tables-nft-save",
		}
	}

	for suffix, cmds := range saveCmds {
		for _, cmd := range cmds {
			exitCode, reader, err := container.Exec(
				context.Background(),
				[]string{cmd},
				exec.Multiplexed(),
			)
			if err != nil {
				errs = append(errs, errors.Wrapf(err, "%s failed", cmd))
				continue
			}

			buf := new(strings.Builder)
			if _, err := io.Copy(buf, reader); err != nil {
				errs = append(errs, errors.Wrapf(
					err,
					"%s: copying output reader to buffer failed",
					cmd,
				))
				continue
			}

			outputMap[suffix] = append(
				outputMap[suffix],
				fmt.Sprintf("# %s", cmd),
			)

			if exitCode != 0 {
				if strings.Contains(buf.String(), "executable file not found") {
					outputMap[suffix] = append(
						outputMap[suffix],
						"# executable not found",
					)
				} else {
					errs = append(errs, errors.Errorf(
						"%s ended with code: %d: %s",
						cmd,
						exitCode,
						buf,
					))
					continue
				}
			} else {
				outputMap[suffix] = append(
					outputMap[suffix],
					utils.CleanIptablesSaveOutput(buf.String()),
				)
			}

			outputMap[suffix] = append(
				outputMap[suffix],
				fmt.Sprintf("# %s end\n", cmd),
			)
		}

		if len(errs) > 0 {
			Fail(fmt.Sprintf(
				"errors encountered during golden files verification: \n%s",
				std_errors.Join(errs...),
			))
		}
	}

	for suffix, output := range outputMap {
		Expect(strings.Join(output, "\n")).To(matchers.MatchGoldenEqual(
			utils.BuildGoldenFileName(
				"install",
				tc.name,
				tc.additionalFlags,
				suffix,
			)...,
		))
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
	var flags [][]string

	for _, flag := range TProxyConfig.InstallFlagsToTest {
		flags = append(flags, strings.Split(flag, " "))
	}

	for name, image := range images {
		entries = append(
			entries,
			genEntriesForImage(name, image, flags, entry, decorators...)...,
		)
	}

	return entries
}

// genEntriesForImage generates Ginkgo test entries for various scenarios
// involving Transparent Proxy (tproxy) installation within a Docker image.
//
// Args:
//
//	name (string): The shortened name of the Docker image for easier reference.
//	image (string): The base Docker image name to use for testing.
//	additionalFlagsToTest ([][]string): A 2D slice of strings representing
//	  optional flags to include during tproxy installation. Each inner slice
//	  represents a set of flags for a single test case (e.g.,
//	  ["--redirect-all-dns-traffic"]).
//	entry (func(description interface{}, args ...interface{}) TableEntry):
//	  A function used to create Ginkgo TableEntry objects. Use PEntry or XEntry
//	  for creating paused or excluded entries.
//	decorators (...interface{}): Optional decorators to apply to each TableEntry
//	  for additional customization or configuration.
//
// Returns:
//
//	[]TableEntry: A slice of Ginkgo TableEntry objects, each representing a
//	  unique test case with the following configuration:
//	  - Image name: The Docker image name used for the test (may include
//	    additional flags).
//	  - testCase: A struct containing detailed test case parameters:
//	    - name: The shortened name of the Docker image.
//	    - image: The base Docker image name.
//	    - postStart: Commands to execute after starting the container (e.g.,
//	      adding a user, updating package lists, installing iptables).
//	    - additionalFlags (optional): Additional flags to pass during tproxy
//	      installation.
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
	additionalFlagsToTest [][]string,
	entry func(description interface{}, args ...interface{}) TableEntry,
	decorators ...interface{},
) []TableEntry {
	var postStart [][]string

	switch {
	case strings.Contains(image, "debian"), strings.Contains(image, "ubuntu"):
		postStart = [][]string{cmdAptUpdate, cmdAptInstallIptables, cmdUserAdd}
	case strings.Contains(image, "alpine"):
		postStart = [][]string{cmdApkAddIptables, cmdAddUser}
	case strings.Contains(image, "redhat/ubi"):
		postStart = [][]string{cmdYumUpdate, cmdYumInstallIptables, cmdUserAdd}
	case strings.Contains(image, "amazonlinux"):
		postStart = [][]string{
			cmdYumUpdate,
			cmdYumInstallIptables,
			cmdYumInstallShadowUtils,
			cmdUserAdd,
		}
	}

	// buildArgs is a helper function that combines test case parameters with
	// any additional decorators.
	buildArgs := func(
		additionalFlags []string,
		decorators []interface{},
	) []interface{} {
		var args []interface{}

		args = append(args, testCase{
			name:            name,
			image:           image,
			postStart:       postStart,
			additionalFlags: additionalFlags,
		})
		args = append(args, decorators...)

		return args
	}

	entries := []TableEntry{
		entry(
			fmt.Sprintf("%s (%s)", name, image),
			buildArgs(nil, decorators)...,
		),
	}

	for _, flags := range additionalFlagsToTest {
		func(flags []string) {
			entries = append(entries, entry(
				fmt.Sprintf(
					"%s (%s) with flags: %s",
					name,
					image,
					strings.Join(flags, " "),
				),
				buildArgs(flags, decorators)...,
			))
		}(flags)
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

// XEntriesForImages generates paused ginkgo test entries (XEntry|PEntry) for
// various Transparent Proxy installation scenarios on a given Docker image.
func XEntriesForImages(images map[string]string) []TableEntry {
	return genEntriesForImages(images, XEntry)
}
