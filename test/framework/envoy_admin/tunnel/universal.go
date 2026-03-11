package tunnel

import (
	"encoding/json"
	"fmt"

	"github.com/kumahq/kuma/v2/test/framework/envoy_admin"
	"github.com/kumahq/kuma/v2/test/framework/envoy_admin/clusters"
	"github.com/kumahq/kuma/v2/test/framework/envoy_admin/config_dump"
	"github.com/kumahq/kuma/v2/test/framework/envoy_admin/stats"
)

type UniversalTunnel struct {
	remoteExec func(cmdName, cmd string) (string, error)
}

var _ envoy_admin.Tunnel = &UniversalTunnel{}

func NewUniversalEnvoyAdminTunnel(remoteExec func(cmdName, cmd string) (string, error)) envoy_admin.Tunnel {
	return &UniversalTunnel{
		remoteExec: remoteExec,
	}
}

// adminCurlCmd builds a curl command that auto-detects whether admin is on
// a Unix domain socket (kuma-envoy-admin.sock) or TCP port 9901.
func adminCurlCmd(flags, path string) string {
	return fmt.Sprintf(
		`sock=$(find /tmp -name kuma-envoy-admin.sock 2>/dev/null | head -1); `+
			`if [ -n "$sock" ]; then `+
			`curl %s --unix-socket "$sock" 'http://localhost%s'; `+
			`else `+
			`curl %s 'http://localhost:9901%s'; `+
			`fi`,
		flags, path, flags, path,
	)
}

// AdminCurlCmd returns a shell command that curls the Envoy admin API,
// auto-detecting UDS vs TCP. Use in Cluster.Exec calls from test code.
func AdminCurlCmd(path string) []string {
	return []string{
		"/bin/bash", "-c",
		fmt.Sprintf(
			`sock=$(find /tmp -name kuma-envoy-admin.sock 2>/dev/null | head -1); `+
				`if [ -n "$sock" ]; then `+
				`curl -s --max-time 5 --fail --unix-socket "$sock" 'http://localhost%s'; `+
				`else `+
				`curl -s --max-time 5 --fail 'http://localhost:9901%s'; `+
				`fi`,
			path, path,
		),
	}
}

func (t *UniversalTunnel) GetStats(name string) (*stats.Stats, error) {
	stdout, err := t.remoteExec(
		"getstats_"+name,
		adminCurlCmd("-s --max-time 3 --fail", fmt.Sprintf("/stats?format=json&filter=%s", name)),
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

func (t *UniversalTunnel) GetClusters() (*clusters.Clusters, error) {
	stdout, err := t.remoteExec(
		"getclusters",
		adminCurlCmd("-s --max-time 3 --fail", "/clusters?format=json"),
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

func (t *UniversalTunnel) GetConfigDump() (*config_dump.EnvoyConfig, error) {
	stdout, err := t.remoteExec(
		"getconfig_dump",
		adminCurlCmd("-s --max-time 3 --fail", "/config_dump?format=json"),
	)
	if err != nil {
		return nil, err
	}

	return config_dump.ParseEnvoyConfig([]byte(stdout))
}

func (t *UniversalTunnel) ResetCounters() error {
	_, err := t.remoteExec(
		"resetcounters",
		adminCurlCmd("-v --max-time 3 --fail -XPOST", "/reset_counters"),
	)
	return err
}
