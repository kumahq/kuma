package tokens_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/emicklei/go-restful"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/app/kumactl/pkg/tokens"
	config_kumactl "github.com/kumahq/kuma/pkg/config/app/kumactl/v1alpha1"
	"github.com/kumahq/kuma/pkg/tokens/builtin/issuer"
	tokens_server "github.com/kumahq/kuma/pkg/tokens/builtin/server"
)

type staticTokenIssuer struct {
}

var _ issuer.DataplaneTokenIssuer = &staticTokenIssuer{}

func (s *staticTokenIssuer) Generate(identity issuer.DataplaneIdentity) (issuer.Token, error) {
	return fmt.Sprintf("token-for-%s-%s", identity.Name, identity.Mesh), nil
}

func (s *staticTokenIssuer) Validate(token issuer.Token, meshName string) (issuer.DataplaneIdentity, error) {
	return issuer.DataplaneIdentity{}, errors.New("not implemented")
}

var _ = Describe("Tokens Client", func() {

	var server *httptest.Server

	BeforeEach(func() {
		container := restful.NewContainer()
		container.Add(tokens_server.NewWebservice(&staticTokenIssuer{}, &zoneIngressStaticTokenIssuer{}))
		server = httptest.NewServer(container.ServeMux)
	})

	AfterEach(func() {
		server.Close()
	})

	It("should return a token", func() {
		// given
		client, err := tokens.NewDataplaneTokenClient(&config_kumactl.ControlPlaneCoordinates_ApiServer{
			Url: server.URL,
		})
		Expect(err).ToNot(HaveOccurred())

		// wait for server
		Eventually(func() error {
			_, err := client.Generate("example", "default", nil, "dataplane")
			return err
		}, "5s", "100ms").ShouldNot(HaveOccurred())

		// when
		token, err := client.Generate("example", "default", nil, "dataplane")

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(token).To(Equal("token-for-example-default"))
	})

	It("should return an error when status code is different than 200", func() {
		// given
		mux := http.NewServeMux()
		server := httptest.NewServer(mux)
		defer server.Close()
		mux.HandleFunc("/tokens", func(writer http.ResponseWriter, req *http.Request) {
			defer GinkgoRecover()
			writer.WriteHeader(500)
			_, err := writer.Write([]byte("Internal Server Error"))
			Expect(err).ToNot(HaveOccurred())
		})
		client, err := tokens.NewDataplaneTokenClient(&config_kumactl.ControlPlaneCoordinates_ApiServer{
			Url: server.URL,
		})
		Expect(err).ToNot(HaveOccurred())

		// when
		_, err = client.Generate("example", "default", nil, "dataplane")

		// then
		Expect(err).To(MatchError("(500): Internal Server Error"))
	})
})
