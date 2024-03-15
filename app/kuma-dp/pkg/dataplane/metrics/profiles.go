package metrics

import (
	"regexp"
	"strings"

	io_prometheus_client "github.com/prometheus/client_model/go"

	"github.com/kumahq/kuma/pkg/plugins/policies/meshmetric/api/v1alpha1"
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
}

func ProfileMutatorGenerator(sidecar *v1alpha1.Sidecar) PrometheusMutator {
	effectiveSelectors := []selectorFunction{alwaysSelect} // default is All
	if sidecar != nil && sidecar.Profiles != nil && sidecar.Profiles.AppendProfiles != nil && len(*sidecar.Profiles.AppendProfiles) == 1 {
		profile := (*sidecar.Profiles.AppendProfiles)[0].Name
		switch profile {
		case v1alpha1.AllProfileName:
			effectiveSelectors = []selectorFunction{alwaysSelect}
		case v1alpha1.NoneProfileName:
			effectiveSelectors = []selectorFunction{neverSelect}
		case v1alpha1.BasicProfileName:
			effectiveSelectors = basicProfile
		}
		logger.Info("selected profile", "name", profile)
	}

	hasInclude := sidecar != nil && sidecar.Profiles != nil && sidecar.Profiles.Include != nil
	hasExclude := sidecar != nil && sidecar.Profiles != nil && sidecar.Profiles.Exclude != nil
	logger.Info("exclude/include", "exclude", hasExclude, "include", hasInclude)

	return func(in map[string]*io_prometheus_client.MetricFamily) error {
		logger.Info("inside mutator")
		for key, metricFamily := range in {
			include := false
			for _, selector := range effectiveSelectors {
				if selector(*metricFamily.Name) {
					include = true
					break
				}
			}

			if hasExclude {
				logger.Info("has exclude")
				for _, selector := range *sidecar.Profiles.Exclude {
					if selectorToFilterFunction(selector)(*metricFamily.Name) {
						include = false
						break
					}
				}
			}

			if hasInclude {
				logger.Info("has include")
				for _, selector := range *sidecar.Profiles.Include {
					if selectorToFilterFunction(selector)(*metricFamily.Name) {
						include = true
						break
					}
				}
			}

			if !include {
				delete(in, key)
			}
		}

		return nil
	}
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
