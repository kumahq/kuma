package generate

import (
	"time"

	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/apimachinery/pkg/util/validation"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/v2/pkg/core/resources/apis/mesh"
)

// dpContribution computes the non-system labels for an auto-generated
// MeshService from a single DataplaneResource and its inbounds.
// It never mutates its inputs and always returns a non-nil map.
func dpContribution(
	dp *core_mesh.DataplaneResource,
	inbounds []*mesh_proto.Dataplane_Networking_Inbound,
	allowSet map[string]struct{},
	metric prometheus.Counter,
	log logr.Logger,
) map[string]string {
	out := map[string]string{}
	if dp == nil {
		return out
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
			metric.Inc()
			log.V(1).Info("dropping inbound tag — intra-DP disagreement", "key", k, "dp", dp.Meta.GetName())
			continue
		}
		if allowSet != nil {
			if _, ok := allowSet[k]; !ok {
				metric.Inc()
				log.V(1).Info("dropping inbound tag not in allow-list", "key", k, "dp", dp.Meta.GetName())
				continue
			}
		}
		if errs := validation.IsQualifiedName(k); len(errs) > 0 {
			metric.Inc()
			log.V(1).Info("dropping invalid label key", "key", k, "dp", dp.Meta.GetName())
			continue
		}
		if errs := validation.IsValidLabelValue(e.value); len(errs) > 0 {
			metric.Inc()
			log.V(1).Info("dropping invalid label value", "key", k, "value", e.value, "dp", dp.Meta.GetName())
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
				metric.Inc()
				log.V(1).Info("dropping DP label not in allow-list", "key", k, "dp", dp.Meta.GetName())
				continue
			}
		}
		if errs := validation.IsQualifiedName(k); len(errs) > 0 {
			metric.Inc()
			log.V(1).Info("dropping invalid DP label key", "key", k, "dp", dp.Meta.GetName())
			continue
		}
		if errs := validation.IsValidLabelValue(v); len(errs) > 0 {
			metric.Inc()
			log.V(1).Info("dropping invalid DP label value", "key", k, "value", v, "dp", dp.Meta.GetName())
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
