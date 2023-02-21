package tunnel

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gruntwork-io/terratest/modules/testing"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/test/framework/envoy_admin"
	"github.com/kumahq/kuma/test/framework/envoy_admin/stats"
)

type K8sTunnel struct {
	t        testing.TestingT
	endpoint string
}

var _ envoy_admin.Tunnel = &K8sTunnel{}

func NewK8sEnvoyAdminTunnel(
	t testing.TestingT,
	endpoint string,
) (envoy_admin.Tunnel, error) {
	return &K8sTunnel{
		endpoint: endpoint,
		t:        t,
	}, nil
}

func (t *K8sTunnel) GetStats(name string) (*stats.Stats, error) {
	url := fmt.Sprintf("http://%s/stats?format=json&filter=%s", t.endpoint, name)

	response, err := http.Post(url, "application/json", nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, errors.Errorf(
			"got response with unexpected status code: %+q, Expected: %+q",
			response.Status,
			http.StatusText(http.StatusOK),
		)
	}

	var s stats.Stats
	if err := json.NewDecoder(response.Body).Decode(&s); err != nil {
		return nil, err
	}

	return &s, nil
}

func (t *K8sTunnel) ResetCounters() error {
	url := fmt.Sprintf("http://%s/reset_counters", t.endpoint)

	response, err := http.Post(url, "text", nil)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return errors.Errorf(
			"got response with unexpected status code: %+q, Expected: %+q",
			response.Status,
			http.StatusText(http.StatusOK),
		)
	}

	return nil
}
