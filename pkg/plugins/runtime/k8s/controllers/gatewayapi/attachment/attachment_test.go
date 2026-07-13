package attachment_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kube_core "k8s.io/api/core/v1"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/kumahq/kuma/v3/pkg/plugins/runtime/k8s/controllers/gatewayapi/attachment"
)

var _ = Describe("EvaluateParentRefAttachment", func() {
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
			serviceRef(kube_core.GroupName),
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(res).To(Equal(attachment.Allowed))
		Expect(kind).To(Equal(attachment.Service))
	})

	It("allows a Service parentRef with the gateway-api group", func() {
		res, kind, err := attachment.EvaluateParentRefAttachment(
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
			gatewayapi.ParentReference{Group: &group, Kind: &kind, Name: "mesh"},
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(res).To(Equal(attachment.Unknown))
		Expect(refKind).To(Equal(attachment.UnknownKind))
	})
})
