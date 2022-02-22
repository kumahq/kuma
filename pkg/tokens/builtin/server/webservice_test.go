package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/emicklei/go-restful"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/core/tokens"
	"github.com/kumahq/kuma/pkg/tokens/builtin/access"
	"github.com/kumahq/kuma/pkg/tokens/builtin/issuer"
	"github.com/kumahq/kuma/pkg/tokens/builtin/server"
	"github.com/kumahq/kuma/pkg/tokens/builtin/server/types"
	"github.com/kumahq/kuma/pkg/tokens/builtin/zone"
	zone_access "github.com/kumahq/kuma/pkg/tokens/builtin/zone/access"
	"github.com/kumahq/kuma/pkg/tokens/builtin/zoneingress"
)

type staticTokenIssuer struct {
	resp string
}

var _ issuer.DataplaneTokenIssuer = &staticTokenIssuer{}

func (s *staticTokenIssuer) Generate(context.Context, issuer.DataplaneIdentity, time.Duration) (tokens.Token, error) {
	return s.resp, nil
}

type zoneIngressStaticTokenIssuer struct {
}

var _ zoneingress.TokenIssuer = &zoneIngressStaticTokenIssuer{}

func (z *zoneIngressStaticTokenIssuer) Generate(ctx context.Context, identity zoneingress.Identity, validFor time.Duration) (zoneingress.Token, error) {
	return fmt.Sprintf("token-for-%s", identity.Zone), nil
}

type zoneStaticTokenIssuer struct {
}

var _ zone.TokenIssuer = &zoneStaticTokenIssuer{}

func (z *zoneStaticTokenIssuer) Generate(ctx context.Context, identity zone.Identity, validFor time.Duration) (zone.Token, error) {
	return fmt.Sprintf("token-for-%s", identity.Zone), nil
}

var _ = Describe("Dataplane Token Webservice", func() {

	const credentials = "test"
	var url string

	BeforeEach(func() {
		ws := server.NewWebservice(
			&staticTokenIssuer{credentials},
			&zoneIngressStaticTokenIssuer{},
			&zoneStaticTokenIssuer{},
			&access.NoopDpTokenAccess{},
			&zone_access.NoopZoneTokenAccess{},
		)

		container := restful.NewContainer()
		container.Add(ws)
		srv := httptest.NewServer(container)
		url = srv.URL

		// wait for the server
		Eventually(func() error {
			_, err := http.DefaultClient.Get(fmt.Sprintf("%s/tokens", srv.URL))
			return err
		}).ShouldNot(HaveOccurred())
	})

	It("should respond with generated token", func() {
		// given
		idReq := types.DataplaneTokenRequest{
			Mesh: "default",
			Name: "dp-1",
		}
		reqBytes, err := json.Marshal(idReq)
		Expect(err).ToNot(HaveOccurred())

		// when
		req, err := http.NewRequest("POST", fmt.Sprintf("%s/tokens", url), bytes.NewReader(reqBytes))
		Expect(err).ToNot(HaveOccurred())
		req.Header.Add("content-type", "application/json")
		resp, err := http.DefaultClient.Do(req)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(200))

		// when
		respBody, err := io.ReadAll(resp.Body)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(string(respBody)).To(Equal(credentials))
	})

	DescribeTable("should return bad request on invalid json",
		func(json string) {
			// given
			req, err := http.NewRequest("POST", fmt.Sprintf("%s/tokens", url), strings.NewReader(json))
			Expect(err).ToNot(HaveOccurred())
			req.Header.Add("content-type", "application/json")

			// when
			resp, err := http.DefaultClient.Do(req)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(400))
		},
		Entry("not valid json", `not-valid-json`),
	)
})
