package issuer_test

import (
	"context"
	"time"

	"github.com/dgrijalva/jwt-go"
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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var _ = Describe("User token issuer", func() {

	var issuer UserTokenIssuer
	var store core_store.ResourceStore
	var signingKeyManager SigningKeyManager

	var now time.Time

	BeforeEach(func() {
		store = memory.NewStore()
		secretManager := secret_manager.NewGlobalSecretManager(secret_store.NewSecretStore(store), cipher.None())
		signingKeyManager = NewSigningKeyManager(secretManager)
		issuer = NewUserTokenIssuer(signingKeyManager, NewTokenRevocations(secretManager))

		Expect(signingKeyManager.CreateDefaultSigningKey()).To(Succeed())
		now = time.Now()
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
			Name:  "john.doe@acme.org",
			Group: "users",
		}

		// when generating many tokens with same signing key
		token1, err := issuer.Generate(id, 0)
		Expect(err).ToNot(HaveOccurred())
		token2, err := issuer.Generate(id, 0)
		Expect(err).ToNot(HaveOccurred())

		// then tokens are equal
		Expect(token1).To(Equal(token2))

		// when new signing key with higher serial number is created
		err = signingKeyManager.CreateSigningKey(2)
		Expect(err).ToNot(HaveOccurred())

		// and a new token is generated
		token3, err := issuer.Generate(id, 0)

		// then the new token is different because a new signing key is used
		Expect(err).ToNot(HaveOccurred())
		Expect(token2).ToNot(Equal(token3))

		// and all tokens are valid because 2 signing keys are present in the system
		_, _, err = issuer.Validate(token1)
		Expect(err).ToNot(HaveOccurred())
		_, _, err = issuer.Validate(token2)
		Expect(err).ToNot(HaveOccurred())
		_, _, err = issuer.Validate(token3)
		Expect(err).ToNot(HaveOccurred())

		// when first signing key is deleted
		err = store.Delete(context.Background(), system.NewGlobalSecretResource(), core_store.DeleteBy(SigningKeyResourceKey(DefaultSerialNumber)))
		Expect(err).ToNot(HaveOccurred())

		// then old tokens are no longer valid
		_, _, err = issuer.Validate(token1)
		Expect(err).To(MatchError("could not get Signing Key with serial number 1. Signing Key most likely has been rotated, regenerate the token: there is no Signing Key in the Control Plane."))
		_, _, err = issuer.Validate(token2)
		Expect(err).To(MatchError("could not get Signing Key with serial number 1. Signing Key most likely has been rotated, regenerate the token: there is no Signing Key in the Control Plane."))

		// and new token is valid because new signing key is present
		_, _, err = issuer.Validate(token3)
		Expect(err).ToNot(HaveOccurred())
	})

	It("should validate out expired tokens", func() {
		// given
		id := user.User{
			Name:  "john.doe@acme.org",
			Group: "users",
		}
		token, err := issuer.Generate(id, 60*time.Second)

		// when
		now = now.Add(60*time.Second + 1*time.Second)
		_, _, err = issuer.Validate(token)

		// then
		Expect(err).To(MatchError("could not parse token: token is expired by 1s"))
	})

	It("should revoke token", func() {
		// given valid token
		id := user.User{
			Name:  "john.doe@acme.org",
			Group: "users",
		}

		token, err := issuer.Generate(id, 60*time.Second)
		_, _, err = issuer.Validate(token)
		Expect(err).ToNot(HaveOccurred())

		// when id of the token is added to revocation list
		c := &jwt.StandardClaims{}
		_, _, err = new(jwt.Parser).ParseUnverified(token, c)
		Expect(err).ToNot(HaveOccurred())

		sec := &system.GlobalSecretResource{
			Spec: &system_proto.Secret{
				Data: &wrapperspb.BytesValue{
					Value: []byte(c.Id),
				},
			},
		}
		err = store.Create(context.Background(), sec, core_store.CreateBy(RevocationsSecretKey))
		Expect(err).ToNot(HaveOccurred())

		// then
		_, _, err = issuer.Validate(token)
		Expect(err).To(MatchError("token is revoked"))
	})
})
