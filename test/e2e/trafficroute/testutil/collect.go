package testutil

import (
	"fmt"
	"strings"

	"github.com/kumahq/kuma/test/framework"
)

type CollectResponsesOpts struct {
	NumberOfRequests int
	Method           string
	Headers          map[string]string
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

func CollectResponses(cluster framework.Cluster, source, destination string, fn ...CollectResponsesOptsFn) (map[string]int, error) {
	opts := DefaultCollectResponsesOpts()
	for _, f := range fn {
		f(&opts)
	}

	responses := map[string]int{}
	for i := 0; i < opts.NumberOfRequests; i++ {
		cmd := []string{"curl", "-X" + opts.Method}
		for key, value := range opts.Headers {
			cmd = append(cmd, fmt.Sprintf("-H'%s: %s'", key, value))
		}
		cmd = append(cmd, "-m", "3", "--fail", destination)
		stdout, _, err := cluster.ExecWithRetries("", "", source, cmd...)
		if err != nil {
			return nil, err
		}
		responses[strings.TrimSpace(stdout)]++
	}
	return responses, nil
}
