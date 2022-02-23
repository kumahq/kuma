package access_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	config_access "github.com/kumahq/kuma/pkg/config/access"
	resources_access "github.com/kumahq/kuma/pkg/core/resources/access"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/user"
)

var _ = Describe("Admin Resource Access", func() {
	resourceAccess := resources_access.NewAdminResourceAccess(config_access.AdminResourcesStaticAccessConfig{
		Users: []string{user.Admin.Name},
	})

	It("should allow regular user to access non admin resource", func() {
		err := resourceAccess.ValidateCreate(
			model.ResourceKey{Name: "xyz", Mesh: "demo"},
			&mesh_proto.CircuitBreaker{},
			mesh.NewCircuitBreakerResource().Descriptor(),
			user.Anonymous,
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
			user.Admin,
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
			user.User{Name: "john doe", Groups: []string{"users"}},
		)

		// then
		Expect(err).To(MatchError(`access denied: user "john doe/users" cannot access the resource of type "Secret"`))
	})

	It("should deny anonymous user to access Create", func() {
		// when
		err := resourceAccess.ValidateCreate(
			model.ResourceKey{Name: "xyz"},
			&system_proto.Secret{},
			system.NewSecretResource().Descriptor(),
			user.Anonymous,
		)

		// then
		Expect(err).To(MatchError(`access denied: user "mesh-system:anonymous/mesh-system:unauthenticated" cannot access the resource of type "Secret"`))
	})

	It("should allow admin to access Update", func() {
		// when
		err := resourceAccess.ValidateUpdate(
			model.ResourceKey{Name: "xyz"},
			&system_proto.Secret{},
			system.NewSecretResource().Descriptor(),
			user.Admin,
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
			user.User{Name: "john doe", Groups: []string{"users"}},
		)

		// then
		Expect(err).To(MatchError(`access denied: user "john doe/users" cannot access the resource of type "Secret"`))
	})

	It("should allow admin to access Get", func() {
		// when
		err := resourceAccess.ValidateGet(
			model.ResourceKey{Name: "xyz"},
			system.NewSecretResource().Descriptor(),
			user.Admin,
		)

		// then
		Expect(err).ToNot(HaveOccurred())
	})

	It("should deny user to access Get", func() {
		// when
		err := resourceAccess.ValidateGet(
			model.ResourceKey{Name: "xyz"},
			system.NewSecretResource().Descriptor(),
			user.User{Name: "john doe", Groups: []string{"users"}},
		)

		// then
		Expect(err).To(MatchError(`access denied: user "john doe/users" cannot access the resource of type "Secret"`))
	})

	It("should allow admin to access List", func() {
		// when
		err := resourceAccess.ValidateList(
			system.NewSecretResource().Descriptor(),
			user.Admin,
		)

		// then
		Expect(err).ToNot(HaveOccurred())
	})

	It("should deny user to access List", func() {
		// when
		err := resourceAccess.ValidateList(
			system.NewSecretResource().Descriptor(),
			user.User{Name: "john doe", Groups: []string{"users"}},
		)

		// then
		Expect(err).To(MatchError(`access denied: user "john doe/users" cannot access the resource of type "Secret"`))
	})
})
