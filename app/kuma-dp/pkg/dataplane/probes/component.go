package probes

import (
	"context"
	"errors"
	"fmt"
	"github.com/bakito/go-log-logr-adapter/adapter"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/probes"
	kube_core "k8s.io/api/core/v1"
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

	localAddrIPv4 = &net.TCPAddr{IP: net.ParseIP("127.0.0.6")}
	localAddrIPv6 = &net.TCPAddr{IP: net.ParseIP("::6")}
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
	listenPort    uint32
	podAddress    string
	isPodAddrIPV6 bool
}

func NewProber(podIPAddr string, listenPort uint32) *Prober {
	ipAddr := net.ParseIP(podIPAddr)
	useIPv6 := false
	if ipAddr != nil {
		useIPv6 = len(ipAddr) == net.IPv6len
	}

	return &Prober{
		listenPort:    listenPort,
		podAddress:    podIPAddr,
		isPodAddrIPV6: useIPv6,
	}
}

func (p *Prober) Start(stop <-chan struct{}) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", p.listenPort))
	if err != nil {
		return err_pkg.Wrap(err, "unable to listen for the virtual probes server")
	}

	logger.Info("starting Virtual Probes Server", "port", p.listenPort)
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
	timeoutParam := req.Header.Get(probes.HeaderNameTimeout)
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

func getScheme(req *http.Request) kube_core.URIScheme {
	schemeParam := req.Header.Get(probes.HeaderNameScheme)
	if schemeParam == "" {
		return kube_core.URISchemeHTTP
	}
	return kube_core.URIScheme(schemeParam)
}

func getGRPCService(req *http.Request) string {
	return req.Header.Get(probes.HeaderNameGRPCService)
}

// copied from https://github.com/kubernetes/kubernetes/blob/v1.27.0-alpha.1/pkg/probe/dialer_others.go#L27
// createProbeDialer returns a dialer optimized for probes to avoid lingering sockets on TIME-WAIT state.
// The dialer reduces the TIME-WAIT period to 1 seconds instead of the OS default of 60 seconds.
// Using 1 second instead of 0 because SO_LINGER socket option to 0 causes pending data to be
// discarded and the connection to be aborted with an RST rather than for the pending data to be
// transmitted and the connection closed cleanly with a FIN.
// Ref: https://issues.k8s.io/89898
func createProbeDialer(isIPv6 bool) *net.Dialer {
	dialer := &net.Dialer{
		Control: func(network, address string, c syscall.RawConn) error {
			return c.Control(func(fd uintptr) {
				_ = syscall.SetsockoptLinger(int(fd), syscall.SOL_SOCKET, syscall.SO_LINGER, &syscall.Linger{Onoff: 1, Linger: 1})
			})
		},
	}
	dialer.LocalAddr = localAddrIPv4
	if isIPv6 {
		dialer.LocalAddr = localAddrIPv6
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
