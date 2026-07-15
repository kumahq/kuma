package tunnel

import (
	"encoding/json"
	"net/url"
	"regexp"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/v3/test/framework/envoy_admin"
	"github.com/kumahq/kuma/v3/test/framework/envoy_admin/clusters"
	"github.com/kumahq/kuma/v3/test/framework/envoy_admin/config_dump"
	"github.com/kumahq/kuma/v3/test/framework/envoy_admin/stats"
)

// CPTunnel reads a dataplane's Envoy admin state through the control plane's
// inspect API (GET {cp}/meshes/{mesh}/dataplanes/{name}/{stats,xds,clusters},
// or the /zoneingresses, /zoneegresses variants) instead of reaching the
// sidecar admin endpoint directly. The CP fetches from the proxy over its
// own mTLS admin channel; it may reshape some responses (config dump is
// sanitized), but the shapes still match what the stats/clusters/config_dump
// parsers expect.
//
// This is the K8s replacement for the removed kuma-dp readiness admin proxy:
// Envoy admin now lives only on a Unix socket in the sidecar, unreachable
// over the pod network, and the CP is the supported channel to inspect it.
type CPTunnel struct {
	// get performs an authenticated GET against the CP API server for the
	// given inspect path (already prefixed for the target resource) and
	// query, returning the raw response body.
	get func(inspectionPath string, query url.Values) ([]byte, error)
}

var _ envoy_admin.Tunnel = &CPTunnel{}

func NewCPTunnel(get func(inspectionPath string, query url.Values) ([]byte, error)) *CPTunnel {
	return &CPTunnel{get: get}
}

// GetStats returns the Envoy stats matching name. The CP inspect API does not
// accept Envoy's `filter` parameter, so we fetch all stats as JSON (including
// unused ones, to preserve the "expect zero" assertions) and filter by name
// client-side, mirroring Envoy's partial-regex match on the stat name.
func (t *CPTunnel) GetStats(name string) (*stats.Stats, error) {
	body, err := t.get("stats", url.Values{"format": {"json"}, "usedonly": {"false"}})
	if err != nil {
		return nil, err
	}

	var all stats.Stats
	if err := json.Unmarshal(body, &all); err != nil {
		return nil, err
	}

	re, err := regexp.Compile(name)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid stats filter %q", name)
	}

	filtered := &stats.Stats{}
	for _, item := range all.Stats {
		if re.MatchString(item.Name) {
			filtered.Stats = append(filtered.Stats, item)
		}
	}

	return filtered, nil
}

func (t *CPTunnel) GetClusters() (*clusters.Clusters, error) {
	body, err := t.get("clusters", url.Values{"format": {"json"}})
	if err != nil {
		return nil, err
	}

	var c clusters.Clusters
	if err := json.Unmarshal(body, &c); err != nil {
		return nil, err
	}

	return &c, nil
}

func (t *CPTunnel) GetConfigDump() (*config_dump.EnvoyConfig, error) {
	body, err := t.get("xds", url.Values{"include_eds": {"true"}})
	if err != nil {
		return nil, err
	}

	return config_dump.ParseEnvoyConfig(body)
}

// ResetCounters is not supported through the CP inspect API: /reset_counters
// is a mutating Envoy admin endpoint the control plane deliberately does not
// proxy. Tests that need it must assert on counter deltas instead of zeroing.
func (t *CPTunnel) ResetCounters() error {
	return errors.New("ResetCounters is not available via the control plane inspect API; " +
		"assert on a before/after counter delta instead")
}
