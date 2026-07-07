package client

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	k8s_exec "k8s.io/client-go/util/exec"

<<<<<<< HEAD
	"github.com/kumahq/kuma/v2/test/framework"
	"github.com/kumahq/kuma/v2/test/framework/utils"
	"github.com/kumahq/kuma/v2/test/server/types"
=======
	"github.com/kumahq/kuma/v3/test/framework"
	"github.com/kumahq/kuma/v3/test/framework/report"
	"github.com/kumahq/kuma/v3/test/framework/utils"
	"github.com/kumahq/kuma/v3/test/server/types"
>>>>>>> 62e80d799d (test(e2e): improve failure diagnostics (#17036))
)

type CollectResponsesOpts struct {
	numberOfRequests            uint
	maxConcurrentRequests       uint
	maxConcurrentRequestDelayMs uint
	maxTime                     uint
	verbose                     bool
	cacert                      string
	URL                         string
	Method                      string
	Headers                     map[string]string

	Flags        []string
	ShellEscaped func(string) string

	namespace   string
	application string

	withoutRetries bool
}

func DefaultCollectResponsesOpts() CollectResponsesOpts {
	return CollectResponsesOpts{
		numberOfRequests:            10,
		maxConcurrentRequests:       5,
		maxConcurrentRequestDelayMs: 100,
		maxTime:                     5,
		verbose:                     false,
		Method:                      "GET",
		Headers:                     map[string]string{},
		ShellEscaped:                utils.ShellEscape,
		Flags: []string{
			"--fail",
		},
	}
}

type CollectResponsesOptsFn func(opts *CollectResponsesOpts)

func WithNumberOfRequests(numberOfRequests uint) CollectResponsesOptsFn {
	return func(opts *CollectResponsesOpts) {
		opts.numberOfRequests = numberOfRequests
	}
}

func WithMaxTime(seconds uint) CollectResponsesOptsFn {
	return func(opts *CollectResponsesOpts) {
		opts.maxTime = seconds
	}
}

func WithVerbose() CollectResponsesOptsFn {
	return func(opts *CollectResponsesOpts) {
		opts.verbose = true
	}
}

func WithCACert(cacert string) CollectResponsesOptsFn {
	return func(opts *CollectResponsesOpts) {
		opts.cacert = cacert
	}
}

func WithMaxConcurrentRequests(maxConcurrentRequests uint) CollectResponsesOptsFn {
	return func(opts *CollectResponsesOpts) {
		opts.maxConcurrentRequests = maxConcurrentRequests
	}
}

// Number of milliseconds as an int
func WithMaxConcurrentRequestDelayMs(maxConcurrentRequestDelayMs uint) CollectResponsesOptsFn {
	return func(opts *CollectResponsesOpts) {
		opts.maxConcurrentRequestDelayMs = maxConcurrentRequestDelayMs
	}
}

func WithMethod(method string) CollectResponsesOptsFn {
	return func(opts *CollectResponsesOpts) {
		opts.Method = method
	}
}

func WithoutRetries() CollectResponsesOptsFn {
	return func(opts *CollectResponsesOpts) {
		opts.withoutRetries = true
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
// Note that the caller of CollectEchoResponse still needs to specify the
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

	if opts.verbose {
		cmd = append(cmd, "--verbose")
	}

	if opts.cacert != "" {
		cmd = append(cmd, "--cacert", opts.cacert)
	}

	cmd = append(cmd, "--max-time", strconv.Itoa(int(opts.maxTime)))
	cmd = append(cmd, opts.Flags...)
	cmd = append(cmd, args...)

	return cmd
}

func CollectTLSResponse(
	cluster framework.Cluster,
	container string,
	destination string,
	stdin string,
	fn ...CollectResponsesOptsFn,
) (string, error) {
	opts := CollectOptions(destination, fn...)
	cmd := []string{"bash", "-c", fmt.Sprintf("echo '%s' | timeout 5 openssl s_client -verify_quiet -quiet -ign_eof -connect %s 2>/dev/null", stdin, opts.ShellEscaped(opts.URL))}

	var appPodName string
	if opts.namespace != "" && opts.application != "" {
		var err error
		appPodName, err = framework.PodNameOfApp(cluster, opts.application, opts.namespace)
		if err != nil {
			return "", err
		}
	}

	stdout, stderr, err := cluster.Exec(opts.namespace, appPodName, container, cmd...)
	if err != nil {
		return "", fmt.Errorf("stderr: '%s', %v", stderr, err)
	}

	return stdout, nil
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

	stdout, stderr, err := cluster.Exec(opts.namespace, appPodName, container, cmd...)
	if err != nil {
		return "", fmt.Errorf("stderr: '%s', %v", stderr, err)
	}

	return stdout, nil
}

func CollectResponse(
	cluster framework.Cluster,
	container string,
	destination string,
	fn ...CollectResponsesOptsFn,
) (string, string, error) {
	execResult := collectResponse(cluster, container, destination, fn...)
	if execResult.err != nil {
		addCollectDiagnostic("exec-error", execResult, execResult.err)
	}
	return execResult.stdout, execResult.stderr, execResult.err
}

type collectResponseResult struct {
	cluster     framework.Cluster
	container   string
	destination string
	opts        CollectResponsesOpts
	command     []string
	namespace   string
	podName     string
	stdout      string
	stderr      string
	err         error
}

func collectResponse(
	cluster framework.Cluster,
	container string,
	destination string,
	fn ...CollectResponsesOptsFn,
) collectResponseResult {
	opts := CollectOptions(destination, fn...)
	cmd := collectCommand(opts, "curl",
		"--request", opts.Method,
		opts.ShellEscaped(opts.URL),
	)

	result := collectResponseResult{
		cluster:     cluster,
		container:   container,
		destination: destination,
		opts:        opts,
		command:     cmd,
		namespace:   opts.namespace,
	}

	var appPodName string
	if opts.namespace != "" && opts.application != "" {
		var err error
		appPodName, err = framework.PodNameOfApp(cluster, opts.application, opts.namespace)
		if err != nil {
			result.err = err
			return result
		}
	}
	result.podName = appPodName

	result.stdout, result.stderr, result.err = cluster.Exec(opts.namespace, appPodName, container, cmd...)
	return result
}

func CollectEchoResponse(
	cluster framework.Cluster,
	container string,
	destination string,
	fn ...CollectResponsesOptsFn,
) (types.EchoResponse, error) {
	execResult := collectResponse(cluster, container, destination, fn...)
	if execResult.err != nil {
		err := fmt.Errorf("stderr: '%s', %v", execResult.stderr, execResult.err)
		return types.EchoResponse{}, withCollectDiagnostic("exec-error", execResult, err)
	}

	response := &types.EchoResponse{}
	if err := json.Unmarshal([]byte(execResult.stdout), response); err != nil {
		err = errors.Wrapf(err, "failed to unmarshal response: %q", execResult.stdout)
		return types.EchoResponse{}, withCollectDiagnostic("unmarshal-error", execResult, err)
	}

	if response.Instance == "" {
		err := errors.New("'instance' field is empty ")
		return types.EchoResponse{}, withCollectDiagnostic("empty-instance", execResult, err)
	}

	return *response, nil
}

type collectDiagnostic struct {
	Reason      string            `json:"reason"`
	Error       string            `json:"error,omitempty"`
	Timestamp   time.Time         `json:"timestamp"`
	Cluster     string            `json:"cluster"`
	Source      string            `json:"source"`
	Namespace   string            `json:"namespace,omitempty"`
	Application string            `json:"application,omitempty"`
	PodName     string            `json:"podName,omitempty"`
	Destination string            `json:"destination"`
	Method      string            `json:"method"`
	URL         string            `json:"url"`
	Headers     map[string]string `json:"headers,omitempty"`
	Flags       []string          `json:"flags,omitempty"`
	Command     []string          `json:"command"`
	Stdout      string            `json:"stdout"`
	Stderr      string            `json:"stderr"`
}

type diagnosticError struct {
	err  error
	file string
}

func (e diagnosticError) Error() string {
	return e.err.Error()
}

func (e diagnosticError) Unwrap() error {
	return e.err
}

func withCollectDiagnostic(reason string, result collectResponseResult, err error) error {
	return diagnosticError{
		err:  err,
		file: addCollectDiagnostic(reason, result, err),
	}
}

func diagnosticFileName(prefix string) string {
	return fmt.Sprintf("%s-%d.json", prefix, time.Now().UnixNano())
}

func addCollectDiagnostic(reason string, result collectResponseResult, err error) string {
	name := diagnosticFileName(
		fmt.Sprintf("client-collect-%s-%s", result.cluster.Name(), result.container),
	)
	diagnostic := collectDiagnostic{
		Reason:      reason,
		Timestamp:   time.Now().UTC(),
		Cluster:     result.cluster.Name(),
		Source:      result.container,
		Namespace:   result.namespace,
		Application: result.opts.application,
		PodName:     result.podName,
		Destination: result.destination,
		Method:      result.opts.Method,
		URL:         result.opts.URL,
		Headers:     redactedHeaders(result.opts.Headers),
		Flags:       result.opts.Flags,
		Command:     redactedCommand(result.command),
		Stdout:      truncateDiagnosticString(redactedCollectStdout(result.stdout)),
		Stderr:      truncateDiagnosticString(result.stderr),
	}
	if err != nil {
		diagnostic.Error = err.Error()
	}

	data, jsonErr := json.MarshalIndent(diagnostic, "", "  ")
	if jsonErr != nil {
		framework.Logf("failed to marshal collect diagnostic: %v", jsonErr)
		return name
	}
	report.AddFileToReportEntry(name, data)
	return name
}

const redactedDiagnosticValue = "[redacted]"

func redactedHeaders(headers map[string]string) map[string]string {
	if len(headers) == 0 {
		return nil
	}

	redacted := make(map[string]string, len(headers))
	for key, value := range headers {
		if shouldRedactHeader(key) {
			redacted[key] = redactedDiagnosticValue
			continue
		}
		redacted[key] = value
	}
	return redacted
}

func redactedEchoHeaders(headers map[string][]string) map[string][]string {
	if len(headers) == 0 {
		return nil
	}

	redacted := make(map[string][]string, len(headers))
	for key, values := range headers {
		if shouldRedactHeader(key) {
			replacement := make([]string, len(values))
			for i := range replacement {
				replacement[i] = redactedDiagnosticValue
			}
			redacted[key] = replacement
			continue
		}
		redacted[key] = append([]string(nil), values...)
	}
	return redacted
}

func redactedEchoResponse(response types.EchoResponse) types.EchoResponse {
	response.Received.Headers = redactedEchoHeaders(response.Received.Headers)
	return response
}

func redactedDiagnosticResponse(response any) any {
	switch value := response.(type) {
	case types.EchoResponse:
		return redactedEchoResponse(value)
	default:
		return value
	}
}

func redactedDiagnosticResponses(responses []any) []any {
	if len(responses) == 0 {
		return nil
	}

	redacted := make([]any, len(responses))
	for i, response := range responses {
		redacted[i] = redactedDiagnosticResponse(response)
	}
	return redacted
}

func redactedCommand(command []string) []string {
	if len(command) == 0 {
		return nil
	}

	redacted := append([]string(nil), command...)
	for i := 0; i < len(redacted); i++ {
		arg := redacted[i]
		switch {
		case (arg == "--header" || arg == "-H") && i+1 < len(redacted):
			i++
			redacted[i] = redactedHeaderArgument(redacted[i])
		case strings.HasPrefix(arg, "--header="):
			redacted[i] = "--header=" + redactedHeaderArgument(strings.TrimPrefix(arg, "--header="))
		case strings.HasPrefix(arg, "-H") && len(arg) > len("-H"):
			redacted[i] = "-H" + redactedHeaderArgument(strings.TrimSpace(strings.TrimPrefix(arg, "-H")))
		}
	}
	return redacted
}

func redactedHeaderArgument(header string) string {
	trimmed := strings.TrimSpace(header)
	if trimmed == "" {
		return header
	}

	quote := byte(0)
	if len(trimmed) >= 2 {
		first := trimmed[0]
		last := trimmed[len(trimmed)-1]
		if (first == '\'' || first == '"') && last == first {
			quote = first
			trimmed = trimmed[1 : len(trimmed)-1]
		}
	}

	name, _, found := strings.Cut(trimmed, ":")
	if !found || !shouldRedactHeader(name) {
		return header
	}

	value := strings.TrimSpace(name) + ": " + redactedDiagnosticValue
	if quote != 0 {
		return string(quote) + value + string(quote)
	}
	return value
}

func shouldRedactHeader(name string) bool {
	normalized := strings.ToLower(strings.TrimSpace(name))
	for _, token := range []string{
		"authorization",
		"token",
		"secret",
		"cookie",
		"credential",
		"api-key",
		"apikey",
	} {
		if strings.Contains(normalized, token) {
			return true
		}
	}
	return false
}

func redactedCollectStdout(stdout string) string {
	if stdout == "" {
		return ""
	}

	var payload any
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		return stdout
	}

	redactJSONEchoHeaders(payload)

	redacted, err := json.Marshal(payload)
	if err != nil {
		return stdout
	}
	return string(redacted)
}

func redactJSONEchoHeaders(payload any) {
	root, ok := payload.(map[string]any)
	if !ok {
		return
	}
	received, ok := root["received"].(map[string]any)
	if !ok {
		return
	}
	headers, ok := received["headers"].(map[string]any)
	if !ok {
		return
	}
	for name, value := range headers {
		if shouldRedactHeader(name) {
			headers[name] = redactedHeaderJSONValue(value)
		}
	}
}

func redactedHeaderJSONValue(value any) any {
	values, ok := value.([]any)
	if !ok {
		return redactedDiagnosticValue
	}
	redacted := make([]any, len(values))
	for i := range values {
		redacted[i] = redactedDiagnosticValue
	}
	return redacted
}

func truncateDiagnosticString(value string) string {
	const maxDiagnosticBytes = 32 * 1024
	if len(value) <= maxDiagnosticBytes {
		return value
	}
	return value[:maxDiagnosticBytes] + "\n...[truncated]"
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
			// #nosec G402 -- Intentionally weak in tests
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
	RedirectURL  string `json:"redirect_url"`
}

// CollectFailure runs Curl to fetch a URL that is expected to fail. The
// Curl JSON output is returned so the caller can inspect the failure to
// see whether it was what was expected.
func CollectFailure(cluster framework.Cluster, container, destination string, fn ...CollectResponsesOptsFn) (FailureResponse, error) {
	opts := CollectOptions(destination, fn...)
	cmd := collectCommand(opts, "curl",
		"--request", opts.Method,
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

<<<<<<< HEAD
	stdout, _, err := cluster.Exec(opts.namespace, appPodName, container, cmd...)
=======
	// Retry on empty stdout with no exec error: the Kubernetes SPDY exec stream can
	// close before stdout is delivered when the process exits immediately (e.g. DNS
	// failure, exit code 6). One retry is enough to recover from the race.
	var stdout string
	var stderr string
	var err error
	for range 2 {
		stdout, stderr, err = cluster.Exec(opts.namespace, appPodName, container, cmd...)
		if stdout != "" || err != nil {
			break
		}
	}
	execResult := collectResponseResult{
		cluster:     cluster,
		container:   container,
		destination: destination,
		opts:        opts,
		command:     cmd,
		namespace:   opts.namespace,
		podName:     appPodName,
		stdout:      stdout,
		stderr:      stderr,
		err:         err,
	}
>>>>>>> 62e80d799d (test(e2e): improve failure diagnostics (#17036))

	// 1. If we fail to decode the JSON status, return the JSON error,
	// but prefer the original error if we have it.
	empty := FailureResponse{}
	response := FailureResponse{}
	if jsonErr := json.Unmarshal([]byte(stdout), &response); jsonErr != nil {
		// Prefer the original error to a JSON decoding error.
		if err == nil {
			return response, withCollectDiagnostic("unmarshal-error", execResult, jsonErr)
		}
		return response, withCollectDiagnostic("exec-error", execResult, err)
	}

	// 2. If there was no error response, we still prefer the original
	// error, but fall back to reporting that the JSON  is missing.
	if response == empty {
		if err != nil {
			return response, withCollectDiagnostic("exec-error", execResult, err)
		}

		err := errors.Errorf("empty JSON response from curl: %q", stdout)
		return response, withCollectDiagnostic("empty-response", execResult, err)
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

func CollectResponsesAndFailures(
	cluster framework.Cluster,
	container, destination string,
	fn ...CollectResponsesOptsFn,
) ([]FailureResponse, error) {
	res, err := callConcurrently(destination, func() (interface{}, error) {
		return CollectFailure(cluster, container, destination, fn...)
	}, fn...)
	if err != nil {
		return nil, err
	}
	responses := make([]FailureResponse, len(res))
	for i := range res {
		responses[i] = res[i].(FailureResponse)
	}
	return responses, nil
}

func CollectResponses(cluster framework.Cluster, source, destination string, fn ...CollectResponsesOptsFn) ([]types.EchoResponse, error) {
	res, err := callConcurrently(destination, func() (interface{}, error) {
		return CollectEchoResponse(cluster, source, destination, fn...)
	}, fn...)
	if err != nil {
		return nil, err
	}
	responses := make([]types.EchoResponse, len(res))
	for i := range res {
		responses[i] = res[i].(types.EchoResponse)
	}
	return responses, nil
}

type result struct {
	idx uint
	res interface{}
	err error
}

func callConcurrently(destination string, call func() (interface{}, error), fn ...CollectResponsesOptsFn) ([]interface{}, error) {
	opts := CollectOptions(destination, fn...)
	var responses []interface{}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	inJobs := make(chan result, opts.numberOfRequests)
	results := make(chan result, opts.numberOfRequests)
	for i := uint(0); i < opts.maxConcurrentRequests; i++ {
		// #nosec G404 - math rand is enough
		delay := time.Duration(rand.Intn(int(opts.maxConcurrentRequestDelayMs))) * time.Millisecond
		go func() {
			for {
				select {
				case res, ok := <-inJobs:
					if !ok {
						return
					}
					res.res, res.err = call()
					// delay between requests
					time.Sleep(delay)
					results <- res
				case <-ctx.Done():
					return
				}
			}
		}()
	}
	for i := uint(0); i < opts.numberOfRequests; i++ {
		inJobs <- result{idx: i}
	}
	close(inJobs)
	for i := uint(0); i < opts.numberOfRequests; i++ {
		res := <-results
		if res.err != nil {
<<<<<<< HEAD
			framework.Logf("got error", "idx", res.idx, "err", res.err)
			return nil, res.err
=======
			framework.Logf("got error idx: %d err: %v", res.idx, res.err)
			addConcurrentDiagnostic(destination, opts, res, responses)
			return nil, fmt.Errorf(
				"request %d/%d failed after %d successful responses: %w",
				res.idx+1,
				opts.numberOfRequests,
				len(responses),
				res.err,
			)
>>>>>>> 62e80d799d (test(e2e): improve failure diagnostics (#17036))
		}
		responses = append(responses, res.res)
	}
	return responses, nil
}

type concurrentDiagnostic struct {
	Timestamp               time.Time         `json:"timestamp"`
	Destination             string            `json:"destination"`
	URL                     string            `json:"url"`
	Method                  string            `json:"method"`
	NumberOfRequests        uint              `json:"numberOfRequests"`
	MaxConcurrentRequests   uint              `json:"maxConcurrentRequests"`
	CompletedResponses      int               `json:"completedResponses"`
	FailedIndex             uint              `json:"failedIndex"`
	Error                   string            `json:"error"`
	FailedRequestDiagnostic string            `json:"failedRequestDiagnostic,omitempty"`
	PartialResponses        []any             `json:"partialResponses,omitempty"`
	Headers                 map[string]string `json:"headers,omitempty"`
	Flags                   []string          `json:"flags,omitempty"`
	InstanceHistogram       map[string]int    `json:"instanceHistogram,omitempty"`
	ResponseCodeHistogram   map[int]int       `json:"responseCodeHistogram,omitempty"`
}

func diagnosticReference(err error) string {
	var diagnosticErr diagnosticError
	if errors.As(err, &diagnosticErr) {
		return diagnosticErr.file
	}
	return ""
}

func buildConcurrentDiagnostic(destination string, opts CollectResponsesOpts, failed result, responses []any) concurrentDiagnostic {
	reference := diagnosticReference(failed.err)
	diagnostic := concurrentDiagnostic{
		Timestamp:               time.Now().UTC(),
		Destination:             destination,
		NumberOfRequests:        opts.numberOfRequests,
		MaxConcurrentRequests:   opts.maxConcurrentRequests,
		CompletedResponses:      len(responses),
		FailedIndex:             failed.idx,
		Error:                   failed.err.Error(),
		FailedRequestDiagnostic: reference,
		PartialResponses:        redactedDiagnosticResponses(responses),
		InstanceHistogram:       instanceHistogram(responses),
		ResponseCodeHistogram:   responseCodeHistogram(responses),
	}
	if reference == "" {
		diagnostic.URL = opts.URL
		diagnostic.Method = opts.Method
		diagnostic.Headers = redactedHeaders(opts.Headers)
		diagnostic.Flags = opts.Flags
	}
	return diagnostic
}

func addConcurrentDiagnostic(destination string, opts CollectResponsesOpts, failed result, responses []any) {
	diagnostic := buildConcurrentDiagnostic(destination, opts, failed, responses)
	data, err := json.MarshalIndent(diagnostic, "", "  ")
	if err != nil {
		framework.Logf("failed to marshal concurrent diagnostic: %v", err)
		return
	}
	report.AddFileToReportEntry(
		diagnosticFileName("client-concurrent"),
		data,
	)
}

func instanceHistogram(responses []any) map[string]int {
	histogram := map[string]int{}
	for _, response := range responses {
		echo, ok := response.(types.EchoResponse)
		if ok && echo.Instance != "" {
			histogram[echo.Instance]++
		}
	}
	if len(histogram) == 0 {
		return nil
	}
	return histogram
}

func responseCodeHistogram(responses []any) map[int]int {
	histogram := map[int]int{}
	for _, response := range responses {
		failure, ok := response.(FailureResponse)
		if ok {
			histogram[failure.ResponseCode]++
		}
	}
	if len(histogram) == 0 {
		return nil
	}
	return histogram
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
