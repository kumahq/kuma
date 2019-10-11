package tokens_test

import (
	"encoding/json"
	"fmt"
	"github.com/Kong/kuma/app/kumactl/pkg/tokens"
	"github.com/Kong/kuma/pkg/tokens/builtin/server/model"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
)

var _ = Describe("Tokens Client", func() {
	It("should return a token", func() {
		// given
		mux := http.NewServeMux()
		server := httptest.NewServer(mux)
		mux.HandleFunc("/token", func(writer http.ResponseWriter, req *http.Request) {
			dpTokenReq := model.DataplaneTokenRequest{}
			reqBytes, err := ioutil.ReadAll(req.Body)
			Expect(err).ToNot(HaveOccurred())
			err = json.Unmarshal(reqBytes, &dpTokenReq)
			Expect(err).ToNot(HaveOccurred())

			token := fmt.Sprintf("token-for-%s-%s", dpTokenReq.Name, dpTokenReq.Mesh)

			_, err = writer.Write([]byte(token))
			Expect(err).ToNot(HaveOccurred())
		})
		client, err := tokens.NewDpTokenClient(server.URL)
		Expect(err).ToNot(HaveOccurred())

		// when
		token, err := client.Generate("example", "default")

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(token).To(Equal("token-for-example-default"))
	})

	It("should return an error when status code is different than 200", func() {
		// given
		mux := http.NewServeMux()
		server := httptest.NewServer(mux)
		mux.HandleFunc("/token", func(writer http.ResponseWriter, req *http.Request) {
			writer.WriteHeader(500)
		})
		client, err := tokens.NewDpTokenClient(server.URL)
		Expect(err).ToNot(HaveOccurred())

		// when
		_, err = client.Generate("example", "default")

		// then
		Expect(err).To(MatchError("unexpected status code 500. Expected 200"))
	})
})
