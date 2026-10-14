package k8s_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kube_core "k8s.io/api/core/v1"
	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	common_api "github.com/kumahq/kuma/v2/api/common/v1alpha1"
	meshtrust_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshtrust/api/v1alpha1"
	_ "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshtrust/k8s/v1alpha1"
	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/core/resources/store"
	meshtimeout_api "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshtimeout/api/v1alpha1"
	_ "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshtimeout/k8s/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/plugins/resources/k8s"
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
			TargetRef: &common_api.TargetRef{Kind: common_api.Mesh},
		}
		return r
	}

	BeforeEach(func() {
		for _, ns := range []string{systemNamespace, "tenant"} {
			err := k8sClient.Create(context.Background(), &kube_core.Namespace{ObjectMeta: kube_meta.ObjectMeta{Name: ns}})
			if !kube_apierrs.IsAlreadyExists(err) {
				Expect(err).ToNot(HaveOccurred())
			}
		}
	})

	DescribeTable("should read only what the admission webhook would have allowed",
		func(isGlobal bool, newResource func() core_model.Resource, mesh string, expected []string) {
			// given
			s, err := k8s.NewStore(k8sClient, k8sClientScheme, k8s.NewSimpleConverter(systemNamespace), systemNamespace, isGlobal)
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

			// and then
			err = s.Get(context.Background(), newResource(), store.GetByKey(mesh+".tenant", mesh))
			if len(expected) == 1 {
				Expect(store.IsNotFound(err)).To(BeTrue())
			} else {
				Expect(err).ToNot(HaveOccurred())
			}
		},
		Entry("hides a system-namespace-only type outside the system namespace on Zone",
			false, meshTrust, "zone-trust", []string{"zone-trust." + systemNamespace}),
		Entry("hides any namespaced type outside the system namespace on Global",
			true, meshTimeout, "global-timeout", []string{"global-timeout." + systemNamespace}),
		Entry("reads a regular type from any namespace on Zone",
			false, meshTimeout, "zone-timeout", []string{"zone-timeout." + systemNamespace, "zone-timeout.tenant"}),
	)
})
