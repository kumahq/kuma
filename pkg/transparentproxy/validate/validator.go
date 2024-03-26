package validate

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/sethvargo/go-retry"
	"net"
	"net/netip"
	"time"
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
	ClientConnectPort   uint16
	ClientRetryInterval time.Duration
}

func NewValidator(useIpv6 bool, port uint16, logger logr.Logger) *Validator {
	// Connect to 127.0.0.6 should be redirected to 127.0.0.1
	// Connect to ::6       should be redirected to ::1
	serverListenIP, _ := netip.AddrFromSlice([]byte{127, 0, 0, 1})

	if useIpv6 {
		serverListenIP, _ = netip.AddrFromSlice([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1})
	}

	return &Validator{
		Config: &Config{
			ServerListenIP:      serverListenIP,
			ServerListenPort:    port,
			ClientRetryInterval: validationInterval,
		},
		Logger: logger,
	}
}

func (validator *Validator) Run() error {
	validator.Logger.Info("Starting iptables validation...")
	sExit := make(chan struct{})

	sError := validator.runServer(sExit)
	select {
	case serverErr := <-sError:
		if serverErr == nil {
			serverErr = fmt.Errorf("server exited unexpectedly")
		}
		validator.Logger.Error(serverErr, "Validation failed")
		return serverErr
	default:
	}

	clientErr := validator.runClient()
	if clientErr != nil {
		validator.Logger.Error(clientErr,
			fmt.Sprintf("Validation failed, client failed to connect to the verification server"))
		close(sExit)
		return clientErr
	} else {
		close(sExit)
		validator.Logger.Info("Validation passed, iptables rules established")
		return nil
	}
}

func (validator *Validator) runServer(sExit chan struct{}) chan error {
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
	return sError
}

type LocalServer struct {
	logger logr.Logger
	config *Config
}

func (s *LocalServer) Run(readiness chan struct{}, exit chan struct{}) error {
	addr := net.JoinHostPort(s.config.ServerListenIP.String(), fmt.Sprintf("%d", s.config.ServerListenPort))
	s.logger.Info(fmt.Sprintf("Listening on %v", addr))

	config := &net.ListenConfig{}
	l, err := config.Listen(context.Background(), "tcp", addr)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("error listening on %v", s.config.ServerListenIP))
		return err
	}

	go s.handleTcpConnections(l, exit)

	readiness <- struct{}{}
	<-exit
	l.Close()
	return nil
}

func (s *LocalServer) handleTcpConnections(l net.Listener, cExit chan struct{}) {
	for {
		conn, err := l.Accept()
		if err != nil {
			s.logger.Error(err, "Listener failed to accept connection")
			return
		}

		s.logger.Error(err, "Server: a connection has been established")
		_, _ = conn.Write([]byte(s.config.ServerListenIP.String()))
		_ = conn.Close()

		select {
		case <-cExit:
			return
		default:
		}
	}
}

func (validator *Validator) runClient() error {
	c := LocalClient{ServerIP: validator.Config.ServerListenIP, ServerPort: validator.Config.ClientConnectPort}
	backoff := retry.WithMaxRetries(validationRetries, retry.NewConstant(validator.Config.ClientRetryInterval))
	return retry.Do(context.TODO(), backoff, func(ctx context.Context) error {
		e := c.Run()
		if e != nil {
			validator.Logger.Error(e, "Client failed to connect to server")
			return retry.RetryableError(e)
		}
		validator.Logger.Error(e, "Client: connection established")
		return nil
	})
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

	if c.ServerPort == 0 {
		// connections to all ports should be redirected to the server
		c.ServerPort = uint16(random.Random(20000, 30000))
	}
	serverAddr := net.JoinHostPort(c.ServerIP.String(), fmt.Sprintf("%d", c.ServerPort))

	dailer := net.Dialer{LocalAddr: laddr, Timeout: 50 * time.Millisecond}
	conn, err := dailer.Dial("tcp", serverAddr)
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}
