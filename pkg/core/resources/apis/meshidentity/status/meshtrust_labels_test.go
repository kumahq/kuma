package status

import (
	"context"
	"maps"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	config_core "github.com/kumahq/kuma/v3/pkg/config/core"
	"github.com/kumahq/kuma/v3/pkg/core/kri"
	meshtrust_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshtrust/api/v1alpha1"
	resource_labels "github.com/kumahq/kuma/v3/pkg/core/resources/labels"
	"github.com/kumahq/kuma/v3/pkg/core/resources/manager"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/store"
	"github.com/kumahq/kuma/v3/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/v3/pkg/test/resources/model"
	"github.com/kumahq/kuma/v3/pkg/test/resources/samples"
)

var _ = Describe("Generated MeshTrust labels", func() {
	DescribeTable("create and repair local labels without changing trust data",
		func(isK8s, syncedIdentity bool) {
			ctx := context.Background()
			resManager := manager.NewResourceManager(memory.NewStore())
			Expect(samples.MeshDefaultBuilder().Create(resManager)).To(Succeed())
			cp := resource_labels.ControlPlane{Mode: config_core.Zone, Zone: "east", IsK8s: isK8s}
			updater, err := New(logr.Discard(), 0, resManager, resManager, nil, cp)
			Expect(err).ToNot(HaveOccurred())

			name := "identity"
			identity := builders.MeshIdentity().WithSpiffeID("default.east.mesh.local", "/service").Build()
			meta := identity.Meta.(*test_model.ResourceMeta)
			meta.Labels = map[string]string{mesh_proto.DisplayName: name, mesh_proto.ZoneTag: "east"}
			if syncedIdentity {
				name += "-hashed"
				meta.Labels[mesh_proto.ResourceOriginLabel] = string(mesh_proto.GlobalResourceOrigin)
				delete(meta.Labels, mesh_proto.ZoneTag)
			}
			meta.Name = name
			namespace := ""
			environment := mesh_proto.UniversalEnvironment
			if isK8s {
				namespace = "kuma-system"
				environment = mesh_proto.KubernetesEnvironment
				meta.Name += "." + namespace
				meta.NameExtensions = model.ResourceNameExtensions{
					model.K8sNameComponent:      name,
					model.K8sNamespaceComponent: namespace,
				}
			}
			ca := []byte(builders.MeshTrust().Build().Spec.CABundles[0].PEM.Value)
			load := func() *meshtrust_api.MeshTrustResource {
				trust := meshtrust_api.NewMeshTrustResource()
				Expect(resManager.Get(ctx, trust, store.GetByKey(meta.Name, meta.Mesh))).To(Succeed())
				return trust
			}

			Expect(updater.createOrUpdateMeshTrust(ctx, identity, ca)).To(Succeed())
			created := load()
			expectedLabels := map[string]string{
				mesh_proto.DisplayName:         name,
				mesh_proto.ZoneTag:             "east",
				mesh_proto.ResourceOriginLabel: string(mesh_proto.ZoneResourceOrigin),
				mesh_proto.EnvTag:              environment,
				mesh_proto.MeshTag:             "default",
			}
			if isK8s {
				expectedLabels[mesh_proto.KubeNamespaceTag] = namespace
			}
			Expect(created.GetMeta().GetLabels()).To(Equal(expectedLabels))
			Expect(kri.From(created)).To(Equal(kri.Identifier{
				ResourceType: meshtrust_api.MeshTrustType,
				Mesh:         "default",
				Zone:         "east",
				Namespace:    namespace,
				Name:         name,
			}))

			// Existing trusts must be repaired even when the CA, domain and status
			// already match. Keep operator labels and avoid subsequent no-op writes.
			for _, oldLabels := range []map[string]string{
				nil,
				{"example.com/owner": "team", mesh_proto.ZoneTag: "old-zone"},
			} {
				Expect(resManager.Update(ctx, load(), store.UpdateWithLabels(oldLabels))).To(Succeed())
				Expect(updater.createOrUpdateMeshTrust(ctx, identity, ca)).To(Succeed())
				repaired := load()
				desired := maps.Clone(expectedLabels)
				if owner, ok := oldLabels["example.com/owner"]; ok {
					desired["example.com/owner"] = owner
				}
				Expect(repaired.GetMeta().GetLabels()).To(Equal(desired))
				Expect(repaired.Spec).To(Equal(created.Spec))
				Expect(repaired.Status).To(Equal(created.Status))
				Expect(updater.createOrUpdateMeshTrust(ctx, identity, ca)).To(Succeed())
				Expect(load().GetMeta().GetVersion()).To(Equal(repaired.GetMeta().GetVersion()))
			}
		},
		Entry("Universal local identity", false, false),
		Entry("Universal global identity", false, true),
		Entry("Kubernetes local identity", true, false),
		Entry("Kubernetes global identity", true, true),
	)
})
