package match_test

import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	model "github.com/kumahq/kuma/pkg/core/resources/model"
	rest_v1alpha1 "github.com/kumahq/kuma/pkg/core/resources/model/rest/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway/match"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway/route"
	"github.com/kumahq/kuma/pkg/test"
)

func TestMatches(t *testing.T) {
	test.RunSpecs(t, "Match Suite")
}

var _ = Describe("Match Destination Policy", func() {
	Named := func(name string, resource model.Resource) model.Resource {
		resource.SetMeta(&rest_v1alpha1.ResourceMeta{
			Type:             string(resource.Descriptor().Name),
			Mesh:             "default",
			Name:             name,
			CreationTime:     time.Now(),
			ModificationTime: time.Now(),
		})

		return resource
	}

	SpecFor := func(svc string) *mesh_proto.Timeout {
		return &mesh_proto.Timeout{
			Destinations: []*mesh_proto.Selector{{
				Match: map[string]string{mesh_proto.ServiceTag: svc},
			}},
		}
	}

	It("should return nil for no policies", func() {
		// given
		d := []route.Destination{{}, {}}
		// then
		Expect(match.BestConnectionPolicyForDestination(d, core_mesh.TimeoutType)).To(BeNil())
	})

	It("should deduplicate a single policy for multiple destinations", func() {
		// given
		p := Named("default-policy", &core_mesh.TimeoutResource{
			Spec: &mesh_proto.Timeout{
				Destinations: []*mesh_proto.Selector{{
					Match: map[string]string{mesh_proto.ServiceTag: "foo-service"},
				}, {
					Match: map[string]string{mesh_proto.ServiceTag: "bar-service"},
				}, {
					Match: map[string]string{mesh_proto.ServiceTag: "baz-service"},
				}},
			},
		})

		d := []route.Destination{{
			Destination: map[string]string{mesh_proto.ServiceTag: "foo-service"},
			Policies:    map[model.ResourceType]model.Resource{core_mesh.TimeoutType: p},
		}, {
			Destination: map[string]string{mesh_proto.ServiceTag: "bar-service"},
			Policies:    map[model.ResourceType]model.Resource{core_mesh.TimeoutType: p},
		}, {
			Destination: map[string]string{mesh_proto.ServiceTag: "baz-service"},
			Policies:    map[model.ResourceType]model.Resource{core_mesh.TimeoutType: p},
		}}

		// when
		policy := match.BestConnectionPolicyForDestination(d, core_mesh.TimeoutType)

		// then
		Expect(policy).ToNot(BeNil())
		Expect(policy.GetMeta().GetName()).To(Equal("default-policy"))
	})

	It("should prefer wildcard policy for multiple destinations", func() {
		// given
		d := []route.Destination{{
			Destination: map[string]string{mesh_proto.ServiceTag: "foo-service"},
			Policies: map[model.ResourceType]model.Resource{
				core_mesh.TimeoutType: Named("foo-policy", &core_mesh.TimeoutResource{
					Spec: SpecFor("foo-service"),
				}),
			},
		}, {
			Destination: map[string]string{mesh_proto.ServiceTag: "bar-service"},
			Policies: map[model.ResourceType]model.Resource{
				core_mesh.TimeoutType: Named("bar-policy", &core_mesh.TimeoutResource{
					Spec: SpecFor("bar-service"),
				}),
			},
		}, {
			Destination: map[string]string{mesh_proto.ServiceTag: "baz-service"},
			Policies: map[model.ResourceType]model.Resource{
				core_mesh.TimeoutType: Named("wildcard-policy", &core_mesh.TimeoutResource{
					Spec: SpecFor(mesh_proto.MatchAllTag),
				}),
			},
		}}

		// when
		policy := match.BestConnectionPolicyForDestination(d, core_mesh.TimeoutType)

		// then
		Expect(policy).ToNot(BeNil())
		Expect(policy.GetMeta().GetName()).To(Equal("wildcard-policy"))
	})

	It("should prefer the oldest policy", func() {
		TimeFrom := func(spec string) time.Time {
			t, err := time.Parse(time.RFC822, spec)
			Expect(err).ToNot(HaveOccurred())
			return t
		}

		SetCreationTime := func(r model.Resource, t time.Time) {
			meta := r.GetMeta().(*rest_v1alpha1.ResourceMeta)
			meta.CreationTime = t
			r.SetMeta(meta)
		}

		// given
		d := []route.Destination{{
			Destination: map[string]string{mesh_proto.ServiceTag: "foo-service"},
			Policies: map[model.ResourceType]model.Resource{
				core_mesh.TimeoutType: Named("foo-policy", &core_mesh.TimeoutResource{
					Spec: SpecFor("foo-service"),
				}),
			},
		}, {
			Destination: map[string]string{mesh_proto.ServiceTag: "bar-service"},
			Policies: map[model.ResourceType]model.Resource{
				core_mesh.TimeoutType: Named("bar-policy", &core_mesh.TimeoutResource{
					Spec: SpecFor("bar-service"),
				}),
			},
		}, {
			Destination: map[string]string{mesh_proto.ServiceTag: "baz-service"},
			Policies: map[model.ResourceType]model.Resource{
				core_mesh.TimeoutType: Named("baz-policy", &core_mesh.TimeoutResource{
					Spec: SpecFor("baz-service"),
				}),
			},
		}}

		SetCreationTime(d[0].Policies[core_mesh.TimeoutType], TimeFrom("02 Jan 06 00:01 UTC"))
		SetCreationTime(d[1].Policies[core_mesh.TimeoutType], TimeFrom("02 Jan 06 00:00 UTC"))
		SetCreationTime(d[2].Policies[core_mesh.TimeoutType], TimeFrom("02 Jan 06 00:03 UTC"))

		// when
		policy := match.BestConnectionPolicyForDestination(d, core_mesh.TimeoutType)

		// then
		Expect(policy).ToNot(BeNil())
		Expect(policy.GetMeta().GetName()).To(Equal("bar-policy"))
	})

})
