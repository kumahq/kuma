package metrics

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/common/expfmt"
	"github.com/prometheus/common/model"

	"github.com/kumahq/kuma/v2/pkg/plugins/policies/meshmetric/api/v1alpha1"
)

// genEnvoyScrape builds a realistic Envoy /stats/prometheus body.
// families: distinct metric family names
// series:   label-permutations per family (envoy_cluster_name like the prod fixtures)
func genEnvoyScrape(families, series int) []byte {
	var b strings.Builder
	prefixes := []string{
		"envoy_cluster_assignment_stale",
		"envoy_cluster_upstream_rq_total",
		"envoy_cluster_upstream_cx_active",
		"envoy_cluster_lb_healthy_panic",
		"envoy_http_downstream_rq_total",
		"envoy_http_downstream_cx_active",
		"envoy_listener_downstream_cx_total",
		"envoy_server_memory_allocated",
		"envoy_rbac_allowed",
		"envoy_rbac_denied",
	}
	clusters := []string{
		"access_log_sink", "ads_cluster",
		"echo-server_kuma-test_svc_8080-94bcdea309bceaff",
		"echo-server_kuma-test_svc_8080-0382aaf295add2eb",
		"echo-server_kuma-test_svc_8080-fe04929f294bcdea",
		"inbound_passthrough_ipv4", "outbound_passthrough_ipv4",
		"localhost_3000", "kuma_envoy_admin",
	}
	for i := range families {
		base := prefixes[i%len(prefixes)]
		name := fmt.Sprintf("%s_%d", base, i)
		fmt.Fprintf(&b, "# TYPE %s counter\n", name)
		for j := range series {
			cluster := clusters[j%len(clusters)]
			fmt.Fprintf(&b, "%s{envoy_cluster_name=\"%s_%d\"} %d\n", name, cluster, j, i*1000+j)
		}
		b.WriteString("\n")
	}
	return []byte(b.String())
}

func sidecarAllExcludeRegex(n int) *v1alpha1.Sidecar {
	excludes := make([]v1alpha1.Selector, n)
	for i := range n {
		// each regex is realistic: matches one family prefix
		excludes[i] = v1alpha1.Selector{Type: v1alpha1.RegexSelectorType, Match: fmt.Sprintf("^envoy_rbac_.*_%d$", i)}
	}
	return &v1alpha1.Sidecar{
		Profiles: &v1alpha1.Profiles{
			AppendProfiles: &[]v1alpha1.Profile{{Name: v1alpha1.AllProfileName}},
			Exclude:        &excludes,
		},
	}
}

func sidecarAllExcludeExact(n int) *v1alpha1.Sidecar {
	excludes := make([]v1alpha1.Selector, n)
	for i := range n {
		excludes[i] = v1alpha1.Selector{Type: v1alpha1.ExactSelectorType, Match: fmt.Sprintf("envoy_rbac_%d", i)}
	}
	return &v1alpha1.Sidecar{
		Profiles: &v1alpha1.Profiles{
			AppendProfiles: &[]v1alpha1.Profile{{Name: v1alpha1.AllProfileName}},
			Exclude:        &excludes,
		},
	}
}

func sidecarNoneIncludeRegex(n int) *v1alpha1.Sidecar {
	includes := make([]v1alpha1.Selector, n)
	for i := range n {
		includes[i] = v1alpha1.Selector{Type: v1alpha1.RegexSelectorType, Match: fmt.Sprintf("^envoy_cluster_.*_%d$", i)}
	}
	return &v1alpha1.Sidecar{
		Profiles: &v1alpha1.Profiles{
			AppendProfiles: &[]v1alpha1.Profile{{Name: v1alpha1.NoneProfileName}},
			Include:        &includes,
		},
	}
}

func sidecarBasic() *v1alpha1.Sidecar {
	return &v1alpha1.Sidecar{
		Profiles: &v1alpha1.Profiles{
			AppendProfiles: &[]v1alpha1.Profile{{Name: v1alpha1.BasicProfileName}},
		},
	}
}

func sidecarAll() *v1alpha1.Sidecar {
	return &v1alpha1.Sidecar{
		Profiles: &v1alpha1.Profiles{
			AppendProfiles: &[]v1alpha1.Profile{{Name: v1alpha1.AllProfileName}},
		},
	}
}

// shape:  ~500 families × 4 series ≈ 2k data points (matches a small sidecar)
// shape2: ~2000 families × 8 series ≈ 16k data points (matches a busy sidecar)
type shape struct {
	name     string
	families int
	series   int
}

var benchShapes = []shape{
	{"small_2k", 500, 4},
	{"large_16k", 2000, 8},
}

func BenchmarkProfileMutator(b *testing.B) {
	for _, sh := range benchShapes {
		body := genEnvoyScrape(sh.families, sh.series)
		b.Run(sh.name+"/all_no_filter", func(b *testing.B) {
			runProfileBench(b, body, sidecarAll())
		})
		b.Run(sh.name+"/basic", func(b *testing.B) {
			runProfileBench(b, body, sidecarBasic())
		})
		for _, n := range []int{1, 10, 50} {
			b.Run(fmt.Sprintf("%s/exclude_regex_%d", sh.name, n), func(b *testing.B) {
				runProfileBench(b, body, sidecarAllExcludeRegex(n))
			})
			b.Run(fmt.Sprintf("%s/exclude_exact_%d", sh.name, n), func(b *testing.B) {
				runProfileBench(b, body, sidecarAllExcludeExact(n))
			})
			b.Run(fmt.Sprintf("%s/include_regex_%d", sh.name, n), func(b *testing.B) {
				runProfileBench(b, body, sidecarNoneIncludeRegex(n))
			})
		}
	}
}

func runProfileBench(b *testing.B, body []byte, sidecar *v1alpha1.Sidecar) {
	b.ReportAllocs()
	b.SetBytes(int64(len(body)))
	mutator := AggregatedMetricsMutator(ProfileMutatorGenerator(sidecar))
	out := new(bytes.Buffer)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		out.Reset()
		if err := mutator(bytes.NewReader(body), out); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkProfileMutator_OnlyFilter(b *testing.B) {
	// Skip parse + serialize, isolate the ProfileMutatorGenerator hot path.
	for _, sh := range benchShapes {
		body := genEnvoyScrape(sh.families, sh.series)
		for _, n := range []int{1, 10, 50} {
			b.Run(fmt.Sprintf("%s/exclude_regex_%d", sh.name, n), func(b *testing.B) {
				runFilterOnly(b, body, sidecarAllExcludeRegex(n))
			})
			b.Run(fmt.Sprintf("%s/exclude_exact_%d", sh.name, n), func(b *testing.B) {
				runFilterOnly(b, body, sidecarAllExcludeExact(n))
			})
		}
	}
}

func runFilterOnly(b *testing.B, body []byte, sidecar *v1alpha1.Sidecar) {
	b.Helper()
	filter := ProfileMutatorGenerator(sidecar)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// re-parse each iter to get a fresh map (filter mutates in place)
		b.StopTimer()
		parser := expfmt.NewTextParser(model.UTF8Validation)
		mf, err := parser.TextToMetricFamilies(bytes.NewReader(body))
		if err != nil {
			b.Fatal(err)
		}
		b.StartTimer()
		if err := filter(mf); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFromPrometheusMetrics(b *testing.B) {
	for _, sh := range benchShapes {
		body := genEnvoyScrape(sh.families, sh.series)
		b.Run(sh.name, func(b *testing.B) {
			parser := expfmt.NewTextParser(model.UTF8Validation)
			now := time.Now()
			b.ReportAllocs()
			b.SetBytes(int64(len(body)))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				mf, err := parser.TextToMetricFamilies(bytes.NewReader(body))
				if err != nil {
					b.Fatal(err)
				}
				b.StartTimer()
				_ = FromPrometheusMetrics(mf, "v0", nil, now)
			}
		})
	}
}
