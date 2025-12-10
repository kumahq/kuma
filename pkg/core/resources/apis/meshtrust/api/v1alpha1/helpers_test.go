package v1alpha1_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshtrust/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/util/pointer"
)

var _ = Describe("MeshTrust Helpers", func() {
	Describe("MigrateOriginToStatus", func() {
		It("should migrate spec.origin to status.origin when status.origin is nil", func() {
			// given
			meshTrust := &v1alpha1.MeshTrustResource{
				Spec: &v1alpha1.MeshTrust{
					Origin: &v1alpha1.Origin{
						KRI: pointer.To("kri://cluster-1/mesh/default/identity/backend-1"),
					},
				},
			}

			// when
			migrated := meshTrust.MigrateOriginToStatus()

			// then
			Expect(migrated).To(BeTrue())
			Expect(meshTrust.Status).ToNot(BeNil())
			Expect(meshTrust.Status.Origin).ToNot(BeNil())
			Expect(meshTrust.Status.Origin.KRI).To(Equal(pointer.To("kri://cluster-1/mesh/default/identity/backend-1")))
		})

		It("should create status when nil during migration", func() {
			// given
			meshTrust := &v1alpha1.MeshTrustResource{
				Spec: &v1alpha1.MeshTrust{
					Origin: &v1alpha1.Origin{
						KRI: pointer.To("kri://cluster-1/mesh/default/identity/backend-1"),
					},
				},
				Status: nil,
			}

			// when
			migrated := meshTrust.MigrateOriginToStatus()

			// then
			Expect(migrated).To(BeTrue())
			Expect(meshTrust.Status).ToNot(BeNil())
			Expect(meshTrust.Status.Origin).ToNot(BeNil())
		})

		It("should not migrate when spec.origin is nil", func() {
			// given
			meshTrust := &v1alpha1.MeshTrustResource{
				Spec: &v1alpha1.MeshTrust{
					Origin: nil,
				},
			}

			// when
			migrated := meshTrust.MigrateOriginToStatus()

			// then
			Expect(migrated).To(BeFalse())
		})

		It("should not migrate when spec is nil", func() {
			// given
			meshTrust := &v1alpha1.MeshTrustResource{
				Spec: nil,
			}

			// when
			migrated := meshTrust.MigrateOriginToStatus()

			// then
			Expect(migrated).To(BeFalse())
		})

		It("should not migrate when status.origin is already set", func() {
			// given
			meshTrust := &v1alpha1.MeshTrustResource{
				Spec: &v1alpha1.MeshTrust{
					Origin: &v1alpha1.Origin{
						KRI: pointer.To("kri://cluster-1/mesh/default/identity/backend-1"),
					},
				},
				Status: &v1alpha1.MeshTrustStatus{
					Origin: &v1alpha1.Origin{
						KRI: pointer.To("kri://cluster-2/mesh/default/identity/backend-2"),
					},
				},
			}

			// when
			migrated := meshTrust.MigrateOriginToStatus()

			// then
			Expect(migrated).To(BeFalse())
			// status.origin should remain unchanged
			Expect(meshTrust.Status.Origin.KRI).To(Equal(pointer.To("kri://cluster-2/mesh/default/identity/backend-2")))
		})
	})
})
