package ws_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/emicklei/go-restful"
	error_types "github.com/kumahq/kuma/pkg/core/rest/errors/types"
	"github.com/kumahq/kuma/pkg/core/secrets/cipher"
	secret_manager "github.com/kumahq/kuma/pkg/core/secrets/manager"
	secret_store "github.com/kumahq/kuma/pkg/core/secrets/store"
	issuer2 "github.com/kumahq/kuma/pkg/plugins/authn/api-server/tokens/issuer"
	client2 "github.com/kumahq/kuma/pkg/plugins/authn/api-server/tokens/ws/client"
	server2 "github.com/kumahq/kuma/pkg/plugins/authn/api-server/tokens/ws/server"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	util_http "github.com/kumahq/kuma/pkg/util/http"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Auth Tokens WS", func() {

	var client client2.UserTokenClient
	var issuer issuer2.UserTokenIssuer

	BeforeEach(func() {
		store := memory.NewStore()
		signingKeyManager := issuer2.NewSigningKeyManager(secret_manager.NewGlobalSecretManager(secret_store.NewSecretStore(store), cipher.None()))
		issuer = issuer2.NewUserTokenIssuer(signingKeyManager)

		Expect(signingKeyManager.CreateDefaultSigningKey()).To(Succeed())
		ws := server2.NewWebService(issuer)

		container := restful.NewContainer()
		container.Add(ws)
		srv := httptest.NewServer(container)

		baseURL, err := url.Parse(srv.URL)
		Expect(err).ToNot(HaveOccurred())
		client = client2.NewHTTPUserTokenClient(util_http.ClientWithBaseURL(http.DefaultClient, baseURL, nil))

		// wait for the server
		Eventually(func() error {
			_, err := client.Generate("john.doe@acme.org", "team-a")
			return err
		}).ShouldNot(HaveOccurred())
	})

	It("should generate token", func() {
		// when
		token, err := client.Generate("john.doe@acme.org", "team-a")

		// then
		Expect(err).ToNot(HaveOccurred())
		u, _, err := issuer.Validate(token)
		Expect(err).ToNot(HaveOccurred())
		Expect(u.Name).To(Equal("john.doe@acme.org"))
		Expect(u.Group).To(Equal("team-a"))
	})

	It("should throw an error when zone is not passed", func() {
		// when
		_, err := client.Generate("", "")

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
