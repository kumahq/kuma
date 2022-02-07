package client

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path"
	"sync"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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
func Resolve(host string, port int, address string) CollectResponsesOptsFn {
	return func(opts *CollectResponsesOpts) {
		opts.Flags = append(opts.Flags,
			"--resolve",
			fmt.Sprintf("%s:%d:%s", host, port, address),
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

func collectOptions(requestURL string, options ...CollectResponsesOptsFn) CollectResponsesOpts {
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

func CollectResponse(
	cluster framework.Cluster,
	container string,
	destination string,
	fn ...CollectResponsesOptsFn,
) (types.EchoResponse, error) {
	opts := collectOptions(destination, fn...)
	cmd := collectCommand(opts, "curl",
		"--request", opts.Method,
		"--max-time", "3",
		opts.ShellEscaped(opts.URL),
	)

	var pod string
	if opts.namespace != "" && opts.application != "" {
		pods, err := k8s.ListPodsE(
			cluster.GetTesting(),
			cluster.GetKubectlOptions(opts.namespace),
			metav1.ListOptions{
				LabelSelector: fmt.Sprintf("app=%s", opts.application),
			},
		)
		if err != nil {
			return types.EchoResponse{}, errors.Wrap(err, "failed to list pods")
		}

		pod = pods[0].Name
	}

	stdout, _, err := cluster.ExecWithRetries(opts.namespace, pod, container, cmd...)
	if err != nil {
		return types.EchoResponse{}, err
	}

	response := &types.EchoResponse{}
	if err := json.Unmarshal([]byte(stdout), response); err != nil {
		return types.EchoResponse{}, errors.Wrapf(err, "failed to unmarshal response: %q", stdout)
	}

	return *response, nil
}

// FailureResponse is the JSON output for a Curl command. Note that the available
// fields depend on the Curl version, which must be at least 7.70.0 for this feature.
//
// See https://curl.se/docs/manpage.html#-w.
type FailureResponse struct {
	Errormsg string `json:"errormsg"`
	Exitcode int    `json:"exitcode"`

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
func CollectFailure(cluster framework.Cluster, source, destination string, fn ...CollectResponsesOptsFn) (FailureResponse, error) {
	opts := collectOptions(destination, fn...)
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

	var pod string
	if opts.namespace != "" && opts.application != "" {
		pods, err := k8s.ListPodsE(
			cluster.GetTesting(),
			cluster.GetKubectlOptions(opts.namespace),
			metav1.ListOptions{
				LabelSelector: fmt.Sprintf("app=%s", opts.application),
			},
		)
		if err != nil {
			return FailureResponse{}, errors.Wrap(err, "failed to list pods")
		}

		pod = pods[0].Name
	}

	stdout, _, err := cluster.Exec(opts.namespace, pod, source, cmd...)

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

	// 3. Finally, report the JSON status and no execution error
	// since the JSON contains all the Curl error information.
	return response, nil
}

func CollectResponses(cluster framework.Cluster, source, destination string, fn ...CollectResponsesOptsFn) ([]types.EchoResponse, error) {
	opts := collectOptions(destination, fn...)

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
