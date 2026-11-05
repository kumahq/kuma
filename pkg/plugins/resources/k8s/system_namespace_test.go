package k8s_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kube_core "k8s.io/api/core/v1"
	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"

	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	meshtrust_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshtrust/api/v1alpha1"
	_ "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshtrust/k8s/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/resources/labels"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/store"
	meshtimeout_api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshtimeout/api/v1alpha1"
	_ "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshtimeout/k8s/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/plugins/resources/k8s"
)

var _ = Describe("KubernetesStore namespace rule on read", func() {
	const systemNamespace = "kuma-system"

	meshTrust := func() core_model.Resource {
		r := meshtrust_api.NewMeshTrustResource()
		r.Spec = &meshtrust_api.MeshTrust{
			TrustDomain: "default.zone-1.mesh.local",
			CABundles: []meshtrust_api.CABundle{{
				Type: meshtrust_api.PemCABundleType,
				PEM:  &meshtrust_api.PEM{Value: "-----BEGIN CERTIFICATE-----"},
			}},
		}
		return r
	}
	meshTimeout := func() core_model.Resource {
		r := meshtimeout_api.NewMeshTimeoutResource()
		r.Spec = &meshtimeout_api.MeshTimeout{
			TargetRef: &common_api.TopLevelTargetRef{Kind: common_api.TopLevelTargetRefKindMesh},
		}
		return r
	}

	BeforeEach(func() {
		for _, ns := range []string{systemNamespace, "tenant"} {
			err := k8sClient.Create(context.Background(), &kube_core.Namespace{Name: ns})
			if !kube_apierrs.IsAlreadyExists(err) {
				Expect(err).ToNot(HaveOccurred())
			}
		}
	})

	DescribeTable("should read only what the admission webhook would have allowed",
		func(isGlobal bool, newResource func() core_model.Resource, mesh string, expected []string) {
			// given
			s, err := k8s.NewStore(k8sClient, k8sClientScheme, k8s.NewSimpleConverter(systemNamespace, labels.ControlPlane{}), systemNamespace, isGlobal)
			Expect(err).ToNot(HaveOccurred())
			Expect(s.Create(context.Background(), newResource(), store.CreateByKey(mesh+"."+systemNamespace, mesh))).To(Succeed())
			Expect(s.Create(context.Background(), newResource(), store.CreateByKey(mesh+".tenant", mesh))).To(Succeed())

			// when
			list := newResource().Descriptor().NewList()
			Expect(s.List(context.Background(), list, store.ListByMesh(mesh))).To(Succeed())

			// then
			var names []string
			for _, item := range list.GetItems() {
				names = append(names, item.GetMeta().GetName())
			}
			Expect(names).To(ConsistOf(expected))
			Expect(list.GetPagination().Total).To(Equal(uint32(len(expected))))

			// and then
			Expect(s.Get(context.Background(), newResource(), store.GetByKey(mesh+"."+systemNamespace, mesh))).To(Succeed())
			err = s.Get(context.Background(), newResource(), store.GetByKey(mesh+".tenant", mesh))
			if len(expected) == 1 {
				Expect(store.IsNotFound(err)).To(BeTrue())
			} else {
				Expect(err).ToNot(HaveOccurred())
			}
		},
		Entry("hides a system-namespace-only type outside the system namespace on Zone",
			false, meshTrust, "zone-trust", []string{"zone-trust." + systemNamespace}),
		Entry("hides a system-namespace-only type outside the system namespace on Global",
			true, meshTrust, "global-trust", []string{"global-trust." + systemNamespace}),
		Entry("hides any namespaced type outside the system namespace on Global",
			true, meshTimeout, "global-timeout", []string{"global-timeout." + systemNamespace}),
		Entry("reads a regular type from any namespace on Zone",
			false, meshTimeout, "zone-timeout", []string{"zone-timeout." + systemNamespace, "zone-timeout.tenant"}),
	)

	It("reads cluster-scoped resources on Global", func() {
		// given
		s, err := k8s.NewStore(k8sClient, k8sClientScheme, k8s.NewSimpleConverter(systemNamespace, labels.ControlPlane{}), systemNamespace, true)
		Expect(err).ToNot(HaveOccurred())
		const name = "global-cluster-scoped"
		Expect(s.Create(context.Background(), core_mesh.NewMeshResource(), store.CreateByKey(name, core_model.NoMesh))).To(Succeed())

		// when
		mesh := core_mesh.NewMeshResource()
		err = s.Get(context.Background(), mesh, store.GetByKey(name, core_model.NoMesh))
		list := &core_mesh.MeshResourceList{}
		listErr := s.List(context.Background(), list)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(mesh.GetMeta().GetName()).To(Equal(name))
		Expect(listErr).ToNot(HaveOccurred())
		Expect(core_model.ResourceListToResourceKeys(list)).To(ContainElement(core_model.WithoutMesh(name)))
		Expect(list.GetPagination().Total).To(Equal(uint32(len(list.Items))))
	})
})
