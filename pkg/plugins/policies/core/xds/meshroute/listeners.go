package meshroute

import (
	"fmt"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/xds/envoy/names"
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
		name = names.GetSplitClusterName(name, sc.GetAndIncrement())
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
