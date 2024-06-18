package probes

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/probes"
	"github.com/kumahq/kuma/pkg/version"
	"io"
	kube_core "k8s.io/api/core/v1"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func (p *Prober) probeHTTP(writer http.ResponseWriter, req *http.Request) {
	// /<port>/<original-path>?timeout=20&host=

	upstreamScheme := getScheme(req)
	timeout := getTimeout(req)

	transport := createHttpTransport(upstreamScheme, p.isPodAddrIPV6)
	client := &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}

	upstreamReq, err := buildUpstreamReq(req, upstreamScheme, p.podAddress)
	if err != nil {
		logger.V(1).Info("unable to create upstream request", "error", err)
		writeProbeResult(writer, Unkown)
		return
	}

	res, err := client.Do(req)
	if err != nil {
		logger.V(1).Info("unable to request upstream server", "error", err)
		writeProbeResult(writer, Unhealthy)
		return
	}
	defer func() {
		err = res.Body.Close()
		if err != nil {
			logger.V(1).Info("error closing upstream request", "error", err)
		}
	}()

	b, err := readAtMost(res.Body, maxRespBodyLength)
	if err != nil {
		if errors.Is(err, errLimitReached) {
			logger.V(1).Info("non fatal body truncation for %s, Response: %v", req.URL.String(), *res)
		} else {
			logger.V(1).Info("failed to read response body", "err", err)
			writeProbeResult(writer, Unhealthy)
			return
		}
	}

	// from [200,400)
	body := string(b)
	if res.StatusCode >= http.StatusOK && res.StatusCode < http.StatusBadRequest {
		logger.V(1).Info(fmt.Sprintf("probe succeeded for %s", upstreamReq.URL.Path),
			"headers", upstreamReq.Header, "body", body)
		writeProbeResult(writer, Healthy)
		return
	}

	logger.V(1).Info(fmt.Sprintf("probe failed for %s", upstreamReq.URL.Path),
		"headers", upstreamReq.Header, "statusCode", res.StatusCode, "body", body)
	writeProbeResult(writer, Unhealthy)
	return
}

func buildUpstreamReq(downstreamReq *http.Request, upstreamScheme kube_core.URIScheme, podAddr string) (*http.Request, error) {
	port, err := getPort(downstreamReq, httpPathPattern)
	if err != nil {
		return nil, err
	}

	// todo: handle query params in original path
	upstreamPath := getUpstreamHTTPPath(downstreamReq.URL.Path)
	upstreamURL := &url.URL{
		Scheme: strings.ToLower(string(upstreamScheme)),
		Path:   upstreamPath,
		Host:   net.JoinHostPort(podAddr, strconv.Itoa(port)),
	}

	upstreamReq, err := http.NewRequest(http.MethodGet, upstreamURL.String(), nil)
	if err != nil {
		return nil, err
	}

	for key, values := range downstreamReq.Header {
		if !strings.HasPrefix(key, probes.KumaProbeHeaderPrefix) {
			upstreamReq.Header[key] = values
		}
	}

	if _, ok := downstreamReq.Header[probes.HeaderNameHost]; ok {
		// User may specify a different Host header, copy it to upstream
		upstreamReq.Header.Set("Host", downstreamReq.Header.Get(probes.HeaderNameHost))
	}
	if _, ok := downstreamReq.Header["User-Agent"]; !ok {
		// explicitly set User-Agent so it's not set to default Go value. K8s use kube-probe.
		upstreamReq.Header.Set("User-Agent", fmt.Sprintf("kuma-probe/%s", version.Build.Version))
	}
	return upstreamReq, nil
}

func createHttpTransport(scheme kube_core.URIScheme, useIPv6 bool) *http.Transport {
	httpTransport := &http.Transport{
		DisableKeepAlives: true,
	}
	if scheme == kube_core.URISchemeHTTPS {
		httpTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	d := createProbeDialer(useIPv6)
	httpTransport.DialContext = d.DialContext
	return httpTransport
}

func readAtMost(r io.Reader, limit int64) ([]byte, error) {
	limitedReader := &io.LimitedReader{R: r, N: limit}
	data, err := io.ReadAll(limitedReader)
	if err != nil {
		return data, err
	}
	if limitedReader.N <= 0 {
		return data, errors.New("the read limit is reached")
	}
	return data, nil
}
