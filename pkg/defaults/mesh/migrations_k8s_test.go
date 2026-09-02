package mesh_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kube_core "k8s.io/api/core/v1"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	kube_client_fake "sigs.k8s.io/controller-runtime/pkg/client/fake"

	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	config_core "github.com/kumahq/kuma/v3/pkg/config/core"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/system"
	resources_manager "github.com/kumahq/kuma/v3/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/v3/pkg/core/resources/store"
	secret_cipher "github.com/kumahq/kuma/v3/pkg/core/secrets/cipher"
	secret_manager "github.com/kumahq/kuma/v3/pkg/core/secrets/manager"
	defaults_mesh "github.com/kumahq/kuma/v3/pkg/defaults/mesh"
	bootstrap_k8s "github.com/kumahq/kuma/v3/pkg/plugins/bootstrap/k8s"
	meshtimeout "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshtimeout/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/plugins/resources/k8s"
	secrets_k8s "github.com/kumahq/kuma/v3/pkg/plugins/secrets/k8s"
)

var _ = Describe("EnsureDefaultMeshResources on Kubernetes", func() {
	const systemNamespace = "kuma-system"

	var (
		rawStore    core_store.ResourceStore
		resManager  resources_manager.ResourceManager
		defaultMesh *core_mesh.MeshResource
	)

	BeforeEach(func() {
		k8sScheme, err := bootstrap_k8s.NewScheme()
		Expect(err).ToNot(HaveOccurred())

		kubeClient := kube_client_fake.NewClientBuilder().
			WithScheme(k8sScheme).
			WithIndex(&kube_core.Secret{}, "type", func(object kube_client.Object) []string {
				secret := object.(*kube_core.Secret)
				return []string{string(secret.Type)}
			}).
			Build()

		rawStore, err = k8s.NewStore(kubeClient, k8sScheme, k8s.NewSimpleConverter(systemNamespace))
		Expect(err).ToNot(HaveOccurred())

		scheme, err := bootstrap_k8s.NewScheme()
		Expect(err).ToNot(HaveOccurred())
		secretStore, err := secrets_k8s.NewStore(kubeClient, kubeClient, scheme, "default")
		Expect(err).ToNot(HaveOccurred())

		baseManager := resources_manager.NewResourceManager(rawStore)
		customizableManager := resources_manager.NewCustomizableResourceManager(baseManager, nil)
		customizableManager.Customize(
			system.SecretType,
			secret_manager.NewSecretManager(secretStore, secret_cipher.None()),
		)
		resManager = customizableManager

		defaultMesh = core_mesh.NewMeshResource()
		err = resManager.Create(context.Background(), defaultMesh, core_store.CreateByKey(core_model.DefaultMesh, core_model.NoMesh))
		Expect(err).ToNot(HaveOccurred())
	})

	It("should migrate a legacy combined MeshTimeout default", func() {
		legacy := meshtimeout.NewMeshTimeoutResource()
		legacy.Spec = &meshtimeout.MeshTimeout{
			TargetRef: &common_api.TopLevelTargetRef{Kind: common_api.TopLevelTargetRefKindMesh},
			To: &[]meshtimeout.To{
				{
					TargetRef: common_api.OutboundTargetRef{Kind: common_api.OutboundTargetRefKindMesh},
					Default:   meshtimeout.Conf{IdleTimeout: &kube_meta.Duration{Duration: time.Hour}},
				},
			},
		}
		err := rawStore.Create(context.Background(), legacy, core_store.CreateByKey("mesh-timeout-all-default.kuma-system", core_model.DefaultMesh))
		Expect(err).ToNot(HaveOccurred())

		err = defaults_mesh.EnsureDefaultMeshResources(context.Background(), resManager, defaultMesh, nil, context.Background(), true, systemNamespace, config_core.Zone, "zone-1", false)
		Expect(err).ToNot(HaveOccurred())

		migrated := meshtimeout.NewMeshTimeoutResource()
		err = resManager.Get(context.Background(), migrated, core_store.GetByKey("mesh-timeout-all-default.kuma-system", core_model.DefaultMesh))
		Expect(err).ToNot(HaveOccurred())
		Expect(migrated.Spec.To).To(BeNil())
		Expect(migrated.Spec.Rules).ToNot(BeNil())
		Expect(*migrated.Spec.Rules).To(Equal(currentDefaultMeshTimeoutRulesForStore(resManager, true, systemNamespace)))

		toResource := meshtimeout.NewMeshTimeoutResource()
		err = resManager.Get(context.Background(), toResource, core_store.GetByKey("mesh-timeout-to-all-default.kuma-system", core_model.DefaultMesh))
		Expect(err).ToNot(HaveOccurred())
		Expect(toResource.Spec.To).ToNot(BeNil())
		Expect(toResource.Spec.Rules).To(BeNil())
	})

	It("should delete legacy gateway-specific default MeshTimeout resources", func() {
		gatewayRules := meshtimeout.NewMeshTimeoutResource()
		gatewayRules.Spec = &meshtimeout.MeshTimeout{
			TargetRef: &common_api.TopLevelTargetRef{Kind: common_api.TopLevelTargetRefKindMesh},
			Rules: &[]meshtimeout.Rule{
				{Default: meshtimeout.Conf{IdleTimeout: &kube_meta.Duration{Duration: time.Minute}}},
			},
		}
		err := rawStore.Create(context.Background(), gatewayRules, core_store.CreateByKey("mesh-gateways-timeout-all-default.kuma-system", core_model.DefaultMesh))
		Expect(err).ToNot(HaveOccurred())

		gatewayTo := meshtimeout.NewMeshTimeoutResource()
		gatewayTo.Spec = &meshtimeout.MeshTimeout{
			TargetRef: &common_api.TopLevelTargetRef{Kind: common_api.TopLevelTargetRefKindMesh},
			To: &[]meshtimeout.To{
				{
					TargetRef: common_api.OutboundTargetRef{Kind: common_api.OutboundTargetRefKindMesh},
					Default:   meshtimeout.Conf{IdleTimeout: &kube_meta.Duration{Duration: time.Minute}},
				},
			},
		}
		err = rawStore.Create(context.Background(), gatewayTo, core_store.CreateByKey("mesh-gateways-timeout-to-all-default.kuma-system", core_model.DefaultMesh))
		Expect(err).ToNot(HaveOccurred())

		err = defaults_mesh.EnsureDefaultMeshResources(context.Background(), resManager, defaultMesh, nil, context.Background(), true, systemNamespace, config_core.Zone, "zone-1", false)
		Expect(err).ToNot(HaveOccurred())

		err = resManager.Get(context.Background(), meshtimeout.NewMeshTimeoutResource(), core_store.GetByKey("mesh-gateways-timeout-all-default.kuma-system", core_model.DefaultMesh))
		Expect(core_store.IsNotFound(err)).To(BeTrue())
		err = resManager.Get(context.Background(), meshtimeout.NewMeshTimeoutResource(), core_store.GetByKey("mesh-gateways-timeout-to-all-default.kuma-system", core_model.DefaultMesh))
		Expect(core_store.IsNotFound(err)).To(BeTrue())
	})
})
