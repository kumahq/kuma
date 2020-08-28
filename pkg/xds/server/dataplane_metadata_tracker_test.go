package server_test

import (
	"context"
	"encoding/base64"

	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	pstruct "github.com/golang/protobuf/ptypes/struct"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"

	"github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/xds/server"
)

var _ = Describe("Dataplane Metadata Tracker", func() {
	rm := manager.NewResourceManager(memory.NewStore())
	tracker := server.NewDataplaneMetadataTracker(rm, core.KubernetesEnvironment)
	_ = rm.Create(context.Background(), &core_mesh.MeshResource{}, store.CreateByKey("default", "default"))

	dp := `
type: Dataplane
mesh: default
name: redis-1
networking:
  address: 127.0.0.1
  inbound:
  - port: 9000
    servicePort: 6379
    tags:
      kuma.io/service: redis
`

	req := v2.DiscoveryRequest{
		Node: &envoy_core.Node{
			Id: "default.example",
			Metadata: &pstruct.Struct{
				Fields: map[string]*pstruct.Value{
					"dataplaneTokenPath": {
						Kind: &pstruct.Value_StringValue{
							StringValue: "/tmp/token",
						},
					},
					"dataplaneResource": {
						Kind: &pstruct.Value_StringValue{
							StringValue: base64.StdEncoding.EncodeToString([]byte(dp)),
						},
					},
				},
			},
		},
	}
	const streamId = 123

	It("should track metadata", func() {
		// when
		err := tracker.OnStreamRequest(streamId, &req)

		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		metadata := tracker.Metadata(streamId)

		// then
		Expect(metadata.GetDataplaneTokenPath()).To(Equal("/tmp/token"))

		// when
		tracker.OnStreamClosed(streamId)

		// then metadata should be deleted
		metadata = tracker.Metadata(streamId)
		Expect(metadata).To(Equal(&xds.DataplaneMetadata{}))
	})

	It("should track metadata with empty Node in consecutive DiscoveryRequests", func() {
		// when
		err := tracker.OnStreamRequest(streamId, &req)

		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		err = tracker.OnStreamRequest(streamId, &v2.DiscoveryRequest{})

		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		metadata := tracker.Metadata(streamId)

		// then
		Expect(metadata.GetDataplaneTokenPath()).To(Equal("/tmp/token"))
	})
})
