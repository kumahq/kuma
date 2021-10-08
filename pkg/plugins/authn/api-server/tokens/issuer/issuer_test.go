package issuer_test

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v4"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/wrapperspb"

	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/secrets/cipher"
	secret_manager "github.com/kumahq/kuma/pkg/core/secrets/manager"
	secret_store "github.com/kumahq/kuma/pkg/core/secrets/store"
	"github.com/kumahq/kuma/pkg/core/user"
	. "github.com/kumahq/kuma/pkg/plugins/authn/api-server/tokens/issuer"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
)

var _ = Describe("User token issuer", func() {

	var issuer UserTokenIssuer
	var store core_store.ResourceStore
	var signingKeyManager SigningKeyManager

	now := time.Now()

	BeforeEach(func() {
		store = memory.NewStore()
		secretManager := secret_manager.NewGlobalSecretManager(secret_store.NewSecretStore(store), cipher.None())
		signingKeyManager = NewSigningKeyManager(secretManager)
		issuer = NewUserTokenIssuer(signingKeyManager, NewTokenRevocations(secretManager))

		Expect(signingKeyManager.CreateDefaultSigningKey()).To(Succeed())
		core.Now = func() time.Time {
			return now
		}
		jwt.TimeFunc = func() time.Time {
			return now
		}
	})

	AfterEach(func() {
		core.Now = time.Now
		jwt.TimeFunc = time.Now
	})

	It("should support rotation", func() {
		// given
		id := user.User{
			Name:   "john.doe@example.com",
			Groups: []string{"users"},
		}

		// when
		token1, err := issuer.Generate(id, time.Second*60)
		Expect(err).ToNot(HaveOccurred())

		// then
		_, err = issuer.Validate(token1)
		Expect(err).ToNot(HaveOccurred())

		// when new signing key with higher serial number is created
		err = signingKeyManager.CreateSigningKey(2)
		Expect(err).ToNot(HaveOccurred())

		// and a new token is generated
		token2, err := issuer.Generate(id, time.Second*60)
		Expect(err).ToNot(HaveOccurred())

		// then all tokens are valid because 2 signing keys are present in the system
		_, err = issuer.Validate(token1)
		Expect(err).ToNot(HaveOccurred())
		_, err = issuer.Validate(token2)
		Expect(err).ToNot(HaveOccurred())

		// when first signing key is deleted
		err = store.Delete(context.Background(), system.NewGlobalSecretResource(), core_store.DeleteBy(SigningKeyResourceKey(DefaultSerialNumber)))
		Expect(err).ToNot(HaveOccurred())

		// then old tokens are no longer valid
		_, err = issuer.Validate(token1)
		Expect(err).To(MatchError("could not parse token: could not get signing key with serial number 1. The signing key most likely has been rotated, regenerate the token: there is no signing key in the Control Plane"))

		// and new token is valid because new signing key is present
		_, err = issuer.Validate(token2)
		Expect(err).ToNot(HaveOccurred())
	})

	It("should validate out expired tokens", func() {
		// given
		id := user.User{
			Name:   "john.doe@example.com",
			Groups: []string{"users"},
		}
		token, err := issuer.Generate(id, 60*time.Second)
		Expect(err).ToNot(HaveOccurred())

		// when
		now = now.Add(60*time.Second + 1*time.Second)
		_, err = issuer.Validate(token)

		// then
		Expect(err.Error()).To(ContainSubstring("could not parse token: token is expired"))
	})

	It("should revoke token", func() {
		// given valid token
		id := user.User{
			Name:   "john.doe@example.com",
			Groups: []string{"users"},
		}

		token, err := issuer.Generate(id, 60*time.Second)
		Expect(err).ToNot(HaveOccurred())
		_, err = issuer.Validate(token)
		Expect(err).ToNot(HaveOccurred())

		// when id of the token is added to revocation list
		c := &jwt.RegisteredClaims{}
		_, _, err = new(jwt.Parser).ParseUnverified(token, c)
		Expect(err).ToNot(HaveOccurred())

		sec := &system.GlobalSecretResource{
			Spec: &system_proto.Secret{
				Data: &wrapperspb.BytesValue{
					Value: []byte(c.ID),
				},
			},
		}
		err = store.Create(context.Background(), sec, core_store.CreateBy(RevocationsSecretKey))
		Expect(err).ToNot(HaveOccurred())

		// then
		_, err = issuer.Validate(token)
		Expect(err).To(MatchError("token is revoked"))
	})
})
