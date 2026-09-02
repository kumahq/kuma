package mesh

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_client_fake "sigs.k8s.io/controller-runtime/pkg/client/fake"

	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	config_core "github.com/kumahq/kuma/v3/pkg/config/core"
	"github.com/kumahq/kuma/v3/pkg/core/resources/manager"
	core_store "github.com/kumahq/kuma/v3/pkg/core/resources/store"
	bootstrap_k8s "github.com/kumahq/kuma/v3/pkg/plugins/bootstrap/k8s"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/meshtimeout/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/plugins/resources/k8s"
)

// These exercise the MeshTimeout default migrations directly against a
// Kubernetes-backed store, bypassing EnsureDefaultMeshResources (and its
// Dataplane Token Signing Key step, which needs a Secret-aware store not
// set up here) to isolate the migration logic itself.
var _ = Describe("MeshTimeout default migrations on Kubernetes", func() {
	var resManager manager.ResourceManager

	BeforeEach(func() {
		scheme, err := bootstrap_k8s.NewScheme()
		Expect(err).ToNot(HaveOccurred())
		kubeClient := kube_client_fake.NewClientBuilder().WithScheme(scheme).Build()
		k8sStore, err := k8s.NewStore(kubeClient, scheme, k8s.NewSimpleConverter("kuma-system"))
		Expect(err).ToNot(HaveOccurred())
		resManager = manager.NewResourceManager(k8sStore)
	})

	It("should migrate a legacy combined MeshTimeout default", func() {
		// given a combined default MeshTimeout stored the way CP 2.14 and
		// earlier wrote it on Kubernetes: 'to' set, no 'rules'
		legacy := v1alpha1.NewMeshTimeoutResource()
		legacy.Spec = &v1alpha1.MeshTimeout{
			TargetRef: &common_api.TopLevelTargetRef{Kind: common_api.TopLevelTargetRefKindMesh},
			To: &[]v1alpha1.To{
				{
					TargetRef: common_api.OutboundTargetRef{Kind: common_api.OutboundTargetRefKindMesh},
					Default:   v1alpha1.Conf{IdleTimeout: &kube_meta.Duration{Duration: time.Hour}},
				},
			},
		}
		err := resManager.Create(context.Background(), legacy, core_store.CreateByKey("mesh-timeout-all-default.kuma-system", "default"))
		Expect(err).ToNot(HaveOccurred())

		// when
		err = migrateCombinedMeshTimeoutDefaults(context.Background(), resManager, "default", true, "kuma-system", config_core.Zone, "zone-1", logr.Discard())
		Expect(err).ToNot(HaveOccurred())

		// then the legacy resource is migrated to 'rules' only
		migrated := v1alpha1.NewMeshTimeoutResource()
		err = resManager.Get(context.Background(), migrated, core_store.GetByKey("mesh-timeout-all-default.kuma-system", "default"))
		Expect(err).ToNot(HaveOccurred())
		Expect(migrated.Spec.To).To(BeNil())
		Expect(migrated.Spec.Rules).ToNot(BeNil())

		// and a separate 'to' resource now carries the outbound defaults
		toResource := v1alpha1.NewMeshTimeoutResource()
		err = resManager.Get(context.Background(), toResource, core_store.GetByKey("mesh-timeout-to-all-default.kuma-system", "default"))
		Expect(err).ToNot(HaveOccurred())
		Expect(toResource.Spec.To).ToNot(BeNil())
	})

	It("should delete legacy gateway-specific default MeshTimeout resources", func() {
		// given gateway-specific default MeshTimeout resources, as an older
		// CP version stored them on Kubernetes
		gatewayRules := v1alpha1.NewMeshTimeoutResource()
		gatewayRules.Spec = &v1alpha1.MeshTimeout{
			TargetRef: &common_api.TopLevelTargetRef{Kind: common_api.TopLevelTargetRefKindMesh},
			Rules: &[]v1alpha1.Rule{
				{Default: v1alpha1.Conf{IdleTimeout: &kube_meta.Duration{Duration: time.Minute}}},
			},
		}
		err := resManager.Create(context.Background(), gatewayRules, core_store.CreateByKey("mesh-gateways-timeout-all-default.kuma-system", "default"))
		Expect(err).ToNot(HaveOccurred())

		gatewayTo := v1alpha1.NewMeshTimeoutResource()
		gatewayTo.Spec = &v1alpha1.MeshTimeout{
			TargetRef: &common_api.TopLevelTargetRef{Kind: common_api.TopLevelTargetRefKindMesh},
			To: &[]v1alpha1.To{
				{
					TargetRef: common_api.OutboundTargetRef{Kind: common_api.OutboundTargetRefKindMesh},
					Default:   v1alpha1.Conf{IdleTimeout: &kube_meta.Duration{Duration: time.Minute}},
				},
			},
		}
		err = resManager.Create(context.Background(), gatewayTo, core_store.CreateByKey("mesh-gateways-timeout-to-all-default.kuma-system", "default"))
		Expect(err).ToNot(HaveOccurred())

		// when
		err = migrateGatewayMeshTimeoutDefaults(context.Background(), resManager, "default", true, "kuma-system", logr.Discard())
		Expect(err).ToNot(HaveOccurred())

		// then the legacy gateway-specific defaults are gone
		err = resManager.Get(context.Background(), v1alpha1.NewMeshTimeoutResource(), core_store.GetByKey("mesh-gateways-timeout-all-default.kuma-system", "default"))
		Expect(core_store.IsNotFound(err)).To(BeTrue())
		err = resManager.Get(context.Background(), v1alpha1.NewMeshTimeoutResource(), core_store.GetByKey("mesh-gateways-timeout-to-all-default.kuma-system", "default"))
		Expect(core_store.IsNotFound(err)).To(BeTrue())
	})
})
