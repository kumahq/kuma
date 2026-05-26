package validate

import (
	"context"
	"fmt"
	"math/rand/v2"
	"net"
	"net/netip"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/sethvargo/go-retry"
)

const (
	ServerPort      uint16 = 15006
	retries                = 10
	retriesInterval        = time.Second
	// ClientConnectIPv4 is the loopback alias the KUMA_MESH_OUTBOUND chain
	// redirects UID 5678 traffic on lo to (when destination is not 127.0.0.1),
	// landing on KUMA_MESH_INBOUND_REDIRECT and finally on Envoy's inbound
	// listener.
	ClientConnectIPv4 = "127.0.0.6"
	ClientConnectIPv6 = "::6"
	// SelfTestPortMin/Max is the destination-port range used by both the
	// init-time validator and the runtime readiness self-test. The exact
	// port is irrelevant because the kernel's REDIRECT --to-ports
	// <inbound-port> rule rewrites it; the range only has to avoid colliding
	// with anything bound on the loopback interface inside the netns.
	SelfTestPortMin = 20000
	SelfTestPortMax = 30000
)

type Validator struct {
	Logger              logr.Logger
	ServerListenIP      netip.Addr
	ServerListenPort    uint16
	ClientConnectIP     netip.Addr
	ClientRetryInterval time.Duration
}

func (v *Validator) getServerAddress() string {
	return net.JoinHostPort(
		v.ServerListenIP.String(),
		fmt.Sprintf("%d", v.ServerListenPort),
	)
}

func NewValidator(useIpv6 bool, port uint16, logger logr.Logger) *Validator {
	v := Validator{
		Logger: logger,
		// Traffic to lo (but not 127.0.0.1) by sidecar will be redirected to
		// KUMA_MESH_INBOUND_REDIRECT, so:
		// connect to 127.0.0.6 should be redirected to 127.0.0.1
		// connect to ::6       should be redirected to ::1
		ServerListenIP:      netip.MustParseAddr("127.0.0.1"),
		ClientConnectIP:     netip.MustParseAddr(ClientConnectIPv4),
		ServerListenPort:    port,
		ClientRetryInterval: retriesInterval,
	}

	if useIpv6 {
		v.ServerListenIP = netip.MustParseAddr("::1")
		v.ClientConnectIP = netip.MustParseAddr(ClientConnectIPv6)
	}

	return &v
}

// SelfTest performs a single TCP dial through the transparent-proxy outbound
// redirect path and returns true when the dial completes the handshake. It is
// the runtime equivalent of the init-time client/server validation but skips
// the local server: a successful handshake means the KUMA_MESH_OUTBOUND ->
// KUMA_MESH_INBOUND_REDIRECT chain rewrote the destination to Envoy's
// inbound listener (or whatever listens on the per-netns inbound port).
//
// Caller must be UID 5678 (the kuma-dp user) for the iptables owner-match to
// apply. timeout bounds the dial; on a healthy loopback redirect the dial
// completes well under 1ms, so a small timeout in the tens of milliseconds is
// safe. useIPv6 picks the v6 redirect target.
func SelfTest(useIPv6 bool, timeout time.Duration) error {
	clientIP := ClientConnectIPv4
	localIP := "127.0.0.1:0"
	if useIPv6 {
		clientIP = ClientConnectIPv6
		localIP = "[::1]:0"
	}

	laddr, err := net.ResolveTCPAddr("tcp", localIP)
	if err != nil {
		return errors.Wrap(err, "resolve local addr")
	}

	port := SelfTestPortMin + rand.IntN(SelfTestPortMax-SelfTestPortMin)
	dst := net.JoinHostPort(clientIP, fmt.Sprintf("%d", port))

	dialer := net.Dialer{LocalAddr: laddr, Timeout: timeout}
	conn, err := dialer.Dial("tcp", dst)
	if err != nil {
		return err
	}
	_ = conn.Close()
	return nil
}

func (v *Validator) RunServer(
	ctx context.Context,
	exitC chan struct{},
) (uint16, error) {
	v.Logger.Info("starting iptables validation")

	s := LocalServer{
		logger:   v.Logger.WithValues("address", v.getServerAddress()),
		listenIP: []byte(v.ServerListenIP.String()),
		address:  v.getServerAddress(),
	}

	readyC := make(chan struct{}, 1)
	errorC := make(chan error, 1)

	go func() {
		errorC <- s.Run(ctx, readyC, exitC)
	}()

	<-readyC

	select {
	case err := <-errorC:
		if err != nil {
			return 0, err
		}

		return 0, errors.New("server exited unexpectedly")
	default:
		return s.port, nil
	}
}

type LocalServer struct {
	logger   logr.Logger
	port     uint16
	address  string
	listenIP []byte
}

func (s *LocalServer) Run(
	ctx context.Context,
	readyC chan struct{},
	exitC chan struct{},
) error {
	s.logger.Info("server is listening")

	config := &net.ListenConfig{}

	l, err := config.Listen(ctx, "tcp", s.address)
	if err != nil {
		close(readyC)
		return err
	}

	s.port = uint16(l.Addr().(*net.TCPAddr).Port)

	go s.handleTcpConnections(l, exitC)
	close(readyC)
	<-exitC
	_ = l.Close()

	return nil
}

func (s *LocalServer) handleTcpConnections(l net.Listener, exitC chan struct{}) {
	for {
		conn, err := l.Accept()
		if err != nil {
			s.logger.Error(err, "failed to accept TCP connection")
			return
		}

		s.logger.Info("connection established")

		_, _ = conn.Write(s.listenIP)
		_ = conn.Close()

		select {
		case <-exitC:
			return
		default:
		}
	}
}

func (v *Validator) RunClient(
	ctx context.Context,
	serverPort uint16,
	exitC chan struct{},
) error {
	defer close(exitC)

	if err := retry.Do(
		ctx,
		retry.WithMaxRetries(retries, retry.NewConstant(v.ClientRetryInterval)), // backoff
		func(context.Context) error {
			if err := runLocalClient(v.ClientConnectIP, serverPort); err != nil {
				v.Logger.Info("failed to connect to the server, retrying", "error", err)
				return retry.RetryableError(err)
			}
			v.Logger.Info("client successfully connected to the server")
			return nil
		},
	); err != nil {
		return errors.Wrap(err, "client failed to connect to the verification server after retries - "+
			"most likely the transparent proxy is not established correctly, please check CNI logs")
	}

	v.Logger.Info("validation successful, iptables rules applied correctly")

	return nil
}

func runLocalClient(serverIP netip.Addr, serverPort uint16) error {
	var serverAddress string
	var laddr *net.TCPAddr
	var err error

	switch {
	case serverIP.Is6():
		if laddr, err = net.ResolveTCPAddr("tcp", "[::1]:0"); err != nil {
			return err
		}
	default:
		if laddr, err = net.ResolveTCPAddr("tcp", "127.0.0.1:0"); err != nil {
			return err
		}
	}

	dialer := net.Dialer{LocalAddr: laddr, Timeout: 50 * time.Millisecond}

	// connections to all ports should be redirected to the server we support
	// a pre-configured port for testing purposes
	switch serverPort {
	case 0:
		port := SelfTestPortMin + rand.IntN(SelfTestPortMax-SelfTestPortMin)
		serverAddress = net.JoinHostPort(
			serverIP.String(),
			fmt.Sprintf("%d", port),
		)
	default:
		serverAddress = net.JoinHostPort(
			serverIP.String(),
			fmt.Sprintf("%d", serverPort),
		)
	}

	conn, err := dialer.Dial("tcp", serverAddress)
	if err != nil {
		return err
	}
	defer conn.Close()

	return nil
}
