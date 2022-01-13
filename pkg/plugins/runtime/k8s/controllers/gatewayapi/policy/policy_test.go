package policy_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	kube_client_fake "sigs.k8s.io/controller-runtime/pkg/client/fake"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1alpha2"

	"github.com/kumahq/kuma/pkg/plugins/bootstrap/k8s"
	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/controllers/gatewayapi/policy"
)

var k8sScheme *kube_runtime.Scheme

var _ = BeforeSuite(func() {
	var err error
	k8sScheme, err = k8s.NewScheme()
	Expect(err).NotTo(HaveOccurred())
})

const (
	defaultNs = "default"
	otherNs   = "other"
)

var (
	simplePolicy = gatewayapi.ReferencePolicy{
		ObjectMeta: kube_meta.ObjectMeta{
			Name:      "basic",
			Namespace: otherNs,
		},
		Spec: gatewayapi.ReferencePolicySpec{
			From: []gatewayapi.ReferencePolicyFrom{
				{
					Group:     gatewayapi.Group(gatewayapi.GroupName),
					Kind:      gatewayapi.Kind("HTTPRoute"),
					Namespace: gatewayapi.Namespace(defaultNs),
				},
			},
			To: []gatewayapi.ReferencePolicyTo{
				{
					Group: gatewayapi.Group(""),
					Kind:  gatewayapi.Kind("Service"),
				},
			},
		},
	}

	// References
	coreGroup = gatewayapi.Group("")
	svcKind   = gatewayapi.Kind("Service")
	somePort  = gatewayapi.PortNumber(80)

	toSameSvc = gatewayapi.BackendObjectReference{
		Group: &coreGroup,
		Kind:  &svcKind,
		Name:  "svc",
		Port:  &somePort,
	}

	toOtherNs  = gatewayapi.Namespace(otherNs)
	toOtherSvc = gatewayapi.BackendObjectReference{
		Group:     &coreGroup,
		Kind:      &svcKind,
		Name:      "svc",
		Namespace: &toOtherNs,
		Port:      &somePort,
	}

	kumaGroup          = gatewayapi.Group(mesh_k8s.GroupVersion.Group)
	externalSvcKind    = gatewayapi.Kind("ExternalService")
	toOtherExternalSvc = gatewayapi.BackendObjectReference{
		Group:     &kumaGroup,
		Kind:      &externalSvcKind,
		Name:      "ext-svc",
		Namespace: &toOtherNs,
	}
)

var _ = Describe("ReferencePolicy support", func() {
	It("permits same namespace references", func() {
		kubeClient := kube_client_fake.NewClientBuilder().WithScheme(k8sScheme).WithObjects(
			&simplePolicy,
		).Build()

		ref := policy.PolicyReferenceBackend(policy.FromHTTPRouteIn(defaultNs), toSameSvc)
		permitted, err := policy.IsReferencePermitted(
			context.Background(),
			kubeClient,
			ref,
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(permitted).To(BeTrue())
	})
	Context("respects basic specs", func() {
		It("permitted for matching .to and .from", func() {
			kubeClient := kube_client_fake.NewClientBuilder().WithScheme(k8sScheme).WithObjects(
				&simplePolicy,
			).Build()

			ref := policy.PolicyReferenceBackend(policy.FromHTTPRouteIn(defaultNs), toOtherSvc)
			permitted, err := policy.IsReferencePermitted(
				context.Background(),
				kubeClient,
				ref,
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(permitted).To(BeTrue())
		})
		It("denies for missing .from GroupKind", func() {
			kubeClient := kube_client_fake.NewClientBuilder().WithScheme(k8sScheme).WithObjects(
				&simplePolicy,
			).Build()

			ref := policy.PolicyReferenceBackend(policy.FromHTTPRouteIn(defaultNs), toOtherExternalSvc)
			permitted, err := policy.IsReferencePermitted(
				context.Background(),
				kubeClient,
				ref,
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(permitted).To(BeFalse())
		})
		It("checks names in .from", func() {
			policyWithName := simplePolicy.DeepCopy()
			permittedToExtSvcName := gatewayapi.ObjectName("specific-permitted-ext-svc")
			policyWithName.Spec.To = append(policyWithName.Spec.To,
				gatewayapi.ReferencePolicyTo{
					Group: kumaGroup,
					Kind:  externalSvcKind,
					Name:  &permittedToExtSvcName,
				},
			)

			kubeClient := kube_client_fake.NewClientBuilder().WithScheme(k8sScheme).WithObjects(
				policyWithName,
			).Build()

			By("denying if the name doesn't match")
			ref := policy.PolicyReferenceBackend(policy.FromHTTPRouteIn(defaultNs), toOtherExternalSvc)
			permitted, err := policy.IsReferencePermitted(
				context.Background(),
				kubeClient,
				ref,
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(permitted).To(BeFalse())

			By("permitting if the name matches")
			toOtherSpecificExternalSvc := toOtherExternalSvc.DeepCopy()
			toOtherSpecificExternalSvc.Name = permittedToExtSvcName

			ref = policy.PolicyReferenceBackend(policy.FromHTTPRouteIn(defaultNs), *toOtherSpecificExternalSvc)
			permitted, err = policy.IsReferencePermitted(
				context.Background(),
				kubeClient,
				ref,
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(permitted).To(BeTrue())
		})
	})
})
