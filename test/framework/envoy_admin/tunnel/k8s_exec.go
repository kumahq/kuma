package tunnel

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/gruntwork-io/terratest/modules/testing"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/v2/test/framework/envoy_admin"
	"github.com/kumahq/kuma/v2/test/framework/envoy_admin/clusters"
	"github.com/kumahq/kuma/v2/test/framework/envoy_admin/config_dump"
	"github.com/kumahq/kuma/v2/test/framework/envoy_admin/stats"
)

// K8sExecTunnel accesses Envoy admin via kubectl exec + curl when admin
// is on a Unix domain socket (no TCP port to port-forward).
type K8sExecTunnel struct {
	t          testing.TestingT
	kubeconfig string
	namespace  string
	pod        string
	container  string
}

var _ envoy_admin.Tunnel = &K8sExecTunnel{}

func NewK8sExecEnvoyAdminTunnel(
	t testing.TestingT,
	kubeconfig string,
	namespace string,
	pod string,
	container string,
) envoy_admin.Tunnel {
	return &K8sExecTunnel{
		t:          t,
		kubeconfig: kubeconfig,
		namespace:  namespace,
		pod:        pod,
		container:  container,
	}
}

func (t *K8sExecTunnel) exec(flags, path string) (string, error) {
	curlCmd := adminCurlCmd(flags, path)

	args := []string{
		"--kubeconfig", t.kubeconfig,
		"-n", t.namespace,
		"exec", t.pod,
		"-c", t.container,
		"--", "sh", "-c", curlCmd,
	}

	cmd := exec.Command("kubectl", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.Wrapf(err, "kubectl exec failed: %s", strings.TrimSpace(string(out)))
	}

	return string(out), nil
}

func (t *K8sExecTunnel) GetStats(name string) (*stats.Stats, error) {
	stdout, err := t.exec(
		"-s --max-time 3 --fail",
		fmt.Sprintf("/stats?format=json&filter=%s", name),
	)
	if err != nil {
		return nil, err
	}

	var s stats.Stats
	if err := json.Unmarshal([]byte(stdout), &s); err != nil {
		return nil, err
	}

	return &s, nil
}

func (t *K8sExecTunnel) GetClusters() (*clusters.Clusters, error) {
	stdout, err := t.exec(
		"-s --max-time 3 --fail",
		"/clusters?format=json",
	)
	if err != nil {
		return nil, err
	}

	var c clusters.Clusters
	if err := json.Unmarshal([]byte(stdout), &c); err != nil {
		return nil, err
	}

	return &c, nil
}

func (t *K8sExecTunnel) GetConfigDump() (*config_dump.EnvoyConfig, error) {
	stdout, err := t.exec(
		"-s --max-time 3 --fail",
		"/config_dump?format=json",
	)
	if err != nil {
		return nil, err
	}

	return config_dump.ParseEnvoyConfig([]byte(stdout))
}

func (t *K8sExecTunnel) ResetCounters() error {
	_, err := t.exec(
		"-v --max-time 3 --fail -XPOST",
		"/reset_counters",
	)
	return err
}
