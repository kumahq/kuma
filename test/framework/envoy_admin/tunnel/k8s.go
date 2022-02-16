package tunnel

import (
	"encoding/json"
	"fmt"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/shell"
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

func (t *K8sTunnel) GetStats(name string) (stats.Stats, error) {
	url := fmt.Sprintf("%s/stats?format=json&filter=%s", t.Endpoint(), name)

	cmd := shell.Command{
		Command: "curl",
		Args:    []string{"--silent", url},
	}

	output, err := shell.RunCommandAndGetOutputE(t.t, cmd)
	if err != nil {
		return stats.Stats{}, err
	}

	var s stats.Stats
	if err := json.Unmarshal([]byte(output), &s); err != nil {
		return stats.Stats{}, err
	}

	return s, nil
}
