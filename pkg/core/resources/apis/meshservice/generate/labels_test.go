package generate_test

import (
	"fmt"
	"maps"
	"time"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/v2/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshservice/generate"
	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/v2/pkg/test/resources/model"
)

func newCounter() prometheus.Counter {
	return prometheus.NewCounter(prometheus.CounterOpts{Name: "test_dropped"})
}

func dpWith(name string, labels map[string]string, created time.Time, inbounds ...*mesh_proto.Dataplane_Networking_Inbound) *core_mesh.DataplaneResource {
	return &core_mesh.DataplaneResource{
		Meta: &test_model.ResourceMeta{
			Name:         name,
			Mesh:         core_model.DefaultMesh,
			CreationTime: created,
			Labels:       labels,
		},
		Spec: &mesh_proto.Dataplane{
			Networking: &mesh_proto.Dataplane_Networking{
				Address: "127.0.0.1",
				Inbound: inbounds,
			},
		},
	}
}

func inboundWithTags(tags map[string]string) *mesh_proto.Dataplane_Networking_Inbound {
	return builders.Inbound().WithPort(80).WithServicePort(8080).WithTags(tags).Build()
}

func identityFor(key string) func(*core_mesh.DataplaneResource) map[string]string {
	return func(dp *core_mesh.DataplaneResource) map[string]string {
		if v, ok := dp.GetMeta().GetLabels()[key]; ok {
			return map[string]string{key: v}
		}
		return nil
	}
}

func sliceOf(n1 int, v1 string, t1 time.Time, n2 int, v2 string, t2 time.Time) func() []*core_mesh.DataplaneResource {
	return func() []*core_mesh.DataplaneResource {
		var out []*core_mesh.DataplaneResource
		for i := range n1 {
			out = append(out, dpWith(fmt.Sprintf("a-%d", i), map[string]string{"version": v1}, t1.Add(time.Duration(i)*time.Second)))
		}
		for i := range n2 {
			out = append(out, dpWith(fmt.Sprintf("b-%d", i), map[string]string{"version": v2}, t2.Add(time.Duration(i)*time.Second)))
		}
		return out
	}
}

var _ = DescribeTable("dpContribution",
	func(dpLabels map[string]string, inboundTags []map[string]string, allow []string, want map[string]string, wantDrops float64) {
		var allowSet map[string]struct{}
		if allow != nil {
			allowSet = map[string]struct{}{}
			for _, k := range allow {
				allowSet[k] = struct{}{}
			}
		}
		var ins []*mesh_proto.Dataplane_Networking_Inbound
		for _, t := range inboundTags {
			ins = append(ins, inboundWithTags(t))
		}
		dp := dpWith("dp-1", dpLabels, time.Unix(0, 0), ins...)
		c := newCounter()
		got := generate.DpContribution(dp, ins, allowSet, c, logr.Discard())
		Expect(got).To(Equal(want))
		Expect(testutil.ToFloat64(c)).To(Equal(wantDrops))
	},
	Entry("1: single inbound custom tag propagates",
		nil,
		[]map[string]string{{mesh_proto.ServiceTag: "backend", "appci": "jeffy"}},
		nil,
		map[string]string{"appci": "jeffy"}, 0.0),
	Entry("2: DP resource label propagates",
		map[string]string{"color": "blu"},
		[]map[string]string{{mesh_proto.ServiceTag: "backend"}},
		nil,
		map[string]string{"color": "blu"}, 0.0),
	Entry("3: DP label beats inbound tag",
		map[string]string{"color": "blu"},
		[]map[string]string{{mesh_proto.ServiceTag: "backend", "color": "red"}},
		nil,
		map[string]string{"color": "blu"}, 0.0),
	Entry("4: two inbounds agree → propagate",
		nil,
		[]map[string]string{
			{mesh_proto.ServiceTag: "backend", "appci": "jeffy"},
			{mesh_proto.ServiceTag: "backend", "appci": "jeffy"},
		},
		nil,
		map[string]string{"appci": "jeffy"}, 0.0),
	Entry("5: two inbounds disagree → drop + metric",
		nil,
		[]map[string]string{
			{mesh_proto.ServiceTag: "backend", "appci": "jeffy"},
			{mesh_proto.ServiceTag: "backend", "appci": "bob"},
		},
		nil,
		map[string]string{}, 1.0),
	Entry("6: kuma.io/protocol reserved → not propagated",
		nil,
		[]map[string]string{{mesh_proto.ServiceTag: "backend", mesh_proto.ProtocolTag: "http"}},
		nil,
		map[string]string{}, 0.0),
	Entry("7: k8s.kuma.io/* reserved → not propagated",
		nil,
		[]map[string]string{{mesh_proto.ServiceTag: "backend", mesh_proto.KubeNamespaceTag: "foo"}},
		nil,
		map[string]string{}, 0.0),
	Entry("8: invalid inbound key skipped, others propagate",
		nil,
		[]map[string]string{{mesh_proto.ServiceTag: "backend", "BAD!KEY": "x", "team": "payments"}},
		nil,
		map[string]string{"team": "payments"}, 1.0),
	Entry("9: allow-list filters non-listed",
		nil,
		[]map[string]string{{mesh_proto.ServiceTag: "backend", "appci": "jeffy", "team": "payments"}},
		[]string{"appci"},
		map[string]string{"appci": "jeffy"}, 0.0),
	Entry("9a: reserved DP label does NOT override valid inbound",
		map[string]string{mesh_proto.ZoneTag: "evil-override", "color": "blu"},
		[]map[string]string{{mesh_proto.ServiceTag: "backend", "color": "red"}},
		nil,
		map[string]string{"color": "blu"}, 0.0),
	Entry("9b: invalid DP label key drops + metric, inbound value remains",
		map[string]string{"BAD!KEY": "x", "color": "blu"},
		[]map[string]string{{mesh_proto.ServiceTag: "backend", "color": "red"}},
		nil,
		map[string]string{"color": "blu"}, 1.0),
	Entry("9c: DP label outside allow-list dropped, inbound allow-listed value wins",
		map[string]string{"team": "infra"},
		[]map[string]string{{mesh_proto.ServiceTag: "backend", "appci": "jeffy"}},
		[]string{"appci"},
		map[string]string{"appci": "jeffy"}, 1.0),
)

var _ = DescribeTable("mergeAcrossDataplanes",
	func(build func() []*core_mesh.DataplaneResource, contribution func(*core_mesh.DataplaneResource) map[string]string, want map[string]string) {
		Expect(generate.MergeAcrossDataplanes(build(), contribution)).To(Equal(want))
	},
	Entry("12: all DPs agree",
		func() []*core_mesh.DataplaneResource {
			var out []*core_mesh.DataplaneResource
			for i := range 3 {
				out = append(out, dpWith(fmt.Sprintf("dp-%d", i), map[string]string{"team": "payments"}, time.Unix(int64(i), 0)))
			}
			return out
		},
		func(dp *core_mesh.DataplaneResource) map[string]string { return dp.GetMeta().GetLabels() },
		map[string]string{"team": "payments"}),
	Entry("13: 7 v5 vs 3 v6 → v5 wins (clear majority)",
		sliceOf(7, "v5", time.Unix(0, 0), 3, "v6", time.Unix(100, 0)),
		identityFor("version"),
		map[string]string{"version": "v5"}),
	Entry("14: 9 v5 vs 1 newer v6 → v5 (canary safety)",
		sliceOf(9, "v5", time.Unix(0, 0), 1, "v6", time.Unix(100, 0)),
		identityFor("version"),
		map[string]string{"version": "v5"}),
	Entry("15: 5 v5 vs 5 newer v6 → v6 (tie, newest wins)",
		sliceOf(5, "v5", time.Unix(0, 0), 5, "v6", time.Unix(100, 0)),
		identityFor("version"),
		map[string]string{"version": "v6"}),
	Entry("16: 5 DPs no tier, 5 DPs tier=critical → critical (single voter)",
		func() []*core_mesh.DataplaneResource {
			var out []*core_mesh.DataplaneResource
			for i := range 5 {
				out = append(out, dpWith(fmt.Sprintf("a-%d", i), nil, time.Unix(int64(i), 0)))
			}
			for i := range 5 {
				out = append(out, dpWith(fmt.Sprintf("b-%d", i), map[string]string{"tier": "critical"}, time.Unix(int64(100+i), 0)))
			}
			return out
		},
		identityFor("tier"),
		map[string]string{"tier": "critical"}),
	Entry("17: 5 old DPs team=payments, 5 new DPs no team → team=payments",
		func() []*core_mesh.DataplaneResource {
			var out []*core_mesh.DataplaneResource
			for i := range 5 {
				out = append(out, dpWith(fmt.Sprintf("a-%d", i), map[string]string{"team": "payments"}, time.Unix(int64(i), 0)))
			}
			for i := range 5 {
				out = append(out, dpWith(fmt.Sprintf("b-%d", i), nil, time.Unix(int64(100+i), 0)))
			}
			return out
		},
		identityFor("team"),
		map[string]string{"team": "payments"}),
	Entry("18: last team-defining DP removed → key absent",
		func() []*core_mesh.DataplaneResource {
			return []*core_mesh.DataplaneResource{
				dpWith("only-other", map[string]string{"other": "x"}, time.Unix(0, 0)),
			}
		},
		identityFor("team"),
		map[string]string{}),
	Entry("19: tied count + tied creation time → lex value wins",
		func() []*core_mesh.DataplaneResource {
			t := time.Unix(0, 0)
			return []*core_mesh.DataplaneResource{
				dpWith("a", map[string]string{"v": "zebra"}, t),
				dpWith("b", map[string]string{"v": "apple"}, t),
			}
		},
		identityFor("v"),
		map[string]string{"v": "apple"}),
	Entry("20: nil DP entries are filtered (no panic)",
		func() []*core_mesh.DataplaneResource {
			return []*core_mesh.DataplaneResource{
				nil,
				dpWith("a", map[string]string{"team": "payments"}, time.Unix(0, 0)),
				nil,
			}
		},
		identityFor("team"),
		map[string]string{"team": "payments"}),
	Entry("21: contribution returns nil map → not a panic, no votes",
		func() []*core_mesh.DataplaneResource {
			return []*core_mesh.DataplaneResource{
				dpWith("a", nil, time.Unix(0, 0)),
			}
		},
		func(*core_mesh.DataplaneResource) map[string]string { return nil },
		map[string]string{}),
)

var _ = Describe("helper purity", func() {
	It("dpContribution does not mutate inbounds, dp labels, or allowSet", func() {
		dpLabels := map[string]string{"color": "blu"}
		dpLabelsSnap := maps.Clone(dpLabels)
		inTags := map[string]string{mesh_proto.ServiceTag: "backend", "appci": "jeffy"}
		inTagsSnap := maps.Clone(inTags)
		allow := map[string]struct{}{"appci": {}}
		allowSnap := maps.Clone(allow)

		in := inboundWithTags(inTags)
		dp := dpWith("dp-1", dpLabels, time.Unix(0, 0), in)
		_ = generate.DpContribution(dp, []*mesh_proto.Dataplane_Networking_Inbound{in}, allow, newCounter(), logr.Discard())

		Expect(dpLabels).To(Equal(dpLabelsSnap))
		Expect(inTags).To(Equal(inTagsSnap))
		Expect(allow).To(Equal(allowSnap))
	})

	It("mergeAcrossDataplanes does not mutate the dps slice", func() {
		dps := []*core_mesh.DataplaneResource{
			dpWith("b", map[string]string{"team": "x"}, time.Unix(100, 0)),
			dpWith("a", map[string]string{"team": "x"}, time.Unix(0, 0)),
		}
		before := []*core_mesh.DataplaneResource{dps[0], dps[1]}
		_ = generate.MergeAcrossDataplanes(dps, func(d *core_mesh.DataplaneResource) map[string]string {
			return d.GetMeta().GetLabels()
		})
		Expect(dps[0]).To(BeIdenticalTo(before[0]))
		Expect(dps[1]).To(BeIdenticalTo(before[1]))
	})
})
