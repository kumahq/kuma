package attachment_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kube_core "k8s.io/api/core/v1"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	kube_client_fake "sigs.k8s.io/controller-runtime/pkg/client/fake"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/kumahq/kuma/v3/pkg/plugins/bootstrap/k8s"
	"github.com/kumahq/kuma/v3/pkg/plugins/runtime/k8s/controllers/gatewayapi/attachment"
)

var k8sScheme *kube_runtime.Scheme

var _ = BeforeSuite(func() {
	var err error
	k8sScheme, err = k8s.NewScheme()
	Expect(err).NotTo(HaveOccurred())
})

var _ = Describe("EvaluateParentRefAttachment", func() {
	var kubeClient kube_client.Client
	var routeNs *kube_core.Namespace

	BeforeEach(func() {
		routeNs = &kube_core.Namespace{
			ObjectMeta: kube_meta.ObjectMeta{Name: "default"},
		}
		kubeClient = kube_client_fake.NewClientBuilder().WithScheme(k8sScheme).WithObjects(routeNs).Build()
	})

	serviceRef := func(group gatewayapi.Group) gatewayapi.ParentReference {
		kind := gatewayapi.Kind("Service")
		return gatewayapi.ParentReference{
			Group: &group,
			Kind:  &kind,
			Name:  "backend",
		}
	}

	It("allows a Service parentRef with the core group", func() {
		res, kind, err := attachment.EvaluateParentRefAttachment(
			context.Background(),
			kubeClient,
			nil,
			routeNs,
			serviceRef(kube_core.GroupName),
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(res).To(Equal(attachment.Allowed))
		Expect(kind).To(Equal(attachment.Service))
	})

	It("allows a Service parentRef with the gateway-api group", func() {
		res, kind, err := attachment.EvaluateParentRefAttachment(
			context.Background(),
			kubeClient,
			nil,
			routeNs,
			serviceRef(gatewayapi.GroupName),
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(res).To(Equal(attachment.Allowed))
		Expect(kind).To(Equal(attachment.Service))
	})

	It("reports Unknown for an unsupported parentRef kind", func() {
		group := gatewayapi.Group(gatewayapi.GroupName)
		kind := gatewayapi.Kind("Mesh")
		res, refKind, err := attachment.EvaluateParentRefAttachment(
			context.Background(),
			kubeClient,
			nil,
			routeNs,
			gatewayapi.ParentReference{Group: &group, Kind: &kind, Name: "mesh"},
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(res).To(Equal(attachment.Unknown))
		Expect(refKind).To(Equal(attachment.UnknownKind))
	})
})
