package providers_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/api/common/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/kri"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	meshidentity_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshidentity/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/meshidentity/providers"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/events"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
)

var _ = Describe("MeshIdentity providers", func() {
	var eventBus events.EventBus
	var noIdentityManager providers.IdentityProviderManager
	var identityManager providers.IdentityProviderManager
	now := time.Now()

	BeforeEach(func() {
		core.Now = func() time.Time {
			return now
		}
		metrics, err := core_metrics.NewMetrics("")
		Expect(err).ToNot(HaveOccurred())

		eventBus, err = events.NewEventBus(10, metrics)
		Expect(err).ToNot(HaveOccurred())

		identityManager = providers.NewIdentityProviderManager(
			providers.IdentityProviders{
				"Bundled": &staticIdentityProvider{},
			}, eventBus,
		)

		noIdentityManager = providers.NewIdentityProviderManager(
			providers.IdentityProviders{
				"Bundled": &noIdentityProvider{},
			}, eventBus,
		)
	})

	AfterEach(func() {
		core.Now = time.Now
	})
	type testCase struct {
		dpp                  *core_mesh.DataplaneResource
		meshIdentities       []*meshidentity_api.MeshIdentityResource
		expectedIdentityName string
	}
	DescribeTable("SelectIdentity",
		func(given testCase) {
			// when
			identity := identityManager.SelectedIdentity(given.dpp, given.meshIdentities)

			// then
			Expect(identity.Meta.GetName()).To(Equal(given.expectedIdentityName))
		},
		Entry("select the most specific identity", testCase{
			dpp: builders.Dataplane().WithLabels(map[string]string{
				"app": "test-app",
			}).AddInboundHttpOfService("test-app").Build(),
			meshIdentities: []*meshidentity_api.MeshIdentityResource{
				builders.MeshIdentity().WithName("not-matching-1").WithSelector(&v1alpha1.LabelSelector{
					MatchLabels: &map[string]string{
						"app":     "test-app",
						"version": "v1",
					},
				}).Build(),
				builders.MeshIdentity().WithName("matching-all").WithSelector(&v1alpha1.LabelSelector{
					MatchLabels: &map[string]string{},
				}).Build(),
				builders.MeshIdentity().WithName("matching-specific").WithSelector(&v1alpha1.LabelSelector{
					MatchLabels: &map[string]string{
						"app": "test-app",
					},
				}).Build(),
			},
			expectedIdentityName: "matching-specific",
		}),
		Entry("select the most specific identity with 2 tags", testCase{
			dpp: builders.Dataplane().WithLabels(map[string]string{
				"app":     "test-app",
				"version": "v1",
			}).AddInboundHttpOfService("test-app").Build(),
			meshIdentities: []*meshidentity_api.MeshIdentityResource{
				builders.MeshIdentity().WithName("matching-1").WithSelector(&v1alpha1.LabelSelector{
					MatchLabels: &map[string]string{
						"app":     "test-app",
						"version": "v1",
					},
				}).Build(),
				builders.MeshIdentity().WithName("matching-all").WithSelector(&v1alpha1.LabelSelector{
					MatchLabels: &map[string]string{},
				}).Build(),
				builders.MeshIdentity().WithName("matching-specific").WithSelector(&v1alpha1.LabelSelector{
					MatchLabels: &map[string]string{
						"app": "test-app",
					},
				}).Build(),
			},
			expectedIdentityName: "matching-1",
		}),
		Entry("select matching all", testCase{
			dpp: builders.Dataplane().WithLabels(map[string]string{
				"app":     "demo-app",
				"version": "v1",
			}).AddInboundHttpOfService("demo-app").Build(),
			meshIdentities: []*meshidentity_api.MeshIdentityResource{
				builders.MeshIdentity().WithName("matching-1").WithSelector(&v1alpha1.LabelSelector{
					MatchLabels: &map[string]string{
						"app":     "test-app",
						"version": "v1",
					},
				}).Build(),
				builders.MeshIdentity().WithName("matching-all").WithSelector(&v1alpha1.LabelSelector{
					MatchLabels: &map[string]string{},
				}).Build(),
				builders.MeshIdentity().WithName("matching-specific").WithSelector(&v1alpha1.LabelSelector{
					MatchLabels: &map[string]string{
						"app": "test-app",
					},
				}).Build(),
			},
			expectedIdentityName: "matching-all",
		}),
	)
	It("should get Identity", func() {
		// given
		createIdentityListener := eventBus.Subscribe(func(event events.Event) bool {
			switch event.(type) {
			case events.WorkloadIdentityChangedEvent:
				return true
			default:
				return false
			}
		})
		defer createIdentityListener.Close()

		dpp := builders.Dataplane().WithLabels(map[string]string{
			"app":                         "test-app",
			"k8s.kuma.io/service-account": "default",
			"k8s.kuma.io/namespace":       "namespace-demo",
		}).AddInboundHttpOfService("test-app").Build()
		meshIdentity := builders.MeshIdentity().WithBundledAutoGenerated().WithInitializedStatus().Build()

		// when
		identity, err := identityManager.GetWorkloadIdentity(context.Background(), &xds.Proxy{
			Dataplane: dpp,
			Metadata:  &xds.DataplaneMetadata{},
		}, meshIdentity)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(identity).ToNot(BeNil())

		Eventually(createIdentityListener.Recv(), 5*time.Second).Should(Receive(Equal(events.WorkloadIdentityChangedEvent{
			ResourceKey: model.MetaToResourceKey(dpp.GetMeta()),
			Operation:   events.Create,
			Origin:      identity.KRI,
		})))
	})

	It("should not get Identity", func() {
		// given
		createIdentityListener := eventBus.Subscribe(func(event events.Event) bool {
			switch event.(type) {
			case events.WorkloadIdentityChangedEvent:
				return true
			default:
				return false
			}
		})
		defer createIdentityListener.Close()

		dpp := builders.Dataplane().WithLabels(map[string]string{
			"app":                         "test-app",
			"k8s.kuma.io/service-account": "default",
			"k8s.kuma.io/namespace":       "namespace-demo",
		}).AddInboundHttpOfService("test-app").Build()

		// when
		identity, err := noIdentityManager.GetWorkloadIdentity(context.Background(), &xds.Proxy{
			Dataplane: dpp,
			Metadata:  &xds.DataplaneMetadata{},
		}, nil)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(identity).To(BeNil())

		Eventually(createIdentityListener.Recv(), 5*time.Second).Should(Receive(Equal(events.WorkloadIdentityChangedEvent{
			ResourceKey: model.MetaToResourceKey(dpp.GetMeta()),
			Operation:   events.Delete,
		})))
	})
})

type noIdentityProvider struct {
	_ *xds.Proxy
}

func (s *noIdentityProvider) Validate(_ context.Context, _ *meshidentity_api.MeshIdentityResource) error {
	return nil
}

func (s *noIdentityProvider) Initialize(_ context.Context, _ *meshidentity_api.MeshIdentityResource) error {
	return nil
}

func (s *noIdentityProvider) CreateIdentity(_ context.Context, _ *meshidentity_api.MeshIdentityResource, _ *xds.Proxy) (*xds.WorkloadIdentity, error) {
	return nil, nil
}

func (s *noIdentityProvider) GetRootCA(_ context.Context, _ *meshidentity_api.MeshIdentityResource) ([]byte, error) {
	return nil, nil
}

type staticIdentityProvider struct {
	_ *xds.Proxy
}

func (s *staticIdentityProvider) Validate(_ context.Context, _ *meshidentity_api.MeshIdentityResource) error {
	return nil
}

func (s *staticIdentityProvider) Initialize(_ context.Context, _ *meshidentity_api.MeshIdentityResource) error {
	return nil
}

func (s *staticIdentityProvider) CreateIdentity(_ context.Context, mid *meshidentity_api.MeshIdentityResource, _ *xds.Proxy) (*xds.WorkloadIdentity, error) {
	return &xds.WorkloadIdentity{
		KRI:            kri.From(mid),
		ManagementMode: xds.KumaManagementMode,
	}, nil
}

func (s *staticIdentityProvider) GetRootCA(_ context.Context, _ *meshidentity_api.MeshIdentityResource) ([]byte, error) {
	return nil, nil
}
