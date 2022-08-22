package client

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	k8s_exec "k8s.io/client-go/util/exec"

	"github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/utils"
	"github.com/kumahq/kuma/test/server/types"
)

type CollectResponsesOpts struct {
	NumberOfRequests int
	URL              string
	Method           string
	Headers          map[string]string

	Flags        []string
	ShellEscaped func(string) string

	namespace   string
	application string
}

func DefaultCollectResponsesOpts() CollectResponsesOpts {
	return CollectResponsesOpts{
		NumberOfRequests: 10,
		Method:           "GET",
		Headers:          map[string]string{},
		ShellEscaped:     utils.ShellEscape,
		Flags: []string{
			"--fail",
		},
	}
}

type CollectResponsesOptsFn func(opts *CollectResponsesOpts)

func WithNumberOfRequests(numberOfRequests int) CollectResponsesOptsFn {
	return func(opts *CollectResponsesOpts) {
		opts.NumberOfRequests = numberOfRequests
	}
}

func WithMethod(method string) CollectResponsesOptsFn {
	return func(opts *CollectResponsesOpts) {
		opts.Method = method
	}
}

func WithHeader(key, value string) CollectResponsesOptsFn {
	return func(opts *CollectResponsesOpts) {
		opts.Headers[key] = value
	}
}

// WithPathPrefix injects prefix at the start of the target URL path.
func WithPathPrefix(prefix string) CollectResponsesOptsFn {
	return func(opts *CollectResponsesOpts) {
		u, err := url.Parse(opts.URL)
		if err != nil {
			panic(fmt.Sprintf("bad URL %q: %s", opts.URL, err))
		}

		u.Path = path.Join(prefix, u.Path)
		opts.URL = u.String()
	}
}

// Resolve sets the curl --resolve flag.
// See https://curl.se/docs/manpage.html#--resolve.
func Resolve(hostPort string, address string) CollectResponsesOptsFn {
	return func(opts *CollectResponsesOpts) {
		opts.Flags = append(opts.Flags,
			"--resolve",
			fmt.Sprintf("%s:%s", hostPort, address),
		)
	}
}

// Insecure sets the curl --insecure flag.
// See https://curl.se/docs/manpage.html#-k.
func Insecure() CollectResponsesOptsFn {
	return func(opts *CollectResponsesOpts) {
		opts.Flags = append(opts.Flags,
			"--insecure",
		)
	}
}

// NoFail removes the default curl --fail flag.
// See https://curl.se/docs/manpage.html#-f.
func NoFail() CollectResponsesOptsFn {
	return func(opts *CollectResponsesOpts) {
		flags := make([]string, 0, len(opts.Flags))
		for _, f := range flags {
			if f != "--fail" {
				flags = append(flags, f)
			}
		}
		opts.Flags = flags
	}
}

// OutputFormat sets the curl --write-out flag.
// See https://everything.curl.dev/usingcurl/verbose/writeout.
func OutputFormat(format string) CollectResponsesOptsFn {
	return func(opts *CollectResponsesOpts) {
		// Setting an output format implicitly silences other
		// kinds of output since the caller presumably needs to
		// parse the format they specified.
		opts.Flags = append(opts.Flags,
			"--silent",
			"--output", os.DevNull,
			"--write-out", opts.ShellEscaped(format),
		)
	}
}

// FromKubernetesPod executes the curl command from a pod belonging to
// the specified Kubernetes deployment. The cluster must be a Kubernetes
// cluster, and the deployment must have an "app" label that matches the
// application parameter.
//
// Note that the caller of CollectResponse still needs to specify the
// source container name within the Pod.
func FromKubernetesPod(namespace string, application string) CollectResponsesOptsFn {
	return func(opts *CollectResponsesOpts) {
		opts.namespace = namespace
		opts.application = application

		// For universal clusters, the curl exec is done in the shell,
		// so we need to quote arguments. For Kubernetes cluster, the
		// exec used the API without the shell, so we must not quote
		// anything.
		opts.ShellEscaped = func(s string) string { return s }
	}
}

func CollectOptions(requestURL string, options ...CollectResponsesOptsFn) CollectResponsesOpts {
	opts := DefaultCollectResponsesOpts()
	opts.URL = requestURL

	for _, o := range options {
		o(&opts)
	}

	return opts
}

func collectCommand(opts CollectResponsesOpts, arg0 string, args ...string) []string {
	var cmd []string

	cmd = append(cmd, arg0)

	for key, value := range opts.Headers {
		cmd = append(cmd, "--header", opts.ShellEscaped(fmt.Sprintf("%s: %s", key, value)))
	}

	cmd = append(cmd, opts.Flags...)
	cmd = append(cmd, args...)

	return cmd
}

func CollectTCPResponse(
	cluster framework.Cluster,
	container string,
	destination string,
	stdin string,
	fn ...CollectResponsesOptsFn,
) (string, error) {
	opts := CollectOptions(destination, fn...)
	cmd := []string{"bash", "-c", fmt.Sprintf("echo '%s' | curl --max-time 3 %s", stdin, opts.ShellEscaped(opts.URL))}

	var appPodName string
	if opts.namespace != "" && opts.application != "" {
		var err error
		appPodName, err = framework.PodNameOfApp(cluster, opts.application, opts.namespace)
		if err != nil {
			return "", err
		}
	}

	stdout, _, err := cluster.ExecWithRetries(opts.namespace, appPodName, container, cmd...)
	if err != nil {
		return "", err
	}

	return stdout, nil
}

func CollectResponse(
	cluster framework.Cluster,
	container string,
	destination string,
	fn ...CollectResponsesOptsFn,
) (types.EchoResponse, error) {
	opts := CollectOptions(destination, fn...)
	cmd := collectCommand(opts, "curl",
		"--request", opts.Method,
		"--max-time", "3",
		opts.ShellEscaped(opts.URL),
	)

	var appPodName string
	if opts.namespace != "" && opts.application != "" {
		var err error
		appPodName, err = framework.PodNameOfApp(cluster, opts.application, opts.namespace)
		if err != nil {
			return types.EchoResponse{}, err
		}
	}

	stdout, _, err := cluster.ExecWithRetries(opts.namespace, appPodName, container, cmd...)
	if err != nil {
		return types.EchoResponse{}, err
	}

	response := &types.EchoResponse{}
	if err := json.Unmarshal([]byte(stdout), response); err != nil {
		return types.EchoResponse{}, errors.Wrapf(err, "failed to unmarshal response: %q", stdout)
	}

	return *response, nil
}

// MakeDirectRequest collects responses using http client that calls the server from outside the cluster
func MakeDirectRequest(
	destination string,
	fn ...CollectResponsesOptsFn,
) (*http.Response, error) {
	opts := CollectOptions(destination, fn...)

	req, err := http.NewRequest(opts.Method, opts.URL, nil)
	if err != nil {
		return nil, err
	}

	for k, v := range opts.Headers {
		req.Header.Set(k, v)
	}
	if host, ok := opts.Headers["host"]; ok {
		req.Host = host
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	if strings.HasPrefix(destination, "https") {
		// When HTTPS is used, we want to send a proper SNI equal to Host header.
		// There is no one property to do this, we need to override DNS resolving and change req.URL
		// https://github.com/golang/go/issues/22704
		u, err := url.Parse(destination)
		if err != nil {
			return nil, err
		}
		dialer := &net.Dialer{}
		client.Transport = &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return dialer.Dial(network, u.Host)
			},
			TLSClientConfig: &tls.Config{
				ServerName:         req.Host,
				InsecureSkipVerify: true,
				NextProtos:         []string{"http/1.1"}, // ALPN is required by Envoy
			},
		}
		req.URL.Host = net.JoinHostPort(req.Host, req.URL.Port())
	}

	return client.Do(req)
}

// CollectResponseDirectly collects responses using http client that calls the server from outside the cluster
func CollectResponseDirectly(
	destination string,
	fn ...CollectResponsesOptsFn,
) (types.EchoResponse, error) {
	resp, err := MakeDirectRequest(destination, fn...)
	if err != nil {
		return types.EchoResponse{}, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return types.EchoResponse{}, err
	}

	response := types.EchoResponse{}
	if err := json.Unmarshal(body, &response); err != nil {
		return types.EchoResponse{}, errors.Wrapf(err, "failed to unmarshal response: %q", string(body))
	}
	return response, nil
}

// FailureResponse is the JSON output for a Curl command. Note that the available
// fields depend on the Curl version, which must be at least 7.70.0 for this feature.
//
// See https://curl.se/docs/manpage.html#-w.
type FailureResponse struct {
	Exitcode int `json:"exitcode"`

	ResponseCode int    `json:"response_code"`
	Method       string `json:"method"`
	Scheme       string `json:"scheme"`
	ContentType  string `json:"content_type"`
	URL          string `json:"url"`
	EffectiveURL string `json:"url_effective"`
}

// CollectFailure runs Curl to fetch a URL that is expected to fail. The
// Curl JSON output is returned so the caller can inspect the failure to
// see whether it was what was expected.
func CollectFailure(cluster framework.Cluster, container, destination string, fn ...CollectResponsesOptsFn) (FailureResponse, error) {
	opts := CollectOptions(destination, fn...)
	cmd := collectCommand(opts, "curl",
		"--request", opts.Method,
		"--max-time", "3",
		"--silent",               // Suppress human-readable errors.
		"--write-out", "%{json}", // Write JSON result. Requires curl 7.70.0, April 2020.
		// Silence output so that we don't try to parse it. A future refactor could try to address this
		// by using "%{stderr}%{json}", but that needs a bit more investigation.
		"--output", os.DevNull,
		opts.ShellEscaped(opts.URL),
	)

	var appPodName string
	if opts.namespace != "" && opts.application != "" {
		var err error
		appPodName, err = framework.PodNameOfApp(cluster, opts.application, opts.namespace)
		if err != nil {
			return FailureResponse{}, err
		}
	}

	stdout, _, err := cluster.Exec(opts.namespace, appPodName, container, cmd...)

	// 1. If we fail to decode the JSON status, return the JSON error,
	// but prefer the original error if we have it.
	empty := FailureResponse{}
	response := FailureResponse{}
	if jsonErr := json.Unmarshal([]byte(stdout), &response); jsonErr != nil {
		// Prefer the original error to a JSON decoding error.
		if err == nil {
			return response, jsonErr
		}
	}

	// 2. If there was no error response, we still prefer the original
	// error, but fall back to reporting that the JSON  is missing.
	if response == empty {
		if err != nil {
			return response, err
		}

		return response, errors.Errorf("empty JSON response from curl: %q", stdout)
	}

	// for k8s
	k8sExitErr := k8s_exec.CodeExitError{}
	if errors.As(err, &k8sExitErr) {
		response.Exitcode = k8sExitErr.Code
	}
	// for universal
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		response.Exitcode = exitErr.ExitCode()
	}

	// 3. Finally, report the JSON status and no execution error
	// since the JSON contains all the Curl error information.
	return response, nil
}

func CollectResponses(cluster framework.Cluster, source, destination string, fn ...CollectResponsesOptsFn) ([]types.EchoResponse, error) {
	opts := CollectOptions(destination, fn...)

	mut := sync.Mutex{}
	var responses []types.EchoResponse

	var wg sync.WaitGroup
	var err error
	for i := 0; i < opts.NumberOfRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			response, localErr := CollectResponse(cluster, source, destination, fn...)
			if localErr != nil {
				err = localErr
			}
			mut.Lock()
			responses = append(responses, response)
			mut.Unlock()
		}()
	}
	wg.Wait()
	if err != nil {
		return nil, err
	}
	return responses, nil
}

func CollectResponsesByInstance(cluster framework.Cluster, source, destination string, fn ...CollectResponsesOptsFn) (map[string]int, error) {
	responses, err := CollectResponses(cluster, source, destination, fn...)
	if err != nil {
		return nil, err
	}
	counter := map[string]int{}
	for _, response := range responses {
		counter[response.Instance]++
	}
	return counter, nil
}
