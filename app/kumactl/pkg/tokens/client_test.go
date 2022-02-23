package tokens_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/emicklei/go-restful"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	kumactl_client "github.com/kumahq/kuma/app/kumactl/pkg/client"
	"github.com/kumahq/kuma/app/kumactl/pkg/tokens"
	config_kumactl "github.com/kumahq/kuma/pkg/config/app/kumactl/v1alpha1"
	core_tokens "github.com/kumahq/kuma/pkg/core/tokens"
	"github.com/kumahq/kuma/pkg/tokens/builtin/access"
	"github.com/kumahq/kuma/pkg/tokens/builtin/issuer"
	tokens_server "github.com/kumahq/kuma/pkg/tokens/builtin/server"
	zone_access "github.com/kumahq/kuma/pkg/tokens/builtin/zone/access"
)

type staticTokenIssuer struct {
}

var _ issuer.DataplaneTokenIssuer = &staticTokenIssuer{}

func (s *staticTokenIssuer) Generate(ctx context.Context, identity issuer.DataplaneIdentity, validFor time.Duration) (core_tokens.Token, error) {
	return fmt.Sprintf("token-for-%s-%s", identity.Name, identity.Mesh), nil
}

var _ = Describe("Tokens Client", func() {

	var server *httptest.Server

	BeforeEach(func() {
		container := restful.NewContainer()
		container.Add(tokens_server.NewWebservice(
			&staticTokenIssuer{},
			&zoneIngressStaticTokenIssuer{},
			&zoneStaticTokenIssuer{},
			access.NoopDpTokenAccess{},
			zone_access.NoopZoneTokenAccess{},
		))
		server = httptest.NewServer(container.ServeMux)
	})

	AfterEach(func() {
		server.Close()
	})

	It("should return a token", func() {
		// given
		baseClient, err := kumactl_client.ApiServerClient(&config_kumactl.ControlPlaneCoordinates_ApiServer{
			Url: server.URL,
		}, time.Second)
		Expect(err).ToNot(HaveOccurred())
		client := tokens.NewDataplaneTokenClient(baseClient)

		// wait for server
		Eventually(func() error {
			_, err := client.Generate("example", "default", nil, "dataplane", 24*time.Hour)
			return err
		}, "5s", "100ms").ShouldNot(HaveOccurred())

		// when
		token, err := client.Generate("example", "default", nil, "dataplane", 24*time.Hour)

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
		baseClient, err := kumactl_client.ApiServerClient(&config_kumactl.ControlPlaneCoordinates_ApiServer{
			Url: server.URL,
		}, time.Second)
		Expect(err).ToNot(HaveOccurred())
		client := tokens.NewDataplaneTokenClient(baseClient)
		Expect(err).ToNot(HaveOccurred())

		// when
		_, err = client.Generate("example", "default", nil, "dataplane", 24*time.Hour)

		// then
		Expect(err).To(MatchError("(500): Internal Server Error"))
	})
})
