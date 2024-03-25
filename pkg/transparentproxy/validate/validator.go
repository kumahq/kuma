package validate

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/sethvargo/go-retry"
	"net"
	"net/netip"
	"time"
)

const (
	ValidationServerPort int = 15010
	validationRetries        = 10
	validationInterval       = 1 * time.Second
)

type Validator struct {
	Config *Config
	Logger logr.Logger
}

type Config struct {
	ServerListenAddress string
	ClientConnectIP     netip.Addr
	ClientRetryInterval time.Duration
}

func NewValidator(useIpv6 bool, logger logr.Logger) *Validator {
	// Connect to 127.0.0.6 should be redirected to 127.0.0.1
	// Connect to ::6       should be redirected to ::1
	serverListenIP, _ := netip.AddrFromSlice([]byte{127, 0, 0, 1})
	clientConnectIP, _ := netip.AddrFromSlice([]byte{127, 0, 0, 6})

	if useIpv6 {
		serverListenIP, _ = netip.AddrFromSlice([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1})
		clientConnectIP, _ = netip.AddrFromSlice([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 6})
	}

	return &Validator{
		Config: &Config{
			ServerListenAddress: net.JoinHostPort(serverListenIP.String(), fmt.Sprintf("%d", ValidationServerPort)),
			ClientConnectIP:     clientConnectIP,
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
			fmt.Sprintf("Validation failed, client failed to connect to server at '%s'",
				validator.Config.ClientConnectIP))
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
	s.logger.Info(fmt.Sprintf("Listening on %v", s.config.ServerListenAddress))

	config := &net.ListenConfig{}
	l, err := config.Listen(context.Background(), "tcp", s.config.ServerListenAddress)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("error listening on %v", s.config.ServerListenAddress))
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

		_, _ = conn.Write([]byte(s.config.ServerListenAddress))
		_ = conn.Close()

		select {
		case <-cExit:
			return
		default:
		}
	}
}

func (validator *Validator) runClient() error {
	c := LocalClient{ServerIP: validator.Config.ClientConnectIP}
	backoff := retry.WithMaxRetries(validationRetries, retry.NewConstant(validator.Config.ClientRetryInterval))
	return retry.Do(context.TODO(), backoff, func(ctx context.Context) error {
		e := c.Run()
		if e != nil {
			validator.Logger.Error(e, "Client failed to connect to server")
			return retry.RetryableError(e)
		}
		return nil
	})
}

type LocalClient struct {
	ServerIP netip.Addr
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

	serverAddr := net.JoinHostPort(c.ServerIP.String(), fmt.Sprintf("%d", ValidationServerPort))

	dailer := net.Dialer{LocalAddr: laddr, Timeout: 50 * time.Millisecond}
	conn, err := dailer.Dial("tcp", serverAddr)
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}
