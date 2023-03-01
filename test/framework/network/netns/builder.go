package netns

import (
	"fmt"
	"math"
	"net"
	"runtime"
	"strconv"
	"time"

	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

const Loopback = "lo"

var suffixes = map[uint8]map[uint8]struct{}{}

func newVeth(nameSeed string, suffixA, suffixB uint8) *netlink.Veth {
	suffix := fmt.Sprintf("-%d%d", suffixA, suffixB)
	la := netlink.NewLinkAttrs()
	la.Name = fmt.Sprintf("%smain%s", nameSeed, suffix)

	return &netlink.Veth{
		LinkAttrs: la,
		PeerName:  fmt.Sprintf("%speer%s", nameSeed, suffix),
	}
}

func NewLinkPair() (netlink.Link, netlink.Link, error) {
	suffixA, suffixB, err := genSuffixes()
	if err != nil {
		return nil, nil, fmt.Errorf("cannot generate suffixes: %s", err)
	}

	veth := newVeth("s-", suffixA, suffixB)
	err = netlink.LinkAdd(veth)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot add link: %s, suffixA %d, suffixB %d", err, suffixA, suffixB)
	}
	mainLink, err := netlink.LinkByName(veth.Name)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot link mainLink by name: %s", err)
	}
	peerLink, err := netlink.LinkByName(veth.PeerName)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot link peerLink by name: %s", err)
	}

	return mainLink, peerLink, err
}

type Builder struct {
	nameSeed string
	ipv6     bool
	// beforeExecFuncs are functions which should be run whenever we want to execute
	// anything inside the network namespace. For example if we want to test
	// the dns conntrack zone splitting we have to reduce the amount of available
	// local ports by writing to /proc/sys/net/ipv4/ip_local_port_range
	// (equivalent of `echo "32768   32770" > /proc/sys/net/ipv4/ip_local_port_range`).
	// By doing so we have to remember that this change is ephemeral and will be
	// applied only for the locked goroutine which it was invoked from
	beforeExecFuncs   []func() error
	sharedLink        *netlink.Link
	sharedLinkAddress *netlink.Addr
}

func (b *Builder) WithNameSeed(seed string) *Builder {
	b.nameSeed = seed

	return b
}

func (b *Builder) WithIPv6(value bool) *Builder {
	b.ipv6 = value

	return b
}

func (b *Builder) WithBeforeExecFuncs(fns ...func() error) *Builder {
	b.beforeExecFuncs = append(b.beforeExecFuncs, fns...)

	return b
}

// we need some values which will make all names we will use to create resources
// (netns name, ip addresses, veth interface names) unique.
// I decided that the easiest way go achieve this uniqueness is to generate
// 2 uint8 values which will be representing second and third octets in the 10.0.0.0/24
// subnet, which will allow us to generate ip (v4) addresses as well as the names.
// genSuffixes will check if any network interface has already assigned subnet
// within the range we are interested in and ignore suffixes in this range
// Example of names regarding generated suffixes:
// suffixes: 123, 254
//
//	netns name:			kmesh-123254
//	veth main name:		kmesh-main-123254
//	veth peer name:		kmesh-peer-123254
//	veth main address:	10.123.254.1
//	veth main cidr:		10.123.254.1/24
//	veth peer address:	10.123.254.2
//	veth peer cidr:		10.123.254.2/24
func genSuffixes() (uint8, uint8, error) {
	ifaceAddresses, err := getIfaceAddresses()
	if err != nil {
		return 0, 0, fmt.Errorf("cannot get network interface addresses: %s", err)
	}

	for i := uint8(1); i < math.MaxUint8; i++ {
		var s map[uint8]struct{}
		var ok bool

		if s, ok = suffixes[i]; ok {
			if len(s) >= math.MaxUint8-1 {
				continue
			}
		} else {
			suffixes[i] = map[uint8]struct{}{
				1: {},
			}

			if ifaceContainsAddress(ifaceAddresses, net.IP{10, i, 1, 0}) {
				continue
			}

			return i, 1, nil
		}

		for j := uint8(1); j < math.MaxUint8; j++ {
			if _, ok := s[j]; !ok {
				s[j] = struct{}{}

				if !ifaceContainsAddress(ifaceAddresses, net.IP{10, i, j, 0}) {
					return i, j, nil
				}
			}
		}
	}

	return 0, 0, fmt.Errorf("out of available suffixes")
}

func getIfaceAddresses() ([]*net.IPNet, error) {
	var addresses []*net.IPNet

	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("cannot list network interfaces: %s", err)
	}

	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			return nil, fmt.Errorf("cannot list network interface's addresses: %s", err)
		}

		for _, addr := range addrs {
			if err != nil {
				return nil, fmt.Errorf("cannot resolve tcp address: %s", err)
			}

			addresses = append(addresses, addr.(*net.IPNet))
		}
	}

	return addresses, nil
}

func ifaceContainsAddress(addresses []*net.IPNet, address net.IP) bool {
	for _, ipNet := range addresses {
		if ipNet.Contains(address) {
			return true
		}
	}

	return false
}

func genIPv4IPNet(octet2, octet3, octet4 uint8) *net.IPNet {
	return &net.IPNet{
		IP:   net.IP{10, octet2, octet3, octet4},
		Mask: net.CIDRMask(24, 32),
	}
}

func genIPv6IPNet(octet1, octet2, octet3 uint8) *net.IPNet {
	hex6 := strconv.FormatInt(int64(octet1), 16)
	hex7 := strconv.FormatInt(int64(octet2), 16)
	hex8 := strconv.FormatInt(int64(octet3), 16)

	address := fmt.Sprintf("fd00::%s:%s:%s", hex6, hex7, hex8)

	return &net.IPNet{
		IP:   net.ParseIP(address),
		Mask: net.CIDRMask(64, 128),
	}
}

func genIPNet(ipv6 bool, octet1, octet2, octet3 uint8) *net.IPNet {
	if ipv6 {
		return genIPv6IPNet(octet1, octet2, octet3)
	}

	return genIPv4IPNet(octet1, octet2, octet3)
}

func genNetNSName(nameSeed string, suffixA, suffixB uint8) string {
	return fmt.Sprintf("%s%d%d", nameSeed, suffixA, suffixB)
}

func (b *Builder) Build() (*NetNS, error) {
	suffixA, suffixB, err := genSuffixes()
	if err != nil {
		return nil, fmt.Errorf("cannot generate suffixes: %s", err)
	}

	var ns *NetNS

	done := make(chan error)

	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		originalNS, err := netns.Get()
		if err != nil {
			done <- fmt.Errorf("cannot get the original network namespace: %s", err)
		}

		// Create a pair of veth interfaces
		veth := newVeth(b.nameSeed, suffixA, suffixB)
		if addLinkErr := netlink.LinkAdd(veth); addLinkErr != nil {
			done <- fmt.Errorf("cannot add veth interfaces: %s", addLinkErr)
		}

		mainLink, err := netlink.LinkByName(veth.Name)
		if err != nil {
			done <- fmt.Errorf("cannot get main veth interface: %s", err)
		}

		// peer link - interface which will be moved to the custom network namespace
		peerLink, err := netlink.LinkByName(veth.PeerName)
		if err != nil {
			done <- fmt.Errorf("cannot get peer veth interface: %s", err)
		}

		peerIPNet := genIPNet(b.ipv6, suffixA, suffixB, 2)
		peerAddr, err := netlink.ParseAddr(peerIPNet.String())
		if err != nil {
			done <- fmt.Errorf("cannot parse peer veth interface address: %s", err)
		}

		mainIPNet := genIPNet(b.ipv6, suffixA, suffixB, 1)
		mainAddr, err := netlink.ParseAddr(mainIPNet.String())
		if err != nil {
			done <- fmt.Errorf("cannot parse main veth interface address: %s", err)
		}

		if err := netlink.AddrAdd(mainLink, mainAddr); err != nil {
			done <- fmt.Errorf("cannot add address to main veth interface: %s", err)
		}

		// I tried to mitigate the problem with IPv6 2s delay (described before
		// the end of the body of this function) by statically adding
		// the neighbor information to the neighbor table, but it didn't fix
		// the issue. I left it even if it doesn't fix the issue as it doesn't
		// have any downsides, but maybe can introduce some performance benefits
		// when running the tests, as the networking stack doesn't have to ask
		// for the neighbor information for our veth-related addresses.
		if err := netlink.NeighAdd(&netlink.Neigh{
			LinkIndex:    mainLink.Attrs().Index,
			State:        NUD_REACHABLE,
			IP:           peerAddr.IP,
			HardwareAddr: peerLink.Attrs().HardwareAddr,
		}); err != nil {
			done <- fmt.Errorf("cannot add neighbor: %s", err)
		}

		if err := netlink.LinkSetUp(mainLink); err != nil {
			done <- fmt.Errorf("cannot set main veth interface up: %s", err)
		}

		// Create a new network namespace (when creating new namespace,
		// we are automatically switching into it)
		//
		// netns.NewNamed calls unix.Unshare(CLONE_NEWNET) which requires CAP_SYS_ADMIN
		// capability (ref. https://man7.org/linux/man-pages/man2/unshare.2.html)
		nsName := genNetNSName(b.nameSeed, suffixA, suffixB)
		newNS, err := netNsNewNamed(nsName)
		if err != nil {
			done <- fmt.Errorf("cannot create new network namespace: %s", err)
		}

		// set the loopback interface up
		lo, err := netlink.LinkByName(Loopback)
		if err != nil {
			done <- fmt.Errorf("cannot get loopback interface: %s", err)
		}

		if err := netlink.LinkSetUp(lo); err != nil {
			done <- fmt.Errorf("cannot set loopback interface up: %s", err)
		}

		// switch to the original namespace to assign veth peer interface
		// to our freshly made namespace
		if err := netns.Set(originalNS); err != nil {
			done <- fmt.Errorf("cannot switch to original network namespace: %s", err)
		}

		if b.sharedLink != nil {
			if err := netlink.LinkSetNsFd(*b.sharedLink, int(newNS)); err != nil {
				done <- fmt.Errorf("cannot put peer shared link inside new network interface: %s", err)
			}
		}

		// Adding an interface to a network namespace will cause the interface
		// to lose its existing IP address, so we cannot assign it earlier.
		if err := netlink.LinkSetNsFd(peerLink, int(newNS)); err != nil {
			done <- fmt.Errorf("cannot put peer veth interface inside new network interface: %s", err)
		}

		if err := netns.Set(newNS); err != nil {
			done <- fmt.Errorf("cannot switch to new network interface: %s", err)
		}

		if err := netlink.AddrAdd(peerLink, peerAddr); err != nil {
			done <- fmt.Errorf("cannot add address to peer veth interface: %s", err)
		}

		if b.sharedLink != nil && b.sharedLinkAddress != nil {
			if err := netlink.LinkSetUp(*b.sharedLink); err != nil {
				done <- fmt.Errorf("cannot set shared link interface up: %s", err)
			}

			if err := netlink.AddrAdd(*b.sharedLink, b.sharedLinkAddress); err != nil {
				done <- fmt.Errorf("cannot add address %s to link %s interface: %s", b.sharedLinkAddress.String(), *(b.sharedLink), err)
			}
		}

		// I tried to mitigate the problem with IPv6 2s delay (described before
		// the end of the body of this function) by statically adding
		// the neighbor information to the neighbor table, but it didn't fix
		// the issue. I left it even if it doesn't fix the issue as it doesn't
		// have any downsides, but maybe can introduce some performance benefits
		// when running the tests, as the networking stack doesn't have to ask
		// for the neighbor information for our veth-related addresses.
		if err := netlink.NeighAdd(&netlink.Neigh{
			LinkIndex:    peerLink.Attrs().Index,
			State:        NUD_REACHABLE,
			IP:           mainAddr.IP,
			HardwareAddr: mainLink.Attrs().HardwareAddr,
		}); err != nil {
			done <- fmt.Errorf("cannot add peer neighbor: %s", err)
		}

		if err := netlink.LinkSetUp(peerLink); err != nil {
			done <- fmt.Errorf("cannot set peer veth interface up: %s", err)
		}

		if err := netlink.RouteAdd(&netlink.Route{Gw: mainAddr.IP}); err != nil {
			done <- fmt.Errorf("cannot set the default route: %s", err)
		}

		if err := netns.Set(originalNS); err != nil {
			done <- fmt.Errorf("cannot switch to original network namespace: %s", err)
		}

		ns = &NetNS{
			name:       nsName,
			ns:         newNS,
			originalNS: originalNS,
			veth: &Veth{
				veth:      veth,
				name:      veth.Name,
				peerName:  veth.PeerName,
				ipNet:     mainIPNet,
				peerIPNet: peerIPNet,
			},
			sharedLinkAddress: b.sharedLinkAddress,
			beforeExecFuncs:   b.beforeExecFuncs,
		}

		// When configuring network namespace with IPv6 addresses, on some
		// machines when sending requests immediately after configuring netns
		// they are time-outing, when on others there is 1-2s delay before
		// receiving the response.
		// bartsmykla: I was able to narrow down the cause to be related to
		// IPv6 neighbor discovery, but was not able to find a real reason
		// or a fix, so we have to introduce this delay.
		// bartsmykla: the other observation I got is when sending requests
		// immediately, they are going through ens5/eth0 interface instead of
		// created by us veth main one
		//
		// tcpdump with the delay:
		// 	12:14:15.019582 IP6 :: > ff02::16: HBH ICMP6, multicast listener report v2, 2 group record(s), length 48
		// 	12:14:15.019602 IP6 :: > ff02::16: HBH ICMP6, multicast listener report v2, 2 group record(s), length 48
		// 	12:14:15.043570 IP6 :: > ff02::1:ff4e:9e8d: ICMP6, neighbor solicitation, who has fe80::fcb3:36ff:fe4e:9e8d, length 32
		// 	12:14:15.307566 IP6 :: > ff02::16: HBH ICMP6, multicast listener report v2, 2 group record(s), length 48
		// 	12:14:15.339567 IP6 :: > ff02::1:ff01:1: ICMP6, neighbor solicitation, who has fd00::1:1:1, length 32
		// 	12:14:15.467565 IP6 :: > ff02::1:ff01:2: ICMP6, neighbor solicitation, who has fd00::1:1:2, length 32
		// 	12:14:15.787572 IP6 :: > ff02::16: HBH ICMP6, multicast listener report v2, 2 group record(s), length 48
		// 	12:14:15.883571 IP6 :: > ff02::1:ff89:31f7: ICMP6, neighbor solicitation, who has fe80::c060:1dff:fe89:31f7, length 32
		// 	12:14:16.075579 IP6 fe80::fcb3:36ff:fe4e:9e8d > ff02::16: HBH ICMP6, multicast listener report v2, 2 group record(s), length 48
		// 	12:14:16.075597 IP6 fe80::fcb3:36ff:fe4e:9e8d > ff02::2: ICMP6, router solicitation, length 16
		// 	12:14:16.087559 IP6 fe80::fcb3:36ff:fe4e:9e8d > ff02::16: HBH ICMP6, multicast listener report v2, 2 group record(s), length 48
		// 	12:14:16.907600 IP6 fe80::c060:1dff:fe89:31f7 > ff02::16: HBH ICMP6, multicast listener report v2, 2 group record(s), length 48
		// 	12:14:16.907621 IP6 fe80::c060:1dff:fe89:31f7 > ff02::2: ICMP6, router solicitation, length 16
		// 	12:14:17.130306 IP6 fd00::1:1:1.41180 > fd00::1:1:2.38469: Flags [S], seq 2227760838, win 64800, options [mss 1440,sackOK,TS val 996411593 ecr 0,nop,wscale 7], length 0
		// 	12:14:17.130354 IP6 fd00::1:1:2.38469 > fd00::1:1:1.41180: Flags [S.], seq 2947266753, ack 2227760839, win 64260, options [mss 1440,sackOK,TS val 729905204 ecr 996411593,nop,wscale 7], length 0
		// 	12:14:17.130366 IP6 fd00::1:1:1.41180 > fd00::1:1:2.38469: Flags [.], ack 1, win 507, options [nop,nop,TS val 996411593 ecr 729905204], length 0
		// 	12:14:17.130533 IP6 fd00::1:1:2.38469 > fd00::1:1:1.41180: Flags [P.], seq 1:20, ack 1, win 503, options [nop,nop,TS val 729905204 ecr 996411593], length 19
		// 	12:14:17.130554 IP6 fd00::1:1:1.41180 > fd00::1:1:2.38469: Flags [.], ack 20, win 507, options [nop,nop,TS val 996411593 ecr 729905204], length 0
		// 	12:14:17.130566 IP6 fd00::1:1:2.38469 > fd00::1:1:1.41180: Flags [F.], seq 20, ack 1, win 503, options [nop,nop,TS val 729905204 ecr 996411593], length 0
		// 	12:14:17.130602 IP6 fd00::1:1:1.41180 > fd00::1:1:2.38469: Flags [F.], seq 1, ack 21, win 507, options [nop,nop,TS val 996411593 ecr 729905204], length 0
		// 	12:14:17.130608 IP6 fd00::1:1:2.38469 > fd00::1:1:1.41180: Flags [.], ack 2, win 503, options [nop,nop,TS val 729905204 ecr 996411593], length 0
		//
		// tcpdump without the delay:
		// 	12:17:06.899580 IP6 :: > ff02::16: HBH ICMP6, multicast listener report v2, 2 group record(s), length 48
		// 	12:17:06.899606 IP6 :: > ff02::16: HBH ICMP6, multicast listener report v2, 2 group record(s), length 48
		// 	12:17:07.010366 IP6 fe80::808:89ff:fe5c:d9df.48138 > fd00::1:1:2.35895: Flags [S], seq 3993000731, win 64800, options [mss 1440,sackOK,TS val 4294921797 ecr 0,nop,wscale 7], length 0
		// 	12:17:07.019578 IP6 :: > ff02::1:ff01:2: ICMP6, neighbor solicitation, who has fd00::1:1:2, length 32
		// 	12:17:07.071570 IP6 :: > ff02::1:ff01:1: ICMP6, neighbor solicitation, who has fd00::1:1:1, length 32
		// 	12:17:07.307569 IP6 :: > ff02::16: HBH ICMP6, multicast listener report v2, 2 group record(s), length 48
		// 	12:17:07.403576 IP6 :: > ff02::1:ffed:1fcf: ICMP6, neighbor solicitation, who has fe80::f087:56ff:feed:1fcf, length 32
		// 	12:17:07.435566 IP6 :: > ff02::16: HBH ICMP6, multicast listener report v2, 2 group record(s), length 48
		// 	12:17:07.659570 IP6 :: > ff02::1:ff14:5ba3: ICMP6, neighbor solicitation, who has fe80::5c9b:e6ff:fe14:5ba3, length 32
		// 	12:17:08.011565 IP6 fe80::808:89ff:fe5c:d9df.48138 > fd00::1:1:2.35895: Flags [S], seq 3993000731, win 64800, options [mss 1440,sackOK,TS val 4294922798 ecr 0,nop,wscale 7], length 0
		// 	12:17:08.427598 IP6 fe80::f087:56ff:feed:1fcf > ff02::16: HBH ICMP6, multicast listener report v2, 2 group record(s), length 48
		// 	12:17:08.427620 IP6 fe80::f087:56ff:feed:1fcf > ff02::2: ICMP6, router solicitation, length 16
		// 	12:17:08.679570 IP6 fe80::f087:56ff:feed:1fcf > ff02::16: HBH ICMP6, multicast listener report v2, 2 group record(s), length 48
		// 	12:17:08.683581 IP6 fe80::5c9b:e6ff:fe14:5ba3 > ff02::16: HBH ICMP6, multicast listener report v2, 2 group record(s), length 48
		// 	12:17:08.683595 IP6 fe80::5c9b:e6ff:fe14:5ba3 > ff02::2: ICMP6, router solicitation, length 16
		// 	12:17:09.163572 IP6 fe80::5c9b:e6ff:fe14:5ba3 > ff02::16: HBH ICMP6, multicast listener report v2, 2 group record(s), length 48
		// 	12:17:10.027570 IP6 fe80::808:89ff:fe5c:d9df.48138 > fd00::1:1:2.35895: Flags [S], seq 3993000731, win 64800, options [mss 1440,sackOK,TS val 4294924814 ecr 0,nop,wscale 7], length 0
		// 	12:17:10.027620 IP6 fd00::1:1:2 > ff02::1:ff5c:d9df: ICMP6, neighbor solicitation, who has fe80::808:89ff:fe5c:d9df, length 32
		// 	12:17:11.051572 IP6 fd00::1:1:2 > ff02::1:ff5c:d9df: ICMP6, neighbor solicitation, who has fe80::808:89ff:fe5c:d9df, length 32
		// 	12:17:12.075570 IP6 fd00::1:1:2 > ff02::1:ff5c:d9df: ICMP6, neighbor solicitation, who has fe80::808:89ff:fe5c:d9df, length 32
		// 	12:17:12.299565 IP6 fe80::f087:56ff:feed:1fcf > ff02::2: ICMP6, router solicitation, length 16
		// 	12:17:12.555561 IP6 fe80::5c9b:e6ff:fe14:5ba3 > ff02::2: ICMP6, router solicitation, length 16
		// 	12:17:14.091569 IP6 fe80::808:89ff:fe5c:d9df.48138 > fd00::1:1:2.35895: Flags [S], seq 3993000731, win 64800, options [mss 1440,sackOK,TS val 4294928878 ecr 0,nop,wscale 7], length 0
		// 	12:17:14.091607 IP6 fd00::1:1:2 > ff02::1:ff5c:d9df: ICMP6, neighbor solicitation, who has fe80::808:89ff:fe5c:d9df, length 32
		// 	12:17:15.115566 IP6 fd00::1:1:2 > ff02::1:ff5c:d9df: ICMP6, neighbor solicitation, who has fe80::808:89ff:fe5c:d9df, length 32
		if b.ipv6 {
			time.Sleep(2 * time.Second)
		}

		close(done)
	}()

	return ns, <-done
}

func (b *Builder) WithSharedLink(link netlink.Link, linkAddress *netlink.Addr) *Builder {
	b.sharedLink = &link
	b.sharedLinkAddress = linkAddress
	return b
}

func NewNetNSBuilder() *Builder {
	return &Builder{
		nameSeed: "kmesh-",
	}
}
