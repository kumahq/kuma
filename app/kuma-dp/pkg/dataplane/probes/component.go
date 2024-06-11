package probes

import (
	"context"
	"errors"
	"fmt"
	"github.com/bakito/go-log-logr-adapter/adapter"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	err_pkg "github.com/pkg/errors"
)

var (
	logger             = core.Log.WithName("virtual-probes")
	tcpGRPCPathPattern = regexp.MustCompile(`^/(tcp|grpc)/(?P<port>[0-9]+)(/.*)?$`)
	httpPathPattern    = regexp.MustCompile(`^/(?P<port>[0-9]+)(?P<path>/.*)?$`)
	errLimitReached    = errors.New("the read limit is reached")
)

const (
	pathPrefixTCP     = "/tcp"
	pathPrefixGRPC    = "/grpc/"
	maxRespBodyLength = 10 * 1 << 10 // 10KB

	Healthy   = "HEALTHY"
	Unhealthy = "UNHEALTHY"
	Unkown    = "UNKNOWN"
)

type Prober struct {
	Port      uint32
	IPAddress string
}

func (p *Prober) Start(stop <-chan struct{}) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", p.Port))
	if err != nil {
		return err_pkg.Wrap(err, "unable to listen for the virtual probes server")
	}

	logger.Info("starting Virtual Probes Server", "port", p.Port)
	server := &http.Server{
		ReadHeaderTimeout: time.Second,
		Handler:           p,
		ErrorLog:          adapter.ToStd(logger),
	}

	errCh := make(chan error)
	go func() {
		if err := server.Serve(lis); err != nil {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-stop:
		logger.Info("stopping Virtual Probes Server")
		return server.Shutdown(context.Background())
	}
}

func (p *Prober) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	// routes:
	// /tcp/<port>
	// /grpc/<port>
	// /<port>/<original-path-query>

	if strings.HasPrefix(req.URL.Path, pathPrefixTCP) {
		p.probeTCP(writer, req)
		return
	}

	if strings.HasPrefix(req.URL.Path, pathPrefixGRPC) {
		p.probeGRPC(writer, req)
		return
	}

	p.probeHTTP(writer, req)
}

func (p *Prober) NeedLeaderElection() bool {
	return false
}

var _ component.Component = &Prober{}

func matchGroups(re *regexp.Regexp, s string) map[string]string {
	results := make(map[string]string)
	match := re.FindStringSubmatch(s)
	if match == nil {
		return results
	}

	for i, name := range re.SubexpNames() {
		if i > 0 && name != "" {
			results[name] = match[i]
		}
	}

	return results
}

func getValidatedPort(portParam string) (int, string) {
	if portParam == "" {
		return 0, "port is required"
	}

	port, err := strconv.Atoi(portParam)
	if err != nil {
		return 0, "invalid port" + err.Error()
	}
	if port < 0 || port > 65535 {
		return 0, "invalid port value " + portParam
	}

	return port, ""
}

func getPort(req *http.Request, re *regexp.Regexp) (int, error) {
	pathParams := matchGroups(re, req.URL.Path)
	portParam := pathParams["port"]
	port, errMsg := getValidatedPort(portParam)
	if errMsg != "" {
		return 0, errors.New(errMsg)
	}
	return port, nil
}

func getTimeout(req *http.Request) time.Duration {
	timeoutParam := req.Header.Get("x-kuma-probes-timeout")
	if timeoutParam == "" {
		timeoutParam = "1"
	}

	timeout, err := strconv.Atoi(timeoutParam)
	if err != nil {
		timeout = 1
	}

	return time.Duration(timeout) * time.Second
}

func getUpstreamHTTPPath(downstreamPath string) string {
	pathParams := matchGroups(httpPathPattern, downstreamPath)
	return pathParams["path"]
}

func getScheme(req *http.Request) string {
	schemeParam := req.Header.Get("x-kuma-probes-scheme")
	if schemeParam == "" {
		schemeParam = "http"
	}
	return schemeParam
}

func getGRPCService(req *http.Request) string {
	return req.Header.Get("x-kuma-probes-grpc-service")
}

// copied from https://github.com/kubernetes/kubernetes/blob/v1.27.0-alpha.1/pkg/probe/dialer_others.go#L27
// createProbeDialer returns a dialer optimized for probes to avoid lingering sockets on TIME-WAIT state.
// The dialer reduces the TIME-WAIT period to 1 seconds instead of the OS default of 60 seconds.
// Using 1 second instead of 0 because SO_LINGER socket option to 0 causes pending data to be
// discarded and the connection to be aborted with an RST rather than for the pending data to be
// transmitted and the connection closed cleanly with a FIN.
// Ref: https://issues.k8s.io/89898
func createProbeDialer() *net.Dialer {
	dialer := &net.Dialer{
		Control: func(network, address string, c syscall.RawConn) error {
			return c.Control(func(fd uintptr) {
				_ = syscall.SetsockoptLinger(int(fd), syscall.SOL_SOCKET, syscall.SO_LINGER, &syscall.Linger{Onoff: 1, Linger: 1})
			})
		},
	}
	return dialer
}

func writeProbeResult(writer http.ResponseWriter, result string) {
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writer.Header().Set("X-Content-Type-Options", "nosniff")

	statusCode := http.StatusServiceUnavailable
	if result == Healthy {
		statusCode = http.StatusOK
	}
	writer.WriteHeader(statusCode)
	_, _ = fmt.Fprintln(writer, result)
}
