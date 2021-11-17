package testutil

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/server/types"
)

type CollectResponsesOpts struct {
	NumberOfRequests int
	Method           string
	Headers          map[string]string

	Flags []string
}

func DefaultCollectResponsesOpts() CollectResponsesOpts {
	return CollectResponsesOpts{
		NumberOfRequests: 10,
		Method:           "GET",
		Headers:          map[string]string{},
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

func CollectResponse(cluster framework.Cluster, source, destination string, fn ...CollectResponsesOptsFn) (types.EchoResponse, error) {
	opts := DefaultCollectResponsesOpts()
	for _, f := range fn {
		f(&opts)
	}
	cmd := []string{
		"curl",
		"--fail",
		"--request", opts.Method,
		"--max-time", "3",
	}
	for key, value := range opts.Headers {
		cmd = append(cmd, "--header", ShellEscape(fmt.Sprintf("%s: %s", key, value)))
	}
	cmd = append(cmd, opts.Flags...)
	cmd = append(cmd, ShellEscape(destination))
	stdout, _, err := cluster.ExecWithRetries("", "", source, cmd...)
	if err != nil {
		return types.EchoResponse{}, err
	}
	response := &types.EchoResponse{}
	if err := json.Unmarshal([]byte(stdout), response); err != nil {
		return types.EchoResponse{}, err
	}
	return *response, nil
}

func ShellEscape(arg string) string {
	return fmt.Sprintf("'%s'", strings.ReplaceAll(arg, "'", "\\'"))
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
	opts := DefaultCollectResponsesOpts()
	for _, f := range fn {
		f(&opts)
	}

	cmd := []string{
		"curl",
		"--request", opts.Method,
		"--max-time", "3",
		"--silent",               // Suppress human-readable errors.
		"--write-out", "%{json}", // Write JSON result. Requires curl 7.70.0, April 2020.
		// Silence output so that we don't try to parse it. A future refactor could try to address this
		// by using "%{stderr}%{json}", but that needs a bit more investigation.
		"--output", os.DevNull,
	}

	for key, value := range opts.Headers {
		cmd = append(cmd, "--header", ShellEscape(fmt.Sprintf("%s: %s", key, value)))
	}

	cmd = append(cmd, ShellEscape(destination))
	stdout, _, err := cluster.Exec("", "", source, cmd...)

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
	opts := DefaultCollectResponsesOpts()
	for _, f := range fn {
		f(&opts)
	}

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
