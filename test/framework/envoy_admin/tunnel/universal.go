package tunnel

import (
	"encoding/json"
	"fmt"

	"github.com/gruntwork-io/terratest/modules/testing"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/test/framework/envoy_admin"
	"github.com/kumahq/kuma/test/framework/envoy_admin/clusters"
	"github.com/kumahq/kuma/test/framework/envoy_admin/stats"
	"github.com/kumahq/kuma/test/framework/ssh"
)

type UniversalTunnel struct {
	t testing.TestingT

	port    string
	verbose bool
}

var _ envoy_admin.Tunnel = &UniversalTunnel{}

func NewUniversalEnvoyAdminTunnel(t testing.TestingT, port string, verbose bool) (envoy_admin.Tunnel, error) {
	return &UniversalTunnel{
		t:       t,
		port:    port,
		verbose: verbose,
	}, nil
}

func (t *UniversalTunnel) GetStats(name string) (*stats.Stats, error) {
	url := fmt.Sprintf("'http://localhost:9901/stats?format=json&filter=%s'", name)

	sshArgs := []string{
		"curl", "--silent", "--max-time", "3", "--fail", url,
	}

	app := ssh.NewApp("tunnel", "", t.verbose, t.port, nil, sshArgs)

	if err := app.Run(); err != nil {
		return nil, err
	}

	if app.Err() != "" {
		return nil, errors.New(app.Err())
	}

	var s stats.Stats
	if err := json.Unmarshal([]byte(app.Out()), &s); err != nil {
		return nil, err
	}

	return &s, nil
}

func (t *UniversalTunnel) GetClusters() (*clusters.Clusters, error) {
	url := "http://localhost:9901/clusters?format=json"

	sshArgs := []string{
		"curl", "--silent", "--max-time", "3", "--fail", url,
	}

	app := ssh.NewApp("tunnel", "", t.verbose, t.port, nil, sshArgs)

	if err := app.Run(); err != nil {
		return nil, err
	}

	if app.Err() != "" {
		return nil, errors.New(app.Err())
	}

	var c clusters.Clusters
	if err := json.Unmarshal([]byte(app.Out()), &c); err != nil {
		return nil, err
	}

	return &c, nil
}

func (t *UniversalTunnel) ResetCounters() error {
	sshArgs := []string{
		"curl", "--verbose", "--max-time", "3", "--fail", "-XPOST",
		"'http://localhost:9901/reset_counters'",
	}

	app := ssh.NewApp("tunnel", "", t.verbose, t.port, nil, sshArgs)

	if err := app.Run(); err != nil {
		return err
	}

	if app.Err() != "" {
		return errors.New(app.Err())
	}

	return nil
}
