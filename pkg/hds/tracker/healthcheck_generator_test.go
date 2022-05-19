package tracker

import (
	"context"
	"time"

	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	dp_server "github.com/kumahq/kuma/pkg/config/dp-server"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/hds/cache"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/test/matchers"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("HDS Snapshot generator", func() {

	var resourceManager manager.ResourceManager

	BeforeEach(func() {
		resourceManager = manager.NewResourceManager(memory.NewStore())

		err := resourceManager.Create(context.Background(), mesh.NewMeshResource(), store.CreateByKey("mesh-1", model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

	})

	type testCase struct {
		goldenFile string
		dataplane  string
		hdsConfig  *dp_server.HdsConfig
	}

	DescribeTable("should generate HDS response",
		func(given testCase) {
			// given
			dp := mesh.NewDataplaneResource()
			err := util_proto.FromYAML([]byte(given.dataplane), dp.Spec)
			Expect(err).ToNot(HaveOccurred())
			err = resourceManager.Create(context.Background(), dp, store.CreateByKey("dp-1", "mesh-1"))
			Expect(err).ToNot(HaveOccurred())
			generator := NewSnapshotGenerator(resourceManager, given.hdsConfig, 9901)

			// when
			snapshot, err := generator.GenerateSnapshot(&envoy_config_core_v3.Node{Id: "mesh-1.dp-1"})

			// then
			Expect(err).ToNot(HaveOccurred())
			actual, err := util_proto.ToYAML(snapshot.GetResources(cache.HealthCheckSpecifierType)["hcs"])
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(matchers.MatchGoldenYAML("testdata", given.goldenFile))
		},
		Entry("should generate HealthCheckSpecifier", testCase{
			goldenFile: "hds.1.yaml",
			dataplane: `
networking:
  address: 10.20.0.1
  inbound:
    - port: 9000
      serviceAddress: 192.168.0.1
      servicePort: 80
      serviceProbe: 
        tcp: {}
      tags:
        kuma.io/service: backend
`,
			hdsConfig: &dp_server.HdsConfig{
				Interval: 8 * time.Second,
				Enabled:  true,
				CheckDefaults: &dp_server.HdsCheck{
					Interval:           1 * time.Second,
					NoTrafficInterval:  2 * time.Second,
					Timeout:            3 * time.Second,
					HealthyThreshold:   4,
					UnhealthyThreshold: 5,
				},
			},
		}),
	)
})
