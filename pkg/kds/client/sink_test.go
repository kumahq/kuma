package client_test

import (
	"time"

	envoy_api_v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/model"
	test_grpc "github.com/Kong/kuma/pkg/test/grpc"
	"github.com/Kong/kuma/pkg/test/kds/samples"
	kds_setup "github.com/Kong/kuma/pkg/test/kds/setup"
	kds_verifier "github.com/Kong/kuma/pkg/test/kds/verifier"
	test_model "github.com/Kong/kuma/pkg/test/resources/model"
)

const (
	defaultTimeout = 3 * time.Second
)

var _ = Describe("KDS Sink", func() {

	var tc kds_verifier.TestContext

	BeforeEach(func() {
		mockClientStream := test_grpc.MakeMockClientStream()
		stopCh := make(chan struct{})
		kds_setup.StartClient([]*test_grpc.MockClientStream{mockClientStream}, []model.ResourceType{mesh.MeshType, mesh.DataplaneType}, stopCh, nil)

		tc = &kds_verifier.TestContextImpl{
			MockClientStream: mockClientStream,
			StopCh:           stopCh,
		}
	})

	It("", func() {
		vrf := kds_verifier.New().
			Exec(kds_verifier.WaitRequest(defaultTimeout, func(req *envoy_api_v2.DiscoveryRequest) {
				Expect(req.TypeUrl).To(Equal(string(mesh.MeshType)))
			})).
			Exec(kds_verifier.WaitRequest(defaultTimeout, func(req *envoy_api_v2.DiscoveryRequest) {
				Expect(req.TypeUrl).To(Equal(string(mesh.DataplaneType)))
			})).
			Exec(kds_verifier.DiscoveryResponse(
				&mesh.MeshResourceList{Items: []*mesh.MeshResource{
					{Meta: &test_model.ResourceMeta{Name: "mesh1", Mesh: "mesh1"}, Spec: samples.Mesh1},
					{Meta: &test_model.ResourceMeta{Name: "mesh2", Mesh: "mesh2"}, Spec: samples.Mesh2},
				}}, "1", "1")).
			Exec(kds_verifier.WaitRequest(defaultTimeout, func(rs *envoy_api_v2.DiscoveryRequest) {
				Expect(rs.VersionInfo).To(Equal("1"))
				Expect(rs.ResponseNonce).To(Equal("1"))
				Expect(rs.TypeUrl).To(Equal(string(mesh.MeshType)))
			}))

		err := vrf.Verify(tc)
		Expect(err).ToNot(HaveOccurred())

		close(tc.Stop())
	})

})
