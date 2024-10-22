package blackbox_network_tests_test

import (
	"context"
	"fmt"
	"io"
	"net"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/vishvananda/netlink"

	"github.com/kumahq/kuma/pkg/transparentproxy/config"
	"github.com/kumahq/kuma/pkg/transparentproxy/consts"
	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/builder"
	"github.com/kumahq/kuma/test/blackbox_network_tests"
	"github.com/kumahq/kuma/test/framework/network/ip"
	"github.com/kumahq/kuma/test/framework/network/netns"
	"github.com/kumahq/kuma/test/framework/network/socket"
	"github.com/kumahq/kuma/test/framework/network/syscall"
	"github.com/kumahq/kuma/test/framework/network/tcp"
)

var _ = Describe("Outbound IPv4 TCP traffic to any address:port", func() {
	var err error
	var ns *netns.NetNS

	BeforeEach(func() {
		ns, err = netns.NewNetNSBuilder().Build()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		Expect(ns.Cleanup()).To(Succeed())
	})

	DescribeTable("should be redirected to outbound port",
		func(serverPort, randomPort uint16) {
			// given
			address := fmt.Sprintf(":%d", serverPort)
			tproxyConfig, err := config.Config{
				Redirect: config.Redirect{
					Inbound: config.TrafficFlow{
						Enabled: true,
					},
					Outbound: config.TrafficFlow{
						Enabled: true,
						Port:    config.Port(serverPort),
					},
				},
				RuntimeStdout: io.Discard,
			}.Initialize(context.Background())
			Expect(err).ToNot(HaveOccurred())

			tcpReadyC, tcpErrC := tcp.UnsafeStartTCPServer(
				ns,
				address,
				tcp.ReplyWithOriginalDstIPv4,
				tcp.CloseConn,
			)
			Eventually(tcpReadyC).Should(BeClosed())
			Consistently(tcpErrC).ShouldNot(Receive())

			// when
			Eventually(ns.UnsafeExec(func() {
				Expect(builder.RestoreIPTables(context.Background(), tproxyConfig)).Error().To(Succeed())
			})).Should(BeClosed())

			// then
			Eventually(ns.UnsafeExec(func() {
				address := ip.GenRandomIPv4()

				Expect(tcp.DialIPWithPortAndGetReply(address, randomPort)).
					To(Equal(fmt.Sprintf("%s:%d", address, randomPort)))
			})).Should(BeClosed())

			// then
			Eventually(tcpErrC).Should(BeClosed())
		},
		func() []TableEntry {
			var entries []TableEntry
			var lockedPorts []uint16

			for i := 0; i < blackbox_network_tests.TestCasesAmount; i++ {
				randomPorts := socket.GenerateRandomPortsSlice(2, lockedPorts...)
				// This gives us more entropy as all generated ports will be
				// different from each other
				lockedPorts = append(lockedPorts, randomPorts...)
				desc := fmt.Sprintf("to port %%d, from port %%d")
				entry := Entry(
					EntryDescription(desc),
					randomPorts[0],
					randomPorts[1],
				)
				entries = append(entries, entry)
			}

			return entries
		}(),
	)
})

var _ = Describe("Outbound IPv6 TCP traffic to any address:port", func() {
	var err error
	var ns *netns.NetNS

	BeforeEach(func() {
		ns, err = netns.NewNetNSBuilder().WithIPv6(true).Build()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		Expect(ns.Cleanup()).To(Succeed())
	})

	DescribeTable("should be redirected to outbound port",
		func(serverPort, randomPort uint16) {
			// given
			address := fmt.Sprintf(":%d", serverPort)
			tproxyConfig, err := config.Config{
				Redirect: config.Redirect{
					Outbound: config.TrafficFlow{
						Enabled: true,
						Port:    config.Port(serverPort),
					},
					Inbound: config.TrafficFlow{
						Enabled: true,
					},
				},
				IPFamilyMode:  config.IPFamilyModeDualStack,
				RuntimeStdout: io.Discard,
			}.Initialize(context.Background())
			Expect(err).ToNot(HaveOccurred())

			tcpReadyC, tcpErrC := tcp.UnsafeStartTCPServer(
				ns,
				address,
				tcp.ReplyWithOriginalDstIPv6,
				tcp.CloseConn,
			)
			Eventually(tcpReadyC).Should(BeClosed())
			Consistently(tcpErrC).ShouldNot(Receive())

			// when
			Eventually(ns.UnsafeExec(func() {
				Expect(builder.RestoreIPTables(context.Background(), tproxyConfig)).Error().To(Succeed())
			})).Should(BeClosed())

			// then
			Eventually(ns.UnsafeExec(func() {
				address := ip.GenRandomIPv6()

				Expect(tcp.DialIPWithPortAndGetReply(address, randomPort)).
					To(Equal(fmt.Sprintf("[%s]:%d", address, randomPort)))
			})).Should(BeClosed())

			// then
			Eventually(tcpErrC).Should(BeClosed())
		},
		func() []TableEntry {
			var entries []TableEntry
			var lockedPorts []uint16

			for i := 0; i < blackbox_network_tests.TestCasesAmount; i++ {
				randomPorts := socket.GenerateRandomPortsSlice(2, lockedPorts...)
				// This gives us more entropy as all generated ports will be
				// different from each other
				lockedPorts = append(lockedPorts, randomPorts...)
				desc := fmt.Sprintf("to port %%d, from port %%d")
				entry := Entry(
					EntryDescription(desc),
					randomPorts[0],
					randomPorts[1],
				)
				entries = append(entries, entry)
			}

			return entries
		}(),
	)
})

var _ = Describe("Outbound IPv4 TCP traffic to any address:port except excluded ones", func() {
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

	DescribeTable("should be redirected to outbound port",
		func(serverPort, randomPort, excludedPort uint16) {
			// given

			tproxyConfig, err := config.Config{
				Redirect: config.Redirect{
					Outbound: config.TrafficFlow{
						Enabled:      true,
						Port:         config.Port(serverPort),
						ExcludePorts: config.Ports{config.Port(excludedPort)},
					},
					Inbound: config.TrafficFlow{
						Enabled: true,
					},
				},
				RuntimeStdout: io.Discard,
				Log: config.Log{
					Enabled: true,
				},
			}.Initialize(context.Background())
			Expect(err).ToNot(HaveOccurred())

			tcpReadyC, tcpErrC := tcp.UnsafeStartTCPServer(
				ns,
				fmt.Sprintf(":%d", serverPort),
				tcp.ReplyWithOriginalDstIPv4,
				tcp.CloseConn,
			)
			Eventually(tcpReadyC).Should(BeClosed())
			Consistently(tcpErrC).ShouldNot(Receive())

			excludedReadyC, excludedErrC := tcp.UnsafeStartTCPServer(
				ns2,
				fmt.Sprintf(":%d", excludedPort),
				tcp.ReplyWith("excluded"),
				tcp.CloseConn,
			)
			Eventually(excludedReadyC).Should(BeClosed())
			Consistently(excludedErrC).ShouldNot(Receive())

			// when
			Eventually(ns.UnsafeExec(func() {
				Expect(builder.RestoreIPTables(context.Background(), tproxyConfig)).Error().To(Succeed())
			})).Should(BeClosed())

			// then
			Eventually(ns.UnsafeExec(func() {
				address := ip.GenRandomIPv4()

				Expect(tcp.DialIPWithPortAndGetReply(address, randomPort)).
					To(Equal(fmt.Sprintf("%s:%d", address, randomPort)))
			})).Should(BeClosed())

			// then
			Eventually(ns.UnsafeExec(func() {
				Expect(tcp.DialIPWithPortAndGetReply(ns2.SharedLinkAddress().IP, excludedPort)).
					To(Equal("excluded"))
			})).Should(BeClosed())

			// then
			Eventually(tcpErrC).Should(BeClosed())
			Eventually(excludedErrC).Should(BeClosed())
		},
		func() []TableEntry {
			var entries []TableEntry
			var lockedPorts []uint16

			for i := 0; i < blackbox_network_tests.TestCasesAmount; i++ {
				randomPorts := socket.GenerateRandomPortsSlice(3, lockedPorts...)
				// This gives us more entropy as all generated ports will be
				// different from each other
				lockedPorts = append(lockedPorts, randomPorts...)
				desc := fmt.Sprintf("to port %%d, from port %%d (excluded: %%d)")
				entry := Entry(
					EntryDescription(desc),
					randomPorts[0],
					randomPorts[1],
					randomPorts[2],
				)
				entries = append(entries, entry)
			}

			return entries
		}(),
	)
})

var _ = Describe("Outbound IPv4 TCP traffic to any address:port except ports excluded by uid ones", func() {
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

	DescribeTable("should be redirected to outbound port",
		func(serverPort, excludedPort uint16) {
			// given
			dnsUserUid := uintptr(4201) // see /.github/workflows/blackbox-tests.yaml:76

			tproxyConfig, err := config.Config{
				Redirect: config.Redirect{
					Outbound: config.TrafficFlow{
						Enabled: true,
						Port:    config.Port(serverPort),
						ExcludePortsForUIDs: []string{
							fmt.Sprintf("tcp:%d:%d", excludedPort, dnsUserUid),
						},
					},
					Inbound: config.TrafficFlow{
						Enabled: true,
					},
				},
				RuntimeStdout: io.Discard,
			}.Initialize(context.Background())
			Expect(err).ToNot(HaveOccurred())

			tcpReadyC, tcpErrC := tcp.UnsafeStartTCPServer(
				ns,
				fmt.Sprintf(":%d", serverPort),
				tcp.ReplyWithOriginalDstIPv4,
				tcp.CloseConn,
			)
			Eventually(tcpReadyC).Should(BeClosed())
			Consistently(tcpErrC).ShouldNot(Receive())

			excludedReadyC, excludedErrC := tcp.UnsafeStartTCPServer(
				ns2,
				fmt.Sprintf(":%d", excludedPort),
				tcp.ReplyWith("excluded"),
				tcp.CloseConn,
			)
			Eventually(excludedReadyC).Should(BeClosed())
			Consistently(excludedErrC).ShouldNot(Receive())

			// when
			Eventually(ns.UnsafeExec(func() {
				Expect(builder.RestoreIPTables(context.Background(), tproxyConfig)).Error().To(Succeed())
			})).Should(BeClosed())

			// then
			Eventually(ns.UnsafeExec(func() {
				address := ip.GenRandomIPv4()

				Expect(tcp.DialIPWithPortAndGetReply(address, excludedPort)).
					To(Equal(fmt.Sprintf("%s:%d", address, excludedPort)))
			})).Should(BeClosed())

			// then
			Eventually(ns.UnsafeExecInLoop(1, 0, func() {
				Expect(tcp.DialIPWithPortAndGetReply(ns2.SharedLinkAddress().IP, excludedPort)).
					To(Equal("excluded"))
			}, syscall.SetUID(dnsUserUid))).Should(BeClosed())

			// then
			Eventually(tcpErrC).Should(BeClosed())
			Eventually(excludedErrC).Should(BeClosed())
		},
		func() []TableEntry {
			var entries []TableEntry
			var lockedPorts []uint16

			for i := 0; i < blackbox_network_tests.TestCasesAmount; i++ {
				randomPorts := socket.GenerateRandomPortsSlice(2, lockedPorts...)
				// This gives us more entropy as all generated ports will be
				// different from each other
				lockedPorts = append(lockedPorts, randomPorts...)
				desc := fmt.Sprintf("to port %%d, excluded: %%d (by uid only)")
				entry := Entry(
					EntryDescription(desc),
					randomPorts[0],
					randomPorts[1],
				)
				entries = append(entries, entry)
			}

			return entries
		}(),
	)
})

var _ = Describe("Outbound IPv6 TCP traffic to any address:port except excluded ones", func() {
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

	DescribeTable("should be redirected to outbound port",
		func(serverPort, randomPort, excludedPort uint16) {
			// given
			tproxyConfig, err := config.Config{
				Redirect: config.Redirect{
					Outbound: config.TrafficFlow{
						Enabled:      true,
						Port:         config.Port(serverPort),
						ExcludePorts: config.Ports{config.Port(excludedPort)},
					},
					Inbound: config.TrafficFlow{
						Enabled: true,
					},
				},
				IPFamilyMode:  config.IPFamilyModeDualStack,
				RuntimeStdout: io.Discard,
				Log: config.Log{
					Enabled: true,
				},
			}.Initialize(context.Background())
			Expect(err).ToNot(HaveOccurred())

			tcpReadyC, tcpErrC := tcp.UnsafeStartTCPServer(
				ns,
				fmt.Sprintf(":%d", serverPort),
				tcp.ReplyWithOriginalDstIPv6,
				tcp.CloseConn,
			)
			Eventually(tcpReadyC).Should(BeClosed())
			Consistently(tcpErrC).ShouldNot(Receive())

			excludedReadyC, excludedErrC := tcp.UnsafeStartTCPServer(
				ns2,
				fmt.Sprintf(":%d", excludedPort),
				tcp.ReplyWith("excluded"),
				tcp.CloseConn,
			)
			Eventually(excludedReadyC).Should(BeClosed())
			Consistently(excludedErrC).ShouldNot(Receive())

			// when
			Eventually(ns.UnsafeExec(func() {
				Expect(builder.RestoreIPTables(context.Background(), tproxyConfig)).Error().To(Succeed())
			})).Should(BeClosed())

			// then
			Eventually(ns.UnsafeExec(func() {
				address := ip.GenRandomIPv6()

				Expect(tcp.DialIPWithPortAndGetReply(address, randomPort)).
					To(Equal(fmt.Sprintf("[%s]:%d", address, randomPort)))
			})).Should(BeClosed())

			// then
			Eventually(ns.UnsafeExec(func() {
				Expect(tcp.DialIPWithPortAndGetReply(ns2.SharedLinkAddress().IP, excludedPort)).
					To(Equal("excluded"))
			})).Should(BeClosed())

			// then
			Eventually(tcpErrC).Should(BeClosed())
			Eventually(excludedErrC).Should(BeClosed())
		},
		func() []TableEntry {
			var entries []TableEntry
			var lockedPorts []uint16

			for i := 0; i < blackbox_network_tests.TestCasesAmount; i++ {
				randomPorts := socket.GenerateRandomPortsSlice(3, lockedPorts...)
				// This gives us more entropy as all generated ports will be
				// different from each other
				lockedPorts = append(lockedPorts, randomPorts...)
				desc := fmt.Sprintf("to port %%d, from port %%d (excluded: %%d)")
				entry := Entry(
					EntryDescription(desc),
					randomPorts[0],
					randomPorts[1],
					randomPorts[2],
				)
				entries = append(entries, entry)
			}

			return entries
		}(),
	)
})

var _ = Describe("Outbound IPv4 TCP traffic only to included port", func() {
	var err error
	var ns *netns.NetNS

	BeforeEach(func() {
		ns, err = netns.NewNetNSBuilder().Build()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		Expect(ns.Cleanup()).To(Succeed())
	})

	DescribeTable("should be redirected to outbound port",
		func(serverPort, includedPort, randomPort uint16) {
			// given

			tproxyConfig, err := config.Config{
				Redirect: config.Redirect{
					Outbound: config.TrafficFlow{
						Enabled:      true,
						Port:         config.Port(serverPort),
						IncludePorts: config.Ports{config.Port(includedPort)},
						ExcludePorts: config.Ports{config.Port(includedPort)},
					},
					Inbound: config.TrafficFlow{
						Enabled: true,
					},
				},
				RuntimeStdout: io.Discard,
			}.Initialize(context.Background())
			Expect(err).ToNot(HaveOccurred())

			tcpReadyC, tcpErrC := tcp.UnsafeStartTCPServer(
				ns,
				fmt.Sprintf(":%d", serverPort),
				tcp.ReplyWithOriginalDstIPv4,
				tcp.CloseConn,
			)
			Eventually(tcpReadyC).Should(BeClosed())
			Consistently(tcpErrC).ShouldNot(Receive())

			randomReadyC, randomErrC := tcp.UnsafeStartTCPServer(
				ns,
				fmt.Sprintf(":%d", randomPort),
				tcp.ReplyWith("random"),
				tcp.CloseConn,
			)
			Eventually(randomReadyC).Should(BeClosed())
			Consistently(randomErrC).ShouldNot(Receive())

			// when
			Eventually(ns.UnsafeExec(func() {
				Expect(builder.RestoreIPTables(context.Background(), tproxyConfig)).Error().To(Succeed())
			})).Should(BeClosed())

			// then
			Eventually(ns.UnsafeExec(func() {
				address := ip.GenRandomIPv4()

				Expect(tcp.DialIPWithPortAndGetReply(address, includedPort)).
					To(Equal(fmt.Sprintf("%s:%d", address, includedPort)))
			})).Should(BeClosed())

			// then
			Eventually(ns.UnsafeExec(func() {
				Expect(tcp.DialIPWithPortAndGetReply(net.IPv4zero, randomPort)).
					To(Equal("random"))
			})).Should(BeClosed())

			// then
			Eventually(tcpErrC).Should(BeClosed())
			Eventually(randomErrC).Should(BeClosed())
		},
		func() []TableEntry {
			var entries []TableEntry
			var lockedPorts []uint16

			for i := 0; i < blackbox_network_tests.TestCasesAmount; i++ {
				randomPorts := socket.GenerateRandomPortsSlice(3, lockedPorts...)
				// This gives us more entropy as all generated ports will be
				// different from each other
				lockedPorts = append(lockedPorts, randomPorts...)
				desc := fmt.Sprintf("to port %%d, from port %%d (random: %%d)")
				entry := Entry(
					EntryDescription(desc),
					randomPorts[0],
					randomPorts[1],
					randomPorts[2],
				)
				entries = append(entries, entry)
			}

			return entries
		}(),
	)
})

var _ = Describe("Outbound IPv6 TCP traffic only to included port", func() {
	var err error
	var ns *netns.NetNS

	BeforeEach(func() {
		ns, err = netns.NewNetNSBuilder().WithIPv6(true).Build()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		Expect(ns.Cleanup()).To(Succeed())
	})

	DescribeTable("should be redirected to outbound port",
		func(serverPort, includedPort, randomPort uint16) {
			// given
			tproxyConfig, err := config.Config{
				Redirect: config.Redirect{
					Outbound: config.TrafficFlow{
						Enabled:      true,
						Port:         config.Port(serverPort),
						IncludePorts: config.Ports{config.Port(includedPort)},
						ExcludePorts: config.Ports{config.Port(includedPort)},
					},
					Inbound: config.TrafficFlow{
						Enabled: true,
					},
				},
				IPFamilyMode:  config.IPFamilyModeDualStack,
				RuntimeStdout: io.Discard,
			}.Initialize(context.Background())
			Expect(err).ToNot(HaveOccurred())

			tcpReadyC, tcpErrC := tcp.UnsafeStartTCPServer(
				ns,
				fmt.Sprintf(":%d", serverPort),
				tcp.ReplyWithOriginalDstIPv6,
				tcp.CloseConn,
			)
			Eventually(tcpReadyC).Should(BeClosed())
			Consistently(tcpErrC).ShouldNot(Receive())

			randomReadyC, randomErrC := tcp.UnsafeStartTCPServer(
				ns,
				fmt.Sprintf(":%d", randomPort),
				tcp.ReplyWith("random"),
				tcp.CloseConn,
			)
			Eventually(randomReadyC).Should(BeClosed())
			Consistently(randomErrC).ShouldNot(Receive())

			// when
			Eventually(ns.UnsafeExec(func() {
				Expect(builder.RestoreIPTables(context.Background(), tproxyConfig)).Error().To(Succeed())
			})).Should(BeClosed())

			// then
			Eventually(ns.UnsafeExec(func() {
				address := ip.GenRandomIPv6()

				Expect(tcp.DialIPWithPortAndGetReply(address, includedPort)).
					To(Equal(fmt.Sprintf("[%s]:%d", address, includedPort)))
			})).Should(BeClosed())

			// then
			Eventually(ns.UnsafeExec(func() {
				Expect(tcp.DialIPWithPortAndGetReply(net.IPv6zero, randomPort)).
					To(Equal("random"))
			})).Should(BeClosed())

			// then
			Eventually(tcpErrC).Should(BeClosed())
			Eventually(randomErrC).Should(BeClosed())
		},
		func() []TableEntry {
			var entries []TableEntry
			var lockedPorts []uint16

			for i := 0; i < blackbox_network_tests.TestCasesAmount; i++ {
				randomPorts := socket.GenerateRandomPortsSlice(3, lockedPorts...)
				// This gives us more entropy as all generated ports will be
				// different from each other
				lockedPorts = append(lockedPorts, randomPorts...)
				desc := fmt.Sprintf("to port %%d, from port %%d (random: %%d)")
				entry := Entry(
					EntryDescription(desc),
					randomPorts[0],
					randomPorts[1],
					randomPorts[2],
				)
				entries = append(entries, entry)
			}

			return entries
		}(),
	)
})

var _ = Describe("Outbound IPv4 TCP traffic to any address:port", func() {
	var err error
	var ns *netns.NetNS

	BeforeEach(func() {
		ns, err = netns.NewNetNSBuilder().Build()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		Expect(ns.Cleanup()).To(Succeed())
	})

	DescribeTable("should not be redirected to outbound port",
		func(serverPort, randomPort uint16) {
			// given
			address := fmt.Sprintf(":%d", randomPort)
			tproxyConfig, err := config.Config{
				Redirect: config.Redirect{
					Outbound: config.TrafficFlow{
						Enabled: false,
						Port:    config.Port(serverPort),
					},
					Inbound: config.TrafficFlow{
						Enabled: true,
					},
				},
				RuntimeStdout: io.Discard,
			}.Initialize(context.Background())
			Expect(err).ToNot(HaveOccurred())

			tcpReadyC, tcpErrC := tcp.UnsafeStartTCPServer(
				ns,
				address,
				tcp.ReplyWith("randomPort"),
				tcp.CloseConn,
			)
			Eventually(tcpReadyC).Should(BeClosed())
			Consistently(tcpErrC).ShouldNot(Receive())

			// when
			Eventually(ns.UnsafeExec(func() {
				Expect(builder.RestoreIPTables(context.Background(), tproxyConfig)).Error().To(Succeed())
			})).Should(BeClosed())

			// then
			Eventually(ns.UnsafeExec(func() {
				Expect(tcp.DialIPWithPortAndGetReply(consts.LocalhostAddress[consts.IPv4].IP, randomPort)).
					To(Equal("randomPort"))
			})).Should(BeClosed())

			// then
			Eventually(tcpErrC).Should(BeClosed())
		},
		func() []TableEntry {
			var entries []TableEntry
			var lockedPorts []uint16

			for i := 0; i < blackbox_network_tests.TestCasesAmount; i++ {
				randomPorts := socket.GenerateRandomPortsSlice(2, lockedPorts...)
				// This gives us more entropy as all generated ports will be
				// different from each other
				lockedPorts = append(lockedPorts, randomPorts...)
				desc := fmt.Sprintf("to port %%d, from port %%d")
				entry := Entry(
					EntryDescription(desc),
					randomPorts[0],
					randomPorts[1],
				)
				entries = append(entries, entry)
			}

			return entries
		}(),
	)
})

var _ = Describe("Outbound IPv6 TCP traffic to any address:port", func() {
	var err error
	var ns *netns.NetNS

	BeforeEach(func() {
		ns, err = netns.NewNetNSBuilder().WithIPv6(true).Build()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		Expect(ns.Cleanup()).To(Succeed())
	})

	DescribeTable("should not be redirected to outbound port",
		func(serverPort, randomPort uint16) {
			// given
			address := fmt.Sprintf(":%d", randomPort)
			tproxyConfig, err := config.Config{
				Redirect: config.Redirect{
					Outbound: config.TrafficFlow{
						Enabled: false,
						Port:    config.Port(serverPort),
					},
					Inbound: config.TrafficFlow{
						Enabled: true,
					},
				},
				IPFamilyMode:  config.IPFamilyModeDualStack,
				RuntimeStdout: io.Discard,
			}.Initialize(context.Background())
			Expect(err).ToNot(HaveOccurred())

			tcpReadyC, tcpErrC := tcp.UnsafeStartTCPServer(
				ns,
				address,
				tcp.ReplyWith("randomPort"),
				tcp.CloseConn,
			)
			Eventually(tcpReadyC).Should(BeClosed())
			Consistently(tcpErrC).ShouldNot(Receive())

			// when
			Eventually(ns.UnsafeExec(func() {
				Expect(builder.RestoreIPTables(context.Background(), tproxyConfig)).Error().To(Succeed())
			})).Should(BeClosed())

			// then
			Eventually(ns.UnsafeExec(func() {
				Expect(tcp.DialIPWithPortAndGetReply(consts.LocalhostAddress[consts.IPv6].IP, randomPort)).
					To(Equal("randomPort"))
			})).Should(BeClosed())

			// then
			Eventually(tcpErrC).Should(BeClosed())
		},
		func() []TableEntry {
			var entries []TableEntry
			var lockedPorts []uint16

			for i := 0; i < blackbox_network_tests.TestCasesAmount; i++ {
				randomPorts := socket.GenerateRandomPortsSlice(2, lockedPorts...)
				// This gives us more entropy as all generated ports will be
				// different from each other
				lockedPorts = append(lockedPorts, randomPorts...)
				desc := fmt.Sprintf("to port %%d, from port %%d")
				entry := Entry(
					EntryDescription(desc),
					randomPorts[0],
					randomPorts[1],
				)
				entries = append(entries, entry)
			}

			return entries
		}(),
	)
})

var _ = Describe("Outbound IPv6 TCP traffic to any address:port except ports excluded by uid ones", func() {
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

	DescribeTable("should be redirected to outbound port",
		func(serverPort, excludedPort uint16) {
			// given
			dnsUserUid := uintptr(4201) // see /.github/workflows/blackbox-tests.yaml:76

			tproxyConfig, err := config.Config{
				Redirect: config.Redirect{
					Outbound: config.TrafficFlow{
						Enabled: true,
						Port:    config.Port(serverPort),
						ExcludePortsForUIDs: []string{
							fmt.Sprintf("tcp:%d:%d", excludedPort, dnsUserUid),
						},
					},
					Inbound: config.TrafficFlow{
						Enabled: true,
					},
				},
				IPFamilyMode:  config.IPFamilyModeDualStack,
				RuntimeStdout: io.Discard,
			}.Initialize(context.Background())
			Expect(err).ToNot(HaveOccurred())

			tcpReadyC, tcpErrC := tcp.UnsafeStartTCPServer(
				ns,
				fmt.Sprintf(":%d", serverPort),
				tcp.ReplyWithOriginalDstIPv6,
				tcp.CloseConn,
			)
			Eventually(tcpReadyC).Should(BeClosed())
			Consistently(tcpErrC).ShouldNot(Receive())

			excludedReadyC, excludedErrC := tcp.UnsafeStartTCPServer(
				ns2,
				fmt.Sprintf(":%d", excludedPort),
				tcp.ReplyWith("excluded"),
				tcp.CloseConn,
			)
			Eventually(excludedReadyC).Should(BeClosed())
			Consistently(excludedErrC).ShouldNot(Receive())

			// when
			Eventually(ns.UnsafeExec(func() {
				Expect(builder.RestoreIPTables(context.Background(), tproxyConfig)).Error().To(Succeed())
			})).Should(BeClosed())

			// then
			Eventually(ns.UnsafeExec(func() {
				address := ip.GenRandomIPv6()

				Expect(tcp.DialIPWithPortAndGetReply(address, excludedPort)).
					To(Equal(fmt.Sprintf("[%s]:%d", address, excludedPort)))
			})).Should(BeClosed())

			// then
			Eventually(ns.UnsafeExecInLoop(1, 0, func() {
				Expect(tcp.DialIPWithPortAndGetReply(ns2.SharedLinkAddress().IP, excludedPort)).
					To(Equal("excluded"))
			}, syscall.SetUID(dnsUserUid))).Should(BeClosed())

			// then
			Eventually(tcpErrC).Should(BeClosed())
			Eventually(excludedErrC).Should(BeClosed())
		},
		func() []TableEntry {
			var entries []TableEntry
			var lockedPorts []uint16

			for i := 0; i < blackbox_network_tests.TestCasesAmount; i++ {
				randomPorts := socket.GenerateRandomPortsSlice(2, lockedPorts...)
				// This gives us more entropy as all generated ports will be
				// different from each other
				lockedPorts = append(lockedPorts, randomPorts...)
				desc := fmt.Sprintf("to port %%d, excluded: %%d (by uid only)")
				entry := Entry(
					EntryDescription(desc),
					randomPorts[0],
					randomPorts[1],
				)
				entries = append(entries, entry)
			}

			return entries
		}(),
	)
})

var _ = Describe("Outbound IPv4 TCP traffic from specific interface to other ip than excluded", func() {
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

	DescribeTable("should be redirected to outbound port",
		func(serverPort, randomPort uint16) {
			// given
			tproxyConfig, err := config.Config{
				Redirect: config.Redirect{
					Outbound: config.TrafficFlow{
						Enabled: true,
						Port:    config.Port(serverPort),
					},
					Inbound: config.TrafficFlow{
						Enabled: true,
					},
					VNet: config.VNet{
						Networks: []string{"s-peer+:192.168.0.1/32"},
					},
				},
				RuntimeStdout: io.Discard,
			}.Initialize(context.Background())
			Expect(err).ToNot(HaveOccurred())

			tcpReadyC, tcpErrC := tcp.UnsafeStartTCPServer(
				ns2,
				fmt.Sprintf(":%d", serverPort),
				tcp.ReplyWithOriginalDstIPv4,
				tcp.CloseConn,
			)
			Eventually(tcpReadyC).Should(BeClosed())
			Consistently(tcpErrC).ShouldNot(Receive())

			// when
			Eventually(ns2.UnsafeExec(func() {
				Expect(builder.RestoreIPTables(context.Background(), tproxyConfig)).Error().To(Succeed())
			})).Should(BeClosed())

			// then
			Eventually(ns.UnsafeExec(func() {
				Expect(tcp.DialIPWithPortAndGetReply(ns2.SharedLinkAddress().IP, randomPort)).
					To(Equal(fmt.Sprintf("%s:%d", ns2.SharedLinkAddress().IP.String(), randomPort)))
			})).Should(BeClosed())

			// then
			Eventually(tcpErrC).Should(BeClosed())
		},
		func() []TableEntry {
			var entries []TableEntry
			var lockedPorts []uint16

			for i := 0; i < blackbox_network_tests.TestCasesAmount; i++ {
				randomPorts := socket.GenerateRandomPortsSlice(2, lockedPorts...)
				// This gives us more entropy as all generated ports will be
				// different from each other
				lockedPorts = append(lockedPorts, randomPorts...)
				desc := fmt.Sprintf("to port %%d, random port: %%d")
				entry := Entry(
					EntryDescription(desc),
					randomPorts[0],
					randomPorts[1],
				)
				entries = append(entries, entry)
			}

			return entries
		}(),
	)
})

var _ = Describe("Outbound IPv6 TCP traffic from specific interface to other ip than excluded", func() {
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

	DescribeTable("should be redirected to outbound port",
		func(serverPort, randomPort uint16) {
			// given
			tproxyConfig, err := config.Config{
				Redirect: config.Redirect{
					Outbound: config.TrafficFlow{
						Enabled: true,
						Port:    config.Port(serverPort),
					},
					Inbound: config.TrafficFlow{
						Enabled: true,
					},
					VNet: config.VNet{
						Networks: []string{"s-peer+:fd00::10:1:1/128"},
					},
				},
				IPFamilyMode:  config.IPFamilyModeDualStack,
				RuntimeStdout: io.Discard,
			}.Initialize(context.Background())
			Expect(err).ToNot(HaveOccurred())

			tcpReadyC, tcpErrC := tcp.UnsafeStartTCPServer(
				ns2,
				fmt.Sprintf(":%d", serverPort),
				tcp.ReplyWithOriginalDstIPv6,
				tcp.CloseConn,
			)
			Eventually(tcpReadyC).Should(BeClosed())
			Consistently(tcpErrC).ShouldNot(Receive())

			// when
			Eventually(ns2.UnsafeExec(func() {
				Expect(builder.RestoreIPTables(context.Background(), tproxyConfig)).Error().To(Succeed())
			})).Should(BeClosed())

			// then
			Eventually(ns.UnsafeExec(func() {
				Expect(tcp.DialIPWithPortAndGetReply(ns2.SharedLinkAddress().IP, randomPort)).
					To(Equal(fmt.Sprintf("[%s]:%d", ns2.SharedLinkAddress().IP.String(), randomPort)))
			})).Should(BeClosed())

			// then
			Eventually(tcpErrC).Should(BeClosed())
		},
		func() []TableEntry {
			var entries []TableEntry
			var lockedPorts []uint16

			for i := 0; i < blackbox_network_tests.TestCasesAmount; i++ {
				randomPorts := socket.GenerateRandomPortsSlice(2, lockedPorts...)
				// This gives us more entropy as all generated ports will be
				// different from each other
				lockedPorts = append(lockedPorts, randomPorts...)
				desc := fmt.Sprintf("to port %%d, random port: %%d")
				entry := Entry(
					EntryDescription(desc),
					randomPorts[0],
					randomPorts[1],
				)
				entries = append(entries, entry)
			}

			return entries
		}(),
	)
})
