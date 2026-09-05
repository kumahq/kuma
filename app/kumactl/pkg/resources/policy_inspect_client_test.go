package resources

import (
	"context"
	"io"
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mtp_api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
)

var _ = Describe("httpPolicyInspectClient", func() {
	DescribeTable("requesting dataplanes for a policy",
		func(size int, offset, expectedURL string) {
			client := httpPolicyInspectClient{
				Client: &http.Client{
					Transport: RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
						Expect(req.Method).To(Equal(http.MethodGet))
						Expect(req.URL.String()).To(Equal(expectedURL))
						return &http.Response{
							StatusCode: http.StatusOK,
							Body: io.NopCloser(strings.NewReader(`{
								"total": 2,
								"next": "http://localhost/next",
								"items": [{"type": "Dataplane", "mesh": "default", "name": "dp-1", "labels": {}}]
							}`)),
						}, nil
					}),
				},
			}

			response, err := client.DataplanesForPolicy(
				context.Background(),
				mtp_api.MeshTrafficPermissionResourceTypeDescriptor,
				"default",
				"policy-1",
				size,
				offset,
			)

			Expect(err).ToNot(HaveOccurred())
			Expect(response.Total).To(Equal(2))
			Expect(response.Next).ToNot(BeNil())
			Expect(*response.Next).To(Equal("http://localhost/next"))
			Expect(response.Items).To(HaveLen(1))
			Expect(response.Items[0].Name).To(Equal("dp-1"))
		},
		Entry(
			"without pagination parameters",
			0,
			"",
			"/meshes/default/meshtrafficpermissions/policy-1/_resources/dataplanes",
		),
		Entry(
			"with pagination parameters",
			25,
			"next page",
			"/meshes/default/meshtrafficpermissions/policy-1/_resources/dataplanes?offset=next+page&size=25",
		),
	)
})
