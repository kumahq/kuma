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
	"github.com/sethvargo/go-retry"
)

const (
	ValidationServerPort uint16 = 15006
	validationRetries           = 10
	validationInterval          = 1 * time.Second
)

type Validator struct {
	Config *Config
	Logger logr.Logger
}

type Config struct {
	ServerListenIP      netip.Addr
	ServerListenPort    uint16
	ClientConnectIP     netip.Addr
	ClientRetryInterval time.Duration
}

func NewValidator(useIpv6 bool, port uint16, logger logr.Logger) *Validator {
	// Traffic to lo (but not 127.0.0.1) by sidecar will be redirected to  KUMA_MESH_INBOUND_REDIRECT, so:
	// connect to 127.0.0.6 should be redirected to 127.0.0.1
	// connect to ::6       should be redirected to ::1
	serverListenIP := netip.MustParseAddr("127.0.0.1")
	clientConnectIP := netip.MustParseAddr("127.0.0.6")

	if useIpv6 {
		serverListenIP = netip.MustParseAddr("::1")
		clientConnectIP = netip.MustParseAddr("::6")
	}

	return &Validator{
		Config: &Config{
			ServerListenIP:      serverListenIP,
			ServerListenPort:    port,
			ClientConnectIP:     clientConnectIP,
			ClientRetryInterval: validationInterval,
		},
		Logger: logger,
	}
}

func (validator *Validator) RunServer(sExit chan struct{}) (uint16, error) {
	validator.Logger.Info("starting iptables validation")

	s := LocalServer{
		logger: validator.Logger,
		config: validator.Config,
	}

	sReady := make(chan struct{}, 1)
	sError := make(chan error, 1)
	go func() {
		sError <- s.Run(sReady, sExit)
	}()

	<-sReady
	select {
	case serverErr := <-sError:
		if serverErr == nil {
			serverErr = fmt.Errorf("server exited unexpectedly")
		}
		serverErr = fmt.Errorf("validation failed: %w", serverErr)
		return 0, serverErr
	default:
		return s.listenedPort, nil
	}
}

type LocalServer struct {
	logger       logr.Logger
	config       *Config
	listenedPort uint16
}

func (s *LocalServer) Run(readiness chan struct{}, exit chan struct{}) error {
	addr := net.JoinHostPort(s.config.ServerListenIP.String(), fmt.Sprintf("%d", s.config.ServerListenPort))
	s.logger.Info(fmt.Sprintf("listening on %v", addr))

	config := &net.ListenConfig{}
	l, err := config.Listen(context.Background(), "tcp", addr)
	if err != nil {
		close(readiness)
		return err
	}
	s.listenedPort = uint16(l.Addr().(*net.TCPAddr).Port)

	go s.handleTcpConnections(l, exit)

	close(readiness)
	<-exit
	_ = l.Close()
	return nil
}

func (s *LocalServer) handleTcpConnections(l net.Listener, cExit chan struct{}) {
	for {
		conn, err := l.Accept()
		if err != nil {
			s.logger.Error(err, "listener failed to accept connection")
			return
		}

		s.logger.Info("server: a connection has been established")
		_, _ = conn.Write([]byte(s.config.ServerListenIP.String()))
		_ = conn.Close()

		select {
		case <-cExit:
			return
		default:
		}
	}
}

func (validator *Validator) RunClient(serverPort uint16, sExit chan struct{}) error {
	c := LocalClient{ServerIP: validator.Config.ClientConnectIP, ServerPort: serverPort}
	backoff := retry.WithMaxRetries(validationRetries, retry.NewConstant(validator.Config.ClientRetryInterval))
	clientErr := retry.Do(context.Background(), backoff, func(ctx context.Context) error {
		e := c.Run()
		if e != nil {
			validator.Logger.Info(fmt.Sprintf("[WARNING] client failed to connect to server: %v", e.Error()))
			return retry.RetryableError(e)
		}
		validator.Logger.Info("client: connection established")
		return nil
	})

	if clientErr != nil {
		clientErr = fmt.Errorf("validation failed, client failed to connect to the verification server: %w", clientErr)
		close(sExit)
		return clientErr
	} else {
		validator.Logger.Info("validation passed, iptables rules applied correctly")
		close(sExit)
		return nil
	}
}

type LocalClient struct {
	ServerIP   netip.Addr
	ServerPort uint16
}

func (c *LocalClient) Run() error {
	laddr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	if c.ServerIP.Is6() {
		laddr, err = net.ResolveTCPAddr("tcp", "[::1]:0")
		if err != nil {
			return err
		}
	}

	// connections to all ports should be redirected to the server
	// we support a pre-configured port for testing purposes
	if c.ServerPort == 0 {
		randPort, _ := rand.Int(rand.Reader, big.NewInt(10000))
		c.ServerPort = uint16(20000 + randPort.Int64())
	}
	serverAddr := net.JoinHostPort(c.ServerIP.String(), fmt.Sprintf("%d", c.ServerPort))

	dailer := net.Dialer{LocalAddr: laddr, Timeout: 50 * time.Millisecond}
	conn, err := dailer.Dial("tcp", serverAddr)
	if err != nil {
		return err
	}
	_ = conn.Close()
	return nil
}
