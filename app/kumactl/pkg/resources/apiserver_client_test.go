package resources

import (
	"io/ioutil"
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ApiServer client", func() {
	It("should return error from the server", func() {
		// given
		client := httpApiServerClient{
			Client: &http.Client{
				Transport: RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusBadRequest,
						Body:       ioutil.NopCloser(strings.NewReader("some error from server")),
					}, nil
				}),
			},
		}

		// when
		_, err := client.GetVersion()

		// then
		Expect(err).To(MatchError("unexpected status code 400"))
	})
})
