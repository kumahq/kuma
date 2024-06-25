package meshservice_test

import (
	"fmt"
	"testing"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/apis/meshservice"
	"github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
)

// To run, remove X from the prefix
// $ cd /Users/jakub/kong/kuma/pkg/core/resources/apis/meshservice
// $ go test -bench=BenchmarkMatchDataplanesWithMeshServices -run=^# -count 10 | tee bench.txt
// $ go install golang.org/x/perf/cmd/benchstat@latest
// $ benchstat bench.txt
// goos: darwin
// goarch: arm64
// pkg: github.com/kumahq/kuma/pkg/core/resources/apis/meshservice
//
//	│     bench.txt     │
//	│     sec/op        │
//
// MatchDataplanesWithMeshServices/faster-matching-ns_200-svc_20-dpp_5-10     0.02471n ± 1%
// MatchDataplanesWithMeshServices/naive-matching-ns_200-svc_20-dpp_5-10        5.633 ± 1%
// MatchDataplanesWithMeshServices/faster-matching-ns_50-svc_5-dpp_5-10      0.001591n ± 3%
// MatchDataplanesWithMeshServices/naive-matching-ns_50-svc_5-dpp_5-10       0.02013n ± 1%
// MatchDataplanesWithMeshServices/faster-matching-ns_10-svc_5-dpp_1-10     0.0001143n ± 9%
// MatchDataplanesWithMeshServices/naive-matching-ns_10-svc_5-dpp_1-10     0.0001701n ± 4%
func XBenchmarkMatchDataplanesWithMeshServices(b *testing.B) {
	var table = []struct {
		dppsPerService       int
		servicesPerNamespace int
		namespaces           int
	}{
		{dppsPerService: 5, servicesPerNamespace: 20, namespaces: 200},
		{dppsPerService: 5, servicesPerNamespace: 5, namespaces: 50},
		{dppsPerService: 1, servicesPerNamespace: 5, namespaces: 10},
	}

	for _, testCase := range table {
		var meshServices []*v1alpha1.MeshServiceResource
		var dpps []*mesh.DataplaneResource
		for nsIdx := 0; nsIdx < testCase.namespaces; nsIdx++ {
			nsName := fmt.Sprintf("ns-%d", nsIdx)
			for svcIdx := 0; svcIdx < testCase.servicesPerNamespace; svcIdx++ {
				svcName := fmt.Sprintf("ms-%d-%d", nsIdx, svcIdx)
				ms := builders.MeshService().
					WithName(svcName).
					WithDataplaneTagsSelectorKV("app", svcName, "k8s.kuma.io/namespace", nsName).
					Build()
				meshServices = append(meshServices, ms)
				for dppIdx := 0; dppIdx < testCase.dppsPerService; dppIdx++ {
					dppName := fmt.Sprintf("dpp-%d-%d-%d", nsIdx, svcIdx, dppIdx)
					dpp := builders.Dataplane().
						WithName(dppName).
						WithInboundOfTags("kuma.io/service", svcName, "app", svcName, "k8s.kuma.io/namespace", nsName).
						Build()
					dpps = append(dpps, dpp)
				}
			}
		}

		suffix := fmt.Sprintf("-ns_%d-svc_%d-dpp_%d", testCase.namespaces, testCase.servicesPerNamespace, testCase.dppsPerService)
		b.Run("faster-matching"+suffix, func(b *testing.B) {
			_ = meshservice.MatchDataplanesWithMeshServices(dpps, meshServices, false)
		})

		b.Run("naive-matching"+suffix, func(b *testing.B) {
			_ = naiveMatching(dpps, meshServices)
		})

	}

}

func naiveMatching(dpps []*mesh.DataplaneResource, meshServices []*v1alpha1.MeshServiceResource) map[*v1alpha1.MeshServiceResource][]*mesh.DataplaneResource {
	// naive O(n^2) matching, works for healthy inbounds matched by tags
	result := map[*v1alpha1.MeshServiceResource][]*mesh.DataplaneResource{}
	for _, ms := range meshServices {
		for _, dpp := range dpps {
			if dpp.Spec.Matches(mesh_proto.TagSelector(ms.Spec.Selector.DataplaneTags)) {
				result[ms] = append(result[ms], dpp)
			}
		}
	}
	return result
}
