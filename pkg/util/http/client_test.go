package http_test

import (
	"net/http"
	"net/url"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	util_http "github.com/kumahq/kuma/pkg/util/http"
)

var _ = Describe("Http Util", func() {
	Describe("ClientWithBaseURL(..)", func() {
		type testCase struct {
			baseURL     string
			requestURL  string
			expectedURL string
		}

		DescribeTable("should rewrite request URL by combining `baseURL` and `requestURL`",
			func(given testCase) {
				// setup
				baseURL, err := url.Parse(given.baseURL)
				Expect(err).ToNot(HaveOccurred())

				// and
				var actualURL *url.URL
				delegate := util_http.ClientFunc(func(req *http.Request) (*http.Response, error) {
					actualURL = req.URL
					return &http.Response{}, nil
				})

				// when
				client := util_http.ClientWithBaseURL(delegate, baseURL, nil)
				// then
				Expect(client).ToNot(BeIdenticalTo(delegate))

				// when
				req, err := http.NewRequest("GET", given.requestURL, nil)
				// then
				Expect(err).ToNot(HaveOccurred())

				// when
				_, err = client.Do(req)
				// then
				Expect(err).ToNot(HaveOccurred())

				// and
				Expect(actualURL.String()).To(Equal(given.expectedURL))
			},
			Entry("baseURL without path", testCase{
				baseURL:     "https://kuma-control-plane:5681",
				requestURL:  "/meshes/default/dataplanes",
				expectedURL: "https://kuma-control-plane:5681/meshes/default/dataplanes",
			}),
			Entry("baseURL without path and request with a relative path", testCase{
				baseURL:     "https://kuma-control-plane:5681",
				requestURL:  "meshes/default/dataplanes",
				expectedURL: "https://kuma-control-plane:5681/meshes/default/dataplanes",
			}),
			Entry("baseURL with path", testCase{
				baseURL:     "https://kuma-control-plane:5681/proxy/foo/bar",
				requestURL:  "/test",
				expectedURL: "https://kuma-control-plane:5681/proxy/foo/bar/test",
			}),
			Entry("baseURL that ends with /", testCase{
				baseURL:     "https://kuma-control-plane:5681/",
				requestURL:  "/meshes/default/dataplanes",
				expectedURL: "https://kuma-control-plane:5681/meshes/default/dataplanes",
			}),
			Entry("baseURL and/or requestURL with double slashes", testCase{
				baseURL:     "https://kuma-control-plane:5681//proxy/foo/bar",
				requestURL:  "/test//baz",
				expectedURL: "https://kuma-control-plane:5681/proxy/foo/bar/test/baz",
			}),
		)

		It("should tolerate nil URL", func() {
			// setup
			baseURL, err := url.Parse("https://kuma-control-plane:5681")
			Expect(err).ToNot(HaveOccurred())

			// and
			var actualURL *url.URL
			delegate := util_http.ClientFunc(func(req *http.Request) (*http.Response, error) {
				actualURL = req.URL
				return &http.Response{}, nil
			})

			// when
			client := util_http.ClientWithBaseURL(delegate, baseURL, nil)
			// then
			Expect(client).ToNot(BeIdenticalTo(delegate))

			// when
			req := &http.Request{
				URL: nil,
			}
			// and
			_, err = client.Do(req)
			// then
			Expect(err).ToNot(HaveOccurred())

			// and
			Expect(actualURL).To(BeNil())
		})
	})
})
