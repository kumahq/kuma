package probes

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/bakito/go-log-logr-adapter/adapter"
	err_pkg "github.com/pkg/errors"
	kube_core "k8s.io/api/core/v1"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/probes"
)

var (
	logger             = core.Log.WithName("virtual-probes")
	tcpGRPCPathPattern = regexp.MustCompile(`^/(tcp|grpc)/(?P<port>[0-9]+)(/.*)?$`)
	httpPathPattern    = regexp.MustCompile(`^/(?P<port>[0-9]+)(?P<path>/.*)?$`)
	errLimitReached    = errors.New("the read limit is reached")

	// LocalAddrIPv4 and LocalAddrIPv6 are the special IP addresses to prevent the probe traffic from being captured by the transparent proxy
	LocalAddrIPv4 = &net.TCPAddr{IP: net.ParseIP("127.0.0.6")}
	LocalAddrIPv6 = &net.TCPAddr{IP: net.ParseIP("::6")}
)

const (
	pathPrefixTCP     = "/tcp/"
	pathPrefixGRPC    = "/grpc/"
	maxRespBodyLength = 10 * 1 << 10 // 10KB

	Healthy   = "HEALTHY"
	Unhealthy = "UNHEALTHY"
	Unknown   = "UNKNOWN"
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
		useIPv6 = len(ipAddr) == net.IPv6len && ipAddr.To4() == nil
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

	// routes:
	// /tcp/<port>
	// /grpc/<port>
	// /<port>/<original-path-query>
	mux := http.NewServeMux()
	mux.HandleFunc(pathPrefixTCP, p.probeTCP)
	mux.HandleFunc(pathPrefixGRPC, p.probeGRPC)
	mux.HandleFunc("/", p.probeHTTP)
	server := &http.Server{
		ReadHeaderTimeout: time.Second,
		Handler:           mux,
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
	if portParam, ok := pathParams["port"]; ok {
		port, errMsg := getValidatedPort(portParam)
		if errMsg != "" {
			return 0, errors.New(errMsg)
		}
		return port, nil
	}

	return 0, errors.New("port is required")
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
	if pathParam, ok := pathParams["path"]; ok {
		return pathParam
	}
	return "/"
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

func writeProbeResult(writer http.ResponseWriter, result string) {
	statusCode := http.StatusServiceUnavailable
	if result == Healthy {
		statusCode = http.StatusOK
	}
	writeHTTPProbeResult(writer, result, statusCode)
}

func writeHTTPProbeResult(writer http.ResponseWriter, result string, statusCode int) {
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writer.Header().Set("X-Content-Type-Options", "nosniff")
	writer.WriteHeader(statusCode)
	_, _ = fmt.Fprintln(writer, result)
}
