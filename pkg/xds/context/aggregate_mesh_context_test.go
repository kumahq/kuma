package context_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
)

var _ = Describe("AggregateMeshContexts", func() {
	It("should ignore meshes that were deleted", func() {
		// given
		resManager := manager.NewResourceManager(memory.NewStore())
		Expect(samples.MeshDefaultBuilder().Create(resManager)).To(Succeed())
		Expect(samples.MeshDefaultBuilder().WithName("other").Create(resManager)).To(Succeed())

		fetcher := func(ctx context.Context, meshName string) (xds_context.MeshContext, error) {
			if meshName == "other" {
				return xds_context.MeshContext{}, core_store.ErrorResourceNotFound(mesh.MeshType, "other", model.NoMesh)
			}
			return xds_context.MeshContext{}, nil
		}

		// when
		ctxs, err := xds_context.AggregateMeshContexts(context.Background(), resManager, fetcher)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(ctxs.Meshes).To(HaveLen(1))
		Expect(ctxs.Meshes[0].GetMeta().GetName()).To(Equal(model.DefaultMesh))
	})
})
