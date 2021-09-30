package rbac_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	config_rbac "github.com/kumahq/kuma/pkg/config/rbac"
	"github.com/kumahq/kuma/pkg/core/rbac"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	resources_rbac "github.com/kumahq/kuma/pkg/core/resources/rbac"
	"github.com/kumahq/kuma/pkg/core/user"
)

var _ = Describe("Admin Resource Access", func() {
	resourceAccess := resources_rbac.NewAdminResourceAccess(rbac.NewStaticRoleAssignments(config_rbac.RBACStaticConfig{
		AdminUsers: []string{"admin"},
	}))

	It("should allow regular user to access non admin resource", func() {
		err := resourceAccess.ValidateCreate(
			model.ResourceKey{Name: "xyz", Mesh: "demo"},
			&mesh_proto.CircuitBreaker{},
			mesh.NewCircuitBreakerResource().Descriptor(),
			nil,
		)

		// then
		Expect(err).ToNot(HaveOccurred())
	})

	It("should allow admin to access Create", func() {
		// when
		err := resourceAccess.ValidateCreate(
			model.ResourceKey{Name: "xyz"},
			&system_proto.Secret{},
			system.NewSecretResource().Descriptor(),
			&user.Admin,
		)

		// then
		Expect(err).ToNot(HaveOccurred())
	})

	It("should deny user to access Create", func() {
		// when
		err := resourceAccess.ValidateCreate(
			model.ResourceKey{Name: "xyz"},
			&system_proto.Secret{},
			system.NewSecretResource().Descriptor(),
			&user.User{Name: "john doe", Group: "users"},
		)

		// then
		Expect(err).To(MatchError(`access denied: user "john doe/users" of role "Member" cannot access the resource of type "Secret"`))
	})

	It("should deny anonymous user to access Create", func() {
		// when
		err := resourceAccess.ValidateCreate(
			model.ResourceKey{Name: "xyz"},
			&system_proto.Secret{},
			system.NewSecretResource().Descriptor(),
			nil,
		)

		// then
		Expect(err).To(MatchError(`access denied: user did not authenticate`))
	})

	It("should allow admin to access Update", func() {
		// when
		err := resourceAccess.ValidateUpdate(
			model.ResourceKey{Name: "xyz"},
			&system_proto.Secret{},
			system.NewSecretResource().Descriptor(),
			&user.Admin,
		)

		// then
		Expect(err).ToNot(HaveOccurred())
	})

	It("should deny user to access Update", func() {
		// when
		err := resourceAccess.ValidateUpdate(
			model.ResourceKey{Name: "xyz"},
			&system_proto.Secret{},
			system.NewSecretResource().Descriptor(),
			&user.User{Name: "john doe", Group: "users"},
		)

		// then
		Expect(err).To(MatchError(`access denied: user "john doe/users" of role "Member" cannot access the resource of type "Secret"`))
	})

	It("should allow admin to access Get", func() {
		// when
		err := resourceAccess.ValidateGet(
			model.ResourceKey{Name: "xyz"},
			system.NewSecretResource().Descriptor(),
			&user.Admin,
		)

		// then
		Expect(err).ToNot(HaveOccurred())
	})

	It("should deny user to access Get", func() {
		// when
		err := resourceAccess.ValidateGet(
			model.ResourceKey{Name: "xyz"},
			system.NewSecretResource().Descriptor(),
			&user.User{Name: "john doe", Group: "users"},
		)

		// then
		Expect(err).To(MatchError(`access denied: user "john doe/users" of role "Member" cannot access the resource of type "Secret"`))
	})

	It("should allow admin to access List", func() {
		// when
		err := resourceAccess.ValidateList(
			system.NewSecretResource().Descriptor(),
			&user.Admin,
		)

		// then
		Expect(err).ToNot(HaveOccurred())
	})

	It("should deny user to access List", func() {
		// when
		err := resourceAccess.ValidateList(
			system.NewSecretResource().Descriptor(),
			&user.User{Name: "john doe", Group: "users"},
		)

		// then
		Expect(err).To(MatchError(`access denied: user "john doe/users" of role "Member" cannot access the resource of type "Secret"`))
	})
})
