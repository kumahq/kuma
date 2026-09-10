package mesh_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/v3/pkg/core/resources/manager"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/v3/pkg/core/resources/store"
	"github.com/kumahq/kuma/v3/pkg/core/tokens"
	"github.com/kumahq/kuma/v3/pkg/defaults/mesh"
	meshcircuitbreaker "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshcircuitbreaker/api/v1alpha1"
	meshretry "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshretry/api/v1alpha1"
	meshtimeout "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshtimeout/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/plugins/resources/memory"
)

var _ = Describe("EnsureDefaultMeshResources", func() {
	var resManager manager.ResourceManager
	var defaultMesh *core_mesh.MeshResource

	signingKeyExists := func() error {
		return resManager.Get(context.Background(), system.NewSecretResource(), core_store.GetBy(tokens.SigningKeyResourceKey(system.DataplaneTokenSigningKey(model.DefaultMesh), tokens.DefaultKeyID, model.DefaultMesh)))
	}

	BeforeEach(func() {
		resManager = manager.NewResourceManager(memory.NewStore())
		defaultMesh = core_mesh.NewMeshResource()

		err := resManager.Create(context.Background(), defaultMesh, core_store.CreateByKey(model.DefaultMesh, model.NoMesh))
		Expect(err).ToNot(HaveOccurred())
	})

	It("should create the Dataplane Token Signing Key", func() {
		// when
		err := mesh.EnsureDefaultMeshResources(context.Background(), resManager, defaultMesh, context.Background())
		Expect(err).ToNot(HaveOccurred())

		// then
		Expect(signingKeyExists()).To(Succeed())
	})

	It("should ignore subsequent calls to EnsureDefaultMeshResources", func() {
		// given already ensured default resources
		err := mesh.EnsureDefaultMeshResources(context.Background(), resManager, defaultMesh, context.Background())
		Expect(err).ToNot(HaveOccurred())

		// when ensuring again
		err = mesh.EnsureDefaultMeshResources(context.Background(), resManager, defaultMesh, context.Background())

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(signingKeyExists()).To(Succeed())
	})

	It("should not create any policy", func() {
		// when
		err := mesh.EnsureDefaultMeshResources(context.Background(), resManager, defaultMesh, context.Background())
		Expect(err).ToNot(HaveOccurred())

		// then a mesh starts with no policies at all
		err = resManager.Get(context.Background(), meshretry.NewMeshRetryResource(), core_store.GetByKey("mesh-retry-all-default", model.DefaultMesh))
		Expect(core_store.IsNotFound(err)).To(BeTrue())

		err = resManager.Get(context.Background(), meshtimeout.NewMeshTimeoutResource(), core_store.GetByKey("mesh-timeout-all-default", model.DefaultMesh))
		Expect(core_store.IsNotFound(err)).To(BeTrue())

		err = resManager.Get(context.Background(), meshcircuitbreaker.NewMeshCircuitBreakerResource(), core_store.GetByKey("mesh-circuit-breaker-all-default", model.DefaultMesh))
		Expect(core_store.IsNotFound(err)).To(BeTrue())
	})
})
