package ws_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/emicklei/go-restful/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	store_config "github.com/kumahq/kuma/pkg/config/core/resources/store"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_tokens "github.com/kumahq/kuma/pkg/core/tokens"
	"github.com/kumahq/kuma/pkg/core/user"
	"github.com/kumahq/kuma/pkg/plugins/authn/api-server/tokens/issuer"
	"github.com/kumahq/kuma/pkg/plugins/authn/api-server/tokens/ws/client"
	"github.com/kumahq/kuma/pkg/plugins/authn/api-server/tokens/ws/server"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/test/matchers"
	util_http "github.com/kumahq/kuma/pkg/util/http"
)

type noopGenerateUserTokenAccess struct{}

func (n *noopGenerateUserTokenAccess) ValidateGenerate(user.User) error {
	return nil
}

var _ = Describe("Auth Tokens WS", func() {
	var userTokenClient client.UserTokenClient
	var userTokenValidator issuer.UserTokenValidator
	var httpClient util_http.Client

	BeforeEach(func() {
		resManager := manager.NewResourceManager(memory.NewStore())
		signingKeyManager := core_tokens.NewSigningKeyManager(resManager, system.UserTokenSigningKeyPrefix)
		tokenIssuer := issuer.NewUserTokenIssuer(core_tokens.NewTokenIssuer(signingKeyManager))
		userTokenValidator = issuer.NewUserTokenValidator(
			core_tokens.NewValidator(
				core.Log.WithName("test"),
				[]core_tokens.SigningKeyAccessor{
					core_tokens.NewSigningKeyAccessor(resManager, system.UserTokenSigningKeyPrefix),
				},
				core_tokens.NewRevocations(resManager, core_model.ResourceKey{Name: system.UserTokenRevocations}),
				store_config.MemoryStore,
			),
		)

		Expect(signingKeyManager.CreateDefaultSigningKey(context.Background())).To(Succeed())
		ws := server.NewWebService(tokenIssuer, &noopGenerateUserTokenAccess{})

		container := restful.NewContainer()
		container.Add(ws)
		srv := httptest.NewServer(container)

		baseURL, err := url.Parse(srv.URL)
		Expect(err).ToNot(HaveOccurred())
		httpClient = util_http.ClientWithBaseURL(http.DefaultClient, baseURL, nil)
		userTokenClient = client.NewHTTPUserTokenClient(httpClient)

		// wait for the server
		Eventually(func() error {
			_, err := userTokenClient.Generate("john.doe@example.com", []string{"team-a"}, time.Hour)
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

	It("should throw an error when name is not passed", func() {
		// when
		_, err := userTokenClient.Generate("", nil, 1*time.Hour)

		// then
		bytes, err := json.MarshalIndent(err, "", "  ")
		Expect(err).ToNot(HaveOccurred())
		Expect(bytes).To(matchers.MatchGoldenJSON(path.Join("testdata", "ws-no-name.golden.json")))
	})

	It("should throw an error with 0 for validFor", func() {
		// when
		_, err := userTokenClient.Generate("foo@example.com", nil, 0)

		// then
		bytes, err := json.MarshalIndent(err, "", "  ")
		Expect(err).ToNot(HaveOccurred())
		Expect(bytes).To(matchers.MatchGoldenJSON(path.Join("testdata", "ws-0-validFor.golden.json")))
	})

	It("should throw an error if validFor is not present", func() {
		// given invalid request (cannot be implemented using UserTokenClient)
		req, err := http.NewRequest("POST", "/tokens/user", strings.NewReader(`{"name": "xyz"}`))
		req.Header.Add("content-type", "application/json")
		Expect(err).ToNot(HaveOccurred())

		// when
		resp, err := httpClient.Do(req)
		Expect(err).ToNot(HaveOccurred())
		defer resp.Body.Close()

		// then
		bytes, err := io.ReadAll(resp.Body)
		Expect(err).ToNot(HaveOccurred())
		Expect(bytes).To(matchers.MatchGoldenJSON(path.Join("testdata", "ws-missing-validFor.golden.json")))
	})

	It("should throw an error when issuer is disabled", func() {
		container := restful.NewContainer()
		ws := server.NewWebService(issuer.DisabledIssuer{}, &noopGenerateUserTokenAccess{})
		container.Add(ws)
		srv := httptest.NewServer(container)

		baseURL, err := url.Parse(srv.URL)
		Expect(err).ToNot(HaveOccurred())
		httpClient := util_http.ClientWithBaseURL(http.DefaultClient, baseURL, nil)
		userTokenClient := client.NewHTTPUserTokenClient(httpClient)

		Eventually(func(g Gomega) {
			_, err := userTokenClient.Generate("john.doe@example.com", []string{"team-a"}, 1*time.Hour)
			bytes, err := json.MarshalIndent(err, "", "  ")
			Expect(err).ToNot(HaveOccurred())
			Expect(bytes).To(matchers.MatchGoldenJSON(path.Join("testdata", "ws-token-issuer-disabled.golden.json")))
		}, "10s", "100ms").Should(Succeed())
	})
})
