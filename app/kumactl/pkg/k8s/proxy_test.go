package k8s_test

import (
	"bytes"
	"io/ioutil"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	kube_rest "k8s.io/client-go/rest"

	"github.com/Kong/kuma/app/kumactl/pkg/k8s"
)

var _ = Describe("Proxy", func() {
	Describe("NewServiceProxyTransport(..)", func() {

		type testCase struct {
			originalPath string
			expectedPath string
		}

		DescribeTable("should rewrite request URI",
			func(given testCase) {
				// given
				req, _ := http.NewRequest("GET", given.originalPath, nil)
				expected := &http.Response{}

				// setup
				rt := RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
					Expect(req.URL.Path).To(Equal(given.expectedPath))
					return expected, nil
				})

				// when
				spt := k8s.NewServiceProxyTransport(rt, "kuma-system", "kuma-control-plane:http-apis-server")
				// and
				resp, err := spt.RoundTrip(req)
				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(resp).To(BeIdenticalTo(expected))
			},
			Entry("", testCase{
				originalPath: "",
				expectedPath: "/api/v1/namespaces/kuma-system/services/kuma-control-plane:http-apis-server/proxy/",
			}),
			Entry("/", testCase{
				originalPath: "/",
				expectedPath: "/api/v1/namespaces/kuma-system/services/kuma-control-plane:http-apis-server/proxy/",
			}),
			Entry("meshes", testCase{
				originalPath: "meshes",
				expectedPath: "/api/v1/namespaces/kuma-system/services/kuma-control-plane:http-apis-server/proxy/meshes",
			}),
			Entry("/meshes/default/dataplanes", testCase{
				originalPath: "/meshes/default/dataplanes",
				expectedPath: "/api/v1/namespaces/kuma-system/services/kuma-control-plane:http-apis-server/proxy/meshes/default/dataplanes",
			}),
		)
	})

	Describe("NewInprocessKubeProxyTransport(..)", func() {
		It("should process requests in-process using a given handler", func() {
			// given
			req, _ := http.NewRequest("GET", "/api/v1/namespaces/kuma-system/services/kuma-control-plane:http-apis-server/proxy/meshes", nil)

			// setup
			handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				defer GinkgoRecover()
				w.WriteHeader(http.StatusOK)
				_, err := w.Write([]byte("hi there!"))
				Expect(err).ToNot(HaveOccurred())
			})

			// when
			kpt := k8s.NewInprocessKubeProxyTransport(handler)
			// and
			resp, err := kpt.RoundTrip(req)
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			// when
			data, err := ioutil.ReadAll(resp.Body)
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(string(data)).To(Equal("hi there!"))
		})

		It("should catch a panic and translate it into an error", func() {
			// given
			req, _ := http.NewRequest("GET", "/api/v1/namespaces/kuma-system/services/kuma-control-plane:http-apis-server/proxy/meshes", nil)

			// setup
			handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				panic("something bad has happened")
			})

			// when
			kpt := k8s.NewInprocessKubeProxyTransport(handler)
			// and
			resp, err := kpt.RoundTrip(req)
			// then
			Expect(err.Error()).To(MatchRegexp("kube proxy paniced: something bad has happened\n.+"))
			// and
			Expect(resp).To(BeNil())
		})
	})

	Describe("NewKubeApiProxyTransport(..)", func() {
		It("should work", func() {
			// given
			req, _ := http.NewRequest("GET", "/api/v1/namespaces/kuma-system/services/kuma-control-plane:http-apis-server/proxy/meshes", nil)
			expected := &http.Response{
				StatusCode: http.StatusCreated,
				Body:       ioutil.NopCloser(&bytes.Buffer{}),
			}

			// setup
			rt := RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer abcdef"))
				return expected, nil
			})

			// when
			apt, err := k8s.NewKubeApiProxyTransport(&kube_rest.Config{
				Host:        "1.2.3.4",
				BearerToken: "abcdef",
				Transport:   rt,
			})
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			resp, err := apt.RoundTrip(req)
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(resp.StatusCode).To(Equal(expected.StatusCode))
		})
	})
})

type RoundTripperFunc func(*http.Request) (*http.Response, error)

func (f RoundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
