package framework

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	std_runtime "runtime"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
)

// savePlatform is the platform to pin an image archive to, empty when it cannot be
// pinned, plus the reason when it cannot. Pinning is harmless on the classic image
// store, which holds one variant per image regardless, so the store itself does not
// need asking about - only whether the CLI can pin, and against which architecture.
var savePlatform = sync.OnceValues(func() (string, string) {
	help, err := exec.Command("docker", "image", "save", "--help").Output()
	if err != nil || !strings.Contains(string(help), "--platform") {
		return "", "cannot pin the archive: `docker image save --platform` did not answer. On a containerd image store the archive then names every platform of a multi-arch image while carrying blobs for one, and `ctr images import --all-platforms` inside k3d and kind fails on the first it cannot find"
	}

	// The daemon resolves the platform, not the client, so ask it rather than assume
	// they match - they need not when it is remote. A template against a missing
	// field yields `<no value>` rather than an error, which would otherwise reach
	// docker as `--platform linux/<no value>`; this binary's own architecture is a
	// better guess than that.
	out, err := exec.Command("docker", "version", "--format", "{{.Server.Arch}}").Output()
	arch := strings.TrimSpace(string(out))
	if err != nil || arch == "" || strings.ContainsAny(arch, " <>\t\n") {
		arch = std_runtime.GOARCH
	}

	return "linux/" + arch, ""
})

func dockerSaveArgs(path string, platform string, images []string) []string {
	args := []string{"image", "save"}
	if platform != "" {
		args = append(args, "--platform", platform)
	}

	return append(append(args, "-o", path), images...)
}

// saveSinglePlatformArchive writes an image archive naming a single platform.
// `k3d image import` and `kind load` both run `ctr images import --all-platforms`
// whether they are given a reference or an archive, so the archive does not avoid
// that walk - it makes the walk succeed, by leaving it one descriptor to find
// rather than one per platform the daemon never pulled blobs for.
func saveSinglePlatformArchive(ctx context.Context, platform string, images ...string) (string, func(), error) {
	noop := func() {}

	file, err := os.CreateTemp("", "kuma-e2e-images-*.tar")
	if err != nil {
		return "", noop, errors.Wrap(err, "create image archive")
	}
	path := file.Name()
	cleanup := func() { _ = os.Remove(path) }
	if err := file.Close(); err != nil {
		cleanup()
		return "", noop, errors.Wrap(err, "create image archive")
	}

	used := platform
	pinned := ""
	out, err := exec.CommandContext(ctx, "docker", dockerSaveArgs(path, platform, images)...).CombinedOutput()
	// Only worth a second shot while the deadline holds: on an expired context the
	// retry cannot run at all, and reporting a pinned save would then name a flag
	// the failing command never carried.
	if err != nil && platform != "" && ctx.Err() == nil {
		// An image with no variant for this platform cannot be pinned, and docker
		// refuses the whole batch over it. Unpinned is what these tools did before,
		// and it works whenever the archive names one platform regardless.
		pinned = fmt.Sprintf(" (pinned to %s first: %s)", platform, strings.TrimSpace(string(out)))
		Logf("docker image save --platform %s failed, retrying unpinned: %s", platform, strings.TrimSpace(string(out)))
		used = ""
		out, err = exec.CommandContext(ctx, "docker", dockerSaveArgs(path, "", images)...).CombinedOutput() //nolint:contextcheck // same deadline, second shot
	}
	if err != nil {
		cleanup()
		// Carries the pinned failure too: a Silent cluster discards the log above,
		// and without it the error names an unpinned save with no sign of the first.
		return "", noop, errors.Wrapf(err, "docker image save (platform=%q images=%v)%s: %s", used, images, pinned, strings.TrimSpace(string(out)))
	}

	return path, cleanup, nil
}

// saveArchive writes the archive once, on its own short retry budget. The flake these
// imports have historically hit is k3d streaming the archive into the node, not the
// local save, so retrying the import must not drag a fresh multi-hundred-megabyte
// save along with it.
func (c *K8sCluster) saveArchive(ctx context.Context, images []string) (string, func(), error) {
	noop := func() {}

	// Logged here rather than inside the probe, which runs once per process and
	// would otherwise report against whichever spec happened to reach it first.
	platform, unpinnable := savePlatform()
	if unpinnable != "" {
		Logf("%s", unpinnable)
	}

	var archive string
	cleanup := noop
	if err := retryKeepingLastError(ctx, c.GetTesting(), "docker image save", 2, 5*time.Second, func() error {
		// Capped like the import attempts are, so a single wedged save cannot sit on
		// the budget they share. Generous against the real work - a few hundred
		// megabytes to some gigabytes written to a disk several slots are sharing -
		// because the budget above is the actual ceiling, not this.
		attempt, cancelAttempt := context.WithTimeout(ctx, 5*time.Minute)
		defer cancelAttempt()

		var err error
		archive, cleanup, err = saveSinglePlatformArchive(attempt, platform, images...)
		return err
	}); err != nil {
		// A failed attempt hands back a noop, but call it rather than rely on that
		// from two functions away: dropping a live one here leaks the archive.
		cleanup()
		return "", noop, err
	}

	return archive, cleanup, nil
}
