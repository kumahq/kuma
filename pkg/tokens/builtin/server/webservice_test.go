package server_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/emicklei/go-restful"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/tokens/builtin/zoneingress"

	"github.com/kumahq/kuma/pkg/tokens/builtin/issuer"
	"github.com/kumahq/kuma/pkg/tokens/builtin/server"
	"github.com/kumahq/kuma/pkg/tokens/builtin/server/types"
)

type staticTokenIssuer struct {
	resp string
}

var _ issuer.DataplaneTokenIssuer = &staticTokenIssuer{}

func (s *staticTokenIssuer) Generate(identity issuer.DataplaneIdentity) (issuer.Token, error) {
	return s.resp, nil
}

func (s *staticTokenIssuer) Validate(token issuer.Token, meshName string) (issuer.DataplaneIdentity, error) {
	return issuer.DataplaneIdentity{}, errors.New("not implemented")
}

type zoneIngressStaticTokenIssuer struct {
}

var _ zoneingress.TokenIssuer = &zoneIngressStaticTokenIssuer{}

func (z *zoneIngressStaticTokenIssuer) Generate(identity zoneingress.Identity) (zoneingress.Token, error) {
	return fmt.Sprintf("token-for-%s", identity.Zone), nil
}

func (z *zoneIngressStaticTokenIssuer) Validate(token zoneingress.Token) (zoneingress.Identity, error) {
	return zoneingress.Identity{}, errors.New("not implemented")
}

var _ = Describe("Dataplane Token Webservice", func() {

	const credentials = "test"
	var url string

	BeforeEach(func() {
		ws := server.NewWebservice(&staticTokenIssuer{credentials}, &zoneIngressStaticTokenIssuer{})

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
		respBody, err := ioutil.ReadAll(resp.Body)

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
