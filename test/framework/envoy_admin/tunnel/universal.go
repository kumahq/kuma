package tunnel

import (
	"encoding/json"
	"fmt"

	"github.com/kumahq/kuma/test/framework/envoy_admin"
	"github.com/kumahq/kuma/test/framework/envoy_admin/clusters"
	"github.com/kumahq/kuma/test/framework/envoy_admin/stats"
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

func (t *UniversalTunnel) GetStats(name string) (*stats.Stats, error) {
	stdout, err := t.remoteExec("getstats_"+name, fmt.Sprintf("curl -s --max-time 3 --fail 'http://localhost:9901/stats?format=json&filter=%s'", name))
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
	stdout, err := t.remoteExec("getclusters", "curl -s --max-time 3 --fail http://localhost:9901/clusters?format=json")
	if err != nil {
		return nil, err
	}
	var c clusters.Clusters
	if err := json.Unmarshal([]byte(stdout), &c); err != nil {
		return nil, err
	}

	return &c, nil
}

func (t *UniversalTunnel) ResetCounters() error {
	_, err := t.remoteExec("resetcounters", "curl -v --max-time 3 --fail -XPOST 'http://localhost:9901/reset_counters'")
	if err != nil {
		return err
	}
	return nil
}
