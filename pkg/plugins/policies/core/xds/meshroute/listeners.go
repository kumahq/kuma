package meshroute

import (
	"fmt"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	plugins_xds "github.com/kumahq/kuma/pkg/plugins/policies/core/xds"
	"github.com/kumahq/kuma/pkg/util/pointer"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_names "github.com/kumahq/kuma/pkg/xds/envoy/names"
	envoy_tags "github.com/kumahq/kuma/pkg/xds/envoy/tags"
)

// SplitCounter
// Whenever `split` is specified in the TrafficRoute which has more than
// kuma.io/service tag we generate a separate Envoy cluster with _X_ suffix.
// SplitCounter ensures that we have different X for every split in one
// Dataplane. Each split is distinct for the whole Dataplane so we can avoid
// accidental cluster overrides.
type SplitCounter struct {
	counter int
}

func (s *SplitCounter) GetAndIncrement() int {
	counter := s.counter
	s.counter++
	return counter
}

func GetClusterName(
	name string,
	tags map[string]string,
	sc *SplitCounter,
) string {
	if len(tags) > 0 {
		name = envoy_names.GetSplitClusterName(name, sc.GetAndIncrement())
	}

	// The mesh tag is present here if this destination is generated
	// from a cross-mesh MeshGateway listener virtual outbound.
	// It is not part of the service tags.
	if mesh, ok := tags[mesh_proto.MeshTag]; ok {
		// The name should be distinct to the service & mesh combination
		name = fmt.Sprintf("%s_%s", name, mesh)
	}

	return name
}

func MakeTCPSplit(
	proxy *core_xds.Proxy,
	clusterCache map[common_api.TargetRefHash]string,
	sc *SplitCounter,
	servicesAcc envoy_common.ServicesAccumulator,
	refs []common_api.BackendRef,
) []envoy_common.Split {
	return makeSplit(
		proxy,
		map[core_mesh.Protocol]struct{}{
			core_mesh.ProtocolUnknown: {},
			core_mesh.ProtocolTCP:     {},
			core_mesh.ProtocolHTTP:    {},
			core_mesh.ProtocolHTTP2:   {},
		},
		clusterCache,
		sc,
		servicesAcc,
		refs,
	)
}

func MakeHTTPSplit(
	proxy *core_xds.Proxy,
	clusterCache map[common_api.TargetRefHash]string,
	sc *SplitCounter,
	servicesAcc envoy_common.ServicesAccumulator,
	refs []common_api.BackendRef,
) []envoy_common.Split {
	return makeSplit(
		proxy,
		map[core_mesh.Protocol]struct{}{
			core_mesh.ProtocolHTTP:  {},
			core_mesh.ProtocolHTTP2: {},
		},
		clusterCache,
		sc,
		servicesAcc,
		refs,
	)
}

func makeSplit(
	proxy *core_xds.Proxy,
	protocols map[core_mesh.Protocol]struct{},
	clusterCache map[common_api.TargetRefHash]string,
	sc *SplitCounter,
	servicesAcc envoy_common.ServicesAccumulator,
	refs []common_api.BackendRef,
) []envoy_common.Split {
	var split []envoy_common.Split

	for _, ref := range refs {
		switch ref.Kind {
		case common_api.MeshService, common_api.MeshServiceSubset:
		default:
			continue
		}

		service := ref.Name
		if pointer.DerefOr(ref.Weight, 1) == 0 {
			continue
		}

		if _, ok := protocols[plugins_xds.InferProtocol(proxy.Routing, service)]; !ok {
			protocol := plugins_xds.InferProtocol(proxy.Routing, service)
			fmt.Println("no protocol", service, protocol)
			return nil
		}

		clusterName := GetClusterName(ref.Name, ref.Tags, sc)
		isExternalService := plugins_xds.HasExternalService(proxy.Routing, service)
		refHash := ref.TargetRef.Hash()

		if existingClusterName, ok := clusterCache[refHash]; ok {
			// cluster already exists, so adding only split
			split = append(split, plugins_xds.NewSplitBuilder().
				WithClusterName(existingClusterName).
				WithWeight(uint32(pointer.DerefOr(ref.Weight, 1))).
				WithExternalService(isExternalService).
				Build())
			continue
		}

		clusterCache[refHash] = clusterName

		split = append(split, plugins_xds.NewSplitBuilder().
			WithClusterName(clusterName).
			WithWeight(uint32(pointer.DerefOr(ref.Weight, 1))).
			WithExternalService(isExternalService).
			Build())

		clusterBuilder := plugins_xds.NewClusterBuilder().
			WithService(service).
			WithName(clusterName).
			WithTags(envoy_tags.Tags(ref.Tags).
				WithTags(mesh_proto.ServiceTag, ref.Name).
				WithoutTags(mesh_proto.MeshTag)).
			WithExternalService(isExternalService)

		if mesh, ok := ref.Tags[mesh_proto.MeshTag]; ok {
			clusterBuilder.WithMesh(mesh)
		}

		servicesAcc.Add(clusterBuilder.Build())
	}

	return split
}
