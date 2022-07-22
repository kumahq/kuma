package transparentproxy

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path"
	"strings"
	"unsafe"

	"golang.org/x/sys/unix"
)

// MaxItemLen is the maximal amount of items like ports or IP ranges to include
// or/and exclude. It's currently hardcoded to 10 as merbridge during creation
// of this map is assigning hardcoded 244 bytes for map values:
//
//  Cidr:        8 bytes
//    Cidr.Net:  4 bytes
//    Cidr.Mask: 1 byte
//    pad:       3 bytes
//
//  PodConfig:                                  244 bytes
//    PodConfig.StatusPort:                       2 bytes
//    pad:                                        2 bytes
//    PodConfig.ExcludeOutRanges (10x Cidr):     80 bytes
//    PodConfig.IncludeOutRanges (10x Cidr):     80 bytes
//    PodConfig.IncludeInPorts   (10x 2 bytes):  20 bytes
//    PodConfig.IncludeOutPorts  (10x 2 bytes):  20 bytes
//    PodConfig.ExcludeInPorts   (10x 2 bytes):  20 bytes
//    PodConfig.ExcludeOutPorts  (10x 2 bytes):  20 bytes
//
// todo (bartsmykla): merbridge flagged this constant to be changed, so if
//                    it will be changed, we have to update it
const MaxItemLen = 10

// LocalPodIPSPinnedMapPathRelativeToBPFFS is a path where the local_pod_ips map
// is pinned, it's hardcoded as "{BPFFS_path}/tc/globals/local_pod_ips" because
// merbridge is hard-coding it as well, and we don't want to allot to change it
// by mistake
const LocalPodIPSPinnedMapPathRelativeToBPFFS = "/tc/globals/local_pod_ips"

type Cidr struct {
	Net  uint32 // network order
	Mask uint8
	_    [3]uint8 // pad
}

type PodConfig struct {
	StatusPort       uint16
	_                uint16 // pad
	ExcludeOutRanges [MaxItemLen]Cidr
	IncludeOutRanges [MaxItemLen]Cidr
	IncludeInPorts   [MaxItemLen]uint16
	IncludeOutPorts  [MaxItemLen]uint16
	ExcludeInPorts   [MaxItemLen]uint16
	ExcludeOutPorts  [MaxItemLen]uint16
}

func IpStrToUint32(ipstr string) (uint32, error) {
	ip := net.ParseIP(ipstr)
	if ip == nil {
		return 0, fmt.Errorf("error when parsing ip string: %s", ipstr)
	}

	return *(*uint32)(unsafe.Pointer(&ip[12])), nil
}

func run(cmdToExec string, args, envVars []string, stdout, stderr io.Writer) error {
	_, _ = stdout.Write([]byte(fmt.Sprintf("Running: %s %s %s\n", strings.Join(envVars, " "), cmdToExec, strings.Join(args, " "))))

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

func runMake(target, sourcePath, bpfFsPath string, stdout, stderr io.Writer) error {
	args := []string{"--directory", sourcePath, target}
	envVars := []string{
		"MESH_MODE=kuma",
		"USE_RECONNECT=1",
		"DEBUG=1",
		"PROG_MOUNT_PATH=" + bpfFsPath,
	}

	return run("make", args, envVars, stdout, stderr)
}

type EbpfProgram struct {
	PinName          string
	SourcePath       string
	BpfFsPath        string
	MakeLoadTarget   string
	MakeAttachTarget string
}

func LoadAndAttachEbpfPrograms(programs []*EbpfProgram, stdout, stderr io.Writer) error {
	var errs []string

	for _, p := range programs {
		if _, err := os.Stat(path.Join(p.BpfFsPath, p.PinName)); err != nil {
			if os.IsNotExist(err) {
				if err := runMake(p.MakeLoadTarget, p.SourcePath, p.BpfFsPath, stdout, stderr); err != nil {
					errs = append(errs, err.Error())
					continue
				}

				if err := runMake(p.MakeAttachTarget, p.SourcePath, p.BpfFsPath, stdout, stderr); err != nil {
					errs = append(errs, err.Error())
				}
			} else {
				errs = append(errs, err.Error())
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("loading and attaching ebpf programs failed:\n\t%s", strings.Join(errs, "\n\t"))
	}

	return nil
}

func isDirEmpty(dirPath string) (bool, error) {
	dir, err := os.ReadDir(dirPath)
	if err != nil {
		return false, err
	}

	var isEmpty []bool
	for _, entry := range dir {
		if !entry.IsDir() {
			return false, nil
		}

		e, err := isDirEmpty(path.Join(dirPath, entry.Name()))
		if err != nil {
			return false, err
		}

		if !e {
			isEmpty = append(isEmpty, e)
		}
	}

	return len(isEmpty) == 0, nil
}

func InitBPFFSMaybe(fsPath string) error {
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

	if err := unix.Mount("bpf", fsPath, "bpf", 0, ""); err != nil {
		return fmt.Errorf("mounting BPF file system failed: %v", err)
	}

	if err := os.MkdirAll(path.Join(fsPath, "tc", "globals"), 0750); err != nil {
		return fmt.Errorf("making directory for tc globals pinning failed: %v", err)
	}

	return nil
}
