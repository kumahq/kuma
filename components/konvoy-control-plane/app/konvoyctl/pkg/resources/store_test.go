package resources

import (
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	konvoyctl_k8s "github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/pkg/k8s"
	konvoyctl_k8s_fake "github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/pkg/k8s/fake"
	config_proto "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config/app/konvoyctl/v1alpha1"
	util_proto "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/util/proto"
)

var _ = Describe("Store", func() {
	Describe("NewResourceStore(..)", func() {

		Context("should support Control Plane installed on Kubernetes", func() {

			var backupGetKubeConfig func(kubeconfig, context, namespace string) (konvoyctl_k8s.KubeConfig, error)

			BeforeEach(func() {
				backupGetKubeConfig = getKubeConfig
			})
			AfterEach(func() {
				getKubeConfig = backupGetKubeConfig
			})

			It("should succeed if configuration is valid", func() {
				// setup
				getKubeConfig = func(kubeconfig, context, namespace string) (konvoyctl_k8s.KubeConfig, error) {
					return &konvoyctl_k8s_fake.FakeKubeConfig{
						ServiceProxyTransport: RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
							return &http.Response{}, nil
						}),
					}, nil
				}

				// given
				config := `
                coordinates:
                  kubernetes:
                    context: minikube
                    kubeconfig: /root/.kube/config
                    namespace: konvoy-system
                name: k8s_minikube
`
				// when
				cp := &config_proto.ControlPlane{}
				err := util_proto.FromYAML([]byte(config), cp)
				// then
				Expect(err).ToNot(HaveOccurred())

				// when
				store, err := NewResourceStore(cp)
				// then
				Expect(store).ToNot(BeNil())
				// and
				Expect(err).To(BeNil())
			})
		})

		Context("should support Control Plane installed elsewhere", func() {

			It("should succeed if configuration is valid", func() {
				// given
				config := `
                coordinates:
                  apiServer:
                    url: https://konvoy-control-plane.internal:5681
                name: vm_test
`
				// when
				cp := &config_proto.ControlPlane{}
				err := util_proto.FromYAML([]byte(config), cp)
				// then
				Expect(err).ToNot(HaveOccurred())

				// when
				store, err := NewResourceStore(cp)
				// then
				Expect(store).ToNot(BeNil())
				// and
				Expect(err).To(BeNil())
			})
		})

		Context("should fail gracefully when Control Plane configuration is unknown", func() {

			It("should fail otherwise", func() {
				// given
				cp := (*config_proto.ControlPlane)(nil)
				// when
				store, err := NewResourceStore(cp)
				// then
				Expect(store).To(BeNil())
				// and
				Expect(err).To(MatchError("Control Plane has coordinates that are not supported yet: <nil>"))
			})
		})
	})
})

type RoundTripperFunc func(*http.Request) (*http.Response, error)

func (f RoundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
