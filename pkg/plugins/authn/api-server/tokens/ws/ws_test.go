package ws_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"time"

	"github.com/emicklei/go-restful"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/core/resources/manager"
	error_types "github.com/kumahq/kuma/pkg/core/rest/errors/types"
	core_tokens "github.com/kumahq/kuma/pkg/core/tokens"
	"github.com/kumahq/kuma/pkg/core/user"
	"github.com/kumahq/kuma/pkg/plugins/authn/api-server/tokens/issuer"
	"github.com/kumahq/kuma/pkg/plugins/authn/api-server/tokens/ws/client"
	"github.com/kumahq/kuma/pkg/plugins/authn/api-server/tokens/ws/server"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	util_http "github.com/kumahq/kuma/pkg/util/http"
)

type noopGenerateUserTokenAccess struct {
}

func (n *noopGenerateUserTokenAccess) ValidateGenerate(user.User) error {
	return nil
}

var _ = Describe("Auth Tokens WS", func() {

	var userTokenClient client.UserTokenClient
	var userTokenValidator issuer.UserTokenValidator

	BeforeEach(func() {
		resManager := manager.NewResourceManager(memory.NewStore())
		signingKeyManager := core_tokens.NewSigningKeyManager(resManager, issuer.UserTokenSigningKeyPrefix)
		tokenIssuer := issuer.NewUserTokenIssuer(core_tokens.NewTokenIssuer(signingKeyManager))
		userTokenValidator = issuer.NewUserTokenValidator(
			core_tokens.NewValidator(
				core_tokens.NewSigningKeyAccessor(resManager, issuer.UserTokenSigningKeyPrefix),
				core_tokens.NewRevocations(resManager, issuer.UserTokenRevocationsGlobalSecretKey),
			),
		)

		Expect(signingKeyManager.CreateDefaultSigningKey(context.Background())).To(Succeed())
		ws := server.NewWebService(tokenIssuer, &noopGenerateUserTokenAccess{})

		container := restful.NewContainer()
		container.Add(ws)
		srv := httptest.NewServer(container)

		baseURL, err := url.Parse(srv.URL)
		Expect(err).ToNot(HaveOccurred())
		userTokenClient = client.NewHTTPUserTokenClient(util_http.ClientWithBaseURL(http.DefaultClient, baseURL, nil))

		// wait for the server
		Eventually(func() error {
			_, err := userTokenClient.Generate("john.doe@example.com", []string{"team-a"}, 0)
			return err
		}).ShouldNot(HaveOccurred())
	})

	It("should generate token", func() {
		// when
		token, err := userTokenClient.Generate("john.doe@example.com", []string{"team-a"}, 1*time.Hour)

		// then
		Expect(err).ToNot(HaveOccurred())
		u, err := userTokenValidator.Validate(context.Background(), token)
		Expect(err).ToNot(HaveOccurred())
		Expect(u.Name).To(Equal("john.doe@example.com"))
		Expect(u.Groups).To(Equal([]string{"team-a"}))
	})

	It("should throw an error when zone is not passed", func() {
		// when
		_, err := userTokenClient.Generate("", nil, 1*time.Hour)

		// then
		Expect(err).To(Equal(&error_types.Error{
			Title:   "Invalid request",
			Details: "Resource is not valid",
			Causes: []error_types.Cause{
				{
					Field:   "name",
					Message: "cannot be empty",
				},
			},
		}))
	})
})
