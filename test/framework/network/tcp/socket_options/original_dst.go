package socket_options

import (
	"fmt"
	"net"
	"syscall"
	"unsafe"
)

// SO_ORIGINAL_DST is the "optname" for "getsockopt" syscall
// ref: https://man7.org/linux/man-pages/man2/getsockopt.2.html
// ref: https://github.com/torvalds/linux/blob/5bfc75d92efd494db37f5c4c173d3639d4772966/include/uapi/linux/netfilter_ipv4.h#L52
const SO_ORIGINAL_DST = 80

// IP6T_SO_ORIGINAL_DST is the "optname" for "ipv6_getorigdst" from
// https://github.com/torvalds/linux/blob/8e0538d8ee061699b7c2cf0b193cc186952cbc21/net/netfilter/nf_conntrack_proto.c#L318
// ref: https://github.com/torvalds/linux/blob/121d1e0941e05c64ee4223064dd83eb24e871739/include/uapi/linux/netfilter_ipv6/ip6_tables.h#L182
const IP6T_SO_ORIGINAL_DST = 80

type OriginalDst struct {
	*net.TCPAddr
}

func (o *OriginalDst) Bytes() []byte {
	if o == nil {
		return nil
	}

	return []byte(o.String())
}

func ExtractOriginalDst(conn *net.TCPConn, ipv6 bool) (*OriginalDst, error) {
	file, err := conn.File()
	if err != nil {
		return nil, fmt.Errorf("cannot get underlying tcp connection's file: %s", err)
	}
	defer file.Close()

	if ipv6 {
		return getOrigDstIPv6(file.Fd())
	}

	return getOrigDstIPv4(file.Fd())
}

func getOrigDstIPv4(fd uintptr) (*OriginalDst, error) {
	mreq, err := syscall.GetsockoptIPv6Mreq(int(fd), syscall.IPPROTO_IP, SO_ORIGINAL_DST)
	if err != nil {
		if errno, ok := err.(syscall.Errno); ok && errno == syscall.ENOENT {
			return nil, nil
		}

		return nil, fmt.Errorf("cannot get socket options: %s", err)
	}

	address := net.IPv4(mreq.Multiaddr[4], mreq.Multiaddr[5], mreq.Multiaddr[6], mreq.Multiaddr[7])
	port := int(uint16(mreq.Multiaddr[2])<<8 + uint16(mreq.Multiaddr[3]))

	return &OriginalDst{
		TCPAddr: &net.TCPAddr{
			IP:   address,
			Port: port,
		},
	}, nil
}

func getOrigDstIPv6(fd uintptr) (*OriginalDst, error) {
	var raw syscall.RawSockaddrInet6
	siz := unsafe.Sizeof(raw)

	if _, _, errno := syscall.Syscall6(
		syscall.SYS_GETSOCKOPT,
		fd,
		syscall.IPPROTO_IPV6,
		IP6T_SO_ORIGINAL_DST,
		uintptr(unsafe.Pointer(&raw)),
		uintptr(unsafe.Pointer(&siz)),
		0,
	); errno != 0 {
		return nil, errno
	}

	// raw.Port is big-endian
	port := (*[2]byte)(unsafe.Pointer(&raw.Port))

	return &OriginalDst{
		TCPAddr: &net.TCPAddr{
			IP:   raw.Addr[:],
			Port: int(port[0])<<8 | int(port[1]),
		},
	}, nil
}
