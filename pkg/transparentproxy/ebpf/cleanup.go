//go:build linux

package ebpf

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/moby/sys/mountinfo"

	"github.com/kumahq/kuma/pkg/transparentproxy/config"
)

func buildOptionSet(rawOptions string) map[string]struct{} {
	options := map[string]struct{}{}

	for _, option := range strings.Split(rawOptions, ",") {
		options[option] = struct{}{}
	}

	return options
}

func CleanPathsRelativeToBPFFS(paths ...string) func(cfg config.Config) error {
	return func(cfg config.Config) error {
		for _, p := range paths {
			if err := os.RemoveAll(path.Join(cfg.Ebpf.BPFFSPath, p)); err != nil {
				return fmt.Errorf(
					"cleaning paths relative to BPF file system failed: %s",
					err,
				)
			}
		}

		return nil
	}
}

func UnloadEbpfPrograms(programs []*Program, cfg config.Config) (string, error) {
	if os.Getuid() != 0 {
		return "", fmt.Errorf("root user in required for this process or container")
	}

	mounts, err := mountinfo.GetMounts(mountinfo.SingleEntryFilter(cfg.Ebpf.BPFFSPath))
	if err != nil {
		return "", fmt.Errorf("getting mounts failed with error: %s", err)
	}

	if len(mounts) == 0 {
		return "", fmt.Errorf(
			"it appears expected BPF FS mount with path %s doesn't exist",
			cfg.Ebpf.BPFFSPath,
		)
	}

	if len(mounts) > 1 {
		return "", fmt.Errorf(
			"unexpected more than one mounts with path %s",
			cfg.Ebpf.BPFFSPath,
		)
	}

	options := buildOptionSet(mounts[0].Options)

	if _, ok := options["ro"]; ok {
		return "it appears provided BPF filesystem path is read only" +
			" - no cleanup necessary", nil
	}

	for _, program := range programs {
		if err := program.Cleanup(cfg); err != nil {
			return "", fmt.Errorf("cleanup of %s failed: %s", program.Name, err)
		}
	}

	return "", nil
}

func Cleanup(cfg config.Config) (string, error) {
	return UnloadEbpfPrograms(programs, cfg)
}
