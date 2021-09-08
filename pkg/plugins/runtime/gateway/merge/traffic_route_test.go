package merge_test

import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway/merge"
	"github.com/kumahq/kuma/pkg/test"
	. "github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/test/resources/model"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

func TestMerge(t *testing.T) {
	test.RunSpecs(t, "Merge")
}

var _ = Describe("TrafficRoute", func() {
	ExactMatch := func(s string) *mesh_proto.TrafficRoute_Http_Match_StringMatcher {
		return &mesh_proto.TrafficRoute_Http_Match_StringMatcher{
			MatcherType: &mesh_proto.TrafficRoute_Http_Match_StringMatcher_Exact{
				Exact: s,
			},
		}
	}

	PrefixMatch := func(s string) *mesh_proto.TrafficRoute_Http_Match_StringMatcher {
		return &mesh_proto.TrafficRoute_Http_Match_StringMatcher{
			MatcherType: &mesh_proto.TrafficRoute_Http_Match_StringMatcher_Prefix{
				Prefix: s,
			},
		}
	}

	RegexMatch := func(s string) *mesh_proto.TrafficRoute_Http_Match_StringMatcher {
		return &mesh_proto.TrafficRoute_Http_Match_StringMatcher{
			MatcherType: &mesh_proto.TrafficRoute_Http_Match_StringMatcher_Regex{
				Regex: s,
			},
		}
	}

	It("should return nil with no routes", func() {
		Expect(merge.TrafficRoute()).To(BeNil())
	})

	It("should keep the oldest default", func() {
		merged := merge.TrafficRoute(
			&core_mesh.TrafficRouteResource{
				Meta: &model.ResourceMeta{
					CreationTime: core.Now(),
				},
				Spec: &mesh_proto.TrafficRoute{
					Conf: &mesh_proto.TrafficRoute_Conf{
						Split: []*mesh_proto.TrafficRoute_Split{
							&mesh_proto.TrafficRoute_Split{
								Weight: util_proto.UInt32(1),
								Destination: map[string]string{
									mesh_proto.ServiceTag: "bar-service",
								},
							},
						},
					}},
			},
			&core_mesh.TrafficRouteResource{
				Meta: &model.ResourceMeta{
					CreationTime: core.Now().Add(-1 * time.Minute),
				},
				Spec: &mesh_proto.TrafficRoute{
					Conf: &mesh_proto.TrafficRoute_Conf{
						Destination: map[string]string{
							mesh_proto.ServiceTag: "foo-service",
						},
					},
				},
			},
		)

		Expect(merged.Spec.Conf).To(MatchProto(&mesh_proto.TrafficRoute_Conf{
			Destination: map[string]string{
				mesh_proto.ServiceTag: "foo-service",
			},
		}))
	})

	It("should order longer path match first", func() {
		merged := merge.TrafficRoute(
			&core_mesh.TrafficRouteResource{
				Meta: &model.ResourceMeta{
					CreationTime: core.Now(),
				},
				Spec: &mesh_proto.TrafficRoute{
					Conf: &mesh_proto.TrafficRoute_Conf{
						Http: []*mesh_proto.TrafficRoute_Http{{
							Match: &mesh_proto.TrafficRoute_Http_Match{
								Path: PrefixMatch("/api"),
							},
						}},
					},
				},
			},
			&core_mesh.TrafficRouteResource{
				Meta: &model.ResourceMeta{
					CreationTime: core.Now().Add(-1 * time.Minute),
				},
				Spec: &mesh_proto.TrafficRoute{
					Conf: &mesh_proto.TrafficRoute_Conf{
						Http: []*mesh_proto.TrafficRoute_Http{{
							Match: &mesh_proto.TrafficRoute_Http_Match{
								Path: PrefixMatch("/api/aaa"),
							},
						}, {
							Match: &mesh_proto.TrafficRoute_Http_Match{
								Path: PrefixMatch("/api/aaab"),
							},
						},
						},
					},
				},
			},
		)

		Expect(merged.Spec.Conf).To(MatchProto(&mesh_proto.TrafficRoute_Conf{
			Http: []*mesh_proto.TrafficRoute_Http{{
				Match: &mesh_proto.TrafficRoute_Http_Match{
					Path: PrefixMatch("/api/aaab"),
				},
			}, {
				Match: &mesh_proto.TrafficRoute_Http_Match{
					Path: PrefixMatch("/api/aaa"),
				},
			}, {
				Match: &mesh_proto.TrafficRoute_Http_Match{
					Path: PrefixMatch("/api"),
				},
			}},
		}))
	})

	It("should order more specific matches first", func() {
		merged := merge.TrafficRoute(
			&core_mesh.TrafficRouteResource{
				Meta: &model.ResourceMeta{
					CreationTime: core.Now(),
				},
				Spec: &mesh_proto.TrafficRoute{
					Conf: &mesh_proto.TrafficRoute_Conf{
						Http: []*mesh_proto.TrafficRoute_Http{{
							Match: &mesh_proto.TrafficRoute_Http_Match{
								Path: RegexMatch("/api"),
							},
						}, {
							Match: &mesh_proto.TrafficRoute_Http_Match{
								Path: PrefixMatch("/api"),
							},
						}, {
							Match: &mesh_proto.TrafficRoute_Http_Match{
								Path: ExactMatch("/api"),
							},
						}, {
							Match: &mesh_proto.TrafficRoute_Http_Match{
								Path: ExactMatch("/api/v2"),
							},
						}},
					},
				},
			},
		)

		Expect(merged.Spec.Conf).To(MatchProto(&mesh_proto.TrafficRoute_Conf{
			Http: []*mesh_proto.TrafficRoute_Http{{
				Match: &mesh_proto.TrafficRoute_Http_Match{
					Path: ExactMatch("/api/v2"),
				},
			}, {
				Match: &mesh_proto.TrafficRoute_Http_Match{
					Path: ExactMatch("/api"),
				},
			}, {
				Match: &mesh_proto.TrafficRoute_Http_Match{
					Path: PrefixMatch("/api"),
				},
			}, {
				Match: &mesh_proto.TrafficRoute_Http_Match{
					Path: RegexMatch("/api"),
				},
			}},
		}))
	})

	It("should order longest header matches first", func() {
		merged := merge.TrafficRoute(
			&core_mesh.TrafficRouteResource{
				Meta: &model.ResourceMeta{
					CreationTime: core.Now(),
				},
				Spec: &mesh_proto.TrafficRoute{
					Conf: &mesh_proto.TrafficRoute_Conf{
						Http: []*mesh_proto.TrafficRoute_Http{{
							Match: &mesh_proto.TrafficRoute_Http_Match{
								Path: ExactMatch("/api"),
								Headers: map[string]*mesh_proto.TrafficRoute_Http_Match_StringMatcher{
									"foo":  ExactMatch("foo"),
									"frog": ExactMatch("foo"),
								},
							},
						}, {
							Match: &mesh_proto.TrafficRoute_Http_Match{
								Path: ExactMatch("/api"),
								Headers: map[string]*mesh_proto.TrafficRoute_Http_Match_StringMatcher{
									"foo":    ExactMatch("foo"),
									"friend": ExactMatch("foo"),
								},
							},
						}, {
							Match: &mesh_proto.TrafficRoute_Http_Match{
								Path: ExactMatch("/api"),
								Headers: map[string]*mesh_proto.TrafficRoute_Http_Match_StringMatcher{
									"foo": ExactMatch("foo"),
									"bar": ExactMatch("foo"),
									"baz": ExactMatch("foo"),
								},
							},
						}},
					},
				},
			},
		)

		Expect(merged.Spec.Conf).To(MatchProto(&mesh_proto.TrafficRoute_Conf{
			Http: []*mesh_proto.TrafficRoute_Http{{
				Match: &mesh_proto.TrafficRoute_Http_Match{
					Path: ExactMatch("/api"),
					Headers: map[string]*mesh_proto.TrafficRoute_Http_Match_StringMatcher{
						"foo": ExactMatch("foo"),
						"bar": ExactMatch("foo"),
						"baz": ExactMatch("foo"),
					},
				},
			}, {
				Match: &mesh_proto.TrafficRoute_Http_Match{
					Path: ExactMatch("/api"),
					// This entry is second because the header maps are equal length, but
					// "friend" is less than "frog".
					Headers: map[string]*mesh_proto.TrafficRoute_Http_Match_StringMatcher{
						"foo":    ExactMatch("foo"),
						"friend": ExactMatch("foo"),
					},
				},
			}, {
				Match: &mesh_proto.TrafficRoute_Http_Match{
					Path: ExactMatch("/api"),
					Headers: map[string]*mesh_proto.TrafficRoute_Http_Match_StringMatcher{
						"foo":  ExactMatch("foo"),
						"frog": ExactMatch("foo"),
					},
				},
			}},
		}))
	})

})
