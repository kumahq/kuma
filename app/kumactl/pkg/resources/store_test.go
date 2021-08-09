package resources

import (
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

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
				// then
				Expect(err).ToNot(HaveOccurred())

				// when
				store, err := NewResourceStore(cp.Coordinates.ApiServer, nil)
				// then
				Expect(store).ToNot(BeNil())
				// and
				Expect(err).To(BeNil())
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
				store, err := NewResourceStore(cp.Coordinates.ApiServer, nil)
				// then
				Expect(store).To(BeNil())
				// and
				Expect(err.Error()).To(ContainSubstring("Failed to parse API Server URL"))
			})
		})
	})
})

type RoundTripperFunc func(*http.Request) (*http.Response, error)

func (f RoundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
