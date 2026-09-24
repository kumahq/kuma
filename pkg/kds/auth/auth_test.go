package auth_test

import (
	"context"
	"errors"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	store_config "github.com/kumahq/kuma/v3/pkg/config/core/resources/store"
	"github.com/kumahq/kuma/v3/pkg/config/multizone"
	config_types "github.com/kumahq/kuma/v3/pkg/config/types"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/v3/pkg/core/secrets/cipher"
	secret_manager "github.com/kumahq/kuma/v3/pkg/core/secrets/manager"
	secret_store "github.com/kumahq/kuma/v3/pkg/core/secrets/store"
	core_tokens "github.com/kumahq/kuma/v3/pkg/core/tokens"
	kds_auth "github.com/kumahq/kuma/v3/pkg/kds/auth"
	"github.com/kumahq/kuma/v3/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/v3/pkg/tokens/builtin"
	"github.com/kumahq/kuma/v3/pkg/tokens/builtin/zone"
)

type serverStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *serverStream) Context() context.Context {
	return s.ctx
}

var _ = Describe("Zone Token authentication", func() {
	var issuer zone.TokenIssuer
	var streamInterceptor grpc.StreamServerInterceptor
	var unaryInterceptor grpc.UnaryServerInterceptor

	incomingCtx := func(zoneName string, tokens ...string) context.Context {
		md := metadata.MD{"client-id": {zoneName}}
		md.Append(kds_auth.TokenHeader, tokens...)
		return metadata.NewIncomingContext(context.Background(), md)
	}

	generate := func(zoneName string, scope ...string) string {
		token, err := issuer.Generate(context.Background(), zone.Identity{Zone: zoneName, Scope: scope}, time.Hour)
		Expect(err).ToNot(HaveOccurred())
		return token
	}

	// every case goes through both interceptors, they have to agree
	authenticate := func(ctx context.Context) error {
		handled := false
		streamErr := streamInterceptor(nil, &serverStream{ctx: ctx}, nil, func(any, grpc.ServerStream) error {
			handled = true
			return nil
		})
		Expect(handled).To(Equal(streamErr == nil))

		handled = false
		_, unaryErr := unaryInterceptor(ctx, nil, nil, func(context.Context, any) (any, error) {
			handled = true
			return nil, nil
		})
		Expect(handled).To(Equal(unaryErr == nil))

		if streamErr == nil {
			Expect(unaryErr).ToNot(HaveOccurred())
		} else {
			Expect(unaryErr).To(MatchError(streamErr))
		}
		return streamErr
	}

	BeforeEach(func() {
		secretManager := secret_manager.NewGlobalSecretManager(secret_store.NewSecretStore(memory.NewStore()), cipher.None())
		signingKeyManager := core_tokens.NewSigningKeyManager(secretManager, system.ZoneTokenSigningKeyPrefix)
		Expect(signingKeyManager.CreateDefaultSigningKey(context.Background())).To(Succeed())
		issuer = builtin.NewZoneTokenIssuer(secretManager)

		validator, err := builtin.NewZoneTokenValidator(secretManager, store_config.MemoryStore, multizone.KDSZoneTokenValidatorConfig{UseSecrets: true})
		Expect(err).ToNot(HaveOccurred())
		authenticator := kds_auth.NewZoneTokenAuthenticator(validator)
		streamInterceptor = kds_auth.StreamServerInterceptor(authenticator)
		unaryInterceptor = kds_auth.UnaryServerInterceptor(authenticator)
	})

	It("should accept a token issued for the zone with cp scope", func() {
		Expect(authenticate(incomingCtx("zone-1", generate("zone-1", zone.CPScope)))).To(Succeed())
	})

	It("should accept a token with cp among other scopes", func() {
		Expect(authenticate(incomingCtx("zone-1", generate("zone-1", "other", zone.CPScope)))).To(Succeed())
	})

	It("should accept a token with Bearer prefix", func() {
		Expect(authenticate(incomingCtx("zone-1", "Bearer "+generate("zone-1", zone.CPScope)))).To(Succeed())
	})

	It("should reject a token issued for another zone", func() {
		err := authenticate(incomingCtx("zone-2", generate("zone-1", zone.CPScope)))

		Expect(err).To(MatchError(status.Error(codes.PermissionDenied, `token is signed for "zone-1" zone, but connected CP advertised as "zone-2"`)))
	})

	DescribeTable("should reject a token without cp scope",
		func(scope []string) {
			err := authenticate(incomingCtx("zone-1", generate("zone-1", scope...)))

			Expect(status.Code(err)).To(Equal(codes.PermissionDenied))
			Expect(err).To(MatchError(ContainSubstring("cp is out of token's scope")))
		},
		Entry("other scope", []string{"other"}),
		Entry("no scope", nil),
	)

	It("should reject a token with invalid signature", func() {
		err := authenticate(incomingCtx("zone-1", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJab25lIjoicmVtb3RlLTEiLCJUb2tlblNlcmlhbE51bWJlciI6MX0.qOgAzuj8LpWX7qhusOKcDhiO9NSNAPbRZL5bZyo7Fx8"))

		Expect(status.Code(err)).To(Equal(codes.Unauthenticated))
		Expect(err).To(MatchError(ContainSubstring("invalid zone token")))
	})

	It("should reject a request without a token", func() {
		err := authenticate(incomingCtx("zone-1"))

		Expect(status.Code(err)).To(Equal(codes.Unauthenticated))
		Expect(err).To(MatchError(ContainSubstring("kumactl generate zone-token --zone=zone-1 --scope=cp")))
	})

	It("should reject a request with many tokens", func() {
		token := generate("zone-1", zone.CPScope)

		err := authenticate(incomingCtx("zone-1", token, token))

		Expect(err).To(MatchError(status.Error(codes.Unauthenticated, "authorization header has 2 values. Expected 1 value")))
	})

	It("should reject a request without client id", func() {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.MD{})

		Expect(status.Code(authenticate(ctx))).To(Equal(codes.InvalidArgument))
	})
})

var _ = Describe("Interceptors", func() {
	type authenticatorFunc = func(ctx context.Context, md metadata.MD, zone string) error

	ctx := metadata.NewIncomingContext(context.Background(), metadata.MD{"client-id": {"zone-1"}})

	run := func(fn authenticatorFunc) error {
		return kds_auth.StreamServerInterceptor(funcAuthenticator(fn))(nil, &serverStream{ctx: ctx}, nil, func(any, grpc.ServerStream) error {
			return nil
		})
	}

	It("should pass the zone and metadata to the authenticator", func() {
		err := run(func(_ context.Context, md metadata.MD, zone string) error {
			Expect(zone).To(Equal("zone-1"))
			Expect(md.Get("client-id")).To(Equal([]string{"zone-1"}))
			return nil
		})

		Expect(err).ToNot(HaveOccurred())
	})

	It("should keep the status returned by the authenticator", func() {
		err := run(func(context.Context, metadata.MD, string) error {
			return status.Error(codes.PermissionDenied, "denied")
		})

		Expect(err).To(MatchError(status.Error(codes.PermissionDenied, "denied")))
	})

	It("should map other errors to Unauthenticated", func() {
		err := run(func(context.Context, metadata.MD, string) error {
			return errors.New("failed")
		})

		Expect(err).To(MatchError(status.Error(codes.Unauthenticated, "failed")))
	})
})

type funcAuthenticator func(ctx context.Context, md metadata.MD, zone string) error

func (f funcAuthenticator) Authenticate(ctx context.Context, md metadata.MD, zone string) error {
	return f(ctx, md, zone)
}

var _ = Describe("Token credentials", func() {
	It("should attach the token", func() {
		creds := kds_auth.NewTokenCredentials(func() (string, error) { return "token", nil }, true)

		md, err := creds.GetRequestMetadata(context.Background())

		Expect(err).ToNot(HaveOccurred())
		Expect(md).To(Equal(map[string]string{"authorization": "token"}))
		Expect(creds.RequireTransportSecurity()).To(BeTrue())
	})

	It("should load the token on every call", func() {
		token := "token-1"
		creds := kds_auth.NewTokenCredentials(func() (string, error) { return token, nil }, false)

		token = "token-2"
		md, err := creds.GetRequestMetadata(context.Background())

		Expect(err).ToNot(HaveOccurred())
		Expect(md).To(HaveKeyWithValue("authorization", "token-2"))
		Expect(creds.RequireTransportSecurity()).To(BeFalse())
	})

	It("should attach nothing without a token", func() {
		creds := kds_auth.NewTokenCredentials(func() (string, error) { return "", nil }, true)

		md, err := creds.GetRequestMetadata(context.Background())

		Expect(err).ToNot(HaveOccurred())
		Expect(md).To(BeEmpty())
	})

	It("should fail when the token cannot be loaded", func() {
		creds := kds_auth.NewTokenCredentials(func() (string, error) { return "", errors.New("failed") }, true)

		_, err := creds.GetRequestMetadata(context.Background())

		Expect(err).To(MatchError("failed"))
	})
})

var _ = Describe("Zone Token signed offline", func() {
	keys := filepath.Join("..", "..", "..", "test", "keys")

	// the same issuer as `kumactl generate zone-token --signing-key-path --kid`
	generate := func(keyFile, kid string) string {
		issuer := zone.NewTokenIssuer(core_tokens.NewTokenIssuer(core_tokens.NewFileSigningKeyManager(filepath.Join(keys, keyFile), kid)))
		token, err := issuer.Generate(context.Background(), zone.Identity{Zone: "zone-1", Scope: []string{zone.CPScope}}, time.Hour)
		Expect(err).ToNot(HaveOccurred())
		return token
	}

	authenticate := func(token string) error {
		// no signing key in the store, only the configured public key can validate the token
		secretManager := secret_manager.NewGlobalSecretManager(secret_store.NewSecretStore(memory.NewStore()), cipher.None())
		validator, err := builtin.NewZoneTokenValidator(secretManager, store_config.MemoryStore, multizone.KDSZoneTokenValidatorConfig{
			UseSecrets: false,
			PublicKeys: []config_types.PublicKey{{KID: "1", KeyFile: filepath.Join(keys, "publickey.pem")}},
		})
		Expect(err).ToNot(HaveOccurred())
		return kds_auth.NewZoneTokenAuthenticator(validator).Authenticate(context.Background(), metadata.Pairs(kds_auth.TokenHeader, token), "zone-1")
	}

	It("should accept a token signed with the key matching the public key", func() {
		Expect(authenticate(generate("samplekey.pem", "1"))).To(Succeed())
	})

	It("should reject a token signed with another key", func() {
		Expect(status.Code(authenticate(generate("samplekey-2.pem", "1")))).To(Equal(codes.Unauthenticated))
	})

	It("should reject a token with unknown kid", func() {
		Expect(status.Code(authenticate(generate("samplekey.pem", "2")))).To(Equal(codes.Unauthenticated))
	})
})

var _ = Describe("Server interceptors", func() {
	authenticators := func(types ...multizone.KDSAuthType) kds_auth.Authenticators {
		authenticators := kds_auth.Authenticators{}
		for _, authType := range types {
			authenticators[authType] = funcAuthenticator(nil)
		}
		return authenticators
	}

	It("should authenticate with the authenticator of the configured type", func() {
		stream, unary, err := kds_auth.ServerInterceptors(multizone.KDSAuthZoneToken, authenticators(multizone.KDSAuthZoneToken))

		Expect(err).ToNot(HaveOccurred())
		Expect(stream).To(HaveLen(1))
		Expect(unary).To(HaveLen(1))
	})

	It("should authenticate with a type a distribution registered", func() {
		stream, _, err := kds_auth.ServerInterceptors("custom", authenticators(multizone.KDSAuthZoneToken, "custom"))

		Expect(err).ToNot(HaveOccurred())
		Expect(stream).To(HaveLen(1))
	})

	DescribeTable("should not authenticate with the none type",
		func(registered kds_auth.Authenticators) {
			stream, unary, err := kds_auth.ServerInterceptors(multizone.KDSAuthNone, registered)

			Expect(err).ToNot(HaveOccurred())
			Expect(stream).To(BeEmpty())
			Expect(unary).To(BeEmpty())
		},
		Entry("nothing registered", nil),
		Entry("authenticators registered", authenticators(multizone.KDSAuthZoneToken, "custom")),
	)

	DescribeTable("should reject a type nothing authenticates with",
		func(authType multizone.KDSAuthType) {
			_, _, err := kds_auth.ServerInterceptors(authType, authenticators("custom"))

			Expect(err).To(MatchError(ContainSubstring("is not supported by this control plane. Supported types: custom, none")))
		},
		Entry("zoneToken without an authenticator", multizone.KDSAuthZoneToken),
		Entry("a type of a distribution that is not installed", multizone.KDSAuthType("other")),
		Entry("a typo", multizone.KDSAuthType("zonetoken")),
	)
})
