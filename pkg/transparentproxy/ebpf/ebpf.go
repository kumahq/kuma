//go:build linux

package ebpf

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"unsafe"

	"github.com/kumahq/kuma/pkg/transparentproxy/config"
	ebpf_programs "github.com/kumahq/kuma/pkg/transparentproxy/ebpf/programs"
)

const (
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
	MaxItemLen = 10
	// MapRelativePathLocalPodIPs is a path where the local_pod_ips map
	// is pinned, it's hardcoded as "{BPFFS_path}/tc/globals/local_pod_ips" because
	// merbridge is hard-coding it as well, and we don't want to allot to change it
	// by mistake
	MapRelativePathLocalPodIPs   = "/local_pod_ips"
	MapRelativePathNetNSPodIPs   = "/netns_pod_ips"
	MapRelativePathCookieOrigDst = "/cookie_orig_dst"
	MapRelativePathProcessIP     = "/process_ip"
	MapRelativePathPairOrigDst   = "/pair_orig_dst"
	MapRelativePathSockPairMap   = "/sock_pair_map"
)

var programs = []*Program{
	{
		Name: "mb_connect",
		Flags: func(
			cfg config.InitializedConfigIPvX,
			cgroup string,
			bpffs string,
		) ([]string, error) {
			return Flags(map[string]string{
				"--cgroup":            cgroup,
				"--sidecar-user-id":   cfg.KumaDPUser.UID,
				"--out-redirect-port": strconv.Itoa(int(cfg.Redirect.Outbound.Port)),
				"--in-redirect-port":  strconv.Itoa(int(cfg.Redirect.Inbound.Port)),
				"--dns-capture-port":  strconv.Itoa(int(cfg.Redirect.DNS.Port)),
			})(cfg, cgroup, bpffs)
		},
		Cleanup: CleanPathsRelativeToBPFFS(
			"connect", // directory
			MapRelativePathCookieOrigDst,
			MapRelativePathNetNSPodIPs,
			MapRelativePathLocalPodIPs,
			MapRelativePathProcessIP,
		),
	},
	{
		Name: "mb_sockops",
		Flags: func(
			cfg config.InitializedConfigIPvX,
			cgroup string,
			bpffs string,
		) ([]string, error) {
			return Flags(map[string]string{
				"--cgroup":            cgroup,
				"--out-redirect-port": strconv.Itoa(int(cfg.Redirect.Outbound.Port)),
				"--in-redirect-port":  strconv.Itoa(int(cfg.Redirect.Inbound.Port)),
			})(cfg, cgroup, bpffs)
		},
		Cleanup: CleanPathsRelativeToBPFFS(
			"sockops",
			MapRelativePathCookieOrigDst,
			MapRelativePathProcessIP,
			MapRelativePathPairOrigDst,
			MapRelativePathSockPairMap,
		),
	},
	{
		Name:  "mb_get_sockopts",
		Flags: CgroupFlags,
		Cleanup: CleanPathsRelativeToBPFFS(
			"get_sockopts",
			MapRelativePathPairOrigDst,
		),
	},
	{
		Name: "mb_sendmsg",
		Flags: func(
			cfg config.InitializedConfigIPvX,
			cgroup string,
			bpffs string,
		) ([]string, error) {
			return Flags(map[string]string{
				"--cgroup":            cgroup,
				"--sidecar-user-id":   cfg.KumaDPUser.UID,
				"--out-redirect-port": strconv.Itoa(int(cfg.Redirect.Outbound.Port)),
				"--dns-capture-port":  strconv.Itoa(int(cfg.Redirect.DNS.Port)),
			})(cfg, cgroup, bpffs)
		},
		Cleanup: CleanPathsRelativeToBPFFS(
			"sendmsg",
			MapRelativePathCookieOrigDst,
		),
	},
	{
		Name: "mb_recvmsg",
		Flags: func(
			cfg config.InitializedConfigIPvX,
			cgroup string,
			bpffs string,
		) ([]string, error) {
			return Flags(map[string]string{
				"--cgroup":            cgroup,
				"--out-redirect-port": strconv.Itoa(int(cfg.Redirect.Outbound.Port)),
				"--dns-capture-port":  strconv.Itoa(int(cfg.Redirect.DNS.Port)),
			})(cfg, cgroup, bpffs)
		},
		Cleanup: CleanPathsRelativeToBPFFS(
			"recvmsg",
			MapRelativePathCookieOrigDst,
		),
	},
	{
		Name:  "mb_redir",
		Flags: Flags(nil),
		Cleanup: CleanPathsRelativeToBPFFS(
			"redir",
			MapRelativePathSockPairMap,
		),
	},
	{
		Name: "mb_tc",
		Flags: func(
			cfg config.InitializedConfigIPvX,
			cgroup string,
			bpffs string,
		) ([]string, error) {
			var err error
			var iface string

			if cfg.Ebpf.TCAttachIface != "" && InterfaceIsUp(cfg.Ebpf.TCAttachIface) {
				iface = cfg.Ebpf.TCAttachIface
			} else if iface, err = GetNonLoopbackRunningInterface(); err != nil {
				return nil, fmt.Errorf("getting non-loopback interface failed: %v", err)
			}

			return Flags(map[string]string{
				"--iface":            iface,
				"--in-redirect-port": strconv.Itoa(int(cfg.Redirect.Inbound.Port)),
			})(cfg, cgroup, bpffs)
		},
		Cleanup: CleanPathsRelativeToBPFFS(
			MapRelativePathLocalPodIPs,
			MapRelativePathPairOrigDst,
		),
	},
}

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

func ipStrToPtr(ipstr string) (unsafe.Pointer, error) {
	var ip net.IP

	if ip = net.ParseIP(ipstr); ip == nil {
		return nil, fmt.Errorf("error parse ip: %s", ipstr)
	} else if ip.To4() != nil {
		// ipv4, we need to clear the bytes
		for i := 0; i < 12; i++ {
			ip[i] = 0
		}
	}

	return unsafe.Pointer(&ip[0]), nil
}

func LoadAndAttachEbpfPrograms(programs []*Program, cfg config.InitializedConfigIPvX) error {
	var errs []string

	cgroup, err := getCgroupPath(cfg)
	if err != nil {
		return fmt.Errorf("getting cgroup failed with error: %s", err)
	}

	bpffs, err := getBpffsPath(cfg)
	if err != nil {
		return fmt.Errorf("getting bpffs failed with error: %s", err)
	}

	if err := os.MkdirAll(cfg.Ebpf.ProgramsSourcePath, os.ModePerm); err != nil {
		return fmt.Errorf("creating directory for ebpf programs failed with error: %s", err)
	}

	for _, p := range programs {
		if err := p.LoadAndAttach(cfg, ebpf_programs.Programs, cgroup, bpffs); err != nil {
			errs = append(errs, err.Error())
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("loading and attaching ebpf programs failed:\n\t%s",
			strings.Join(errs, "\n\t"))
	}

	return nil
}
