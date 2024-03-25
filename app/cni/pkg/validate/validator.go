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
	validationServerPort int = 15010
	validationRetries        = 10
	validationInterval       = 1 * time.Second
)

type Validator struct {
	Config *Config
	logger logr.Logger
}

type Config struct {
	ServerListenAddress string
	ClientConnectIP     netip.Addr
}

func NewValidator(useIpv6 bool) *Validator {
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
			ServerListenAddress: net.JoinHostPort(serverListenIP.String(), fmt.Sprintf("%d", validationServerPort)),
			ClientConnectIP:     clientConnectIP,
		},
	}
}

func (validator *Validator) Run() error {
	validator.logger.Info("Starting iptables validation...")
	sExit := make(chan struct{})

	sError := validator.runServer(sExit)
	clientErr := validator.runClient()
	if clientErr != nil {
		validator.logger.Error(clientErr,
			fmt.Sprintf("Validation failed, client failed to connect to server at '%s'",
				validator.Config.ClientConnectIP))
		close(sExit)
		return clientErr
	}

	select {
	case serverErr := <-sError:
		if serverErr == nil {
			validator.logger.Info("Validation passed, iptables rules established")
		} else {
			validator.logger.Error(serverErr, "Validation failed")
		}
		return serverErr
	}
}

func (validator *Validator) runServer(sExit chan struct{}) chan error {
	s := LocalServer{
		logger: validator.logger,
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
	s.logger.Info("Listening on %v", s.config.ServerListenAddress)

	config := &net.ListenConfig{}
	l, err := config.Listen(context.Background(), "tcp", s.config.ServerListenAddress)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("error listening on %v", s.config.ServerListenAddress))
		return err
	}

	go s.handleTcpConnections(l, exit)

	readiness <- struct{}{}
	<-exit
	return nil
}

func (s *LocalServer) handleTcpConnections(l net.Listener, cExit chan struct{}) {
	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			s.logger.Error(err, "Listener failed to accept connection")
			continue
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
	backoff := retry.WithMaxRetries(validationRetries, retry.NewConstant(validationInterval))
	return retry.Do(context.TODO(), backoff, func(ctx context.Context) error {
		e := c.Run()
		if e != nil {
			validator.logger.Error(e, "Client failed to connect to server")
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

	serverAddr := net.JoinHostPort(c.ServerIP.String(), fmt.Sprintf("%d", validationServerPort))
	raddr, err := net.ResolveTCPAddr("tcp", serverAddr)
	if err != nil {
		return err
	}

	conn, err := net.DialTCP("tcp", laddr, raddr)
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}
