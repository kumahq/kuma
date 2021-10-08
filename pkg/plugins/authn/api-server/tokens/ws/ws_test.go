package ws_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"time"

	"github.com/emicklei/go-restful"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	error_types "github.com/kumahq/kuma/pkg/core/rest/errors/types"
	"github.com/kumahq/kuma/pkg/core/secrets/cipher"
	secret_manager "github.com/kumahq/kuma/pkg/core/secrets/manager"
	secret_store "github.com/kumahq/kuma/pkg/core/secrets/store"
	"github.com/kumahq/kuma/pkg/plugins/authn/api-server/tokens/issuer"
	"github.com/kumahq/kuma/pkg/plugins/authn/api-server/tokens/ws/client"
	"github.com/kumahq/kuma/pkg/plugins/authn/api-server/tokens/ws/server"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	util_http "github.com/kumahq/kuma/pkg/util/http"
)

var _ = Describe("Auth Tokens WS", func() {

	var userTokenClient client.UserTokenClient
	var userTokenIssuer issuer.UserTokenIssuer

	BeforeEach(func() {
		store := memory.NewStore()
		manager := secret_manager.NewGlobalSecretManager(secret_store.NewSecretStore(store), cipher.None())
		signingKeyManager := issuer.NewSigningKeyManager(manager)
		userTokenIssuer = issuer.NewUserTokenIssuer(signingKeyManager, issuer.NewTokenRevocations(manager))

		Expect(signingKeyManager.CreateDefaultSigningKey()).To(Succeed())
		ws := server.NewWebService(userTokenIssuer)

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
		u, err := userTokenIssuer.Validate(token)
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
