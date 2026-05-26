package xds_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	common_api "github.com/kumahq/kuma/v2/api/common/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/plugins/policies/core/rules/inbound"
	api "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshaccesslog/api/v1alpha1"
	. "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshaccesslog/plugin/xds"
	"github.com/kumahq/kuma/v2/pkg/util/pointer"
	listeners_v3 "github.com/kumahq/kuma/v2/pkg/xds/envoy/listeners/v3"
)

var _ = Describe("BuildAccessLogBuildersFromRules", func() {
	fileBackend := func(path string) api.Backend {
		return api.Backend{
			Type: api.FileBackendType,
			File: &api.FileBackend{Path: path},
		}
	}

	confWith := func(backends ...api.Backend) any {
		return api.Conf{Backends: pointer.To(backends)}
	}

	values := listeners_v3.KumaValues{Mesh: "default"}
	endpoints := &EndpointAccumulator{}

	build := func(rules []*inbound.Rule) []map[string]any {
		builders := BuildAccessLogBuildersFromRules(rules, "", endpoints, values, "/tmp/sock")
		out := make([]map[string]any, 0, len(builders))
		for _, b := range builders {
			al, err := b.Build()
			Expect(err).ToNot(HaveOccurred())
			entry := map[string]any{"name": al.Name}
			if al.Filter != nil {
				entry["hasFilter"] = true
				entry["filterName"] = al.Filter.GetExtensionFilter().GetName()
			} else {
				entry["hasFilter"] = false
			}
			out = append(out, entry)
		}
		return out
	}

	It("catch-all rule produces an unfiltered entry", func() {
		rules := []*inbound.Rule{
			{Match: nil, Conf: confWith(fileBackend("/var/log/a"))},
		}
		Expect(build(rules)).To(ConsistOf(
			HaveKeyWithValue("hasFilter", false),
		))
	})

	It("each rule's match becomes its own CEL filter", func() {
		spiffeMatch := common_api.Match{
			SpiffeID: &common_api.SpiffeIDMatch{Type: common_api.ExactMatchType, Value: "spiffe://default/sa/x"},
		}
		sniMatch := common_api.Match{
			SNI: &common_api.SNIMatch{Type: common_api.SNIExactMatchType, Value: "sni-1"},
		}
		rules := []*inbound.Rule{
			{Match: &spiffeMatch, Conf: confWith(fileBackend("/var/log/a"))},
			{Match: &sniMatch, Conf: confWith(fileBackend("/var/log/b"))},
		}
		out := build(rules)
		Expect(out).To(HaveLen(2))
		Expect(out[0]).To(HaveKeyWithValue("hasFilter", true))
		Expect(out[0]).To(HaveKeyWithValue("filterName", "envoy.access_loggers.extension_filters.cel"))
		Expect(out[1]).To(HaveKeyWithValue("hasFilter", true))
		Expect(out[1]).To(HaveKeyWithValue("filterName", "envoy.access_loggers.extension_filters.cel"))
	})

	It("catch-all alongside a specific rule stays unfiltered", func() {
		spiffeMatch := common_api.Match{
			SpiffeID: &common_api.SpiffeIDMatch{Type: common_api.ExactMatchType, Value: "spiffe://default/sa/x"},
		}
		rules := []*inbound.Rule{
			{Match: &spiffeMatch, Conf: confWith(fileBackend("/var/log/a"))},
			{Match: nil, Conf: confWith(fileBackend("/var/log/b"))},
		}
		out := build(rules)
		Expect(out).To(HaveLen(2))
		Expect(out[0]).To(HaveKeyWithValue("hasFilter", true))
		Expect(out[1]).To(HaveKeyWithValue("hasFilter", false))
	})

	It("one rule with two backends generates two entries with the same filter", func() {
		spiffeMatch := common_api.Match{
			SpiffeID: &common_api.SpiffeIDMatch{Type: common_api.ExactMatchType, Value: "spiffe://default/sa/x"},
		}
		rules := []*inbound.Rule{
			{Match: &spiffeMatch, Conf: confWith(fileBackend("/var/log/a"), fileBackend("/var/log/b"))},
		}
		Expect(build(rules)).To(HaveLen(2))
	})

	It("conf with unexpected type is skipped", func() {
		spiffeMatch := common_api.Match{
			SpiffeID: &common_api.SpiffeIDMatch{Type: common_api.ExactMatchType, Value: "spiffe://default/sa/x"},
		}
		sniMatch := common_api.Match{
			SNI: &common_api.SNIMatch{Type: common_api.SNIExactMatchType, Value: "sni-1"},
		}
		rules := []*inbound.Rule{
			{Match: &spiffeMatch, Conf: "not-a-conf"},
			{Match: &sniMatch, Conf: confWith(fileBackend("/var/log/b"))},
		}
		out := build(rules)
		Expect(out).To(HaveLen(1))
		Expect(out[0]).To(HaveKeyWithValue("hasFilter", true))
	})
})
