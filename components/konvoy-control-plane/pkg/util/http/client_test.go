package http_test

import (
	"net/http"
	"net/url"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	util_http "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/util/http"
)

var _ = Describe("Http Util", func() {
	Describe("ClientWithBaseURL(..)", func() {
		It("should rewrite request URL", func() {
			// given
			baseURL, _ := url.Parse("https://konvoy-control-plane:5681")
			// and
			delegate := util_http.ClientFunc(func(req *http.Request) (*http.Response, error) {
				Expect(req.URL.Scheme).To(Equal("https"))
				Expect(req.URL.Host).To(Equal("konvoy-control-plane:5681"))
				return &http.Response{}, nil
			})

			// when
			client := util_http.ClientWithBaseURL(delegate, baseURL)
			// then
			Expect(client).ToNot(BeIdenticalTo(delegate))

			// when
			req, _ := http.NewRequest("GET", "/meshes/default/dataplanes", nil)
			// and
			_, err := client.Do(req)
			// then
			Expect(err).ToNot(HaveOccurred())
		})

		It("should tolerate nil URL", func() {
			// given
			baseURL, _ := url.Parse("https://konvoy-control-plane:5681")
			// and
			delegate := util_http.ClientFunc(func(req *http.Request) (*http.Response, error) {
				Expect(req.URL).To(BeNil())
				return &http.Response{}, nil
			})

			// when
			client := util_http.ClientWithBaseURL(delegate, baseURL)
			// then
			Expect(client).ToNot(BeIdenticalTo(delegate))

			// when
			req := &http.Request{}
			// and
			req.URL = nil
			// and
			_, err := client.Do(req)
			// then
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
