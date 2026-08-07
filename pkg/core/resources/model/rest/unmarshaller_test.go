package rest_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v3/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/v3/pkg/core/validators"
	mtp_api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
)

var _ = Describe("Unmarshal", func() {
	It("should report schema violations in a stable order", func() {
		// given a resource breaking more than one schema rule
		input := []byte(`
type: MeshTrafficPermission
mesh: default
name: sample
spec:
  targetRef:
    kind: NotAKind
    sectionName: 123
`)

		// when unmarshalling it repeatedly
		var orders [][]validators.Violation
		for range 20 {
			_, err := rest.YAML.Unmarshal(input, mtp_api.MeshTrafficPermissionResourceTypeDescriptor)
			Expect(err).To(HaveOccurred())
			verr, ok := err.(*validators.ValidationError)
			Expect(ok).To(BeTrue())
			orders = append(orders, verr.Violations)
		}

		// then every run reports the same violations in the same order
		Expect(orders[0]).To(Equal([]validators.Violation{
			{Field: "spec.targetRef.kind", Message: "in body should be one of [Mesh Dataplane]"},
			{Field: "spec.targetRef.sectionName", Message: `in body must be of type string: "number"`},
		}))
		for _, order := range orders {
			Expect(order).To(Equal(orders[0]))
		}
	})
})
