package blackbox_network_tests_test

import (
	"context"
	"fmt"
	"io"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/transparentproxy/config"
	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/builder"
	"github.com/kumahq/kuma/test/blackbox_network_tests"
	"github.com/kumahq/kuma/test/framework/network/netns"
	"github.com/kumahq/kuma/test/framework/network/socket"
	"github.com/kumahq/kuma/test/framework/network/tcp"
)

var _ = Describe("Inbound IPv4 TCP traffic from any ports", func() {
	var err error
	var ns *netns.NetNS

	BeforeEach(func() {
		ns, err = netns.NewNetNSBuilder().Build()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		Expect(ns.Cleanup()).To(Succeed())
	})

	DescribeTable("should be redirected to the inbound_redirection port",
		func(serverPort, randomPort uint16) {
			// given
			tcpServerAddress := fmt.Sprintf(":%d", serverPort)
			peerAddress := ns.Veth().PeerAddress()
			tproxyConfig, err := config.Config{
				Redirect: config.Redirect{
					Inbound: config.TrafficFlow{
						Enabled: true,
						Port:    config.Port(serverPort),
					},
					Outbound: config.TrafficFlow{
						Enabled: true,
					},
				},
				RuntimeStdout: io.Discard,
				RuntimeStderr: io.Discard,
				Log: config.Log{
					Enabled: true,
				},
			}.Initialize(context.Background())
			Expect(err).ToNot(HaveOccurred())

			tcpReadyC, tcpErrC := tcp.UnsafeStartTCPServer(
				ns,
				tcpServerAddress,
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
			Expect(tcp.DialIPWithPortAndGetReply(peerAddress, randomPort)).
				To(Equal(fmt.Sprintf("%s:%d", peerAddress, randomPort)))

			// and, then
			Consistently(tcpErrC).ShouldNot(Receive())
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

var _ = Describe("Inbound IPv6 TCP traffic from any ports", func() {
	var err error
	var ns *netns.NetNS

	BeforeEach(func() {
		ns, err = netns.NewNetNSBuilder().WithIPv6(true).Build()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		Expect(ns.Cleanup()).To(Succeed())
	})

	DescribeTable("should be redirected to the inbound_redirection port",
		func(serverPort, randomPort uint16) {
			// given
			tcpServerAddress := fmt.Sprintf(":%d", serverPort)
			peerAddress := ns.Veth().PeerAddress()
			tproxyConfig, err := config.Config{
				Redirect: config.Redirect{
					Inbound: config.TrafficFlow{
						Enabled: true,
						Port:    config.Port(serverPort),
					},
					Outbound: config.TrafficFlow{
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
				tcpServerAddress,
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
			Expect(tcp.DialIPWithPortAndGetReply(peerAddress, randomPort)).
				To(Equal(fmt.Sprintf("[%s]:%d", peerAddress, randomPort)))

			// and, then
			Consistently(tcpErrC).ShouldNot(Receive())
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

var _ = Describe("Inbound IPv4 TCP traffic from any ports except excluded ones", func() {
	var err error
	var ns *netns.NetNS

	BeforeEach(func() {
		ns, err = netns.NewNetNSBuilder().Build()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		Expect(ns.Cleanup()).To(Succeed())
	})

	DescribeTable("should be redirected to the inbound_redirection port",
		func(serverPort, randomPort, excludedPort uint16) {
			// given
			tproxyConfig, err := config.Config{
				Redirect: config.Redirect{
					Inbound: config.TrafficFlow{
						Enabled:      true,
						Port:         config.Port(serverPort),
						ExcludePorts: config.Ports{config.Port(excludedPort)},
					},
					Outbound: config.TrafficFlow{
						Enabled: true,
					},
				},
				RuntimeStdout: io.Discard,
			}.Initialize(context.Background())
			Expect(err).ToNot(HaveOccurred())

			peerAddress := ns.Veth().PeerAddress()

			redirectReadyC, redirectErrC := tcp.UnsafeStartTCPServer(
				ns,
				fmt.Sprintf(":%d", serverPort),
				tcp.ReplyWithOriginalDstIPv4,
				tcp.CloseConn,
			)
			Eventually(redirectReadyC).Should(BeClosed())
			Consistently(redirectErrC).ShouldNot(Receive())

			excludedReadyC, excludedErrC := tcp.UnsafeStartTCPServer(
				ns,
				fmt.Sprintf(":%d", excludedPort),
				tcp.ReplyWith("foobar"),
				tcp.CloseConn,
			)
			Eventually(excludedReadyC).Should(BeClosed())
			Consistently(excludedErrC).ShouldNot(Receive())

			// when
			Eventually(ns.UnsafeExec(func() {
				Expect(builder.RestoreIPTables(context.Background(), tproxyConfig)).Error().To(Succeed())
			})).Should(BeClosed())

			// then
			Expect(tcp.DialIPWithPortAndGetReply(peerAddress, excludedPort)).To(Equal("foobar"))

			// then
			Expect(tcp.DialIPWithPortAndGetReply(peerAddress, randomPort)).
				To(Equal(fmt.Sprintf("%s:%d", peerAddress, randomPort)))

			// then
			Eventually(redirectErrC).Should(BeClosed())
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

var _ = Describe("Inbound IPv6 TCP traffic from any ports except excluded ones", func() {
	var err error
	var ns *netns.NetNS

	BeforeEach(func() {
		ns, err = netns.NewNetNSBuilder().WithIPv6(true).Build()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		Expect(ns.Cleanup()).To(Succeed())
	})

	DescribeTable("should be redirected to the inbound_redirection port",
		func(serverPort, randomPort, excludedPort uint16) {
			// given
			tproxyConfig, err := config.Config{
				Redirect: config.Redirect{
					Inbound: config.TrafficFlow{
						Enabled:      true,
						Port:         config.Port(serverPort),
						ExcludePorts: config.Ports{config.Port(excludedPort)},
					},
					Outbound: config.TrafficFlow{
						Enabled: true,
					},
				},
				IPFamilyMode:  config.IPFamilyModeDualStack,
				RuntimeStdout: io.Discard,
			}.Initialize(context.Background())
			Expect(err).ToNot(HaveOccurred())

			peerAddress := ns.Veth().PeerAddress()

			redirectReadyC, redirectErrC := tcp.UnsafeStartTCPServer(
				ns,
				fmt.Sprintf(":%d", serverPort),
				tcp.ReplyWithOriginalDstIPv6,
				tcp.CloseConn,
			)
			Eventually(redirectReadyC).Should(BeClosed())
			Consistently(redirectErrC).ShouldNot(Receive())

			excludedReadyC, excludedErrC := tcp.UnsafeStartTCPServer(
				ns,
				fmt.Sprintf(":%d", excludedPort),
				tcp.ReplyWith("foobar"),
				tcp.CloseConn,
			)
			Eventually(excludedReadyC).Should(BeClosed())
			Consistently(excludedErrC).ShouldNot(Receive())

			// when
			Eventually(ns.UnsafeExec(func() {
				Expect(builder.RestoreIPTables(context.Background(), tproxyConfig)).Error().To(Succeed())
			})).Should(BeClosed())

			// then
			Expect(tcp.DialIPWithPortAndGetReply(peerAddress, excludedPort)).To(Equal("foobar"))

			// then
			Expect(tcp.DialIPWithPortAndGetReply(peerAddress, randomPort)).
				To(Equal(fmt.Sprintf("[%s]:%d", peerAddress, randomPort)))

			// then
			Eventually(redirectErrC).Should(BeClosed())
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

var _ = Describe("Inbound IPv4 TCP traffic only from included ports", func() {
	var err error
	var ns *netns.NetNS

	BeforeEach(func() {
		ns, err = netns.NewNetNSBuilder().Build()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		Expect(ns.Cleanup()).To(Succeed())
	})

	DescribeTable("should be redirected to the inbound_redirection port",
		func(serverPort, randomPort, includedPort uint16) {
			// given
			tproxyConfig, err := config.Config{
				Redirect: config.Redirect{
					Inbound: config.TrafficFlow{
						Enabled:      true,
						Port:         config.Port(serverPort),
						IncludePorts: config.Ports{config.Port(includedPort)},
						ExcludePorts: config.Ports{config.Port(includedPort)},
					},
					Outbound: config.TrafficFlow{
						Enabled: true,
					},
				},
				RuntimeStdout: io.Discard,
			}.Initialize(context.Background())
			Expect(err).ToNot(HaveOccurred())

			peerAddress := ns.Veth().PeerAddress()

			redirectReadyC, redirectErrC := tcp.UnsafeStartTCPServer(
				ns,
				fmt.Sprintf(":%d", serverPort),
				tcp.ReplyWithOriginalDstIPv4,
				tcp.CloseConn,
			)
			Eventually(redirectReadyC).Should(BeClosed())
			Consistently(redirectErrC).ShouldNot(Receive())

			randomReadyC, randomErrC := tcp.UnsafeStartTCPServer(
				ns,
				fmt.Sprintf(":%d", randomPort),
				tcp.ReplyWith("foobar"),
				tcp.CloseConn,
			)
			Eventually(randomReadyC).Should(BeClosed())
			Consistently(randomErrC).ShouldNot(Receive())

			// when
			Eventually(ns.UnsafeExec(func() {
				Expect(builder.RestoreIPTables(context.Background(), tproxyConfig)).Error().To(Succeed())
			})).Should(BeClosed())

			// then
			Expect(tcp.DialIPWithPortAndGetReply(peerAddress, includedPort)).
				To(Equal(fmt.Sprintf("%s:%d", peerAddress, includedPort)))

			// then
			Expect(tcp.DialIPWithPortAndGetReply(peerAddress, randomPort)).
				To(Equal("foobar"))

			// then
			Eventually(redirectErrC).Should(BeClosed())
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
				desc := fmt.Sprintf("to port %%d, from port %%d (included: %%d)")
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

var _ = Describe("Inbound IPv6 TCP traffic only from included ports", func() {
	var err error
	var ns *netns.NetNS

	BeforeEach(func() {
		ns, err = netns.NewNetNSBuilder().WithIPv6(true).Build()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		Expect(ns.Cleanup()).To(Succeed())
	})

	DescribeTable("should be redirected to the inbound_redirection port",
		func(serverPort, randomPort, includedPort uint16) {
			// given
			tproxyConfig, err := config.Config{
				Redirect: config.Redirect{
					Inbound: config.TrafficFlow{
						Enabled:      true,
						Port:         config.Port(serverPort),
						IncludePorts: config.Ports{config.Port(includedPort)},
						ExcludePorts: config.Ports{config.Port(includedPort)},
					},
					Outbound: config.TrafficFlow{
						Enabled: true,
					},
				},
				IPFamilyMode:  config.IPFamilyModeDualStack,
				RuntimeStdout: io.Discard,
			}.Initialize(context.Background())
			Expect(err).ToNot(HaveOccurred())

			peerAddress := ns.Veth().PeerAddress()

			redirectReadyC, redirectErrC := tcp.UnsafeStartTCPServer(
				ns,
				fmt.Sprintf(":%d", serverPort),
				tcp.ReplyWithOriginalDstIPv6,
				tcp.CloseConn,
			)
			Eventually(redirectReadyC).Should(BeClosed())
			Consistently(redirectErrC).ShouldNot(Receive())

			randomReadyC, randomErrC := tcp.UnsafeStartTCPServer(
				ns,
				fmt.Sprintf(":%d", randomPort),
				tcp.ReplyWith("foobar"),
				tcp.CloseConn,
			)
			Eventually(randomReadyC).Should(BeClosed())
			Consistently(randomErrC).ShouldNot(Receive())

			// when
			Eventually(ns.UnsafeExec(func() {
				Expect(builder.RestoreIPTables(context.Background(), tproxyConfig)).Error().To(Succeed())
			})).Should(BeClosed())

			// then
			Expect(tcp.DialIPWithPortAndGetReply(peerAddress, includedPort)).
				To(Equal(fmt.Sprintf("[%s]:%d", peerAddress, includedPort)))

			// then
			Expect(tcp.DialIPWithPortAndGetReply(peerAddress, randomPort)).
				To(Equal("foobar"))

			// then
			Eventually(redirectErrC).Should(BeClosed())
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

var _ = Describe("Inbound IPv4 TCP traffic from any ports", func() {
	var err error
	var ns *netns.NetNS

	BeforeEach(func() {
		ns, err = netns.NewNetNSBuilder().Build()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		Expect(ns.Cleanup()).To(Succeed())
	})

	DescribeTable("should not be redirected to the inbound_redirection port",
		func(serverPort, randomPort uint16) {
			// given
			peerAddress := ns.Veth().PeerAddress()
			address := fmt.Sprintf("%s:%d", peerAddress.To4(), randomPort)
			tproxyConfig, err := config.Config{
				Redirect: config.Redirect{
					Inbound: config.TrafficFlow{
						Enabled: false,
						Port:    config.Port(serverPort),
					},
					Outbound: config.TrafficFlow{
						Enabled: true,
					},
				},
				RuntimeStdout: io.Discard,
			}.Initialize(context.Background())
			Expect(err).ToNot(HaveOccurred())

			tcpReadyC, tcpErrC := tcp.UnsafeStartTCPServer(
				ns,
				address,
				tcp.ReplyWith("randomServer"),
				tcp.CloseConn,
			)
			Eventually(tcpReadyC).Should(BeClosed())
			Consistently(tcpErrC).ShouldNot(Receive())

			// when
			Eventually(ns.UnsafeExec(func() {
				Expect(builder.RestoreIPTables(context.Background(), tproxyConfig)).Error().To(Succeed())
			})).Should(BeClosed())

			// then
			Expect(tcp.DialIPWithPortAndGetReply(peerAddress, randomPort)).
				To(Equal("randomServer"))

			// and, then
			Consistently(tcpErrC).ShouldNot(Receive())
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

var _ = Describe("Inbound IPv6 TCP traffic from any ports", func() {
	var err error
	var ns *netns.NetNS

	BeforeEach(func() {
		ns, err = netns.NewNetNSBuilder().WithIPv6(true).Build()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		Expect(ns.Cleanup()).To(Succeed())
	})

	DescribeTable("should not be redirected to the inbound_redirection port",
		func(serverPort, randomPort uint16) {
			// given
			peerAddress := ns.Veth().PeerAddress()
			address := fmt.Sprintf(":%d", randomPort)
			tproxyConfig, err := config.Config{
				Redirect: config.Redirect{
					Inbound: config.TrafficFlow{
						Enabled: false,
						Port:    config.Port(serverPort),
					},
					Outbound: config.TrafficFlow{
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
			Expect(tcp.DialIPWithPortAndGetReply(peerAddress, randomPort)).
				To(Equal("randomPort"))

			// and, then
			Consistently(tcpErrC).ShouldNot(Receive())
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
