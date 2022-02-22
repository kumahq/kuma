package tokens_test

import (
	"context"
	"crypto/x509"
	"time"

	"github.com/golang-jwt/jwt/v4"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/tokens"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	util_rsa "github.com/kumahq/kuma/pkg/util/rsa"
)

var _ = Describe("Compatibility with old ASN.1 format", func() {

	var issuer tokens.Issuer
	var validator tokens.Validator
	var signingKeyManager tokens.SigningKeyManager
	var resManager manager.ResourceManager

	var legacySigningKey []byte
	var ctx context.Context

	BeforeEach(func() {
		ctx = context.Background()
		resManager = manager.NewResourceManager(memory.NewStore())
		signingKeyManager = tokens.NewMeshedSigningKeyManager(resManager, TestTokenSigningKeyPrefix, model.DefaultMesh)
		issuer = tokens.NewTokenIssuer(signingKeyManager)
		validator = tokens.NewValidator(
			tokens.NewMeshedSigningKeyAccessor(resManager, TestTokenSigningKeyPrefix, model.DefaultMesh),
			tokens.NewRevocations(resManager, TokenRevocationsGlobalSecretKey),
		)

		Expect(resManager.Create(ctx, mesh.NewMeshResource(), core_store.CreateByKey(model.DefaultMesh, model.NoMesh))).To(Succeed())

		// setup ASN.1 signing key (created in Kuma 1.3.x)
		key, err := util_rsa.GenerateKey(util_rsa.DefaultKeySize)
		Expect(err).ToNot(HaveOccurred())
		legacySigningKey = x509.MarshalPKCS1PrivateKey(key)
		secret := system.NewSecretResource()
		secret.Spec.Data = &wrapperspb.BytesValue{
			Value: legacySigningKey,
		}
		Expect(resManager.Create(ctx, secret, core_store.CreateByKey(TestTokenSigningKeyPrefix, model.DefaultMesh))).To(Succeed())
	})

	It("support old tokens without KID, no expiration and and signed with HMAC256 instead of RSA256", func() {
		// given
		claims := &TestClaims{}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString(legacySigningKey)
		Expect(err).ToNot(HaveOccurred())

		// when
		err = validator.ParseWithValidation(ctx, tokenString, &TestClaims{})

		// then
		Expect(err).ToNot(HaveOccurred())
	})

	It("support new tokens with ASN.1 signing key", func() {
		// given
		token, err := issuer.Generate(ctx, &TestClaims{}, time.Hour)
		Expect(err).ToNot(HaveOccurred())

		// when
		err = validator.ParseWithValidation(ctx, token, &TestClaims{})

		// then
		Expect(err).ToNot(HaveOccurred())
	})

	It("should support rotation to the new PEM-encoded signing keys", func() {
		// given legacy token
		claims := &TestClaims{}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString(legacySigningKey)
		Expect(err).ToNot(HaveOccurred())

		// when new PEM-encoded signing key is created
		err = signingKeyManager.CreateSigningKey(ctx, 1)
		Expect(err).ToNot(HaveOccurred())

		// then old token is valid
		Expect(validator.ParseWithValidation(ctx, tokenString, &TestClaims{})).To(Succeed())

		// when
		newToken, err := issuer.Generate(ctx, &TestClaims{}, time.Hour)

		// then new token is valid
		Expect(err).ToNot(HaveOccurred())
		Expect(validator.ParseWithValidation(ctx, newToken, &TestClaims{})).To(Succeed())

		// when old ASN.1 signing key is deleted
		Expect(resManager.Delete(context.Background(), system.NewSecretResource(), core_store.DeleteByKey(TestTokenSigningKeyPrefix, model.DefaultMesh))).To(Succeed())

		// then old token is not valid
		Expect(validator.ParseWithValidation(ctx, tokenString, &TestClaims{})).ToNot(Succeed())

		// and new token is still valid
		Expect(validator.ParseWithValidation(ctx, newToken, &TestClaims{})).To(Succeed())
	})
})
