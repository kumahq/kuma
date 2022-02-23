package resources

import (
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/app/kumactl/pkg/client"
	config_proto "github.com/kumahq/kuma/pkg/config/app/kumactl/v1alpha1"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("Store", func() {
	Describe("NewResourceStore(..)", func() {

		Context("should support Control Plane installed anywhere", func() {

			It("should succeed if configuration is valid", func() {
				// given
				config := `
                coordinates:
                  apiServer:
                    url: https://kuma-control-plane.internal:5681
                name: vm_test
`
				// when
				cp := &config_proto.ControlPlane{}
				err := util_proto.FromYAML([]byte(config), cp)
				Expect(err).ToNot(HaveOccurred())
				_, err = client.ApiServerClient(cp.Coordinates.ApiServer, time.Second)

				// then
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("should fail gracefully when Control Plane url is unparsable", func() {

			It("should fail otherwise", func() {
				// given
				cp := config_proto.ControlPlane{
					Name: "test1",
					Coordinates: &config_proto.ControlPlaneCoordinates{
						ApiServer: &config_proto.ControlPlaneCoordinates_ApiServer{
							Url: "\r\nbadbadurl",
						},
					},
				}

				// when
				_, err := client.ApiServerClient(cp.Coordinates.ApiServer, time.Second)

				// then
				Expect(err.Error()).To(ContainSubstring("Failed to parse API Server URL"))
			})
		})
	})
})

type RoundTripperFunc func(*http.Request) (*http.Response, error)

func (f RoundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
