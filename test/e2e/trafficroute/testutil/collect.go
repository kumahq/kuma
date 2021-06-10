package testutil

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/server/types"
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

func CollectResponse(cluster framework.Cluster, source, destination string, fn ...CollectResponsesOptsFn) (types.EchoResponse, error) {
	opts := DefaultCollectResponsesOpts()
	for _, f := range fn {
		f(&opts)
	}
	cmd := []string{"curl", "-X" + opts.Method}
	for key, value := range opts.Headers {
		cmd = append(cmd, fmt.Sprintf("-H'%s: %s'", key, value))
	}
	cmd = append(cmd, "-m", "3", "--fail", destination)
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
