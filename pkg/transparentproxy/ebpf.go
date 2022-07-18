package transparentproxy

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strings"
	"unsafe"
)

// MaxItemLen is the maximal amount of items like ports or IP ranges to include
// or/and exclude. It's currently hardcoded to 10 as merbridge during creation
// of this map is assigning hardcoded 244 bytes for map values:
//
//  Cidr:			8 bytes
//    Cidr.Net:		4 bytes
//    Cidr.Mask:	1 byte
//    pad:			3 bytes
//
//  PodConfig:										244 bytes
//    PodConfig.StatusPort:							2 bytes
//    pad:											2 bytes
//    PodConfig.ExcludeOutRanges	(10x Cidr):		80 bytes
//    PodConfig.IncludeOutRanges	(10x Cidr):		80 bytes
//    PodConfig.IncludeInPorts		(10x 2 bytes):	20 bytes
//    PodConfig.IncludeOutPorts		(10x 2 bytes):	20 bytes
//    PodConfig.ExcludeInPorts		(10x 2 bytes):	20 bytes
//    PodConfig.ExcludeOutPorts		(10x 2 bytes):	20 bytes
//
// todo (bartsmykla): merbridge flagged this constant to be changed, so if
//                    it will be changed, we have to update it
const MaxItemLen = 10

// LocalPodIPSPinnedMapPath is a path where the local_pod_ips map is pinned,
// it's hardcoded as "/sys/fs/bpf/tc/globals" because merbridge is hard-coding
// it as well, and we don't want to allot to change it by mistake
const LocalPodIPSPinnedMapPath = "/sys/fs/bpf/tc/globals/local_pod_ips"

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

func RunMake(target, directory string, stdout, stderr io.Writer) error {
	envVars := []string{"MESH_MODE=kuma", "USE_RECONNECT=1", "DEBUG=1"}
	args := []string{"--directory", directory, target}

	_, _ = stdout.Write([]byte(fmt.Sprintf("Running: make %v %v\n", strings.Join(args, " "), strings.Join(envVars, " "))))

	cmd := exec.Command("make", args...)
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
