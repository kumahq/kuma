package generate

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/apimachinery/pkg/util/validation"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/v2/pkg/core/resources/apis/mesh"
)

// propagationTrackingPrefix is the prefix for labels that record which
// non-system keys were written by the previous propagation cycle. Used to
// distinguish previously-propagated labels (candidates for removal) from
// operator-managed labels (which must be preserved).
const propagationTrackingPrefix = "kuma.io/pkey-"

// trackingKeyHash returns an 8-hex-char fingerprint of labelKey. Storing the
// hash in the label key (instead of the key name in the label value) avoids the
// label-value character restriction: valid Kubernetes label keys can contain "/"
// (e.g. "foo.example.com/bar") which is not allowed in label values, causing
// such keys to be silently skipped and never cleaned up after DP removal.
func trackingKeyHash(labelKey string) string {
	h := sha256.Sum256([]byte(labelKey))
	return fmt.Sprintf("%x", h[:4]) // 8 hex chars, collision negligible for <1000 keys
}

// addPropagationTracking clears any existing tracking labels in labels and
// writes one tracking label per key in propagated. The key name is hashed into
// the label key suffix so that qualified names with "/" are handled correctly.
func addPropagationTracking(labels map[string]string, propagated map[string]string) {
	for k := range labels {
		if strings.HasPrefix(k, propagationTrackingPrefix) {
			delete(labels, k)
		}
	}
	for k := range propagated {
		labels[propagationTrackingPrefix+trackingKeyHash(k)] = ""
	}
}

// extractPropagatedKeys returns a function that reports whether a given label
// key was recorded as propagated in the previous reconcile cycle.
//
// Handles two formats for backward compatibility during upgrades:
//   - New format (current): kuma.io/pkey-<8hexchars> = ""  (key is hashed)
//   - Old format (pre-hash): kuma.io/pkey-N = "<label-key>"  (key stored as value)
func extractPropagatedKeys(labels map[string]string) func(string) bool {
	hashes := map[string]bool{}
	oldKeys := map[string]bool{}
	for k, v := range labels {
		if !strings.HasPrefix(k, propagationTrackingPrefix) {
			continue
		}
		if v == "" {
			// New format: suffix is a hash of the label key.
			hashes[k[len(propagationTrackingPrefix):]] = true
		} else {
			// Old format: value holds the label key directly.
			oldKeys[v] = true
		}
	}
	return func(key string) bool {
		return hashes[trackingKeyHash(key)] || oldKeys[key]
	}
}

// dpContribution computes the non-system labels for an auto-generated
// MeshService from a single DataplaneResource and its inbounds.
// It never mutates its inputs and always returns a non-nil map.
func dpContribution(
	dp *core_mesh.DataplaneResource,
	inbounds []*mesh_proto.Dataplane_Networking_Inbound,
	allowSet map[string]struct{},
	droppedLabels *prometheus.CounterVec,
	log logr.Logger,
	service string,
) map[string]string {
	out := map[string]string{}
	if dp == nil {
		return out
	}

	drop := func(reason, key string) {
		droppedLabels.WithLabelValues(reason).Inc()
		log.Info("dropping label during MeshService generation",
			"reason", reason,
			"service", service,
			"dataplane", dp.GetMeta().GetName(),
			"mesh", dp.GetMeta().GetMesh(),
			"key", key,
		)
	}

	debugDrop := func(reason, key string) {
		log.V(1).Info("dropping label during MeshService generation",
			"reason", reason,
			"service", service,
			"dataplane", dp.GetMeta().GetName(),
			"mesh", dp.GetMeta().GetMesh(),
			"key", key,
		)
	}

	// Step 1: cross-inbound consensus on non-reserved tags.
	type tagEntry struct {
		value string
		seen  bool
		drop  bool
	}
	consensus := map[string]*tagEntry{}
	for _, ib := range inbounds {
		for k, v := range ib.GetTags() {
			if mesh_proto.IsReservedLabelKey(k) {
				continue
			}
			e, ok := consensus[k]
			if !ok {
				consensus[k] = &tagEntry{value: v, seen: true}
				continue
			}
			if e.value != v {
				e.drop = true
			}
		}
	}

	// Step 2: validate and write consensus values.
	for k, e := range consensus {
		if e.drop {
			drop("inbound_conflict", k)
			continue
		}
		if allowSet != nil {
			if _, ok := allowSet[k]; !ok {
				debugDrop("not_allowed", k)
				continue
			}
		}
		if errs := validation.IsQualifiedName(k); len(errs) > 0 {
			drop("invalid", k)
			continue
		}
		if errs := validation.IsValidLabelValue(e.value); len(errs) > 0 {
			drop("invalid", k)
			continue
		}
		out[k] = e.value
	}

	// Step 3: DP resource labels overlay — runs same validation pipeline.
	for k, v := range dp.GetMeta().GetLabels() {
		if mesh_proto.IsReservedLabelKey(k) {
			continue
		}
		if allowSet != nil {
			if _, ok := allowSet[k]; !ok {
				debugDrop("not_allowed", k)
				continue
			}
		}
		if errs := validation.IsQualifiedName(k); len(errs) > 0 {
			drop("invalid", k)
			continue
		}
		if errs := validation.IsValidLabelValue(v); len(errs) > 0 {
			drop("invalid", k)
			continue
		}
		out[k] = v
	}

	return out
}

type valueStats struct {
	count  int
	newest time.Time
}

// mergeAcrossDataplanes merges per-DP label contributions into a single label
// map. Majority-count wins; ties broken by newest creation time, then lex
// value. Never mutates the input slice and always returns a non-nil map.
func mergeAcrossDataplanes(
	dps []*core_mesh.DataplaneResource,
	contribution func(*core_mesh.DataplaneResource) map[string]string,
) map[string]string {
	filtered := make([]*core_mesh.DataplaneResource, 0, len(dps))
	for _, dp := range dps {
		if dp != nil {
			filtered = append(filtered, dp)
		}
	}

	perKey := map[string]map[string]*valueStats{}
	for _, dp := range core_mesh.SortDataplanes(filtered) {
		ct := dp.GetMeta().GetCreationTime()
		for k, v := range contribution(dp) {
			byVal, ok := perKey[k]
			if !ok {
				byVal = map[string]*valueStats{}
				perKey[k] = byVal
			}
			stats, ok := byVal[v]
			if !ok {
				stats = &valueStats{}
				byVal[v] = stats
			}
			stats.count++
			if ct.After(stats.newest) {
				stats.newest = ct
			}
		}
	}

	out := make(map[string]string, len(perKey))
	for k, byVal := range perKey {
		var bestVal string
		var bestStats *valueStats
		for v, stats := range byVal {
			switch {
			case bestStats == nil:
				bestVal, bestStats = v, stats
			case stats.count > bestStats.count:
				bestVal, bestStats = v, stats
			case stats.count == bestStats.count && stats.newest.After(bestStats.newest):
				bestVal, bestStats = v, stats
			case stats.count == bestStats.count && stats.newest.Equal(bestStats.newest) && v < bestVal:
				bestVal, bestStats = v, stats
			}
		}
		out[k] = bestVal
	}
	return out
}
