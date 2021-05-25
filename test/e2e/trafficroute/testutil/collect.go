package testutil

import (
	"strings"

	"github.com/kumahq/kuma/test/framework"
)

const numberOrRequests = 10

func CollectResponses(cluster framework.Cluster, source, destination string, args ...int) (map[string]int, error) {
	responses := map[string]int{}
	attempts := numberOrRequests
	if len(args) > 0 {
		attempts = args[0]
	}
	for i := 0; i < attempts; i++ {
		stdout, _, err := cluster.ExecWithRetries("", "", source, "curl", "-m", "3", "--fail", destination)
		if err != nil {
			return nil, err
		}
		responses[strings.TrimSpace(stdout)]++
	}
	return responses, nil
}
