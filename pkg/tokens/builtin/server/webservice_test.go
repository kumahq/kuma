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

	"github.com/Kong/kuma/pkg/core/xds"
	"github.com/Kong/kuma/pkg/sds/auth"
	"github.com/Kong/kuma/pkg/tokens/builtin/issuer"
	"github.com/Kong/kuma/pkg/tokens/builtin/server"
	"github.com/Kong/kuma/pkg/tokens/builtin/server/types"
)

type staticTokenIssuer struct {
	resp string
}

var _ issuer.DataplaneTokenIssuer = &staticTokenIssuer{}

func (s *staticTokenIssuer) Generate(proxyId xds.ProxyId) (auth.Credential, error) {
	return auth.Credential(s.resp), nil
}

func (s *staticTokenIssuer) Validate(credential auth.Credential) (xds.ProxyId, error) {
	return xds.ProxyId{}, errors.New("not implemented")
}

var _ = Describe("Dataplane Token Webservice", func() {

	const credentials = "test"
	var url string

	BeforeEach(func() {
		ws := server.NewWebservice(&staticTokenIssuer{credentials})

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
			Mesh: "defualt",
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
		Entry("json does not contain name", `{"mesh": "default"}`),
		Entry("json does not contain mesh", `{"name": "default"}`),
		Entry("not valid json", `not-valid-json`),
	)
})
