package vips

import (
	"encoding/json"
	"path/filepath"
	"sort"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/util/pointer"
)

var _ = Describe("Test Tag first Virtual Outbound view", func() {
	type testCase struct {
		meshView             *VirtualOutboundMeshView
		expectedTagFirstView string
	}

	DescribeTable("convert to tag first view", func(given testCase) {
		// when
		view := NewTagFirstOutboundView(given.meshView)

		// then
		marshal, err := json.Marshal(ordered(*view))
		Expect(err).ToNot(HaveOccurred())
		Expect(marshal).To(matchers.MatchGoldenJSON(filepath.Join(".", "testdata", given.expectedTagFirstView)))

		// then convert back
		result := view.ToVirtualOutboundView()
		Expect(result).To(Equal(given.meshView))
	},
		Entry("full example", testCase{
			meshView: &VirtualOutboundMeshView{
				byHostname: map[HostnameEntry]*VirtualOutbound{
					{
						Type: 0,
						Name: "demo-app_kuma-demo_svc_5000",
					}: {
						Address: "240.0.0.1",
						Outbounds: []OutboundEntry{
							{
								TagSet: map[string]string{
									"kuma.io/service": "demo-app_kuma-demo_svc_5000",
								},
								Origin: "service",
							},
						},
					},
					{
						Type: 0,
						Name: "demo-app_kuma-demo_svc_5000",
					}: {
						Address: "240.0.0.2",
						Outbounds: []OutboundEntry{
							{
								TagSet: map[string]string{
									"kuma.io/service": "demo-app_kuma-demo_svc_5000",
								},
								Origin: "service",
							},
						},
					},
					{
						Type: 0,
						Name: "redis_kuma-demo_svc_6379",
					}: {
						Address: "240.0.0.1",
						Outbounds: []OutboundEntry{
							{
								TagSet: map[string]string{
									"kuma.io/service": "redis_kuma-demo_svc_6379",
								},
								Origin: "service",
							},
						},
					},
					{
						Type: 1,
						Name: "10.43.224.70",
					}: {
						Address: "10.43.224.70",
						Outbounds: []OutboundEntry{
							{
								Port: 6379,
								TagSet: map[string]string{
									"kuma.io/service": "redis_kuma-demo_svc_6379",
								},
								Origin: "kubernetes",
							},
						},
					},
					{
						Type: 1,
						Name: "10.43.96.209",
					}: {
						Address: "10.43.96.209",
						Outbounds: []OutboundEntry{
							{
								Port: 5000,
								TagSet: map[string]string{
									"kuma.io/service": "demo-app_kuma-demo_svc_5000",
								},
								Origin: "kubernetes",
							},
						},
					},
					{
						Type: 0,
						Name: "httpbin",
					}: {
						Address: "240.0.0.2",
						Outbounds: []OutboundEntry{
							{
								Port: 443,
								TagSet: map[string]string{
									"kuma.io/service": "httpbin",
								},
								Origin: "service",
							},
						},
					},
					{
						Type: 1,
						Name: "httpbin.org",
					}: {
						Address: "240.0.0.3",
						Outbounds: []OutboundEntry{
							{
								Port: 443,
								TagSet: map[string]string{
									"kuma.io/service": "httpbin",
								},
								Origin: "external-service:httpbin",
							},
						},
					},
					{
						Type: 2,
						Name: "redis_kuma-demo_svc_6379.mesh",
					}: {
						Address: "240.0.0.6",
						Outbounds: []OutboundEntry{
							{
								Port: 80,
								TagSet: map[string]string{
									"kuma.io/service": "redis_kuma-demo_svc_6379",
								},
								Origin: "virtual-outbound:default",
							},
						},
					},
					{
						Type: 2,
						Name: "httpbin.mesh",
					}: {
						Address: "240.0.0.5",
						Outbounds: []OutboundEntry{
							{
								Port: 80,
								TagSet: map[string]string{
									"kuma.io/service": "httpbin",
								},
								Origin: "virtual-outbound:default",
							},
						},
					},
				},
			},
			expectedTagFirstView: "full_example_tag_first.golden.json",
		}),
		Entry("single service", testCase{
			meshView: &VirtualOutboundMeshView{
				byHostname: map[HostnameEntry]*VirtualOutbound{
					{
						Type: 0,
						Name: "demo-app_kuma-demo_svc_5000",
					}: {
						Address: "240.0.0.2",
						Outbounds: []OutboundEntry{
							{
								TagSet: map[string]string{
									"kuma.io/service": "demo-app_kuma-demo_svc_5000",
								},
								Origin: "service",
							},
						},
					},
					{
						Type: 1,
						Name: "10.43.96.209",
					}: {
						Address: "10.43.96.209",
						Outbounds: []OutboundEntry{
							{
								Port: 5000,
								TagSet: map[string]string{
									"kuma.io/service": "demo-app_kuma-demo_svc_5000",
								},
								Origin: "kubernetes",
							},
						},
					},
				},
			},
			expectedTagFirstView: "single_service_tag_first.golden.json",
		}),
		Entry("multiple services", testCase{
			meshView: &VirtualOutboundMeshView{
				byHostname: map[HostnameEntry]*VirtualOutbound{
					{
						Type: 0,
						Name: "demo-app_kuma-demo_svc_5000",
					}: {
						Address: "240.0.0.2",
						Outbounds: []OutboundEntry{
							{
								TagSet: map[string]string{
									"kuma.io/service": "demo-app_kuma-demo_svc_5000",
								},
								Origin: "service",
							},
						},
					},
					{
						Type: 1,
						Name: "10.43.96.209",
					}: {
						Address: "10.43.96.209",
						Outbounds: []OutboundEntry{
							{
								Port: 5000,
								TagSet: map[string]string{
									"kuma.io/service": "demo-app_kuma-demo_svc_5000",
								},
								Origin: "kubernetes",
							},
						},
					},
					{
						Type: 0,
						Name: "redis_kuma-demo_svc_6379",
					}: {
						Address: "240.0.0.1",
						Outbounds: []OutboundEntry{
							{
								TagSet: map[string]string{
									"kuma.io/service": "redis_kuma-demo_svc_6379",
								},
								Origin: "service",
							},
						},
					},
				},
			},
			expectedTagFirstView: "multiple_services_tag_first.golden.json",
		}),
		Entry("different tags on single service", testCase{
			meshView: &VirtualOutboundMeshView{
				byHostname: map[HostnameEntry]*VirtualOutbound{
					{
						Type: 0,
						Name: "demo-app_kuma-demo_svc_5000",
					}: {
						Address: "240.0.0.2",
						Outbounds: []OutboundEntry{
							{
								TagSet: map[string]string{
									"kuma.io/service": "demo-app_kuma-demo_svc_5000",
								},
								Origin: "service",
							},
						},
					},
					{
						Type: 2,
						Name: "demo-app_kuma-demo_svc_5000.mesh",
					}: {
						Address: "240.0.0.3",
						Outbounds: []OutboundEntry{
							{
								TagSet: map[string]string{
									"kuma.io/service": "demo-app_kuma-demo_svc_5000",
									"kuma.io/zone":    "zone-1",
								},
								Origin: "virtual-outbound:default",
							},
						},
					},
					{
						Type: 1,
						Name: "10.43.96.209",
					}: {
						Address: "10.43.96.209",
						Outbounds: []OutboundEntry{
							{
								Port: 5000,
								TagSet: map[string]string{
									"kuma.io/service": "demo-app_kuma-demo_svc_5000",
									"kuma.io/zone":    "zone-1",
								},
								Origin: "kubernetes",
							},
						},
					},
				},
			},
			expectedTagFirstView: "different_tags_tag_first.golden.json",
		}),
	)
})

func ordered(view TagFirstVirtualOutboundView) TagFirstVirtualOutboundView {
	for _, v := range view.PerService {
		for _, outbound := range v.Outbounds {
			sort.SliceStable(outbound.AddressPorts, func(i, j int) bool {
				return pointer.Deref(outbound.AddressPorts[i].Address) < pointer.Deref(outbound.AddressPorts[j].Address)
			})
		}
		sort.SliceStable(v.Outbounds, func(i, j int) bool {
			return len(v.Outbounds[i].Tags) < len(v.Outbounds[j].Tags)
		})
	}
	return view
}
