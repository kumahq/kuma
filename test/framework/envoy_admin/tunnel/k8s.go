package tunnel

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/testing"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/test/framework/envoy_admin"
	"github.com/kumahq/kuma/test/framework/envoy_admin/stats"
	"github.com/kumahq/kuma/test/framework/utils"
)

type K8sTunnel struct {
	t testing.TestingT

	*k8s.Tunnel
}

var _ envoy_admin.Tunnel = &K8sTunnel{}

func NewK8sEnvoyAdminTunnel(
	t testing.TestingT,
	kubectlOptions *k8s.KubectlOptions,
	resourceType k8s.KubeResourceType,
	resourceName string,
) (envoy_admin.Tunnel, error) {
	port := 9901

	localPort, err := utils.GetFreePort()
	if err != nil {
		return nil, errors.Wrapf(err, "getting free port for the new tunnel failed")
	}

	tunnel := k8s.NewTunnel(kubectlOptions, resourceType, resourceName, localPort, port)

	if err := tunnel.ForwardPortE(t); err != nil {
		return nil, errors.Wrapf(err, "port forwarding for %d:%d failed", localPort, port)
	}

	return &K8sTunnel{
		Tunnel: tunnel,
		t:      t,
	}, nil
}

func (t *K8sTunnel) GetStats(name string) (*stats.Stats, error) {
	url := fmt.Sprintf("http://%s/stats?format=json&filter=%s", t.Endpoint(), name)

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
	url := fmt.Sprintf("http://%s/reset_counters", t.Endpoint())

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

func (t *K8sTunnel) Close() {
	t.Tunnel.Close()
}
