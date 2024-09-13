package validate

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
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
		ClientConnectIP:     netip.MustParseAddr("127.0.0.6"),
		ServerListenPort:    port,
		ClientRetryInterval: retriesInterval,
	}

	if useIpv6 {
		v.ServerListenIP = netip.MustParseAddr("::1")
		v.ClientConnectIP = netip.MustParseAddr("::6")
	}

	return &v
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
				v.Logger.Info("failed to connect to the server, retrying...", "error", err)
				return retry.RetryableError(err)
			}
			v.Logger.Info("client successfully connected to the server")
			return nil
		},
	); err != nil {
		return errors.Wrap(err, "client failed to connect to the verification server after retries")
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
		randPort, _ := rand.Int(rand.Reader, big.NewInt(10000))
		serverAddress = net.JoinHostPort(
			serverIP.String(),
			fmt.Sprintf("%d", 20000+randPort.Int64()),
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
