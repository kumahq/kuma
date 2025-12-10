package v1alpha1_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshtrust/api/v1alpha1"
)

var _ = Describe("MeshTrust", func() {
	It("should support status", func() {
		t := v1alpha1.NewMeshTrustResource()
		val := "kri:foo"
		t.Status.Origin = &v1alpha1.Origin{
			KRI: &val,
		}
		Expect(*t.Status.Origin.KRI).To(Equal("kri:foo"))
	})
})
