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
	tokens_server "github.com/kumahq/kuma/pkg/tokens/builtin/server"
	"github.com/kumahq/kuma/pkg/tokens/builtin/zoneingress"
)

type zoneIngressStaticTokenIssuer struct {
}

var _ zoneingress.TokenIssuer = &zoneIngressStaticTokenIssuer{}

func (z *zoneIngressStaticTokenIssuer) Generate(identity zoneingress.Identity) (zoneingress.Token, error) {
	return fmt.Sprintf("token-for-%s", identity.Zone), nil
}

func (z *zoneIngressStaticTokenIssuer) Validate(token zoneingress.Token) (zoneingress.Identity, error) {
	return zoneingress.Identity{}, errors.New("not implemented")
}

var _ = Describe("Zone Ingress Tokens Client", func() {

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
		client, err := tokens.NewZoneIngressTokenClient(&config_kumactl.ControlPlaneCoordinates_ApiServer{
			Url: server.URL,
		})
		Expect(err).ToNot(HaveOccurred())

		// wait for server
		Eventually(func() error {
			_, err := client.Generate("my-zone-1")
			return err
		}, "5s", "100ms").ShouldNot(HaveOccurred())

		// when
		token, err := client.Generate("my-zone-1")

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(token).To(Equal("token-for-my-zone-1"))
	})

	It("should return an error when status code is different than 200", func() {
		// given
		mux := http.NewServeMux()
		server := httptest.NewServer(mux)
		defer server.Close()
		mux.HandleFunc("/tokens/zone-ingress", func(writer http.ResponseWriter, req *http.Request) {
			defer GinkgoRecover()
			writer.WriteHeader(500)
			_, err := writer.Write([]byte("Internal Server Error"))
			Expect(err).ToNot(HaveOccurred())
		})
		client, err := tokens.NewZoneIngressTokenClient(&config_kumactl.ControlPlaneCoordinates_ApiServer{
			Url: server.URL,
		})
		Expect(err).ToNot(HaveOccurred())

		// when
		_, err = client.Generate("my-zone-2")

		// then
		Expect(err).To(MatchError("(500): Internal Server Error"))
	})
})
