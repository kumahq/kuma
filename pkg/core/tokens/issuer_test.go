package tokens_test

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v4"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/wrapperspb"

	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/secrets/cipher"
	secret_manager "github.com/kumahq/kuma/pkg/core/secrets/manager"
	secret_store "github.com/kumahq/kuma/pkg/core/secrets/store"
	"github.com/kumahq/kuma/pkg/core/tokens"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
)

type TestClaims struct {
	jwt.RegisteredClaims
}

func (t *TestClaims) ID() string {
	return t.RegisteredClaims.ID
}

func (t *TestClaims) KeyIDFallback() (int, error) {
	return 0, nil
}

func (t *TestClaims) SetRegisteredClaims(claims jwt.RegisteredClaims) {
	t.RegisteredClaims = claims
}

var _ tokens.Claims = &TestClaims{}

const TestTokenSigningKeyPrefix = "test-token-signing-key"

var TokenRevocationsGlobalSecretKey = core_model.ResourceKey{
	Name: "test-token-revocations",
	Mesh: core_model.NoMesh,
}

func TokenRevocationsSecretKey(mesh string) core_model.ResourceKey {
	return core_model.ResourceKey{
		Name: "test-token-revocations",
		Mesh: mesh,
	}
}

var _ = Describe("Token issuer", func() {

	var issuer tokens.Issuer
	var validator tokens.Validator
	var store core_store.ResourceStore
	var signingKeyManager tokens.SigningKeyManager

	now := time.Now()
	var ctx context.Context

	BeforeEach(func() {
		ctx = context.Background()
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

	Context("Global Scoped tokens", func() {
		BeforeEach(func() {
			store = memory.NewStore()
			secretManager := secret_manager.NewGlobalSecretManager(secret_store.NewSecretStore(store), cipher.None())
			signingKeyManager = tokens.NewSigningKeyManager(secretManager, TestTokenSigningKeyPrefix)
			issuer = tokens.NewTokenIssuer(signingKeyManager)
			validator = tokens.NewValidator(
				tokens.NewSigningKeyAccessor(secretManager, TestTokenSigningKeyPrefix),
				tokens.NewRevocations(secretManager, TokenRevocationsGlobalSecretKey),
			)

			Expect(signingKeyManager.CreateDefaultSigningKey(ctx)).To(Succeed())
		})

		It("should support rotation", func() {
			// given
			id := &TestClaims{}

			// when
			token1, err := issuer.Generate(ctx, id, time.Minute)
			Expect(err).ToNot(HaveOccurred())

			// then
			err = validator.ParseWithValidation(ctx, token1, id)
			Expect(err).ToNot(HaveOccurred())

			// when new signing key with higher serial number is created
			err = signingKeyManager.CreateSigningKey(ctx, 2)
			Expect(err).ToNot(HaveOccurred())

			// and a new token is generated
			token2, err := issuer.Generate(ctx, id, time.Minute)
			Expect(err).ToNot(HaveOccurred())

			// then all tokens are valid because 2 signing keys are present in the system
			err = validator.ParseWithValidation(ctx, token1, id)
			Expect(err).ToNot(HaveOccurred())
			err = validator.ParseWithValidation(ctx, token2, id)
			Expect(err).ToNot(HaveOccurred())

			// when first signing key is deleted
			err = store.Delete(ctx, system.NewGlobalSecretResource(), core_store.DeleteBy(tokens.SigningKeyResourceKey(TestTokenSigningKeyPrefix, tokens.DefaultSerialNumber, core_model.NoMesh)))
			Expect(err).ToNot(HaveOccurred())

			// then old tokens are no longer valid
			err = validator.ParseWithValidation(ctx, token1, id)
			Expect(err).To(HaveOccurred())

			// and new token is valid because new signing key is present
			err = validator.ParseWithValidation(ctx, token2, id)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should validate out expired tokens", func() {
			// given
			id := &TestClaims{}
			token, err := issuer.Generate(ctx, id, time.Minute)
			Expect(err).ToNot(HaveOccurred())

			// when
			now = now.Add(time.Minute + 1*time.Second)
			err = validator.ParseWithValidation(ctx, token, id)

			// then
			Expect(err).To(HaveOccurred())
		})

		It("should revoke token", func() {
			// given valid token
			id := &TestClaims{}

			token, err := issuer.Generate(ctx, id, time.Minute)
			Expect(err).ToNot(HaveOccurred())
			err = validator.ParseWithValidation(ctx, token, id)
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
			err = store.Create(ctx, sec, core_store.CreateBy(TokenRevocationsGlobalSecretKey))
			Expect(err).ToNot(HaveOccurred())

			// then
			err = validator.ParseWithValidation(ctx, token, id)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("Mesh Scoped tokens", func() {
		BeforeEach(func() {
			store = memory.NewStore()
			secretManager := manager.NewResourceManager(store)
			signingKeyManager = tokens.NewMeshedSigningKeyManager(secretManager, TestTokenSigningKeyPrefix, core_model.DefaultMesh)
			issuer = tokens.NewTokenIssuer(signingKeyManager)
			validator = tokens.NewValidator(
				tokens.NewMeshedSigningKeyAccessor(secretManager, TestTokenSigningKeyPrefix, core_model.DefaultMesh),
				tokens.NewRevocations(secretManager, TokenRevocationsSecretKey(core_model.DefaultMesh)),
			)

			Expect(secretManager.Create(ctx, mesh.NewMeshResource(), core_store.CreateByKey(core_model.DefaultMesh, core_model.NoMesh))).To(Succeed())
			Expect(signingKeyManager.CreateDefaultSigningKey(ctx)).To(Succeed())
		})

		It("should revoke token", func() {
			// given valid token
			id := &TestClaims{}

			token, err := issuer.Generate(ctx, id, time.Minute)
			Expect(err).ToNot(HaveOccurred())
			err = validator.ParseWithValidation(ctx, token, id)
			Expect(err).ToNot(HaveOccurred())

			// when id of the token is added to revocation list
			c := &jwt.RegisteredClaims{}
			_, _, err = new(jwt.Parser).ParseUnverified(token, c)
			Expect(err).ToNot(HaveOccurred())

			sec := &system.SecretResource{
				Spec: &system_proto.Secret{
					Data: &wrapperspb.BytesValue{
						Value: []byte(c.ID),
					},
				},
			}
			err = store.Create(context.Background(), sec, core_store.CreateBy(TokenRevocationsSecretKey(core_model.DefaultMesh)))
			Expect(err).ToNot(HaveOccurred())

			// then
			err = validator.ParseWithValidation(ctx, token, id)
			Expect(err).To(HaveOccurred())
		})
	})
})
