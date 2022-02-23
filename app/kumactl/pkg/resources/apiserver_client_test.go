package resources

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("httpApiServerClient", func() {
	Describe("GetVersion()", func() {
		It("should parse response", func() {
			hostname, err := os.Hostname()
			Expect(err).ToNot(HaveOccurred())
			buildVersion := fmt.Sprintf(`
			{
				"hostname": "%s",
				"tagline": "Kuma",
				"version": "0.4.0"
			}`, hostname)
			client := httpApiServerClient{
				Client: &http.Client{
					Transport: RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
						Expect(req.URL.String()).To(Equal("/"))
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(bytes.NewBufferString(buildVersion)),
						}, nil
					}),
				},
			}

			// when
			version, err := client.GetVersion(context.Background())
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(version.Version).To(Equal("0.4.0"))
			Expect(version.Hostname).To(Equal(hostname))
			Expect(version.Tagline).To(Equal("Kuma"))
		})
		It("should return error from the server", func() {
			// given
			client := httpApiServerClient{
				Client: &http.Client{
					Transport: RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: http.StatusBadRequest,
							Body:       io.NopCloser(strings.NewReader("some error from server")),
						}, nil
					}),
				},
			}

			// when
			_, err := client.GetVersion(context.Background())

			// then
			Expect(err).To(MatchError("(400): some error from server"))
		})
	})
})
