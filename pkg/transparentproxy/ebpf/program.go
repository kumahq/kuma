//go:build linux

package ebpf

import (
	"embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/moby/sys/mountinfo"
	"golang.org/x/sys/unix"

	"github.com/kumahq/kuma/pkg/transparentproxy/config"
)

const (
	FSTypeCgroup2 = "cgroup2"
	FSTypeBPF     = "bpf"
)

type FlagGenerator = func(
	cfg config.InitializedConfigIPvX,
	cgroup string,
	bpffs string,
) ([]string, error)

type Program struct {
	Name    string
	Flags   FlagGenerator
	Cleanup func(cfg config.InitializedConfigIPvX) error
}

func (p Program) LoadAndAttach(cfg config.InitializedConfigIPvX, programs embed.FS, cgroup, bpffs string) error {
	programBytes, err := programs.ReadFile(p.Name)
	if err != nil {
		return fmt.Errorf("reading ebpf program bytes failed: %s", err)
	}

	programPath := path.Join(cfg.Ebpf.ProgramsSourcePath, p.Name)
	if err := os.WriteFile(
		programPath,
		programBytes,
		0o744, // #nosec G306
	); err != nil {
		return fmt.Errorf("writing program bytes to file failed with error: %s", err)
	}

	flags, err := p.Flags(cfg, cgroup, bpffs)
	if err != nil {
		return err
	}

	return run(programPath, flags, nil, cfg.RuntimeStdout, cfg.RuntimeStderr)
}

func run(cmdToExec string, args, envVars []string, stdout, stderr io.Writer) error {
	_, _ = fmt.Fprintf(stdout, "Running: %s %s %s\n", strings.Join(envVars, " "), cmdToExec, strings.Join(args, " "))

	cmd := exec.Command(cmdToExec, args...)
	cmd.Env = append(os.Environ(), envVars...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err := cmd.Run()
	if code := cmd.ProcessState.ExitCode(); code != 0 || err != nil {
		return fmt.Errorf("unexpected exit code: %d, err: %v", code, err)
	}

	_, _ = stdout.Write([]byte("\n"))

	return nil
}

func isDirEmpty(dirPath string) (bool, error) {
	dir, err := os.ReadDir(dirPath)
	if err != nil {
		return false, err
	}

	for _, entry := range dir {
		if !entry.IsDir() {
			return false, nil
		}

		fullPath := path.Join(dirPath, entry.Name())

		if isEmpty, err := isDirEmpty(fullPath); err != nil || !isEmpty {
			return false, err
		}
	}

	return true, nil
}

func initBPFFSMaybe(fsPath string) error {
	stat, err := os.Stat(fsPath)
	if err != nil {
		return err
	}

	if !stat.IsDir() {
		return fmt.Errorf("bpf fs path (%s) is not a directory", fsPath)
	}

	isEmpty, err := isDirEmpty(fsPath)
	if err != nil {
		return fmt.Errorf("checking if BPF file system path is empty failed: %v", err)
	}

	// if directory is not empty, we are assuming BPF filesystem was already
	// initialized, so we won't do it again
	if !isEmpty {
		return nil
	}

	if err := unix.Mount(FSTypeBPF, fsPath, FSTypeBPF, 0, ""); err != nil {
		return fmt.Errorf("mounting BPF file system failed: %v", err)
	}

	return nil
}

func getCgroupPath(cfg config.InitializedConfigIPvX) (string, error) {
	cgroupPath := cfg.Ebpf.CgroupPath

	if cgroupPath != "" {
		mounts, err := mountinfo.GetMounts(mountinfo.SingleEntryFilter(cgroupPath))
		if err != nil {
			return "", fmt.Errorf(
				"getting mount (%s) failed with error: %s", cgroupPath, err)
		}

		if len(mounts) == 1 {
			fsType, mountpoint := mounts[0].FSType, mounts[0].Mountpoint

			if fsType == FSTypeCgroup2 {
				return mountpoint, nil
			}

			_, _ = fmt.Fprintf(cfg.RuntimeStderr,
				"warning: found mount %s, but it's type (%s) is not %s - ignoring\n",
				cgroupPath, fsType, FSTypeCgroup2,
			)
		} else {
			_, _ = fmt.Fprintf(cfg.RuntimeStderr,
				"warning: cannot find mount %s - ignoring\n", cgroupPath)
		}
	}

	mounts, err := mountinfo.GetMounts(mountinfo.FSTypeFilter(FSTypeCgroup2))
	if err != nil {
		return "", fmt.Errorf("getting mounts failed with error: %s", err)
	}

	if len(mounts) == 0 {
		return "", fmt.Errorf("cannot find any %s mounts", FSTypeCgroup2)
	}

	if len(mounts) > 1 {
		var mountpoints []string

		for _, mount := range mounts {
			mountpoints = append(mountpoints, mount.Mountpoint)
		}

		_, _ = fmt.Fprintf(cfg.RuntimeStderr,
			"warning: found %d %s mounts, only first one (%s) will be used (ignored: [%s])\n",
			len(mounts), FSTypeCgroup2, mountpoints[0], strings.Join(mountpoints, ", "),
		)
	}

	_, _ = fmt.Fprintf(cfg.RuntimeStdout,
		"Using %s mount: %s\n", FSTypeCgroup2, mounts[0].Mountpoint)

	return mounts[0].Mountpoint, nil
}

func getBpffsPath(cfg config.InitializedConfigIPvX) (string, error) {
	bpffsPath := cfg.Ebpf.BPFFSPath

	if bpffsPath != "" {
		mounts, err := mountinfo.GetMounts(mountinfo.SingleEntryFilter(bpffsPath))
		if err != nil {
			return "", fmt.Errorf("getting mount (%s) failed with error: %s",
				bpffsPath, err,
			)
		}

		if len(mounts) == 1 {
			fsType := mounts[0].FSType

			if fsType == FSTypeBPF {
				return bpffsPath, nil
			}

			_, _ = fmt.Fprintf(cfg.RuntimeStderr,
				"warning: found mount %s, but its type (%s) is not %s - ignoring\n",
				bpffsPath, fsType, FSTypeBPF,
			)
		} else {
			_, _ = fmt.Fprintf(cfg.RuntimeStderr,
				"warning: expected %s mount: %s cannot be found - ignoring\n",
				FSTypeBPF, bpffsPath,
			)
		}
	}

	mounts, err := mountinfo.GetMounts(mountinfo.FSTypeFilter(FSTypeBPF))
	if err != nil {
		return "", fmt.Errorf("getting mounts failed with error: %s", err)
	}

	if len(mounts) == 0 {
		_, _ = fmt.Fprintf(cfg.RuntimeStderr,
			"warning: cannot find any %s mounts - will try to manually mount %s\n",
			FSTypeBPF, bpffsPath,
		)

		return bpffsPath, initBPFFSMaybe(bpffsPath)
	}

	if len(mounts) > 1 {
		var mountpoints []string

		for _, mount := range mounts {
			mountpoints = append(mountpoints, mount.Mountpoint)
		}

		_, _ = fmt.Fprintf(cfg.RuntimeStderr,
			"found %d %s mounts, only first one (%s) will be used (ignored: [%s])\n",
			len(mounts), FSTypeBPF, mountpoints[0], strings.Join(mountpoints, ", "),
		)
	}

	return mounts[0].Mountpoint, nil
}

func Flags(flags map[string]string) FlagGenerator {
	return func(cfg config.InitializedConfigIPvX, _, bpffs string) ([]string, error) {
		f := map[string]string{
			"--bpffs": bpffs,
		}

		if cfg.Verbose {
			f["--verbose"] = ""
		}

		for k, v := range flags {
			f[k] = v
		}

		return mapFlagsToSlice(f), nil
	}
}

func CgroupFlags(
	cfg config.InitializedConfigIPvX,
	cgroup string,
	bpffs string,
) ([]string, error) {
	return Flags(map[string]string{
		"--cgroup": cgroup,
	})(cfg, cgroup, bpffs)
}

func mapFlagsToSlice(flags map[string]string) []string {
	var result []string

	for k, v := range flags {
		result = append(result, k)

		if v != "" {
			result = append(result, v)
		}
	}

	return result
}
