package tunnel

import (
	"encoding/json"

	"github.com/gruntwork-io/terratest/modules/testing"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/test/framework/envoy_admin"
	"github.com/kumahq/kuma/test/framework/envoy_admin/stats"
	"github.com/kumahq/kuma/test/framework/ssh"
)

type UniversalTunnel struct {
	t testing.TestingT

	port string
}

func NewUniversalEnvoyAdminTunnel(t testing.TestingT, port string) (envoy_admin.Tunnel, error) {
	return &UniversalTunnel{
		t:    t,
		port: port,
	}, nil
}

func (t *UniversalTunnel) GetStats(name string) (stats.Stats, error) {
	sshArgs := []string{
		"curl", "--silent", "--max-time", "3",
		"localhost:9901/stats?format=json",
	}

	app := ssh.NewApp(false, t.port, nil, sshArgs)

	if err := app.Run(); err != nil {
		return stats.Stats{}, err
	}

	if app.Err() != "" {
		return stats.Stats{}, errors.New(app.Err())
	}

	var s stats.Stats
	if err := json.Unmarshal([]byte(app.Out()), &s); err != nil {
		return stats.Stats{}, err
	}

	var filtered []stats.StatItem
	for _, stat := range s.Stats {
		if stat.Name == name {
			filtered = append(filtered, stat)
		}
	}

	return stats.Stats{Stats: filtered}, nil
}
