package metrics

import (
	"regexp"
	"slices"
	"strings"

	io_prometheus_client "github.com/prometheus/client_model/go"

	"github.com/kumahq/kuma/pkg/plugins/policies/meshmetric/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/xds/envoy/names"
)

type (
	selectorFunction = func(value string) bool
)

var neverSelect = func(value string) bool {
	return false
}

var alwaysSelect = func(value string) bool {
	return true
}

var basicProfile = []selectorFunction{
	// start of golden signals
	selectorToFilterFunction(v1alpha1.Selector{
		Type:  v1alpha1.ContainsSelectorType,
		Match: "rq_time",
	}),
	selectorToFilterFunction(v1alpha1.Selector{
		Type:  v1alpha1.ContainsSelectorType,
		Match: "cx_length_ms",
	}),
	selectorToFilterFunction(v1alpha1.Selector{
		Type:  v1alpha1.ContainsSelectorType,
		Match: "cx_count",
	}),
	selectorToFilterFunction(v1alpha1.Selector{
		Type:  v1alpha1.ContainsSelectorType,
		Match: "_rq",
	}),
	selectorToFilterFunction(v1alpha1.Selector{
		Type:  v1alpha1.ContainsSelectorType,
		Match: "bytes",
	}),
	selectorToFilterFunction(v1alpha1.Selector{
		Type:  v1alpha1.ContainsSelectorType,
		Match: "timeout",
	}),
	selectorToFilterFunction(v1alpha1.Selector{
		Type:  v1alpha1.ContainsSelectorType,
		Match: "health_check",
	}),
	selectorToFilterFunction(v1alpha1.Selector{
		Type:  v1alpha1.ContainsSelectorType,
		Match: "lb_healthy_panic",
	}),
	selectorToFilterFunction(v1alpha1.Selector{
		Type:  v1alpha1.ContainsSelectorType,
		Match: "cx_destroy",
	}),
	selectorToFilterFunction(v1alpha1.Selector{
		Type:  v1alpha1.ExactSelectorType,
		Match: "envoy_cluster_membership_degraded",
	}),
	selectorToFilterFunction(v1alpha1.Selector{
		Type:  v1alpha1.ExactSelectorType,
		Match: "envoy_cluster_membership_healthy",
	}),
	selectorToFilterFunction(v1alpha1.Selector{
		Type:  v1alpha1.ContainsSelectorType,
		Match: "error",
	}),
	selectorToFilterFunction(v1alpha1.Selector{
		Type:  v1alpha1.ContainsSelectorType,
		Match: "fail",
	}),
	selectorToFilterFunction(v1alpha1.Selector{
		Type:  v1alpha1.ContainsSelectorType,
		Match: "reset",
	}),
	selectorToFilterFunction(v1alpha1.Selector{
		Type:  v1alpha1.ContainsSelectorType,
		Match: "eject",
	}),
	selectorToFilterFunction(v1alpha1.Selector{
		Type:  v1alpha1.ContainsSelectorType,
		Match: "overflow",
	}),
	selectorToFilterFunction(v1alpha1.Selector{
		Type:  v1alpha1.ContainsSelectorType,
		Match: "cancelled",
	}),
	selectorToFilterFunction(v1alpha1.Selector{
		Type:  v1alpha1.ContainsSelectorType,
		Match: "max_duration_reached",
	}),
	selectorToFilterFunction(v1alpha1.Selector{
		Type:  v1alpha1.ContainsSelectorType,
		Match: "no_cluster",
	}),
	selectorToFilterFunction(v1alpha1.Selector{
		Type:  v1alpha1.ContainsSelectorType,
		Match: "no_route",
	}),
	selectorToFilterFunction(v1alpha1.Selector{
		Type:  v1alpha1.ContainsSelectorType,
		Match: "no_filter_chain",
	}),
	selectorToFilterFunction(v1alpha1.Selector{
		Type:  v1alpha1.ContainsSelectorType,
		Match: "reject",
	}),
	selectorToFilterFunction(v1alpha1.Selector{
		Type:  v1alpha1.ContainsSelectorType,
		Match: "denied",
	}),
	selectorToFilterFunction(v1alpha1.Selector{
		Type:  v1alpha1.ContainsSelectorType,
		Match: "days_until_first_cert_expiring",
	}),
	// end of golden signals
	// start of dashboards
	selectorToFilterFunction(v1alpha1.Selector{
		Type:  v1alpha1.PrefixSelectorType,
		Match: "envoy_server",
	}),
	selectorToFilterFunction(v1alpha1.Selector{
		Type:  v1alpha1.PrefixSelectorType,
		Match: "envoy_control_plane",
	}),
	selectorToFilterFunction(v1alpha1.Selector{
		Type:  v1alpha1.ContainsSelectorType,
		Match: "rbac",
	}),
	selectorToFilterFunction(v1alpha1.Selector{
		Type:  v1alpha1.ContainsSelectorType,
		Match: "cx_active",
	}),
	selectorToFilterFunction(v1alpha1.Selector{
		Type:  v1alpha1.ContainsSelectorType,
		Match: "cx_connect",
	}),
	selectorToFilterFunction(v1alpha1.Selector{
		Type:  v1alpha1.ExactSelectorType,
		Match: "envoy_cluster_ssl_handshake",
	}),
	selectorToFilterFunction(v1alpha1.Selector{
		Type:  v1alpha1.ExactSelectorType,
		Match: "envoy_cluster_membership_total",
	}),
	// end of dashboards
	// start of dns stats
	selectorToFilterFunction(v1alpha1.Selector{
		Type:  v1alpha1.PrefixSelectorType,
		Match: "envoy_dns_filter",
	}),
	selectorToFilterFunction(v1alpha1.Selector{
		Type:  v1alpha1.PrefixSelectorType,
		Match: "envoy_dns_cares",
	}),
	// end of dns stats
}

var basicProfileLabels = []selectorFunction{
	selectorToFilterFunction(v1alpha1.Selector{
		Type:  v1alpha1.PrefixSelectorType,
		Match: names.GetInternalClusterNamePrefix(),
	}),
	selectorToFilterFunction(v1alpha1.Selector{
		Type:  v1alpha1.ExactSelectorType,
		Match: names.GetAdsClusterName(),
	}),
	selectorToFilterFunction(v1alpha1.Selector{
		Type:  v1alpha1.ExactSelectorType,
		Match: names.GetAccessLogSinkClusterName(),
	}),
	selectorToFilterFunction(v1alpha1.Selector{
		Type:  v1alpha1.ExactSelectorType,
		Match: names.GetEnvoyAdminClusterName(),
	}),
	selectorToFilterFunction(v1alpha1.Selector{
		Type:  v1alpha1.ExactSelectorType,
		Match: names.GetMetricsHijackerClusterName(),
	}),
	selectorToFilterFunction(v1alpha1.Selector{
		Type:  v1alpha1.PrefixSelectorType,
		Match: names.GetOpenTelemetryClusterPrefix(),
	}),
	selectorToFilterFunction(v1alpha1.Selector{
		Type:  v1alpha1.PrefixSelectorType,
		Match: names.GetTracingClusterPrefix(),
	}),
}

var profileOrder = map[v1alpha1.ProfileName]int{
	v1alpha1.AllProfileName:   0,
	v1alpha1.BasicProfileName: 1,
	v1alpha1.NoneProfileName:  2,
}

func ProfileMutatorGenerator(sidecar *v1alpha1.Sidecar) PrometheusMutator {
	// setup default Basic
	effectiveSelectors := basicProfile
	effectiveLabelsSelectors := basicProfileLabels
	effectiveProfile := v1alpha1.BasicProfileName

	if sidecar != nil && sidecar.Profiles != nil && sidecar.Profiles.AppendProfiles != nil && len(*sidecar.Profiles.AppendProfiles) > 0 {
		effectiveSelectors = []selectorFunction{}
		effectiveLabelsSelectors = []selectorFunction{}

		var profiles []string
		for _, profile := range *sidecar.Profiles.AppendProfiles {
			profiles = append(profiles, string(profile.Name))
		}

		// this sorting is done to pick the most broad profile, so we do the minimal work that's needed
		slices.SortFunc(profiles, func(a, b string) int {
			return profileOrder[v1alpha1.ProfileName(a)] - profileOrder[v1alpha1.ProfileName(b)]
		})
		for _, profile := range profiles {
			switch profile {
			case string(v1alpha1.AllProfileName):
				effectiveSelectors = append(effectiveSelectors, alwaysSelect)
			case string(v1alpha1.NoneProfileName):
				effectiveSelectors = append(effectiveSelectors, neverSelect)
			case string(v1alpha1.BasicProfileName):
				effectiveSelectors = append(effectiveSelectors, basicProfile...)
				effectiveLabelsSelectors = append(effectiveSelectors, basicProfileLabels...)
			}
		}

		logger.V(1).Info("selected profiles", "names", profiles)
		// now that profiles contain each other (All contains Basic contains None) we can skip doing additional work
		// by just selecting the most broad profile, if this changes in the future the algorithm must change!
		effectiveProfile = v1alpha1.ProfileName(profiles[0])
	}

	hasInclude := sidecar != nil && sidecar.Profiles != nil && sidecar.Profiles.Include != nil
	hasExclude := sidecar != nil && sidecar.Profiles != nil && sidecar.Profiles.Exclude != nil

	return func(in map[string]*io_prometheus_client.MetricFamily) error {
		for key, metricFamily := range in {
			include := false
			for _, selector := range effectiveSelectors {
				if selector(*metricFamily.Name) {
					include = true
					break
				}
			}

			if hasExclude {
				for _, selector := range *sidecar.Profiles.Exclude {
					if selectorToFilterFunction(selector)(*metricFamily.Name) {
						include = false
						break
					}
				}
			}

			if hasInclude {
				for _, selector := range *sidecar.Profiles.Include {
					if selectorToFilterFunction(selector)(*metricFamily.Name) {
						include = true
						break
					}
				}
			}

			// filter out internal clusters
			// only activate this on basic profile so there is no overhead on other profiles
			if effectiveProfile == v1alpha1.BasicProfileName {
				var metrics []*io_prometheus_client.Metric
				for _, m := range metricFamily.Metric {
					includeMetric := true
					for _, l := range m.Label {
						if metricFromInternalCluster(l, effectiveLabelsSelectors) {
							includeMetric = false
							break
						}
					}
					if includeMetric {
						metrics = append(metrics, m)
					}
				}

				// if after cleaning there is no metrics remove this metric family
				if len(metrics) == 0 {
					include = false
				} else {
					metricFamily.Metric = metrics
				}
			}

			if !include {
				delete(in, key)
			}
		}

		return nil
	}
}

func metricFromInternalCluster(l *io_prometheus_client.LabelPair, effectiveLabelsSelectors []selectorFunction) bool {
	for _, selector := range effectiveLabelsSelectors {
		if l.GetName() == EnvoyClusterLabelName && selector(l.GetValue()) {
			return true
		}
	}
	return false
}

func selectorToFilterFunction(selector v1alpha1.Selector) selectorFunction {
	switch selector.Type {
	case v1alpha1.ContainsSelectorType:
		return func(value string) bool {
			return strings.Contains(value, selector.Match)
		}
	case v1alpha1.PrefixSelectorType:
		return func(value string) bool {
			return strings.HasPrefix(value, selector.Match)
		}
	case v1alpha1.RegexSelectorType:
		compiled, err := regexp.Compile(selector.Match)
		if err != nil {
			// validation prevents compilation errors
			return neverSelect
		}
		return func(value string) bool {
			return compiled.MatchString(value)
		}
	case v1alpha1.ExactSelectorType:
		return func(value string) bool {
			return value == selector.Match
		}
	}
	return neverSelect
}
