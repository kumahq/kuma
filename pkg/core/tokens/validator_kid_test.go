package tokens_test

import (
	"context"
	"encoding/base64"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	store_config "github.com/kumahq/kuma/v3/pkg/config/core/resources/store"
	"github.com/kumahq/kuma/v3/pkg/core"
	"github.com/kumahq/kuma/v3/pkg/core/secrets/cipher"
	secret_manager "github.com/kumahq/kuma/v3/pkg/core/secrets/manager"
	secret_store "github.com/kumahq/kuma/v3/pkg/core/secrets/store"
	"github.com/kumahq/kuma/v3/pkg/core/tokens"
	"github.com/kumahq/kuma/v3/pkg/plugins/resources/memory"
)

// The JWT header is decoded into a map[string]interface{}, so a "kid" written
// as a JSON number arrives as a float64. The validator must reject it with a
// normal error rather than doing an unchecked string type assertion.
var _ = Describe("Validator with a non-string kid header", func() {
	var validator tokens.Validator
	var ctx context.Context

	BeforeEach(func() {
		ctx = context.Background()
		store := memory.NewStore()
		secretManager := secret_manager.NewGlobalSecretManager(secret_store.NewSecretStore(store), cipher.None())
		validator = tokens.NewValidator(
			core.Log.WithName("test"),
			[]tokens.SigningKeyAccessor{
				tokens.NewSigningKeyAccessor(secretManager, TestTokenSigningKeyPrefix),
			},
			tokens.NewRevocations(secretManager, TokenRevocationsGlobalSecretKey),
			store_config.MemoryStore,
		)
	})

	It("returns an error when kid is a JSON number", func() {
		// Header {"alg":"HS256","typ":"JWT","kid":1337} - kid is a number.
		enc := base64.RawURLEncoding.EncodeToString
		header := enc([]byte(`{"alg":"HS256","typ":"JWT","kid":1337}`))
		payload := enc([]byte(`{}`))
		signature := enc([]byte("signature"))
		token := header + "." + payload + "." + signature

		var err error
		Expect(func() {
			err = validator.ParseWithValidation(ctx, token, &TestClaims{})
		}).ToNot(Panic())
		Expect(err).To(HaveOccurred())
	})
})
