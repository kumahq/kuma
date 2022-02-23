package v1alpha1_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/api/mesh/v1alpha1"
)

var _ = Describe("MultiValueTagSet", func() {

	Describe("HostnameEntries()", func() {
		type testCase struct {
			value    MultiValueTagSet
			expected []string
		}

		DescribeTable("should return a sorted list of keys",
			func(given testCase) {
				Expect(given.value.Keys()).To(Equal(given.expected))
			},
			Entry("`service` and `services` tags", testCase{
				value: MultiValueTagSet{
					"versions":        map[string]bool{},
					"version":         map[string]bool{},
					"services":        map[string]bool{},
					"kuma.io/service": map[string]bool{},
				},
				expected: []string{"kuma.io/service", "services", "version", "versions"},
			}),
		)
	})
})

var _ = Describe("Dataplane_Networking", func() {

	Describe("GetOutboundInterfaces()", func() {
		Context("valid input values", func() {
			type testCase struct {
				input    *Dataplane_Networking
				expected []OutboundInterface
			}

			DescribeTable("should parse valid input values",
				func(given testCase) {
					// when
					ofaces := given.input.GetOutboundInterfaces()
					// then
					Expect(ofaces).To(Equal(given.expected))
				},
				Entry("nil", testCase{
					input:    nil,
					expected: nil,
				}),
				Entry("empty", testCase{
					input:    &Dataplane_Networking{},
					expected: []OutboundInterface{},
				}),
				Entry("2 outbound interfaces", testCase{
					input: &Dataplane_Networking{
						Outbound: []*Dataplane_Networking_Outbound{
							{
								Port: 8080,
							},
							{
								Address: "192.168.0.1",
								Port:    443,
							},
						},
					},
					expected: []OutboundInterface{
						{DataplaneIP: "127.0.0.1", DataplanePort: 8080},
						{DataplaneIP: "192.168.0.1", DataplanePort: 443},
					},
				}),
				Entry("2 outbound interfaces IPv6", testCase{
					input: &Dataplane_Networking{
						Outbound: []*Dataplane_Networking_Outbound{
							{
								Port: 8080,
							},
							{
								Address: "fd00::1",
								Port:    443,
							},
						},
					},
					expected: []OutboundInterface{
						{DataplaneIP: "127.0.0.1", DataplanePort: 8080},
						{DataplaneIP: "fd00::1", DataplanePort: 443},
					},
				}),
			)
		})
	})

	Describe("GetInboundInterfaces()", func() {

		Context("valid input values", func() {
			type testCase struct {
				input    *Dataplane_Networking
				expected []InboundInterface
			}

			DescribeTable("should parse valid input values",
				func(given testCase) {
					// when
					ifaces, err := given.input.GetInboundInterfaces()
					// then
					Expect(err).ToNot(HaveOccurred())
					// and
					Expect(ifaces).To(ConsistOf(given.expected))
				},
				Entry("nil", testCase{
					input:    nil,
					expected: nil,
				}),
				Entry("empty", testCase{
					input:    &Dataplane_Networking{},
					expected: []InboundInterface{},
				}),
				Entry("2 inbound interfaces", testCase{
					input: &Dataplane_Networking{
						Address: "192.168.0.1",
						Inbound: []*Dataplane_Networking_Inbound{
							{
								Port: 80,
							},
							{
								Address:        "192.168.0.2",
								Port:           443,
								ServiceAddress: "192.168.0.3",
								ServicePort:    8443,
							},
						},
					},
					expected: []InboundInterface{
						{DataplaneAdvertisedIP: "192.168.0.1", DataplaneIP: "192.168.0.1", DataplanePort: 80, WorkloadIP: "127.0.0.1", WorkloadPort: 80},
						{DataplaneAdvertisedIP: "192.168.0.2", DataplaneIP: "192.168.0.2", DataplanePort: 443, WorkloadIP: "192.168.0.3", WorkloadPort: 8443},
					},
				}),
			)
		})
	})

	Describe("GetHealthyInbounds()", func() {

		It("should return only healty inbounds", func() {
			networking := &Dataplane_Networking{
				Inbound: []*Dataplane_Networking_Inbound{
					{
						Health:      nil,
						Port:        8080,
						ServicePort: 80,
					},
					{
						Health:      &Dataplane_Networking_Inbound_Health{Ready: true},
						Port:        9090,
						ServicePort: 90,
					},
					{
						Health:      &Dataplane_Networking_Inbound_Health{Ready: false},
						Port:        7070,
						ServicePort: 70,
					},
				},
			}

			actual := networking.GetHealthyInbounds()
			Expect(actual).To(HaveLen(2))
			Expect(actual).To(ConsistOf(
				&Dataplane_Networking_Inbound{
					Health:      &Dataplane_Networking_Inbound_Health{Ready: true},
					Port:        9090,
					ServicePort: 90,
				},
				&Dataplane_Networking_Inbound{
					Port:        8080,
					ServicePort: 80,
				}))
		})
	})
})

var _ = Describe("Dataplane_Networking_Outbound", func() {
	type testCase struct {
		serviceTag    string
		selector      TagSelector
		expectedMatch bool
	}
	DescribeTable("MatchTags()",
		func(given testCase) {
			// given
			outbound := Dataplane_Networking_Outbound{
				Service: given.serviceTag,
			}

			// when
			matched := outbound.MatchTags(given.selector)

			// then
			Expect(matched).To(Equal(given.expectedMatch))
		},
		Entry("it should match *", testCase{
			serviceTag: "backend",
			selector: map[string]string{
				"kuma.io/service": "*",
			},
			expectedMatch: true,
		}),
		Entry("it should match service", testCase{
			serviceTag: "backend",
			selector: map[string]string{
				"kuma.io/service": "backend",
			},
			expectedMatch: true,
		}),
		Entry("it shouldn't match tag other than service", testCase{
			serviceTag: "backend",
			selector: map[string]string{
				"version": "1.0",
			},
			expectedMatch: false,
		}),
	)
})

var _ = Describe("Dataplane_Networking_Inbound", func() {

	DescribeTable("GetService()", func() {

		type testCase struct {
			inbound  *Dataplane_Networking_Inbound
			expected string
		}

		DescribeTable("should infer service name from `service` tag",
			func(given testCase) {
				Expect(given.inbound.GetService()).To(Equal(given.expected))
			},
			Entry("inbound is `nil`", testCase{
				inbound:  nil,
				expected: "",
			}),
			Entry("inbound has no `service` tag", testCase{
				inbound:  &Dataplane_Networking_Inbound{},
				expected: "",
			}),
			Entry("inbound has `service` tag", testCase{
				inbound: &Dataplane_Networking_Inbound{
					Tags: map[string]string{
						"kuma.io/service": "backend",
					},
				},
				expected: "backend",
			}),
		)
	})

	DescribeTable("GetProtocol()", func() {

		type testCase struct {
			inbound  *Dataplane_Networking_Inbound
			expected string
		}

		DescribeTable("should infer protocol from `protocol` tag",
			func(given testCase) {
				Expect(given.inbound.GetProtocol()).To(Equal(given.expected))
			},
			Entry("inbound is `nil`", testCase{
				inbound:  nil,
				expected: "",
			}),
			Entry("inbound has no `protocol` tag", testCase{
				inbound:  &Dataplane_Networking_Inbound{},
				expected: "",
			}),
			Entry("inbound has `protocol` tag with a known value", testCase{
				inbound: &Dataplane_Networking_Inbound{
					Tags: map[string]string{
						"protocol": "http",
					},
				},
				expected: "http",
			}),
			Entry("inbound has `protocol` tag with an unknown value", testCase{
				inbound: &Dataplane_Networking_Inbound{
					Tags: map[string]string{
						"protocol": "not-yet-supported-protocol",
					},
				},
				expected: "not-yet-supported-protocol",
			}),
		)
	})
})

var _ = Describe("Dataplane with inbound", func() {
	d := Dataplane{
		Networking: &Dataplane_Networking{
			Inbound: []*Dataplane_Networking_Inbound{
				{
					Tags: map[string]string{
						"kuma.io/service": "backend",
						"version":         "v1",
					},
				},
				{
					Tags: map[string]string{
						"kuma.io/service": "backend-metrics",
						"version":         "v1",
						"role":            "metrics",
					},
				},
			},
		},
	}

	Describe("TagSet()", func() {
		It("should provide combined tags", func() {
			// when
			tags := d.TagSet()

			// then
			Expect(tags.Values("kuma.io/service")).To(Equal([]string{"backend", "backend-metrics"}))
			Expect(tags.Values("version")).To(Equal([]string{"v1"}))
			Expect(tags.Values("role")).To(Equal([]string{"metrics"}))
		})
	})

	Describe("MatchTags()", func() {
		It("should match any inbound", func() {
			// when
			selector := TagSelector{
				"kuma.io/service": "backend",
				"version":         "v1",
			}

			// then
			Expect(d.MatchTags(selector)).To(BeTrue())
		})

		It("should not match if all inbounds did not match", func() {
			// when
			selector := TagSelector{
				"kuma.io/service": "unknown",
			}

			// then
			Expect(d.MatchTags(selector)).To(BeFalse())
		})
	})
})

var _ = Describe("Dataplane classification", func() {
	Describe("with normal networking", func() {
		It("should be a dataplane", func() {
			dp := Dataplane{
				Networking: &Dataplane_Networking{},
			}
			Expect(dp.IsDelegatedGateway()).To(BeFalse())
			Expect(dp.IsBuiltinGateway()).To(BeFalse())
		})
	})

	Describe("with gateway networking", func() {
		It("should be a gateway", func() {
			gw := Dataplane{
				Networking: &Dataplane_Networking{
					Gateway: &Dataplane_Networking_Gateway{},
				},
			}
			Expect(gw.IsDelegatedGateway()).To(BeTrue())
			Expect(gw.IsBuiltinGateway()).To(BeFalse())
		})
	})

	Describe("with delegated gateway networking", func() {
		It("should be a gateway", func() {
			gw := Dataplane{
				Networking: &Dataplane_Networking{
					Gateway: &Dataplane_Networking_Gateway{
						Type: Dataplane_Networking_Gateway_DELEGATED,
					},
				},
			}
			Expect(gw.IsDelegatedGateway()).To(BeTrue())
			Expect(gw.IsBuiltinGateway()).To(BeFalse())
		})
	})

	Describe("with builtin gateway networking", func() {
		It("should be a gateway", func() {
			gw := Dataplane{
				Networking: &Dataplane_Networking{
					Gateway: &Dataplane_Networking_Gateway{
						Type: Dataplane_Networking_Gateway_BUILTIN,
					},
				},
			}
			Expect(gw.IsDelegatedGateway()).To(BeFalse())
			Expect(gw.IsBuiltinGateway()).To(BeTrue())
		})
	})
})

var _ = Describe("Dataplane with gateway", func() {
	d := Dataplane{
		Networking: &Dataplane_Networking{
			Gateway: &Dataplane_Networking_Gateway{
				Tags: map[string]string{
					"kuma.io/service": "backend",
					"version":         "v1",
				},
			},
		},
	}

	Describe("Tags()", func() {
		It("should provide combined tags", func() {
			// when
			tags := d.TagSet()

			// then
			Expect(tags.Values("kuma.io/service")).To(Equal([]string{"backend"}))
		})
	})

	Describe("MatchTags()", func() {
		It("should match gateway", func() {
			// when
			selector := TagSelector{
				"kuma.io/service": "backend",
				"version":         "v1",
			}

			// then
			Expect(d.MatchTags(selector)).To(BeTrue())
		})

		It("should not match if gateway did not match", func() {
			// when
			selector := TagSelector{
				"kuma.io/service": "unknown",
			}

			// then
			Expect(d.MatchTags(selector)).To(BeFalse())
		})
	})
})

var _ = Describe("TagSelector", func() {

	Describe("Matches()", func() {
		type testCase struct {
			tags  map[string]string
			match bool
		}
		DescribeTable("should Match tags",
			func(given testCase) {
				// given
				dpTags := map[string]string{
					"kuma.io/service": "mobile",
					"version":         "v1",
				}

				// when
				match := TagSelector(given.tags).Matches(dpTags)

				// then
				Expect(match).To(Equal(given.match))
			},
			Entry("should match 0 tags", testCase{
				tags:  map[string]string{},
				match: true,
			}),
			Entry("should match 1 tag", testCase{
				tags:  map[string]string{"kuma.io/service": "mobile"},
				match: true,
			}),
			Entry("should match all tags", testCase{
				tags: map[string]string{
					"kuma.io/service": "mobile",
					"version":         "v1",
				},
				match: true,
			}),
			Entry("should match * tag", testCase{
				tags:  map[string]string{"kuma.io/service": "*"},
				match: true,
			}),
			Entry("should not match on one mismatch", testCase{
				tags: map[string]string{
					"kuma.io/service": "backend",
					"version":         "v1",
				},
				match: false,
			}),
		)
	})

	Describe("Equal()", func() {
		type testCase struct {
			one      TagSelector
			another  TagSelector
			expected bool
		}

		DescribeTable("should correctly determine if two selectors are equal",
			func(given testCase) {
				// expect
				Expect(given.one.Equal(given.another)).To(Equal(given.expected))
			},
			Entry("two nil selectors", testCase{
				one:      nil,
				another:  nil,
				expected: true,
			}),
			Entry("nil selector and empty selector", testCase{
				one:      nil,
				another:  TagSelector{},
				expected: true,
			}),
			Entry("empty selector and nil selector", testCase{
				one:      TagSelector{},
				another:  nil,
				expected: true,
			}),
			Entry("two empty selectors", testCase{
				one:      TagSelector{},
				another:  TagSelector{},
				expected: true,
			}),
			Entry("equal selectors of 1 tag", testCase{
				one:      TagSelector{"kuma.io/service": "backend"},
				another:  TagSelector{"kuma.io/service": "backend"},
				expected: true,
			}),
			Entry("equal selectors of 2 tag", testCase{
				one:      TagSelector{"kuma.io/service": "backend", "version": "v1"},
				another:  TagSelector{"kuma.io/service": "backend", "version": "v1"},
				expected: true,
			}),
			Entry("unequal selectors of 1 tag", testCase{
				one:      TagSelector{"kuma.io/service": "backend"},
				another:  TagSelector{"kuma.io/service": "redis"},
				expected: false,
			}),
			Entry("one 1 tag selector and one 2 tags selector", testCase{
				one:      TagSelector{"kuma.io/service": "backend"},
				another:  TagSelector{"kuma.io/service": "redis", "version": "v1"},
				expected: false,
			}),
		)
	})
})

var _ = Describe("Tags", func() {
	It("should print tags", func() {
		// given
		tags := map[string]map[string]bool{
			"kuma.io/service": {
				"backend-api":   true,
				"backend-admin": true,
			},
			"version": {
				"v1": true,
			},
		}

		// when
		result := MultiValueTagSet(tags).String()

		// then
		Expect(result).To(Equal("kuma.io/service=backend-admin,backend-api version=v1"))
	})
})

var _ = Describe("TagSelectorRank", func() {

	Describe("CompareTo()", func() {
		type testCase struct {
			rank1    TagSelectorRank
			rank2    TagSelectorRank
			expected int
		}
		DescribeTable("should correctly compare two ranks",
			func(given testCase) {
				// expect
				Expect(given.rank1.CompareTo(given.rank2)).To(Equal(given.expected))
			},
			Entry("0 ranks are equal", testCase{
				rank1:    TagSelectorRank{},
				rank2:    TagSelectorRank{},
				expected: 0,
			}),
			Entry("matches by the same number of exact values (1) are equal", testCase{
				rank1:    TagSelectorRank{ExactMatches: 1},
				rank2:    TagSelectorRank{ExactMatches: 1},
				expected: 0,
			}),
			Entry("matches by the same number of wildcard values (2) are equal", testCase{
				rank1:    TagSelectorRank{WildcardMatches: 2},
				rank2:    TagSelectorRank{WildcardMatches: 2},
				expected: 0,
			}),
			Entry("equal ranks by non-0 ExactMatches and WildcardMatches", testCase{
				rank1:    TagSelectorRank{ExactMatches: 1, WildcardMatches: 2},
				rank2:    TagSelectorRank{ExactMatches: 1, WildcardMatches: 2},
				expected: 0,
			}),
			Entry("match by an exact value (1) is more specific than match by a wildcard", testCase{
				rank1:    TagSelectorRank{ExactMatches: 1, WildcardMatches: 0},
				rank2:    TagSelectorRank{ExactMatches: 0, WildcardMatches: 1},
				expected: 1,
			}),
			Entry("match by a wildcard is less specific than match by an exact value (1)", testCase{
				rank1:    TagSelectorRank{ExactMatches: 0, WildcardMatches: 1},
				rank2:    TagSelectorRank{ExactMatches: 1, WildcardMatches: 0},
				expected: -1,
			}),
			Entry("match by an exact value (2) is more specific than match by a wildcard", testCase{
				rank1:    TagSelectorRank{ExactMatches: 2, WildcardMatches: 0},
				rank2:    TagSelectorRank{ExactMatches: 0, WildcardMatches: 2},
				expected: 2,
			}),
			Entry("match by a wildcard is less specific than match by an exact value (2)", testCase{
				rank1:    TagSelectorRank{ExactMatches: 0, WildcardMatches: 2},
				rank2:    TagSelectorRank{ExactMatches: 2, WildcardMatches: 0},
				expected: -2,
			}),
			Entry("match by an exact value (3) is more specific than match by a wildcard", testCase{
				rank1:    TagSelectorRank{ExactMatches: 3, WildcardMatches: 0},
				rank2:    TagSelectorRank{ExactMatches: 0, WildcardMatches: 3},
				expected: 3,
			}),
			Entry("match by a wildcard is less specific than match by an exact value (3)", testCase{
				rank1:    TagSelectorRank{ExactMatches: 0, WildcardMatches: 3},
				rank2:    TagSelectorRank{ExactMatches: 3, WildcardMatches: 0},
				expected: -3,
			}),
			Entry("match by an exact value is more specific than match by a wildcard", testCase{
				rank1:    TagSelectorRank{ExactMatches: 2, WildcardMatches: 1},
				rank2:    TagSelectorRank{ExactMatches: 1, WildcardMatches: 1},
				expected: 1,
			}),
			Entry("match by a wildcard is less specific than match by an exact value", testCase{
				rank1:    TagSelectorRank{ExactMatches: 2, WildcardMatches: 1},
				rank2:    TagSelectorRank{ExactMatches: 1, WildcardMatches: 1},
				expected: 1,
			}),
		)
	})
	Describe("CombinedWith()", func() {
		type testCase struct {
			rank1    TagSelectorRank
			rank2    TagSelectorRank
			expected TagSelectorRank
		}
		DescribeTable("should correctly aggregate two ranks",
			func(given testCase) {
				// expect
				Expect(given.rank1.CombinedWith(given.rank2)).To(Equal(given.expected))
			},
			Entry("combination of two 0 ranks is zero rank", testCase{
				rank1:    TagSelectorRank{},
				rank2:    TagSelectorRank{},
				expected: TagSelectorRank{},
			}),
			Entry("cobination of a match by an exact value with a match by a wildcard", testCase{
				rank1:    TagSelectorRank{ExactMatches: 1},
				rank2:    TagSelectorRank{WildcardMatches: 2},
				expected: TagSelectorRank{ExactMatches: 1, WildcardMatches: 2},
			}),
			Entry("cobination of two mixed matches", testCase{
				rank1:    TagSelectorRank{ExactMatches: 1, WildcardMatches: 2},
				rank2:    TagSelectorRank{ExactMatches: 10, WildcardMatches: 20},
				expected: TagSelectorRank{ExactMatches: 11, WildcardMatches: 22},
			}),
		)
	})
})
