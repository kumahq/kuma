package blackbox_network_tests_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"path/filepath"
	"strconv"
	"time"

	"github.com/miekg/dns"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/vishvananda/netlink"

	"github.com/kumahq/kuma/pkg/transparentproxy/config"
	"github.com/kumahq/kuma/pkg/transparentproxy/consts"
	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/builder"
	"github.com/kumahq/kuma/test/blackbox_network_tests"
	"github.com/kumahq/kuma/test/framework/network/netns"
	"github.com/kumahq/kuma/test/framework/network/socket"
	"github.com/kumahq/kuma/test/framework/network/syscall"
	"github.com/kumahq/kuma/test/framework/network/sysctl"
	"github.com/kumahq/kuma/test/framework/network/tcp"
	"github.com/kumahq/kuma/test/framework/network/udp"
)

var _ = Describe("Outbound IPv4 DNS/UDP traffic to port 53", func() {
	var err error
	var ns *netns.NetNS

	BeforeEach(func() {
		ns, err = netns.NewNetNSBuilder().Build()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		Expect(ns.Cleanup()).To(Succeed())
	})

	DescribeTable("should be redirected to provided port",
		func(randomPort uint16) {
			// given
			address := udp.GenRandomAddressIPv4(consts.DNSPort)
			tproxyConfig, err := config.Config{
				Redirect: config.Redirect{
					DNS: config.DNS{
						Enabled:    true,
						Port:       config.Port(randomPort),
						CaptureAll: true,
					},
					Inbound: config.TrafficFlow{
						Enabled: true,
					},
					Outbound: config.TrafficFlow{
						Enabled: true,
					},
				},
				RuntimeStdout: io.Discard,
			}.Initialize(context.Background())
			Expect(err).ToNot(HaveOccurred())

			serverAddress := fmt.Sprintf("%s:%d", consts.LocalhostAddress[consts.IPv4].IP, randomPort)

			readyC, errC := udp.UnsafeStartUDPServer(ns, serverAddress, udp.ReplyWithReceivedMsg)
			Consistently(errC).ShouldNot(Receive())
			Eventually(readyC).Should(BeClosed())

			// when
			Eventually(ns.UnsafeExec(func() {
				Expect(builder.RestoreIPTables(context.Background(), tproxyConfig)).Error().To(Succeed())
			})).Should(BeClosed())

			// and
			Eventually(ns.UnsafeExec(func() {
				Expect(udp.DialUDPAddrWithHelloMsgAndGetReply(address, address)).
					To(Equal(address.String()))
			})).Should(BeClosed())

			// then
			Consistently(errC).ShouldNot(Receive())
		},
		func() []TableEntry {
			var entries []TableEntry
			lockedPorts := []uint16{consts.DNSPort}

			for i := 0; i < blackbox_network_tests.TestCasesAmount; i++ {
				randomPorts := socket.GenerateRandomPortsSlice(1, lockedPorts...)
				// This gives us more entropy as all generated ports will be
				// different from each other
				lockedPorts = append(lockedPorts, randomPorts...)
				desc := fmt.Sprintf("to port %%d, from port %d", consts.DNSPort)
				entry := Entry(EntryDescription(desc), randomPorts[0])
				entries = append(entries, entry)
			}

			return entries
		}(),
	)
})

var _ = Describe("Outbound IPv4 DNS/UDP traffic to port 53", func() {
	var err error
	var ns *netns.NetNS

	BeforeEach(func() {
		ns, err = netns.NewNetNSBuilder().Build()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		Expect(ns.Cleanup()).To(Succeed())
	})

	DescribeTable("should be redirected to provided port except for traffic excluded by uid",
		func(randomPort uint16) {
			dnsUserUid := uintptr(4201) // see /.github/workflows/blackbox-tests.yaml:76
			// given
			tproxyConfig, err := config.Config{
				Redirect: config.Redirect{
					DNS: config.DNS{
						Enabled:    true,
						Port:       config.Port(randomPort),
						CaptureAll: true,
					},
					Inbound: config.TrafficFlow{
						Enabled: true,
					},
					Outbound: config.TrafficFlow{
						Enabled: true,
						ExcludePortsForUIDs: []string{
							fmt.Sprintf("udp:%d:%d", consts.DNSPort, dnsUserUid),
						},
					},
				},
				RuntimeStdout: io.Discard,
			}.Initialize(context.Background())
			Expect(err).ToNot(HaveOccurred())

			originalAddress := &net.UDPAddr{IP: consts.LocalhostAddress[consts.IPv4].IP, Port: int(consts.DNSPort)}
			redirectedToAddress := fmt.Sprintf("%s:%d", consts.LocalhostAddress[consts.IPv4].IP, randomPort)

			redirectedC, redirectedErr := udp.UnsafeStartUDPServer(ns, redirectedToAddress, udp.ReplyWithReceivedMsg)
			Consistently(redirectedErr).ShouldNot(Receive())
			Eventually(redirectedC).Should(BeClosed())

			originalC, originalErr := udp.UnsafeStartUDPServer(ns, originalAddress.String(), udp.ReplyWithMsg("excluded"))
			Consistently(originalErr).ShouldNot(Receive())
			Eventually(originalC).Should(BeClosed())

			// when
			Eventually(ns.UnsafeExec(func() {
				Expect(builder.RestoreIPTables(context.Background(), tproxyConfig)).Error().To(Succeed())
			})).Should(BeClosed())

			// and
			Eventually(ns.UnsafeExec(func() {
				Expect(udp.DialUDPAddrWithHelloMsgAndGetReply(originalAddress, originalAddress)).
					To(Equal(originalAddress.String()))
			})).Should(BeClosed())

			// and
			Eventually(ns.UnsafeExecInLoop(1, 0, func() {
				Expect(udp.DialUDPAddrWithHelloMsgAndGetReply(originalAddress, originalAddress)).
					To(Equal("excluded"))
			}, syscall.SetUID(dnsUserUid))).Should(BeClosed())

			// then
			Consistently(redirectedErr).ShouldNot(Receive())
			Consistently(originalErr).ShouldNot(Receive())
		},
		func() []TableEntry {
			var entries []TableEntry
			lockedPorts := []uint16{consts.DNSPort}

			for i := 0; i < blackbox_network_tests.TestCasesAmount; i++ {
				randomPorts := socket.GenerateRandomPortsSlice(2, lockedPorts...)
				// This gives us more entropy as all generated ports will be
				// different from each other
				lockedPorts = append(lockedPorts, randomPorts...)
				desc := fmt.Sprintf("to port %%d, from port %d", consts.DNSPort)
				entry := Entry(EntryDescription(desc), randomPorts[0])
				entries = append(entries, entry)
			}

			return entries
		}(),
	)
})

var _ = Describe("Outbound IPv6 DNS/UDP traffic to port 53", func() {
	var err error
	var ns *netns.NetNS

	BeforeEach(func() {
		ns, err = netns.NewNetNSBuilder().WithIPv6(true).Build()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		Expect(ns.Cleanup()).To(Succeed())
	})

	DescribeTable("should be redirected to provided port except for traffic excluded by uid",
		func(randomPort uint16) {
			// given
			dnsUserUid := uintptr(4201) // see /.github/workflows/blackbox-tests.yaml:76
			address := udp.GenRandomAddressIPv6(consts.DNSPort)
			tproxyConfig, err := config.Config{
				Redirect: config.Redirect{
					DNS: config.DNS{
						Enabled:    true,
						Port:       config.Port(randomPort),
						CaptureAll: true,
					},
					Inbound: config.TrafficFlow{
						Enabled: true,
					},
					Outbound: config.TrafficFlow{
						Enabled: true,
						ExcludePortsForUIDs: []string{
							fmt.Sprintf("udp:%d:%d", consts.DNSPort, dnsUserUid),
						},
					},
				},
				IPFamilyMode:  config.IPFamilyModeDualStack,
				RuntimeStdout: io.Discard,
			}.Initialize(context.Background())
			Expect(err).ToNot(HaveOccurred())

			redirectedAddress := fmt.Sprintf("%s:%d", consts.LocalhostAddress[consts.IPv6].IP, randomPort)
			originalAddress := &net.UDPAddr{IP: consts.LocalhostAddress[consts.IPv6].IP, Port: int(consts.DNSPort)}

			redirectedC, redirectedErr := udp.UnsafeStartUDPServer(ns, redirectedAddress, udp.ReplyWithReceivedMsg)
			Consistently(redirectedErr).ShouldNot(Receive())
			Eventually(redirectedC).Should(BeClosed())

			originalC, originalErr := udp.UnsafeStartUDPServer(ns, originalAddress.String(), udp.ReplyWithMsg("excluded"))
			Consistently(originalErr).ShouldNot(Receive())
			Eventually(originalC).Should(BeClosed())

			// when
			Eventually(ns.UnsafeExec(func() {
				Expect(builder.RestoreIPTables(context.Background(), tproxyConfig)).Error().To(Succeed())
			})).Should(BeClosed())

			// and
			Eventually(ns.UnsafeExec(func() {
				Expect(udp.DialUDPAddrWithHelloMsgAndGetReply(address, address)).
					To(Equal(address.String()))
			})).Should(BeClosed())

			// and
			Eventually(ns.UnsafeExecInLoop(1, 0, func() {
				Expect(udp.DialUDPAddrWithHelloMsgAndGetReply(originalAddress, originalAddress)).
					To(Equal("excluded"))
			}, syscall.SetUID(dnsUserUid))).Should(BeClosed())

			// then
			Consistently(redirectedErr).ShouldNot(Receive())
			Consistently(originalErr).ShouldNot(Receive())
		},
		func() []TableEntry {
			var entries []TableEntry
			lockedPorts := []uint16{consts.DNSPort}

			for i := 0; i < blackbox_network_tests.TestCasesAmount; i++ {
				randomPorts := socket.GenerateRandomPortsSlice(1, lockedPorts...)
				// This gives us more entropy as all generated ports will be
				// different from each other
				lockedPorts = append(lockedPorts, randomPorts...)
				desc := fmt.Sprintf("to port %%d, from port %d", consts.DNSPort)
				entry := Entry(EntryDescription(desc), randomPorts[0])
				entries = append(entries, entry)
			}

			return entries
		}(),
	)
})

var _ = Describe("Outbound IPv4 DNS/TCP traffic to port 53", func() {
	var err error
	var ns *netns.NetNS

	BeforeEach(func() {
		ns, err = netns.NewNetNSBuilder().Build()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		Expect(ns.Cleanup()).To(Succeed())
	})

	DescribeTable("should be redirected to provided port",
		func(dnsPort, outboundPort uint16) {
			// given
			address := tcp.GenRandomAddressIPv4(consts.DNSPort)
			tproxyConfig, err := config.Config{
				Redirect: config.Redirect{
					DNS: config.DNS{
						Enabled:    true,
						Port:       config.Port(dnsPort),
						CaptureAll: true,
					},
					Outbound: config.TrafficFlow{
						Port:    config.Port(outboundPort),
						Enabled: true,
					},
					Inbound: config.TrafficFlow{
						Enabled: true,
					},
				},
				RuntimeStdout: io.Discard,
			}.Initialize(context.Background())
			Expect(err).ToNot(HaveOccurred())

			serverAddress := fmt.Sprintf("%s:%d", consts.LocalhostAddress[consts.IPv4].IP, dnsPort)

			readyC, errC := tcp.UnsafeStartTCPServer(
				ns,
				serverAddress,
				tcp.ReplyWithOriginalDstIPv4,
				tcp.CloseConn,
			)
			Consistently(errC).ShouldNot(Receive())
			Eventually(readyC).Should(BeClosed())

			// when
			Eventually(ns.UnsafeExec(func() {
				Expect(builder.RestoreIPTables(context.Background(), tproxyConfig)).Error().To(Succeed())
			})).Should(BeClosed())

			// and
			Eventually(ns.UnsafeExec(func() {
				Expect(tcp.DialTCPAddrAndGetReply(address)).To(Equal(address.String()))
			})).Should(BeClosed())

			// then
			Eventually(errC).Should(BeClosed())
		},
		func() []TableEntry {
			var entries []TableEntry
			lockedPorts := []uint16{consts.DNSPort}

			for i := 0; i < blackbox_network_tests.TestCasesAmount; i++ {
				// We are drawing two ports instead of one as the first one will be used
				// to expose TCP server inside the namespace, which will be pretending
				// a DNS server which should intercept all DNS traffic on port TCP#53,
				// and the second one will be set as an outbound redirection port,
				// which wound intercept the packet if no DNS redirection would be set,
				// and we don't want them to be the same
				randomPorts := socket.GenerateRandomPortsSlice(2, lockedPorts...)
				// This gives us more entropy as all generated ports will be
				// different from each other
				lockedPorts = append(lockedPorts, randomPorts...)
				desc := fmt.Sprintf(
					"to port %d, from port %d",
					randomPorts[0],
					consts.DNSPort,
				)
				entry := Entry(EntryDescription(desc), randomPorts[0], randomPorts[1])
				entries = append(entries, entry)
			}

			return entries
		}(),
	)
})

var _ = Describe("Outbound IPv6 DNS/UDP traffic to port 53", func() {
	var err error
	var ns *netns.NetNS

	BeforeEach(func() {
		ns, err = netns.NewNetNSBuilder().WithIPv6(true).Build()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		Expect(ns.Cleanup()).To(Succeed())
	})

	DescribeTable("should be redirected to provided port",
		func(randomPort uint16) {
			// given
			address := udp.GenRandomAddressIPv6(consts.DNSPort)
			tproxyConfig, err := config.Config{
				Redirect: config.Redirect{
					DNS: config.DNS{
						Enabled:    true,
						Port:       config.Port(randomPort),
						CaptureAll: true,
					},
					Inbound: config.TrafficFlow{
						Enabled: true,
					},
					Outbound: config.TrafficFlow{
						Enabled: true,
					},
				},
				IPFamilyMode:  config.IPFamilyModeDualStack,
				RuntimeStdout: io.Discard,
			}.Initialize(context.Background())
			Expect(err).ToNot(HaveOccurred())

			serverAddress := fmt.Sprintf("%s:%d", consts.LocalhostAddress[consts.IPv6].IP, randomPort)

			readyC, errC := udp.UnsafeStartUDPServer(ns, serverAddress, udp.ReplyWithReceivedMsg)
			Consistently(errC).ShouldNot(Receive())
			Eventually(readyC).Should(BeClosed())

			// when
			Eventually(ns.UnsafeExec(func() {
				Expect(builder.RestoreIPTables(context.Background(), tproxyConfig)).Error().To(Succeed())
			})).Should(BeClosed())

			// and
			Eventually(ns.UnsafeExec(func() {
				Expect(udp.DialUDPAddrWithHelloMsgAndGetReply(address, address)).
					To(Equal(address.String()))
			})).Should(BeClosed())

			// then
			Consistently(errC).ShouldNot(Receive())
		},
		func() []TableEntry {
			var entries []TableEntry
			lockedPorts := []uint16{consts.DNSPort}

			for i := 0; i < blackbox_network_tests.TestCasesAmount; i++ {
				randomPorts := socket.GenerateRandomPortsSlice(1, lockedPorts...)
				// This gives us more entropy as all generated ports will be
				// different from each other
				lockedPorts = append(lockedPorts, randomPorts...)
				desc := fmt.Sprintf("to port %%d, from port %d", consts.DNSPort)
				entry := Entry(EntryDescription(desc), randomPorts[0])
				entries = append(entries, entry)
			}

			return entries
		}(),
	)
})

var _ = Describe("Outbound IPv6 DNS/TCP traffic to port 53", func() {
	var err error
	var ns *netns.NetNS

	BeforeEach(func() {
		ns, err = netns.NewNetNSBuilder().WithIPv6(true).Build()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		Expect(ns.Cleanup()).To(Succeed())
	})

	DescribeTable("should be redirected to provided port",
		func(dnsPort, outboundPort uint16) {
			// given
			address := tcp.GenRandomAddressIPv6(consts.DNSPort)
			tproxyConfig, err := config.Config{
				Redirect: config.Redirect{
					DNS: config.DNS{
						Enabled:    true,
						Port:       config.Port(dnsPort),
						CaptureAll: true,
					},
					Outbound: config.TrafficFlow{
						Port:    config.Port(outboundPort),
						Enabled: true,
					},
					Inbound: config.TrafficFlow{
						Enabled: true,
					},
				},
				IPFamilyMode:  config.IPFamilyModeDualStack,
				RuntimeStdout: io.Discard,
			}.Initialize(context.Background())
			Expect(err).ToNot(HaveOccurred())

			serverAddress := fmt.Sprintf("%s:%d", consts.LocalhostAddress[consts.IPv6].IP, dnsPort)

			readyC, errC := tcp.UnsafeStartTCPServer(
				ns,
				serverAddress,
				tcp.ReplyWithOriginalDstIPv6,
				tcp.CloseConn,
			)
			Consistently(errC).ShouldNot(Receive())
			Eventually(readyC).Should(BeClosed())

			// when
			Eventually(ns.UnsafeExec(func() {
				Expect(builder.RestoreIPTables(context.Background(), tproxyConfig)).Error().To(Succeed())
			})).Should(BeClosed())

			// and
			Eventually(ns.UnsafeExec(func() {
				Expect(tcp.DialTCPAddrAndGetReply(address)).To(Equal(address.String()))
			})).Should(BeClosed())

			// then
			Eventually(errC).Should(BeClosed())
		},
		func() []TableEntry {
			var entries []TableEntry
			lockedPorts := []uint16{consts.DNSPort}

			for i := 0; i < blackbox_network_tests.TestCasesAmount; i++ {
				// We are drawing two ports instead of one as the first one will be used
				// to expose TCP server inside the namespace, which will be pretending
				// a DNS server which should intercept all DNS traffic on port TCP#53,
				// and the second one will be set as an outbound redirection port,
				// which wound intercept the packet if no DNS redirection would be set,
				// and we don't want them to be the same
				randomPorts := socket.GenerateRandomPortsSlice(2, lockedPorts...)
				// This gives us more entropy as all generated ports will be
				// different from each other
				lockedPorts = append(lockedPorts, randomPorts...)
				desc := fmt.Sprintf(
					"to port %d, from port %d",
					randomPorts[0],
					consts.DNSPort,
				)
				entry := Entry(EntryDescription(desc), randomPorts[0], randomPorts[1])
				entries = append(entries, entry)
			}

			return entries
		}(),
	)
})

var _ = Describe("Outbound IPv4 DNS/UDP conntrack zone splitting", func() {
	var err error
	var ns *netns.NetNS

	BeforeEach(func() {
		ns, err = netns.NewNetNSBuilder().
			WithBeforeExecFuncs(sysctl.SetLocalPortRange(32768, 32770)).
			Build()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		Expect(ns.Cleanup()).To(Succeed())
	})

	DescribeTable("should be redirected to provided port",
		func(port uint16) {
			// given
			uid := uintptr(5678)
			s1Address := fmt.Sprintf("%s:%d", ns.Veth().PeerAddress(), consts.DNSPort)
			s2Address := fmt.Sprintf("%s:%d", consts.LocalhostAddress[consts.IPv4].IP, port)
			tproxyConfig, err := config.Config{
				Redirect: config.Redirect{
					DNS: config.DNS{
						Enabled:                true,
						Port:                   config.Port(port),
						SkipConntrackZoneSplit: false,
						CaptureAll:             true,
					},
					Outbound: config.TrafficFlow{
						Enabled: true,
					},
					Inbound: config.TrafficFlow{
						Enabled: true,
					},
				},
				KumaDPUser:    strconv.Itoa(int(uid)),
				RuntimeStdout: io.Discard,
			}.Initialize(context.Background())
			Expect(err).ToNot(HaveOccurred())

			want := map[string]uint{
				s1Address: blackbox_network_tests.DNSConntrackZoneSplittingStressCallsAmount,
				s2Address: blackbox_network_tests.DNSConntrackZoneSplittingStressCallsAmount,
			}

			s1ReadyC, s1ErrC := udp.UnsafeStartUDPServer(
				ns,
				s1Address,
				udp.ReplyWithLocalAddr,
			)
			Consistently(s1ErrC).ShouldNot(Receive())
			Eventually(s1ReadyC).Should(BeClosed())

			s2ReadyC, s2ErrC := udp.UnsafeStartUDPServer(
				ns,
				s2Address,
				udp.ReplyWithLocalAddr,
				sysctl.SetUnprivilegedPortStart(0),
				syscall.SetUID(uid),
			)
			Consistently(s2ErrC).ShouldNot(Receive())
			Eventually(s2ReadyC).Should(BeClosed())

			// when
			Eventually(ns.UnsafeExec(func() {
				Expect(builder.RestoreIPTables(context.Background(), tproxyConfig)).Error().To(Succeed())
			})).Should(BeClosed())

			results := udp.NewResultMap()

			exec1ErrC := ns.UnsafeExecInLoop(
				blackbox_network_tests.DNSConntrackZoneSplittingStressCallsAmount,
				time.Millisecond,
				func() {
					Expect(udp.DialAddrAndIncreaseResultMap(s1Address, results)).To(Succeed())
				},
				syscall.SetUID(uid),
			)

			exec2ErrC := ns.UnsafeExecInLoop(
				blackbox_network_tests.DNSConntrackZoneSplittingStressCallsAmount,
				time.Millisecond,
				func() {
					Expect(udp.DialAddrAndIncreaseResultMap(s1Address, results)).To(Succeed())
				},
			)

			Consistently(exec1ErrC).ShouldNot(Receive())
			Consistently(exec2ErrC).ShouldNot(Receive())
			Eventually(exec1ErrC, blackbox_network_tests.DNSConntrackZoneSplittingTestTimeout).
				Should(BeClosed())
			Eventually(exec2ErrC, blackbox_network_tests.DNSConntrackZoneSplittingTestTimeout).
				Should(BeClosed())

			Expect(results.GetFinalResults()).To(BeEquivalentTo(want))
		},
		func() []TableEntry {
			var entries []TableEntry
			lockedPorts := []uint16{consts.DNSPort}

			for i := 0; i < blackbox_network_tests.TestCasesAmount; i++ {
				ports := socket.GenerateRandomPortsSlice(1, lockedPorts...)
				// This gives us more entropy as all generated ports will be
				// different from each other
				lockedPorts = append(lockedPorts, ports...)
				desc := fmt.Sprintf("to port %%d, from port %d", consts.DNSPort)
				entry := Entry(EntryDescription(desc), ports[0])
				entries = append(entries, entry)
			}

			return entries
		}(),
	)
})

var _ = Describe("Outbound IPv6 DNS/UDP conntrack zone splitting", func() {
	var err error
	var ns *netns.NetNS

	BeforeEach(func() {
		ns, err = netns.NewNetNSBuilder().
			WithIPv6(true).
			WithBeforeExecFuncs(sysctl.SetLocalPortRange(32768, 32770)).
			Build()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		Expect(ns.Cleanup()).To(Succeed())
	})

	DescribeTable("should be redirected to provided port",
		func(port uint16) {
			// given
			uid := uintptr(5678)
			s1Address := fmt.Sprintf("%s:%d", consts.LocalhostAddress[consts.IPv6].IP, consts.DNSPort)
			s2Address := fmt.Sprintf("%s:%d", consts.LocalhostAddress[consts.IPv6].IP, port)
			tproxyConfig, err := config.Config{
				Redirect: config.Redirect{
					DNS: config.DNS{
						Enabled:                true,
						Port:                   config.Port(port),
						SkipConntrackZoneSplit: false,
						CaptureAll:             true,
					},
					Outbound: config.TrafficFlow{
						Enabled: true,
					},
					Inbound: config.TrafficFlow{
						Enabled: true,
					},
				},
				IPFamilyMode:  config.IPFamilyModeDualStack,
				KumaDPUser:    strconv.Itoa(int(uid)),
				RuntimeStdout: io.Discard,
			}.Initialize(context.Background())
			Expect(err).ToNot(HaveOccurred())

			want := map[string]uint{
				s1Address: blackbox_network_tests.DNSConntrackZoneSplittingStressCallsAmount,
				s2Address: blackbox_network_tests.DNSConntrackZoneSplittingStressCallsAmount,
			}

			s1ReadyC, s1ErrC := udp.UnsafeStartUDPServer(
				ns,
				s1Address,
				udp.ReplyWithLocalAddr,
			)
			Consistently(s1ErrC).ShouldNot(Receive())
			Eventually(s1ReadyC).Should(BeClosed())

			s2ReadyC, s2ErrC := udp.UnsafeStartUDPServer(
				ns,
				s2Address,
				udp.ReplyWithLocalAddr,
				sysctl.SetUnprivilegedPortStart(0),
				syscall.SetUID(uid),
			)
			Consistently(s2ErrC).ShouldNot(Receive())
			Eventually(s2ReadyC).Should(BeClosed())

			// when
			Eventually(ns.UnsafeExec(func() {
				Expect(builder.RestoreIPTables(context.Background(), tproxyConfig)).Error().To(Succeed())
			})).Should(BeClosed())

			results := udp.NewResultMap()

			exec1ErrC := ns.UnsafeExecInLoop(
				blackbox_network_tests.DNSConntrackZoneSplittingStressCallsAmount,
				time.Millisecond,
				func() {
					Expect(udp.DialAddrAndIncreaseResultMap(s1Address, results)).To(Succeed())
				},
				syscall.SetUID(uid),
			)

			exec2ErrC := ns.UnsafeExecInLoop(
				blackbox_network_tests.DNSConntrackZoneSplittingStressCallsAmount,
				time.Millisecond,
				func() {
					Expect(udp.DialAddrAndIncreaseResultMap(s1Address, results)).To(Succeed())
				},
			)

			Consistently(exec1ErrC).ShouldNot(Receive())
			Consistently(exec2ErrC).ShouldNot(Receive())
			Eventually(exec1ErrC, blackbox_network_tests.DNSConntrackZoneSplittingTestTimeout).
				Should(BeClosed())
			Eventually(exec2ErrC, blackbox_network_tests.DNSConntrackZoneSplittingTestTimeout).
				Should(BeClosed())

			Expect(results.GetFinalResults()).To(BeEquivalentTo(want))
		},
		func() []TableEntry {
			var entries []TableEntry
			lockedPorts := []uint16{consts.DNSPort}

			for i := 0; i < blackbox_network_tests.TestCasesAmount; i++ {
				ports := socket.GenerateRandomPortsSlice(1, lockedPorts...)
				// This gives us more entropy as all generated ports will be
				// different from each other
				lockedPorts = append(lockedPorts, ports...)
				desc := fmt.Sprintf("to port %%d, from port %d", consts.DNSPort)
				entry := Entry(EntryDescription(desc), ports[0])
				entries = append(entries, entry)
			}

			return entries
		}(),
	)
})

var _ = Describe("Outbound IPv4 DNS/UDP traffic to port 53 only for addresses in configuration ", func() {
	var err error
	var ns *netns.NetNS

	BeforeEach(func() {
		ns, err = netns.NewNetNSBuilder().Build()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		Expect(ns.Cleanup()).To(Succeed())
	})

	DescribeTable("should be redirected to provided port",
		func(randomPort uint16) {
			// given
			dnsServers := getDnsServers("testdata/resolv4.conf", 2, false)
			randomAddressDnsRequest := udp.GenRandomAddressIPv4(consts.DNSPort)
			tproxyConfig, err := config.Config{
				Redirect: config.Redirect{
					DNS: config.DNS{
						Enabled:          true,
						CaptureAll:       false,
						Port:             config.Port(randomPort),
						ResolvConfigPath: "testdata/resolv4.conf",
					},
				},
				RuntimeStdout: io.Discard,
			}.Initialize(context.Background())
			Expect(err).NotTo(HaveOccurred())

			serverAddress := fmt.Sprintf("%s:%d", consts.LocalhostAddress[consts.IPv4].IP, randomPort)

			readyC, errC := udp.UnsafeStartUDPServer(ns, serverAddress, udp.ReplyWithReceivedMsg)
			Consistently(errC).ShouldNot(Receive())
			Eventually(readyC).Should(BeClosed())

			// when
			Eventually(ns.UnsafeExec(func() {
				Expect(builder.RestoreIPTables(context.Background(), tproxyConfig)).Error().To(Succeed())
			})).Should(BeClosed())

			// and
			for _, dnsServer := range dnsServers {
				Eventually(ns.UnsafeExec(func() {
					Expect(udp.DialUDPAddrWithHelloMsgAndGetReply(dnsServer, dnsServer)).
						To(Equal(dnsServer.String()))
				})).Should(BeClosed())
			}

			// and do not redirect any dns request
			Eventually(ns.UnsafeExec(func() {
				Expect(udp.DialUDPAddrWithHelloMsgAndGetReply(randomAddressDnsRequest, randomAddressDnsRequest)).To(Succeed())
			})).ShouldNot(BeClosed())

			// then
			Consistently(errC).ShouldNot(Receive())
		},
		func() []TableEntry {
			var entries []TableEntry
			lockedPorts := []uint16{consts.DNSPort}

			for i := 0; i < blackbox_network_tests.TestCasesAmount; i++ {
				randomPorts := socket.GenerateRandomPortsSlice(1, lockedPorts...)
				// This gives us more entropy as all generated ports will be
				// different from each other
				lockedPorts = append(lockedPorts, randomPorts...)
				desc := fmt.Sprintf("to port %%d, from port %d", consts.DNSPort)
				entry := Entry(EntryDescription(desc), randomPorts[0])
				entries = append(entries, entry)
			}

			return entries
		}(),
	)
})

var _ = Describe("Outbound IPv6 DNS/UDP traffic to port 53 only for addresses in configuration ", func() {
	var err error
	var ns *netns.NetNS

	BeforeEach(func() {
		ns, err = netns.NewNetNSBuilder().WithIPv6(true).Build()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		Expect(ns.Cleanup()).To(Succeed())
	})

	DescribeTable("should be redirected to provided port",
		func(randomPort uint16) {
			// given
			dnsServers := getDnsServers("testdata/resolv6.conf", 2, true)
			randomAddressDnsRequest := udp.GenRandomAddressIPv6(consts.DNSPort)
			tproxyConfig, err := config.Config{
				Redirect: config.Redirect{
					DNS: config.DNS{
						Enabled:          true,
						CaptureAll:       false,
						Port:             config.Port(randomPort),
						ResolvConfigPath: "testdata/resolv6.conf",
					},
				},
				RuntimeStdout: io.Discard,
				IPFamilyMode:  config.IPFamilyModeDualStack,
			}.Initialize(context.Background())
			Expect(err).ToNot(HaveOccurred())

			serverAddress := fmt.Sprintf("%s:%d", consts.LocalhostAddress[consts.IPv6].IP, randomPort)

			readyC, errC := udp.UnsafeStartUDPServer(ns, serverAddress, udp.ReplyWithReceivedMsg)
			Consistently(errC).ShouldNot(Receive())
			Eventually(readyC).Should(BeClosed())

			// when
			Eventually(ns.UnsafeExec(func() {
				Expect(builder.RestoreIPTables(context.Background(), tproxyConfig)).Error().To(Succeed())
			})).Should(BeClosed())

			// and
			for _, dnsServer := range dnsServers {
				Eventually(ns.UnsafeExec(func() {
					Expect(udp.DialUDPAddrWithHelloMsgAndGetReply(dnsServer, dnsServer)).
						To(Equal(dnsServer.String()))
				})).Should(BeClosed())
			}

			// and do not redirect any dns request
			Eventually(ns.UnsafeExec(func() {
				Expect(udp.DialUDPAddrWithHelloMsgAndGetReply(randomAddressDnsRequest, randomAddressDnsRequest)).To(Succeed())
			})).ShouldNot(BeClosed())

			// then
			Consistently(errC).ShouldNot(Receive())
		},
		func() []TableEntry {
			var entries []TableEntry
			lockedPorts := []uint16{consts.DNSPort}

			for i := 0; i < blackbox_network_tests.TestCasesAmount; i++ {
				randomPorts := socket.GenerateRandomPortsSlice(1, lockedPorts...)
				// This gives us more entropy as all generated ports will be
				// different from each other
				lockedPorts = append(lockedPorts, randomPorts...)
				desc := fmt.Sprintf("to port %%d, from port %d", consts.DNSPort)
				entry := Entry(EntryDescription(desc), randomPorts[0])
				entries = append(entries, entry)
			}

			return entries
		}(),
	)
})

var _ = Describe("Outbound IPv4 DNS/UDP conntrack zone splitting with specific IP", func() {
	var err error
	var ns *netns.NetNS

	BeforeEach(func() {
		ns, err = netns.NewNetNSBuilder().
			WithBeforeExecFuncs(sysctl.SetLocalPortRange(32768, 32770)).
			Build()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		Expect(ns.Cleanup()).To(Succeed())
	})

	DescribeTable("should be redirected to provided port",
		func(port uint16) {
			// given
			uid := uintptr(5678)
			dnsServers := getDnsServers("testdata/resolv4-conntrack.conf", 1, false)
			s1Address := fmt.Sprintf("%s:%d", dnsServers[0].IP.String(), consts.DNSPort)
			s2Address := fmt.Sprintf("%s:%d", consts.LocalhostAddress[consts.IPv4].IP, port)
			notRedirected := udp.GenRandomAddressIPv4(consts.DNSPort).AddrPort().String()
			tproxyConfig, err := config.Config{
				Redirect: config.Redirect{
					DNS: config.DNS{
						Enabled:                true,
						Port:                   config.Port(port),
						SkipConntrackZoneSplit: false,
						CaptureAll:             false,
						ResolvConfigPath:       "testdata/resolv4.conf",
					},
				},
				KumaDPUser:    strconv.Itoa(int(uid)),
				RuntimeStdout: io.Discard,
			}.Initialize(context.Background())
			Expect(err).ToNot(HaveOccurred())

			want := map[string]uint{
				s1Address: blackbox_network_tests.DNSConntrackZoneSplittingStressCallsAmount,
				s2Address: blackbox_network_tests.DNSConntrackZoneSplittingStressCallsAmount,
			}

			s1ReadyC, s1ErrC := udp.UnsafeStartUDPServer(
				ns,
				s1Address,
				udp.ReplyWithLocalAddr,
			)
			Consistently(s1ErrC).ShouldNot(Receive())
			Eventually(s1ReadyC).Should(BeClosed())

			s2ReadyC, s2ErrC := udp.UnsafeStartUDPServer(
				ns,
				s2Address,
				udp.ReplyWithLocalAddr,
				sysctl.SetUnprivilegedPortStart(0),
				syscall.SetUID(uid),
			)
			Consistently(s2ErrC).ShouldNot(Receive())
			Eventually(s2ReadyC).Should(BeClosed())

			// when
			Eventually(ns.UnsafeExec(func() {
				Expect(builder.RestoreIPTables(context.Background(), tproxyConfig)).Error().To(Succeed())
			})).Should(BeClosed())

			results := udp.NewResultMap()

			exec1ErrC := ns.UnsafeExecInLoop(
				blackbox_network_tests.DNSConntrackZoneSplittingStressCallsAmount,
				time.Millisecond,
				func() {
					Expect(udp.DialAddrAndIncreaseResultMap(s1Address, results)).To(Succeed())
				},
				syscall.SetUID(uid),
			)

			exec2ErrC := ns.UnsafeExecInLoop(
				blackbox_network_tests.DNSConntrackZoneSplittingStressCallsAmount,
				time.Millisecond,
				func() {
					Expect(udp.DialAddrAndIncreaseResultMap(s1Address, results)).To(Succeed())
				},
			)

			exec3ErrC := ns.UnsafeExecInLoop(
				blackbox_network_tests.DNSConntrackZoneSplittingStressCallsAmount,
				time.Millisecond,
				func() {
					Expect(udp.DialAddrAndIncreaseResultMap(notRedirected, results)).ToNot(Succeed())
				},
			)

			Consistently(exec1ErrC).ShouldNot(Receive())
			Consistently(exec2ErrC).ShouldNot(Receive())
			Consistently(exec3ErrC).ShouldNot(Receive())
			Eventually(exec1ErrC, blackbox_network_tests.DNSConntrackZoneSplittingTestTimeout).
				Should(BeClosed())
			Eventually(exec2ErrC, blackbox_network_tests.DNSConntrackZoneSplittingTestTimeout).
				Should(BeClosed())
			Eventually(exec3ErrC, blackbox_network_tests.DNSConntrackZoneSplittingTestTimeout).
				ShouldNot(BeClosed())

			Expect(results.GetFinalResults()).To(BeEquivalentTo(want))
		},
		func() []TableEntry {
			var entries []TableEntry
			lockedPorts := []uint16{consts.DNSPort}

			for i := 0; i < blackbox_network_tests.TestCasesAmount; i++ {
				ports := socket.GenerateRandomPortsSlice(1, lockedPorts...)
				// This gives us more entropy as all generated ports will be
				// different from each other
				lockedPorts = append(lockedPorts, ports...)
				desc := fmt.Sprintf("to port %%d, from port %d", consts.DNSPort)
				entry := Entry(EntryDescription(desc), ports[0])
				entries = append(entries, entry)
			}

			return entries
		}(),
	)
})

var _ = Describe("Outbound IPv4 DNS/UDP traffic to port 53 from specific input interface", func() {
	var err error
	var ns *netns.NetNS
	var ns2 *netns.NetNS

	BeforeEach(func() {
		mainLink, peerLink, linkErr := netns.NewLinkPair()
		Expect(linkErr).ToNot(HaveOccurred())

		ns1Address, addrErr := netlink.ParseAddr("192.168.0.1/24")
		Expect(addrErr).ToNot(HaveOccurred())
		ns, err = netns.NewNetNSBuilder().WithSharedLink(mainLink, ns1Address).Build()
		Expect(err).ToNot(HaveOccurred())

		ns2Address, addrErr := netlink.ParseAddr("192.168.0.2/24")
		Expect(addrErr).ToNot(HaveOccurred())
		ns2, err = netns.NewNetNSBuilder().WithSharedLink(peerLink, ns2Address).Build()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		Expect(ns.Cleanup()).To(Succeed())
		Expect(ns2.Cleanup()).To(Succeed())
	})

	DescribeTable("should be redirected to provided port",
		func(randomPort uint16) {
			// given
			address := udp.GenRandomAddressIPv4(consts.DNSPort)
			tproxyConfig, err := config.Config{
				Redirect: config.Redirect{
					DNS: config.DNS{
						Enabled:    true,
						Port:       config.Port(randomPort),
						CaptureAll: true,
					},
					Inbound: config.TrafficFlow{
						Enabled: true,
					},
					Outbound: config.TrafficFlow{
						Enabled: true,
					},
					// interface name and its network
					VNet: config.VNet{Networks: []string{"s-peer+:192.168.0.2/16"}},
				},
				RuntimeStdout: io.Discard,
			}.Initialize(context.Background())
			Expect(err).ToNot(HaveOccurred())

			serverAddress := fmt.Sprintf(":%d", randomPort)
			readyC, errC := udp.UnsafeStartUDPServer(ns2, serverAddress, udp.ReplyWithReceivedMsg)
			Consistently(errC).ShouldNot(Receive())
			Eventually(readyC).Should(BeClosed())

			// when
			Eventually(ns2.UnsafeExec(func() {
				Expect(builder.RestoreIPTables(context.Background(), tproxyConfig)).Error().To(Succeed())
			})).Should(BeClosed())

			// and
			Eventually(ns.UnsafeExec(func() {
				Expect(udp.DialUDPAddrWithHelloMsgAndGetReply(&net.UDPAddr{
					IP:   ns2.SharedLinkAddress().IP,
					Port: int(consts.DNSPort),
				}, address)).
					To(Equal(address.String()))
			})).Should(BeClosed())

			// then
			Consistently(errC).ShouldNot(Receive())
		},
		func() []TableEntry {
			var entries []TableEntry
			lockedPorts := []uint16{consts.DNSPort}

			for i := 0; i < blackbox_network_tests.TestCasesAmount; i++ {
				randomPorts := socket.GenerateRandomPortsSlice(1, lockedPorts...)
				// This gives us more entropy as all generated ports will be
				// different from each other
				lockedPorts = append(lockedPorts, randomPorts...)
				desc := fmt.Sprintf("to port %%d, from port %d", consts.DNSPort)
				entry := Entry(EntryDescription(desc), randomPorts[0])
				entries = append(entries, entry)
			}

			return entries
		}(),
	)
})

var _ = Describe("Outbound IPv6 DNS/UDP traffic to port 53 from specific input interface", func() {
	var err error
	var ns *netns.NetNS
	var ns2 *netns.NetNS

	BeforeEach(func() {
		mainLink, peerLink, linkErr := netns.NewLinkPair()
		Expect(linkErr).ToNot(HaveOccurred())

		ns1Address, addrErr := netlink.ParseAddr("fd00::10:1:1/64")
		Expect(addrErr).ToNot(HaveOccurred())
		ns, err = netns.NewNetNSBuilder().WithIPv6(true).WithSharedLink(mainLink, ns1Address).Build()
		Expect(err).ToNot(HaveOccurred())

		ns2Address, addrErr := netlink.ParseAddr("fd00::10:1:2/64")
		Expect(addrErr).ToNot(HaveOccurred())
		ns2, err = netns.NewNetNSBuilder().WithIPv6(true).WithSharedLink(peerLink, ns2Address).Build()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		Expect(ns.Cleanup()).To(Succeed())
		Expect(ns2.Cleanup()).To(Succeed())
	})

	DescribeTable("should be redirected to provided port",
		func(randomPort uint16) {
			// given
			address := udp.GenRandomAddressIPv6(consts.DNSPort)
			tproxyConfig, err := config.Config{
				Redirect: config.Redirect{
					DNS: config.DNS{
						Enabled:    true,
						Port:       config.Port(randomPort),
						CaptureAll: true,
					},
					Inbound: config.TrafficFlow{
						Enabled: true,
					},
					Outbound: config.TrafficFlow{
						Enabled: true,
					},
					// interface name and its network
					VNet: config.VNet{Networks: []string{"s-peer+:fd00::10:1:2/64"}},
				},
				IPFamilyMode:  config.IPFamilyModeDualStack,
				RuntimeStdout: io.Discard,
			}.Initialize(context.Background())
			Expect(err).ToNot(HaveOccurred())

			serverAddress := fmt.Sprintf(":%d", randomPort)
			readyC, errC := udp.UnsafeStartUDPServer(ns2, serverAddress, udp.ReplyWithReceivedMsg)
			Consistently(errC).ShouldNot(Receive())
			Eventually(readyC).Should(BeClosed())

			// when
			Eventually(ns2.UnsafeExec(func() {
				Expect(builder.RestoreIPTables(context.Background(), tproxyConfig)).Error().To(Succeed())
			})).Should(BeClosed())

			// and
			Eventually(ns.UnsafeExec(func() {
				Expect(udp.DialUDPAddrWithHelloMsgAndGetReply(&net.UDPAddr{
					IP:   ns2.SharedLinkAddress().IP,
					Port: int(consts.DNSPort),
				}, address)).
					To(Equal(address.String()))
			})).Should(BeClosed())

			// then
			Consistently(errC).ShouldNot(Receive())
		},
		func() []TableEntry {
			var entries []TableEntry
			lockedPorts := []uint16{consts.DNSPort}

			for i := 0; i < blackbox_network_tests.TestCasesAmount; i++ {
				randomPorts := socket.GenerateRandomPortsSlice(1, lockedPorts...)
				// This gives us more entropy as all generated ports will be
				// different from each other
				lockedPorts = append(lockedPorts, randomPorts...)
				desc := fmt.Sprintf("to port %%d, from port %d", consts.DNSPort)
				entry := Entry(EntryDescription(desc), randomPorts[0])
				entries = append(entries, entry)
			}

			return entries
		}(),
	)
})

func getDnsServers(configPath string, expectedServers int, isIpv6 bool) []*net.UDPAddr {
	var dnsServers []*net.UDPAddr
	configPath, err := filepath.Abs(configPath)
	Expect(err).ToNot(HaveOccurred())

	dnsConfig, err := dns.ClientConfigFromFile(configPath)
	Expect(err).ToNot(HaveOccurred())

	var ipv4 []string
	var ipv6 []string

	for _, address := range dnsConfig.Servers {
		parsed := net.ParseIP(address)
		if parsed.To4() != nil {
			ipv4 = append(ipv4, address)
		} else {
			ipv6 = append(ipv6, address)
		}
	}

	dnsAddresses := ipv4
	if isIpv6 {
		dnsAddresses = ipv6
	}
	Expect(dnsAddresses).To(HaveLen(expectedServers))
	for _, dnsServer := range dnsAddresses {
		dnsServers = append(dnsServers, &net.UDPAddr{
			IP:   net.ParseIP(dnsServer),
			Port: int(consts.DNSPort),
		})
	}
	return dnsServers
}
